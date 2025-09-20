package collector

import (
	"net/http"
	"net/url"

	connectivity "github.com/GrammaTonic/experia-v10-exporter/internal/collector/connectivity"
)

// setCookiesFromResponse is a thin forwarding wrapper to the connectivity
// package's exported helper. Keeping this wrapper preserves existing local
// call sites while allowing the implementation to live in connectivity.
func setCookiesFromResponse(jar http.CookieJar, resp *http.Response, reqURL *url.URL, fallbackURL string) {
	connectivity.SetCookiesFromResponse(jar, resp, reqURL, fallbackURL)
}
