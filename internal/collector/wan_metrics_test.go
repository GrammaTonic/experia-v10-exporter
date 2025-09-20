package collector

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	metrics "github.com/GrammaTonic/experia-v10-exporter/internal/collector/metrics"
	"github.com/GrammaTonic/experia-v10-exporter/internal/testutil"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

// Test that when getWANStatus MAC matches a candidate's LLAddress the collector
// emits the explicit WAN families (wan_ifname + wan_rx_bytes_total).
func TestCollect_EmitsWanMetricsOnMatch(t *testing.T) {
	// Force candidate ordering so ETH2 (which in SampleMibJSON has the
	// LLAddress matching our test WAN MAC) is processed first and maps to
	// canonical label "eth1".
	_ = os.Setenv("EXPERIA_EXPECT_NETDEV_IFACES", "ETH2,ETH3,ETH0,ETH1")
	defer os.Unsetenv("EXPERIA_EXPECT_NETDEV_IFACES")

	c := NewCollector(net.ParseIP("127.0.0.1"), "u", "p", 1*time.Second)
	// Prepare transport that responds to the sequence of requests the
	// collector performs.
	c.client.Transport = testutil.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		b, _ := io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewReader(b))

		var respBody string
		switch {
		case bytes.Contains(b, []byte("createContext")):
			respBody = `{"data":{"contextID":"CTX-TEST"}}`
		case bytes.Contains(b, []byte("getWANStatus")):
			// Use MAC that appears in testutil.SampleMibJSON (88:D2:74:AB:05:D0)
			respBody = `{"status":true,"data":{"ConnectionState":"Connected","LinkType":"ETH","Protocol":"DHCP","IPAddress":"1.2.3.4","MACAddress":"88:D2:74:AB:05:D0"}}`
		case bytes.Contains(b, []byte("getMIBs")):
			respBody = testutil.SampleMibJSON
		case bytes.Contains(b, []byte("getNetDevStats")):
			// Return distinct values per requested candidate so we can assert
			// the WAN counters belong to the expected canonical label.
			if bytes.Contains(b, []byte("ETH2")) {
				// For ETH2 choose RxBytes = 5555 so we can assert later.
				respBody = fmt.Sprintf(testutil.SampleStatsFmt, 201, 301, 5555, 6666, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17)
			} else if bytes.Contains(b, []byte("ETH3")) {
				respBody = fmt.Sprintf(testutil.SampleStatsFmt, 202, 302, 1002, 2002, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18)
			} else if bytes.Contains(b, []byte("ETH0")) {
				respBody = fmt.Sprintf(testutil.SampleStatsFmt, 203, 303, 1003, 2003, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19)
			} else if bytes.Contains(b, []byte("ETH1")) {
				respBody = fmt.Sprintf(testutil.SampleStatsFmt, 204, 304, 1004, 2004, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20)
			} else {
				respBody = `{"status":true}`
			}
		default:
			respBody = `{"status":true}`
		}

		r := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader([]byte(respBody))),
			Header:     make(http.Header),
		}
		return r, nil
	})

	reg := prometheus.NewRegistry()
	if err := reg.Register(c); err != nil {
		t.Fatalf("failed to register collector: %v", err)
	}

	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("gather failed: %v", err)
	}

	// Helper to find a metric family by full name
	findMF := func(name string) *dto.MetricFamily {
		for _, mf := range mfs {
			if mf.GetName() == metrics.MetricPrefix+name {
				return mf
			}
		}
		return nil
	}

	// WanIfname should be present and mark eth1 (because ETH2 was first in the
	// candidate list and maps to canonical eth1).
	mf := findMF("wan_ifname")
	if mf == nil {
		t.Fatalf("expected metric family %swan_ifname to be present", metrics.MetricPrefix)
	}
	if len(mf.GetMetric()) == 0 {
		t.Fatalf("wan_ifname has no metrics")
	}
	// Find which ifname label was emitted
	var wanIf string
	for _, m := range mf.GetMetric() {
		for _, lp := range m.GetLabel() {
			if lp.GetName() == "ifname" {
				wanIf = lp.GetValue()
			}
		}
	}
	if wanIf != "eth1" {
		t.Fatalf("expected wan_ifname ifname=eth1 but got %q", wanIf)
	}

	// Ensure a WAN traffic family exists and contains the expected RxBytes
	mfRx := findMF("wan_rx_bytes_total")
	if mfRx == nil {
		t.Fatalf("expected metric family %swan_rx_bytes_total to be present", metrics.MetricPrefix)
	}
	// Build map of ifname->value
	got := map[string]float64{}
	for _, m := range mfRx.GetMetric() {
		var ifn string
		var val float64
		if m.GetGauge() != nil {
			val = m.GetGauge().GetValue()
		}
		for _, lp := range m.GetLabel() {
			if lp.GetName() == "ifname" {
				ifn = lp.GetValue()
			}
		}
		got[ifn] = val
	}

	if v, ok := got["eth1"]; !ok || int(v) != 5555 {
		t.Fatalf("expected wan_rx_bytes_total for eth1 = 5555 present=%v got=%v", ok, v)
	}
}
