package collector

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"testing"
)

func TestCookiesForHostAndSessionToken(t *testing.T) {
	c := NewCollector(nil, "u", "p", 0)
	// attach a jar and set a cookie
	jar, _ := cookiejar.New(nil)
	u, _ := url.Parse("http://192.0.2.1/")
	jar.SetCookies(u, []*http.Cookie{{Name: "s", Value: "v"}})
	c.client.Jar = jar

	got := c.CookiesForHost("http://192.0.2.1/")
	if len(got) != 1 || got[0].Name != "s" || got[0].Value != "v" {
		t.Fatalf("unexpected cookies: %#v", got)
	}

	// session token should respect the lock-protected field
	c.sessionMu.Lock()
	c.session = sessionContext{Token: "CTX"}
	c.sessionMu.Unlock()
	if c.SessionToken() != "CTX" {
		t.Fatalf("expected session token CTX, got %q", c.SessionToken())
	}
}
