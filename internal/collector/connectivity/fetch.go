package connectivity

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
)

// FetchURL performs an HTTP request using the provided client. It mirrors the
// behavior previously implemented on the collector type: returns response
// body and treats non-2xx responses as errors.
func FetchURL(client *http.Client, ctx context.Context, method, url string, headers map[string]string, body []byte) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("http status %d: %s", resp.StatusCode, string(respBody))
	}
	return respBody, nil
}
