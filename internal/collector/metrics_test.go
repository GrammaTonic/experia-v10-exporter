package collector

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"testing"

	"bytes"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"

	connectivity "github.com/GrammaTonic/experia-v10-exporter/internal/collector/connectivity"
	"github.com/GrammaTonic/experia-v10-exporter/internal/testutil"
	"github.com/prometheus/client_golang/prometheus"
)

func TestDescribeEmitsDescriptors(t *testing.T) {
	c := NewCollector(nil, "", "", 0)
	ch := make(chan *prometheus.Desc, 10)
	go func() {
		c.Describe(ch)
		close(ch)
	}()
	foundIfup := false
	foundPermission := false
	for d := range ch {
		if d.String() == ifupTime.String() {
			foundIfup = true
		}
		if d == permissionErrors.Desc() {
			foundPermission = true
		}
	}
	if !foundIfup {
		t.Fatalf("ifupTime descriptor not emitted")
	}
	if !foundPermission {
		t.Fatalf("permissionErrors descriptor not emitted")
	}
}

func TestAuthenticateMalformedJSON(t *testing.T) {
	// server returns invalid JSON for auth
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not-json"))
	}))
	defer ts.Close()

	c := NewCollector(nil, "u", "p", 1)
	c.client.Transport = testutil.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		// redirect to test server
		newReq, _ := http.NewRequest(req.Method, ts.URL, req.Body)
		newReq.Header = req.Header.Clone()
		return http.DefaultTransport.RoundTrip(newReq)
	})

	_, err := c.authenticate()
	if err == nil {
		t.Fatalf("expected authenticate to fail on malformed JSON")
	}
}

func TestFetchURLReadError(t *testing.T) {
	// Make client with Transport that returns a response with a body that errors on Read
	c := NewCollector(nil, "", "", 1)
	c.client.Transport = testutil.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		r := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader([]byte("ok"))),
		}
		return r, nil
	})
	// fetchURL should succeed with our body; assert content
	b, err := connectivity.FetchURL(c.client, context.Background(), "GET", "http://example/", nil, nil)
	if err != nil {
		t.Fatalf("fetchURL failed: %v", err)
	}
	if string(b) != "ok" {
		t.Fatalf("unexpected body: %s", string(b))
	}
}

func TestFetchURLErrors(t *testing.T) {
	c := NewCollector(nil, "", "", 1)

	// case: empty method -> http.NewRequest should error
	if _, err := connectivity.FetchURL(c.client, context.Background(), "", "http://example", nil, nil); err == nil {
		t.Fatalf("expected error for empty method")
	}

	// case: client.Do returns error
	c.client.Transport = testutil.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return nil, &testutil.SimpleErr{S: "network error"}
	})
	if _, err := connectivity.FetchURL(c.client, context.Background(), "GET", "http://example", nil, nil); err == nil {
		t.Fatalf("expected error when client.Do fails")
	}

	// case: resp.Body.Read returns error
	c.client.Transport = testutil.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		r := &http.Response{
			StatusCode: http.StatusOK,
			Body:       &testutil.ErrReadCloser{},
		}
		return r, nil
	})
	if _, err := connectivity.FetchURL(c.client, context.Background(), "GET", "http://example", nil, nil); err == nil {
		t.Fatalf("expected error when reading body fails")
	}
}

func TestAuthenticateDoAndReadErrors(t *testing.T) {
	// client.Do returns error
	c := NewCollector(net.ParseIP("192.0.2.10"), "u", "p", 1)
	c.client.Transport = testutil.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return nil, &testutil.SimpleErr{S: "do error"}
	})
	if _, err := c.authenticate(); err == nil {
		t.Fatalf("expected authenticate to fail when client.Do errors")
	}

	// client.Do returns response with body read error
	c.client.Transport = testutil.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		r := &http.Response{StatusCode: http.StatusOK, Body: &testutil.ErrReadCloser{}}
		return r, nil
	})
	if _, err := c.authenticate(); err == nil {
		t.Fatalf("expected authenticate to fail when reading body errors")
	}
}

// TestAuthenticateEmptyContextID ensures authenticate fails when contextID is empty
func TestAuthenticateEmptyContextID(t *testing.T) {
	c := NewCollector(net.ParseIP("192.0.2.20"), "u", "p", 1)
	// Transport returns a valid JSON but empty contextID
	c.client.Transport = testutil.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		r := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader([]byte(`{"data":{"contextID":""}}`))),
			Header:     make(http.Header),
		}
		return r, nil
	})
	if _, err := c.authenticate(); err == nil {
		t.Fatalf("expected authenticate to fail when contextID is empty")
	}
}

// TestAuthenticateCookieJarSet ensures cookies from the auth response are stored in the client's Jar
func TestAuthenticateCookieJarSet(t *testing.T) {
	ip := net.ParseIP("192.0.2.21")
	c := NewCollector(ip, "u", "p", 1)
	// Create a response that sets a cookie and returns a valid contextID
	c.client.Transport = testutil.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		h := make(http.Header)
		h.Add("Set-Cookie", "sessionid=abc123; Path=/")
		r := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader([]byte(`{"data":{"contextID":"ctx-1"}}`))),
			Header:     h,
		}
		return r, nil
	})

	ctx, err := c.authenticate()
	if err != nil {
		t.Fatalf("authenticate failed: %v", err)
	}
	if ctx.Token != "ctx-1" {
		t.Fatalf("unexpected token: %s", ctx.Token)
	}
	// Verify cookie jar contains the cookie for the API URL
	u, _ := url.Parse(fmt.Sprintf(apiUrl, ip.String()))
	cookies := c.client.Jar.Cookies(u)
	if len(cookies) == 0 {
		t.Fatalf("expected cookie jar to contain cookies for %s", u.String())
	}
}

