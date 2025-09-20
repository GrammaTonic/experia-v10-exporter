package collector

import (
	"bytes"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// use sampleMibJSON from testhelpers_test.go

func TestCollect_ParsesMIBsAndEmitsNetdevMetrics(t *testing.T) {
	// Create collector and override client transport to return sampleMibJSON
	c := NewCollector(net.ParseIP("127.0.0.1"), "u", "p", 1*time.Second)
	c.client.Transport = roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		// Read request body to determine which API is being called
		b, _ := io.ReadAll(req.Body)
		// restore body for any callers that might read it again
		req.Body = io.NopCloser(bytes.NewReader(b))
		var respBody string
		if bytes.Contains(b, []byte("createContext")) {
			// authentication response with contextID
			respBody = `{"data":{"contextID":"CTX-TEST"}}`
		} else if bytes.Contains(b, []byte("getWANStatus")) {
			respBody = `{"status":true,"data":{"ConnectionState":"Connected","LinkType":"ETH","Protocol":"DHCP","IPAddress":"1.2.3.4","MACAddress":"aa:bb:cc:dd:ee:ff"}}`
		} else if bytes.Contains(b, []byte("getMIBs")) {
			respBody = sampleMibJSON
		} else {
			// Default: return empty success
			respBody = `{"status":true}`
		}

		r := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader([]byte(respBody))),
			Header:     make(http.Header),
		}
		return r, nil
	})

	// Register collector with a fresh registry and call Gather
	reg := prometheus.NewRegistry()
	if err := reg.Register(c); err != nil {
		t.Fatalf("failed to register collector: %v", err)
	}

	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("gather failed: %v", err)
	}

	// Find netdev_up metric family and assert eth2 and eth3 are present and up
	found := map[string]float64{}
	for _, mf := range mfs {
		if mf.GetName() == metricPrefix+"netdev_up" {
			for _, m := range mf.GetMetric() {
				for _, lp := range m.GetLabel() {
					if lp.GetName() == "ifname" {
						found[lp.GetValue()] = m.GetGauge().GetValue()
					}
				}
			}
		}
	}

	// helper to lookup by lowercased key
	getL := func(m map[string]float64, name string) (float64, bool) {
		v, ok := m[strings.ToLower(name)]
		return v, ok
	}

	if got, ok := getL(found, "eth2"); !ok || got != 1 {
		t.Fatalf("expected eth2 up==1, got %v present=%v", got, ok)
	}
	if got, ok := getL(found, "eth3"); !ok || got != 1 {
		t.Fatalf("expected eth3 up==1, got %v present=%v", got, ok)
	}

	// check MTU metric for eth3
	// Search mfs for netdev_mtu
	mtuFound := map[string]float64{}
	for _, mf := range mfs {
		if mf.GetName() == metricPrefix+"netdev_mtu" {
			for _, m := range mf.GetMetric() {
				var ifn string
				var val float64
				for _, lp := range m.GetLabel() {
					if lp.GetName() == "ifname" {
						ifn = lp.GetValue()
					}
				}
				val = m.GetGauge().GetValue()
				mtuFound[ifn] = val
			}
		}
	}
	if v, ok := getL(mtuFound, "eth3"); !ok || int(v) != 1500 {
		t.Fatalf("expected eth3 mtu 1500 got %v present=%v", v, ok)
	}
}
