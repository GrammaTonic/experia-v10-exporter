package collector

import (
	"net/http"
	"net/url"
)

// setCookiesFromResponse chooses the correct URL to set cookies for and sets them on the Jar.
// It prefers resp.Request.URL, then reqURL, then a parsed fallbackURL.
func setCookiesFromResponse(jar http.CookieJar, resp *http.Response, reqURL *url.URL, fallbackURL string) {
	if jar == nil || resp == nil {
		return
	}
	var u *url.URL
	if resp.Request != nil && resp.Request.URL != nil {
		u = resp.Request.URL
	} else {
		u = reqURL
	}
	if u == nil {
		u, _ = url.Parse(fallbackURL)
	}
	jar.SetCookies(u, resp.Cookies())
}