// TestFetchURLHeaders ensures headers passed into fetchURL are forwarded on the outgoing request
func TestFetchURLHeaders(t *testing.T) {
	c := NewCollector(nil, "", "", 1)
	c.client.Transport = testutil.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		if req.Header.Get("X-Test-Header") != "yes" {
			return nil, &testutil.SimpleErr{S: "header missing"}
		}
		r := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader([]byte("resp"))),
			Header:     make(http.Header),
		}
		return r, nil
	})
	b, err := connectivity.FetchURL(c.client, context.Background(), "GET", "http://example/", map[string]string{"X-Test-Header": "yes"}, nil)
	if err != nil {
		t.Fatalf("fetchURL failed: %v", err)
	}
	if string(b) != "resp" {
		t.Fatalf("unexpected body: %s", string(b))
	}
}

// TestFetchURLEmptyURL ensures http.NewRequest errors when URL is empty
func TestFetchURLEmptyURL(t *testing.T) {
	c := NewCollector(nil, "", "", 1)
	if _, err := connectivity.FetchURL(c.client, context.Background(), "GET", "", nil, nil); err == nil {
		t.Fatalf("expected error for empty URL")
	}
}

// TestAuthenticateNoCookieJar ensures authenticate works even when the client's Jar is nil
func TestAuthenticateNoCookieJar(t *testing.T) {
	ip := net.ParseIP("192.0.2.99")
	c := NewCollector(ip, "u", "p", 1)
	// ensure Jar is nil
	c.client.Jar = nil
	// Transport returns valid JSON with Set-Cookie header and contextID
	c.client.Transport = testutil.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		h := make(http.Header)
		h.Add("Set-Cookie", "sessionid=xyz; Path=/")
		r := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader([]byte(`{"data":{"contextID":"ctx-nojar"}}`))),
			Header:     h,
		}
		return r, nil
	})

	ctx, err := c.authenticate()
	if err != nil {
		t.Fatalf("authenticate failed: %v", err)
	}
	if ctx.Token != "ctx-nojar" {
		t.Fatalf("unexpected token: %s", ctx.Token)
	}
}

// TestFetchURLInvalidURL_NonTag verifies fetchURL returns an error for an invalid URL (non-test build)
func TestFetchURLInvalidURL_NonTag(t *testing.T) {
	c := NewCollector(nil, "", "", 1)
	if _, err := connectivity.FetchURL(c.client, context.Background(), "GET", "http://\x00/", nil, nil); err == nil {
		t.Fatalf("expected fetchURL to fail for invalid URL")
	}
}

// TestAuthenticateJSONMarshalError_NonTag forces json marshal failure in non-test build
func TestAuthenticateJSONMarshalError_NonTag(t *testing.T) {
	old := jsonMarshal
	defer func() { jsonMarshal = old }()
	jsonMarshal = func(v any) ([]byte, error) { return nil, &testutil.SimpleErr{S: "marshal fail"} }

	c := NewCollector(net.ParseIP("192.0.2.40"), "u", "p", 1)
	if _, err := c.authenticate(); err == nil {
		t.Fatalf("expected authenticate to fail when json.Marshal errors")
	}
}

func TestSetCookiesFromResponseFallback(t *testing.T) {
	// Simulate resp with nil Request and nil reqURL so fallback parsing is used
	jar, _ := cookiejar.New(nil)
	resp := &http.Response{
		Header: make(http.Header),
		Body:   io.NopCloser(bytes.NewReader([]byte("ok"))),
	}
	resp.Header.Add("Set-Cookie", "a=b; Path=/")

	// Call helper with nil reqURL and a fallback URL
	connectivity.SetCookiesFromResponse(jar, resp, nil, "http://example.local/")
	u, _ := url.Parse("http://example.local/")
	cookies := jar.Cookies(u)
	if len(cookies) == 0 {
		t.Fatalf("expected cookies to be set on fallback URL")
	}
}

func TestSetCookiesFromResponseNilInputs(t *testing.T) {
	// Verify no panic on nil jar or nil resp
	connectivity.SetCookiesFromResponse(nil, nil, nil, "http://example.local/")
}

// TestAuthenticateNewRequestError_NonTag forces http.NewRequest to fail via newRequest override
func TestAuthenticateNewRequestError_NonTag(t *testing.T) {
	old := newRequest
	defer func() { newRequest = old }()
	newRequest = func(method, url string, body io.Reader) (*http.Request, error) {
		return nil, &testutil.SimpleErr{S: "bad request"}
	}

	c := NewCollector(net.ParseIP("192.0.2.31"), "u", "p", 1)
	if _, err := c.authenticate(); err == nil {
		t.Fatalf("expected authenticate to fail when newRequest errors")
	}
}

// Additional tests moved to test-only override files where appropriate.

// TestAuthenticateNewRequestError forces an invalid apiUrl to trigger http.NewRequest error in authenticate
// moved to test-only file metrics_test_overrides.go

// TestFetchURLInvalidURL forces fetchURL to fail on NewRequest by passing an invalid URL
// moved to test-only file metrics_test_overrides.go
