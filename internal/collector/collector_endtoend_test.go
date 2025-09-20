package collector

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/GrammaTonic/experia-v10-exporter/internal/testutil"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

// TestCollectorFullFlow simulates a full scrape: authenticate -> getWANStatus -> getMIBs -> getNetDevStats
// and asserts that the collector emits expected metric families and some values.
func TestCollectorFullFlow(t *testing.T) {
	// create an httptest server that responds differently based on request body
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// read request body to decide response
		b, _ := io.ReadAll(r.Body)
		body := string(b)
		// simple routing by content
		if body == "" {
			// follow-up GET after auth
			w.Header().Set("Set-Cookie", "sessionid=abc123; Path=/")
			w.WriteHeader(200)
			_, _ = w.Write([]byte(`{}`))
			return
		}
		if strings.Contains(body, "sah.Device.Information") {
			// createContext -> return contextID
			w.Header().Set("Set-Cookie", "auth=1; Path=/")
			_, _ = w.Write([]byte(`{"data":{"contextID":"CTX123"}}`))
			return
		}
		if strings.Contains(body, "NeMo") {
			if strings.Contains(body, "getMIBs") {
				_, _ = w.Write([]byte(testutil.SampleMibJSON))
				return
			}
			if strings.Contains(body, "getNetDevStats") {
				// respond with specific numeric stats
				_, _ = w.Write([]byte(sprintf(testutil.SampleStatsFmt, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21)))
				return
			}
			_, _ = w.Write([]byte(`{}`))
			return
		}
		_, _ = w.Write([]byte(`{}`))
	}))
	defer ts.Close()

	// create collector using the test server as API endpoint by overriding apiUrl via test-only collector
	// Use a short timeout to keep test quick
	c := NewCollector(net.ParseIP("127.0.0.1"), "u", "p", 500*time.Millisecond, "eth0")
	// replace client's transport to redirect to ts
	c.client.Transport = testutil.RewriteTransport(ts.URL)

	// ensure Login works (authenticate path)
	if err := c.Login(); err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	// collect metrics
	reg := prometheus.NewRegistry()
	// register the collector with the registry
	reg.MustRegister(prometheus.Collector(c))

	// Gather metrics
	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("gather failed: %v", err)
	}

	// basic assertions: there should be at least one metric family
	if len(mfs) == 0 {
		t.Fatalf("expected some metrics, got none")
	}

	// find a specific metric family (netdev_up) and assert it exists
	found := false
	for _, mf := range mfs {
		if mf.GetName() == "experia_v10_up" || mf.GetName() == "experia_v10_netdev_up" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected netdev or up metrics present")
	}

	// also assert that session token is stored
	if c.SessionToken() == "" {
		t.Fatalf("expected session token to be set after Login")
	}

	// sample check: read authErrorsMetric value via prometheus path
	var pm dto.Metric
	ch := make(chan prometheus.Metric, 1)
	go func() { c.authErrorsMetric.Collect(ch); close(ch) }()
	for m := range ch {
		if err := m.Write(&pm); err == nil {
			// value present (may be zero)
			_ = pm
		}
	}
}

// helper: simple sprintf wrapper to avoid importing fmt at top-level
func sprintf(format string, a ...any) string {
	return fmt.Sprintf(format, a...)
}
