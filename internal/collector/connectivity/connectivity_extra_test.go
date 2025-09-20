package connectivity

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"testing"
)

func TestSetCookiesFromResponse_FallbackParsing(t *testing.T) {
	// Create a response with a non-standard Set-Cookie header
	hdr := http.Header{}
	hdr.Add("Set-Cookie", "weird/name=val; Path=/; HttpOnly")
	reqURL, _ := url.Parse("http://example.com/")
	r := &http.Response{Header: hdr, Request: &http.Request{URL: reqURL}}
	jar, _ := cookiejar.New(nil)
	SetCookiesFromResponse(jar, r, r.Request.URL, "http://example.com/")
	// ensure cookie present
	if len(jar.Cookies(r.Request.URL)) == 0 {
		t.Fatalf("expected cookie set from header fallback")
	}
}
