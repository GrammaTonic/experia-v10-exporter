package collector

import (
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/GrammaTonic/experia-v10-exporter/internal/testutil"
	"github.com/prometheus/client_golang/prometheus"
)

// When getMIBs returns an empty body, the collector should emit zeroed
// netdev metrics for the candidate so families remain present.
func TestCollect_EmptyMIBsEmitsZeroedNetdevMetrics(t *testing.T) {
	c := NewCollector(net.ParseIP("127.0.0.1"), "u", "p", 1*time.Second)
	c.client.Transport = testutil.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		// read body to allow re-use in downstream handlers
		_, _ = io.ReadAll(req.Body)
		// Always return an empty body for getMIBs so collector treats it as
		// resp == "" and emits zeroed netdev metrics.
		if req.Method == "POST" {
			// Return minimal auth/createContext and empty for other bodies
			body := `{"data":{"contextID":"CTX"}}`
			return testutil.MakeResp(body), nil
		}
		return testutil.MakeResp(""), nil
	})

	reg := prometheus.NewRegistry()
	if err := reg.Register(c); err != nil {
		t.Fatalf("register failed: %v", err)
	}
	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("gather failed: %v", err)
	}

	// Look for netdev_mtu metric for eth1 and ensure its value is 0
	found := false
	for _, mf := range mfs {
		if mf.GetName() == "experia_v10_netdev_mtu" {
			for _, m := range mf.GetMetric() {
				for _, lp := range m.GetLabel() {
					if lp.GetName() == "ifname" && lp.GetValue() == "eth1" {
						if m.GetGauge().GetValue() == 0 {
							found = true
						}
					}
				}
			}
		}
	}
	if !found {
		t.Fatalf("expected netdev_mtu for eth1 to be present with value 0 when MIBs empty")
	}
}
