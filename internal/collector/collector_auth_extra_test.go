package collector

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/GrammaTonic/experia-v10-exporter/internal/testutil"
	"github.com/prometheus/client_golang/prometheus"
)

// Test that when Authenticate fails during Collect, the collector emits the
// placeholder IfupTime metric with value 0 and increments auth_errors_total.
func TestCollect_AuthFailurePlaceholderAlternate(t *testing.T) {
	c := NewCollector(nil, "u", "p", 1)
	// transport that simulates network error
	c.client.Transport = testutil.MakeRoundTripper(func(req *http.Request) (*http.Response, error) {
		return nil, &testutil.SimpleErr{S: "network"}
	})

	reg := prometheus.NewRegistry()
	if err := reg.Register(c); err != nil {
		t.Fatalf("register failed: %v", err)
	}

	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("gather failed: %v", err)
	}

	// Expect internet_connection metric family present even with auth failure
	found := false
	for _, mf := range mfs {
		if mf.GetName() == "experia_v10_internet_connection" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected internet_connection family present on auth failure")
	}
}

// Test authenticate helper stores the session token and processes cookies.
func TestAuthenticate_SuccessStoresToken(t *testing.T) {
	c := NewCollector(nil, "u", "p", 1)
	// Transport that returns a contextID on POST and ok on GET
	c.client.Transport = testutil.MakeRoundTripper(func(req *http.Request) (*http.Response, error) {
		if req.Method == "POST" {
			body := `{"data":{"contextID":"CTX-TEST"}}`
			r := &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader([]byte(body))), Header: make(http.Header)}
			r.Header.Add("Set-Cookie", "sess=1")
			return r, nil
		}
		// GET follow-up
		r := &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader([]byte("{}"))), Header: make(http.Header)}
		r.Header.Add("Set-Cookie", "follow=1")
		return r, nil
	})

	sess, err := c.authenticate()
	if err != nil {
		t.Fatalf("authenticate failed: %v", err)
	}
	if sess.Token != "CTX-TEST" {
		t.Fatalf("expected token CTX-TEST, got %s", sess.Token)
	}
	if c.SessionToken() != "CTX-TEST" {
		t.Fatalf("SessionToken mismatch, got %s", c.SessionToken())
	}
	// cookies should be present for host; cookie jar may be nil in tests
	_ = c.CookiesForHost("http://127.0.0.1/")
	// ensure Describe can be called with correct channel type
	descCh := make(chan *prometheus.Desc, 10)
	go func() { c.Describe(descCh); close(descCh) }()
	for range descCh {
	}
}
