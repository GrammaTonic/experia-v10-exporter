package collector

import (
	"net/http"
	"net/url"
	"strings"
)

// setCookiesFromResponse chooses the correct URL to set cookies for and sets them on the Jar.
// It prefers resp.Request.URL, then reqURL, then a parsed fallbackURL.
// Some firmwares emit Set-Cookie headers with non-standard cookie names (for
// example containing '/'). net/http's cookie parser may ignore these; when
// that happens we fall back to parsing the raw Set-Cookie header strings and
// constructing http.Cookie values so they can be stored in the jar.
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

	cookies := resp.Cookies()
	if len(cookies) == 0 {
		// Fallback: parse raw Set-Cookie header strings.
		for _, sc := range resp.Header["Set-Cookie"] {
			// Extract the name=value before the first ';'
			parts := strings.SplitN(sc, ";", 2)
			if len(parts) == 0 {
				continue
			}
			kv := strings.TrimSpace(parts[0])
			if kv == "" {
				continue
			}
			eq := strings.Index(kv, "=")
			if eq <= 0 {
				continue
			}
			name := strings.TrimSpace(kv[:eq])
			val := strings.TrimSpace(kv[eq+1:])
			cookies = append(cookies, &http.Cookie{Name: name, Value: val, Path: "/"})
		}
	}

	if len(cookies) > 0 {
		jar.SetCookies(u, cookies)
	}
}
