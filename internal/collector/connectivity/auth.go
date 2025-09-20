package collector

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
)

func (c *Experiav10Collector) authenticate() (sessionContext, error) {
	apiURL := fmt.Sprintf(apiUrl, c.ip.String())
	payload := map[string]any{
		"service": "sah.Device.Information",
		"method":  "createContext",
		"parameters": map[string]any{
			"applicationName": "webui",
			"username":        c.username,
			"password":        c.password,
		},
	}
	body, err := jsonMarshal(payload)
	if err != nil {
		return sessionContext{}, err
	}
	req, err := newRequest("POST", apiURL, bytes.NewBuffer(body))
	if err != nil {
		return sessionContext{}, err
	}
	req.Header.Set("content-type", "application/x-sah-ws-4-call+json")
	req.Header.Set("Authorization", "X-Sah-Login")
	req.Header.Set("accept", "*/*")
	req.Header.Set("accept-language", "en-US,en;q=0.7")
	req.Header.Set("sec-gpc", "1")
	// Add browser-like headers so the router may emit cookies in the response
	req.Header.Set("Origin", "http://192.168.2.254")
	req.Header.Set("Referer", "http://192.168.2.254/")

	resp, err := c.client.Do(req)
	if err != nil {
		return sessionContext{}, err
	}
	defer resp.Body.Close()
	// Store cookies from authentication response. Use the actual request URL so tests that
	// rewrite the transport (redirecting to an httptest.Server) still set cookies under
	// the correct host:port.
	// Store cookies using helper to make fallback behavior testable.
	if os.Getenv("EXPERIA_E2E") == "1" {
		log.Printf("AUTH RESP status=%s headers=%v cookies=%v", resp.Status, resp.Header, resp.Cookies())
	}
	// store cookies using helper in connectivity package
	setCookiesFromResponse(c.client.Jar, resp, req.URL, apiURL)
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return sessionContext{}, err
	}
	var result struct {
		Data struct {
			ContextID string `json:"contextID"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return sessionContext{}, fmt.Errorf("failed to parse auth response: %w", err)
	}
	if result.Data.ContextID == "" {
		return sessionContext{}, fmt.Errorf("authentication failed: no contextID")
	}
	// Some firmware sets an HTTP cookie after the context is created; ensure
	// we capture any Set-Cookie by performing a follow-up GET to the base
	// router URL using the newly obtained token. This mirrors browser
	// behaviour and ensures the cookie jar contains values like
	// "121adbc0/sessid" when present.
	base := fmt.Sprintf("http://%s/", c.ip.String())
	if c.client != nil {
		// build a GET request and include the token so the router will emit
		// cookies for this session if it usually does so for authenticated
		// requests. Use an empty non-nil body to accommodate test transports
		// that call io.ReadAll(req.Body).
		getReq, err := newRequest("GET", base, bytes.NewBuffer([]byte{}))
		if err == nil {
			getReq.Header.Set("Authorization", "X-Sah "+result.Data.ContextID)
			getReq.Header.Set("x-context", result.Data.ContextID)
			getReq.Header.Set("accept", "*/*")
			getReq.Header.Set("accept-language", "en-US,en;q=0.7")
			getReq.Header.Set("sec-gpc", "1")
			getReq.Header.Set("Origin", "http://192.168.2.254")
			getReq.Header.Set("Referer", "http://192.168.2.254/")
			resp2, err2 := c.client.Do(getReq)
			if err2 == nil {
				if os.Getenv("EXPERIA_E2E") == "1" {
					log.Printf("FOLLOWUP GET RESP status=%s headers=%v cookies=%v", resp2.Status, resp2.Header, resp2.Cookies())
				}
				// store cookies from this response into the jar using the
				// helper which selects the correct URL to attach cookies to.
				setCookiesFromResponse(c.client.Jar, resp2, getReq.URL, base)
				// drain and close
				_, _ = io.ReadAll(resp2.Body)
				resp2.Body.Close()
			}
		} else {
			// If newRequest failed, attempt to parse URL and clear cookies
			if u, err := url.Parse(base); err == nil && c.client.Jar != nil {
				// no-op, leave jar as-is
				_ = u
			}
		}
	}

	return sessionContext{Token: result.Data.ContextID}, nil
}
