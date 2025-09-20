package connectivity

import (
	"net/http"
	"net/http/cookiejar"
	"time"
)

// NewHTTPClient constructs a default *http.Client used by the collector.
// Exported so the parent package can create a configured client without
// duplicating cookiejar setup logic.
func NewHTTPClient(timeout time.Duration) *http.Client {
	jar, _ := cookiejar.New(nil)
	return &http.Client{
		Timeout: timeout,
		Jar:     jar,
	}
}
