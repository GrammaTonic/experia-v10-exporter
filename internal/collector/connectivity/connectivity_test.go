package connectivity

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/GrammaTonic/experia-v10-exporter/internal/testutil"
)

func TestFetchURL_SuccessAndNon2xx(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "bad") {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("internal error"))
			return
		}
		_, _ = w.Write([]byte("ok"))
	}))
	defer ts.Close()

	client := &http.Client{Timeout: 1 * time.Second}
	client.Transport = testutil.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		// preserve path when forwarding to test server
		newURL := ts.URL + req.URL.Path
		newReq, err := http.NewRequest(req.Method, newURL, req.Body)
		if err != nil {
			return nil, err
		}
		newReq.Header = req.Header.Clone()
		return http.DefaultTransport.RoundTrip(newReq)
	})

	b, err := FetchURL(client, context.Background(), "GET", "http://example/", map[string]string{"X-Test": "1"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(b) != "ok" {
		t.Fatalf("unexpected body: %s", string(b))
	}

	// non-2xx
	_, err = FetchURL(client, context.Background(), "GET", "http://example/bad", nil, nil)
	if err == nil {
		t.Fatalf("expected error for non-2xx status")
	}
}

func TestSetCookiesFromResponse_FallbackAndNil(t *testing.T) {
	jar, _ := cookiejar.New(nil)
	resp := &http.Response{
		Header: make(http.Header),
		Body:   io.NopCloser(strings.NewReader("ok")),
	}
	resp.Header.Add("Set-Cookie", "a=b; Path=/")

	// fallback URL should receive the cookie
	SetCookiesFromResponse(jar, resp, nil, "http://example.local/")
	u, _ := url.Parse("http://example.local/")
	cookies := jar.Cookies(u)
	if len(cookies) == 0 {
		t.Fatalf("expected cookies to be set on fallback URL")
	}

	// nil inputs shouldn't panic
	SetCookiesFromResponse(nil, nil, nil, "http://example.local/")
}

func TestAuthenticate_SuccessAndCookieJar(t *testing.T) {
	// Auth server: respond to POST with contextID and Set-Cookie, and accept follow-up GET
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			w.Header().Add("Set-Cookie", "sessionid=abc123; Path=/")
			_, _ = w.Write([]byte(`{"data":{"contextID":"ctx-1"}}`))
			return
		}
		// follow-up GET should include Authorization header
		if r.Header.Get("Authorization") == "" {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("missing auth"))
			return
		}
		w.Header().Add("Set-Cookie", "follow=1; Path=/")
		_, _ = w.Write([]byte("ok"))
	}))
	defer ts.Close()

	jar, _ := cookiejar.New(nil)
	client := &http.Client{Transport: testutil.RewriteTransport(ts.URL), Jar: jar, Timeout: 1 * time.Second}

	apiURL := "http://192.0.2.50/ws/NeMo/Intf/lan:getMIBs"
	token, err := Authenticate(client, apiURL, "u", "p", http.NewRequest, json.Marshal)
	if err != nil {
		t.Fatalf("Authenticate failed: %v", err)
	}
	if token != "ctx-1" {
		t.Fatalf("unexpected token: %s", token)
	}

	// Cookies may be set under either the request URL (test server host) or the apiURL host
	tsURL, _ := url.Parse(ts.URL)
	apiParsed, _ := url.Parse(apiURL)
	if len(jar.Cookies(tsURL)) == 0 && len(jar.Cookies(apiParsed)) == 0 {
		t.Fatalf("expected cookie jar to contain cookies for either %s or %s", tsURL.String(), apiParsed.String())
	}
}

func TestAuthenticate_MalformedJSONAndErrors(t *testing.T) {
	// server returns invalid JSON
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("not-json"))
	}))
	defer ts.Close()

	jar, _ := cookiejar.New(nil)
	client := &http.Client{Transport: testutil.RewriteTransport(ts.URL), Jar: jar, Timeout: 1 * time.Second}
	_, err := Authenticate(client, "http://192.0.2.60/ws/", "u", "p", http.NewRequest, json.Marshal)
	if err == nil {
		t.Fatalf("expected error for malformed JSON")
	}
}
