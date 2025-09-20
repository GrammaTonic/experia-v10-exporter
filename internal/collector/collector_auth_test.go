package collector

import (
	"errors"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type errRT struct{}

func (e errRT) RoundTrip(r *http.Request) (*http.Response, error) { return nil, errors.New("rt error") }

func TestCollect_AuthFailureEmitsPlaceholder(t *testing.T) {
	c := NewCollector(net.ParseIP("127.0.0.1"), "u", "p", 100*time.Millisecond)
	// client that errors on any request
	c.client.Transport = errRT{}

	reg := prometheus.NewRegistry()
	reg.MustRegister(prometheus.Collector(c))

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
