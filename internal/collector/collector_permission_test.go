package collector

import (
	"bytes"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/GrammaTonic/experia-v10-exporter/internal/collector/metrics"
	"github.com/GrammaTonic/experia-v10-exporter/internal/testutil"
	"github.com/prometheus/client_golang/prometheus"
)

func TestCollect_PermissionDeniedIncrementsPermissionErrors(t *testing.T) {
	c := NewCollector(net.ParseIP("127.0.0.1"), "u", "p", 1*time.Second)
	c.client.Transport = testutil.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		b, _ := io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewReader(b))
		body := string(b)
		if strings.Contains(body, "createContext") {
			return testutil.MakeResp(`{"data":{"contextID":"CTX"}}`), nil
		}
		if strings.Contains(body, "getWANStatus") {
			return testutil.MakeResp(`{"status":false,"errors":[{"error":1,"description":"Permission denied","info":""}]}`), nil
		}
		// Default return empty
		return testutil.MakeResp(`{"status":true}`), nil
	})

	before := testutil.ReadCounterValue(metrics.PermissionErrors)
	reg := prometheus.NewRegistry()
	if err := reg.Register(c); err != nil {
		t.Fatalf("register failed: %v", err)
	}
	if _, err := reg.Gather(); err != nil {
		t.Fatalf("gather failed: %v", err)
	}
	after := testutil.ReadCounterValue(metrics.PermissionErrors)
	if after <= before {
		t.Fatalf("expected PermissionErrors to increase, before=%v after=%v", before, after)
	}
}

func TestCollect_StatusBoolOverridesNetDevState(t *testing.T) {
	c := NewCollector(net.ParseIP("127.0.0.1"), "u", "p", 1*time.Second)
	// For ETH0 (label eth1) return MIBs with NetDevState=down but top-level Status=true
	mib := `{"status":{"Status":true,"base":{"ETH0":{"NetDevState":"down","MTU":1400}},"alias":{"ETH0":{"Alias":"wan0"}}}}`
	c.client.Transport = testutil.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		b, _ := io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewReader(b))
		body := string(b)
		if strings.Contains(body, "createContext") {
			return testutil.MakeResp(`{"data":{"contextID":"CTX"}}`), nil
		}
		if strings.Contains(body, "getMIBs") {
			// return mib for ETH0, empty for others
			if strings.Contains(body, "ETH0") {
				return testutil.MakeResp(mib), nil
			}
			return testutil.MakeResp(``), nil
		}
		if strings.Contains(body, "getNetDevStats") {
			return testutil.MakeResp(`{"status":true}`), nil
		}
		return testutil.MakeResp(`{"status":true}`), nil
	})

	reg := prometheus.NewRegistry()
	if err := reg.Register(c); err != nil {
		t.Fatalf("register failed: %v", err)
	}
	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("gather failed: %v", err)
	}

	// Look for netdev_up for ifname=eth1
	found := false
	for _, mf := range mfs {
		if mf.GetName() == metrics.MetricPrefix+"netdev_up" {
			for _, m := range mf.GetMetric() {
				for _, lp := range m.GetLabel() {
					if lp.GetName() == "ifname" && lp.GetValue() == "eth1" {
						if m.GetGauge().GetValue() == 1 {
							found = true
						}
					}
				}
			}
		}
	}
	if !found {
		t.Fatalf("expected netdev_up for eth1 to be 1 due to Status boolean override")
	}
}
