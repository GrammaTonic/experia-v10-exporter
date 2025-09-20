package testutil

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func TestMakeRespAndHandlerAndRewrite(t *testing.T) {
	r := MakeResp("hello")
	if r.StatusCode != 200 {
		t.Fatalf("expected 200 got %d", r.StatusCode)
	}
	b, _ := io.ReadAll(r.Body)
	if string(b) != "hello" {
		t.Fatalf("expected body hello got %s", string(b))
	}

	ts := httptest.NewServer(MakeJSONHandler("world"))
	defer ts.Close()
	// RewriteTransport should forward requests to ts
	client := &http.Client{Transport: RewriteTransport(ts.URL)}
	resp, err := client.Get("http://example.invalid/")
	if err != nil {
		t.Fatalf("client get failed: %v", err)
	}
	bb, _ := io.ReadAll(resp.Body)
	if string(bb) != "world" {
		t.Fatalf("expected world got %s", string(bb))
	}
}

func TestReadCounterValueAndFmtError(t *testing.T) {
	c := prometheus.NewCounter(prometheus.CounterOpts{Name: "t_counter"})
	c.Add(2)
	v := ReadCounterValue(c)
	if v != 2 {
		t.Fatalf("expected 2 got %v", v)
	}
	if FmtError("e").Error() == "" {
		t.Fatalf("FmtError returned empty")
	}
}

func TestErrReadCloser(t *testing.T) {
	var rc io.ReadCloser = &ErrReadCloser{}
	_, err := rc.Read([]byte{0})
	if err == nil {
		t.Fatalf("expected read error")
	}
	if rc.Close() != nil {
		t.Fatalf("expected close nil")
	}
}

func TestRewriteTransportUrlParsing(t *testing.T) {
	// ensure RewriteTransport preserves headers
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Test-Header") != "1" {
			t.Fatalf("expected header present")
		}
		_, _ = w.Write([]byte("ok"))
	}))
	defer ts.Close()
	client := &http.Client{Transport: RewriteTransport(ts.URL)}
	req, _ := http.NewRequest("GET", "http://unused/", nil)
	req.Header.Set("X-Test-Header", "1")
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("do failed: %v", err)
	}
	_, _ = io.ReadAll(resp.Body)
	resp.Body.Close()
}
