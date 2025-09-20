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

	// also assert that session token is stored
	if c.SessionToken() == "" {
		t.Fatalf("expected session token to be set after Login")
	}

	// Look for concrete metrics: netdev_up and netdev_rx_bytes_total for label ifname="eth1"
	wantIf := "eth1"
	var gotNetdevUp *dto.Metric
	var gotRxBytes *dto.Metric
	for _, mf := range mfs {
		name := mf.GetName()
		if name != "experia_v10_netdev_up" && name != "experia_v10_netdev_rx_bytes_total" {
			continue
		}
		for _, m := range mf.Metric {
			// find label 'ifname' == wantIf
			for _, lp := range m.Label {
				if lp.GetName() == "ifname" && lp.GetValue() == wantIf {
					if name == "experia_v10_netdev_up" {
						gotNetdevUp = m
					}
					if name == "experia_v10_netdev_rx_bytes_total" {
						gotRxBytes = m
					}
				}
			}
		}
	}

	if gotNetdevUp == nil {
		t.Fatalf("netdev_up metric for %s not found", wantIf)
	}
	if gotRxBytes == nil {
		t.Fatalf("netdev_rx_bytes_total metric for %s not found", wantIf)
	}

	// Assert expected values: netdev_up == 1, rx_bytes == 3 (from SampleStatsFmt args)
	if gotNetdevUp.Gauge == nil {
		t.Fatalf("netdev_up metric for %s is not a gauge", wantIf)
	}
	if gotNetdevUp.Gauge.GetValue() != 1.0 {
		t.Fatalf("unexpected netdev_up value for %s: got %v, want 1.0", wantIf, gotNetdevUp.Gauge.GetValue())
	}
	if gotRxBytes.Gauge == nil {
		t.Fatalf("netdev_rx_bytes_total metric for %s is not a gauge", wantIf)
	}
	if gotRxBytes.Gauge.GetValue() != 3.0 {
		t.Fatalf("unexpected netdev_rx_bytes_total for %s: got %v, want 3.0", wantIf, gotRxBytes.Gauge.GetValue())
	}
}

// helper: simple sprintf wrapper to avoid importing fmt at top-level
func sprintf(format string, a ...any) string {
	return fmt.Sprintf(format, a...)
}
