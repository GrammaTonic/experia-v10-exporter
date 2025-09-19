package collector

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
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
	body, err := json.Marshal(payload)
	if err != nil {
		return sessionContext{}, err
	}
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(body))
	if err != nil {
		return sessionContext{}, err
	}
	req.Header.Set("content-type", "application/x-sah-ws-4-call+json")
	req.Header.Set("Authorization", "X-Sah-Login")
	req.Header.Set("accept", "*/*")
	req.Header.Set("accept-language", "en-US,en;q=0.7")
	req.Header.Set("sec-gpc", "1")

	resp, err := c.client.Do(req)
	if err != nil {
		return sessionContext{}, err
	}
	defer resp.Body.Close()
	// Store cookies from authentication response. Use the actual request URL so tests that
	// rewrite the transport (redirecting to an httptest.Server) still set cookies under
	// the correct host:port.
	if c.client.Jar != nil {
		var u *url.URL
		if req != nil && req.URL != nil {
			u = req.URL
		} else {
			u, _ = url.Parse(apiURL)
		}
		c.client.Jar.SetCookies(u, resp.Cookies())
	}
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
	return sessionContext{Token: result.Data.ContextID}, nil
}
