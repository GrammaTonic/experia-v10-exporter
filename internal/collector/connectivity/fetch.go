package collector

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
)

// fetchURL performs an HTTP request with the given context, method, url,
// headers, and body, and returns the response body. It returns an error for
// non-2xx HTTP responses so callers can explicitly detect permission-denied
// or other HTTP errors.
func (c *Experiav10Collector) fetchURL(ctx context.Context, method, url string, headers map[string]string, body []byte) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	// Treat non-2xx responses as errors so callers can react accordingly.
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("http status %d: %s", resp.StatusCode, string(respBody))
	}
	return respBody, nil
}
