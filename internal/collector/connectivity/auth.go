package connectivity

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

// Authenticate performs the Experia V10 login flow using the provided
// http.Client and helper functions. It returns the session token (contextID)
// on success. Helpers are injected to keep this package testable and avoid
// depending on package-level indirections from the collector package.
// - newRequest mirrors http.NewRequest signature.
// - jsonMarshal mirrors json.Marshal signature.
func Authenticate(client *http.Client, apiURL, username, password string,
	newRequest func(string, string, io.Reader) (*http.Request, error),
	jsonMarshal func(any) ([]byte, error)) (string, error) {

	payload := map[string]any{
		"service": "sah.Device.Information",
		"method":  "createContext",
		"parameters": map[string]any{
			"applicationName": "webui",
			"username":        username,
			"password":        password,
		},
	}
	body, err := jsonMarshal(payload)
	if err != nil {
		return "", err
	}
	req, err := newRequest("POST", apiURL, bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("content-type", "application/x-sah-ws-4-call+json")
	req.Header.Set("Authorization", "X-Sah-Login")
	req.Header.Set("accept", "*/*")
	req.Header.Set("accept-language", "en-US,en;q=0.7")
	req.Header.Set("sec-gpc", "1")
	req.Header.Set("Origin", "http://192.168.2.254")
	req.Header.Set("Referer", "http://192.168.2.254/")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if os.Getenv("EXPERIA_E2E") == "1" {
		log.Printf("AUTH RESP status=%s headers=%v cookies=%v", resp.Status, resp.Header, resp.Cookies())
	}
	// store cookies from authentication response
	SetCookiesFromResponse(client.Jar, resp, req.URL, apiURL)

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var result struct {
		Data struct {
			ContextID string `json:"contextID"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("failed to parse auth response: %w", err)
	}
	if result.Data.ContextID == "" {
		return "", fmt.Errorf("authentication failed: no contextID")
	}

	// follow-up GET to capture any additional Set-Cookie headers emitted
	base := fmt.Sprintf("http://%s/", req.URL.Host)
	if client != nil {
		getReq, err := newRequest("GET", base, bytes.NewBuffer([]byte{}))
		if err == nil {
			getReq.Header.Set("Authorization", "X-Sah "+result.Data.ContextID)
			getReq.Header.Set("x-context", result.Data.ContextID)
			getReq.Header.Set("accept", "*/*")
			getReq.Header.Set("accept-language", "en-US,en;q=0.7")
			getReq.Header.Set("sec-gpc", "1")
			getReq.Header.Set("Origin", "http://192.168.2.254")
			getReq.Header.Set("Referer", "http://192.168.2.254/")
			resp2, err2 := client.Do(getReq)
			if err2 == nil {
				if os.Getenv("EXPERIA_E2E") == "1" {
					log.Printf("FOLLOWUP GET RESP status=%s headers=%v cookies=%v", resp2.Status, resp2.Header, resp2.Cookies())
				}
				SetCookiesFromResponse(client.Jar, resp2, getReq.URL, base)
				_, _ = io.ReadAll(resp2.Body)
				resp2.Body.Close()
			}
		}
	}

	return result.Data.ContextID, nil
}
