package collector

import (
	"bytes"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/GrammaTonic/experia-v10-exporter/internal/testutil"
	"github.com/prometheus/client_golang/prometheus"
)

func TestCollect_StatsUnderTopLevelStatus(t *testing.T) {
	c := NewCollector(net.ParseIP("127.0.0.1"), "u", "p", 1*time.Second)
	c.client.Transport = testutil.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		b, _ := io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewReader(b))
		body := string(b)
		if strings.Contains(body, "createContext") {
			return testutil.MakeResp(`{"data":{"contextID":"CTX"}}`), nil
		}
		if strings.Contains(body, "getWANStatus") {
			return testutil.MakeResp(`{"status":true,"data":{}}`), nil
		}
		if strings.Contains(body, "getMIBs") {
			return testutil.MakeResp(testutil.SampleMibJSON), nil
		}
		if strings.Contains(body, "getNetDevStats") {
			// Return metrics under top-level 'status' map to exercise that branch
			return testutil.MakeResp(`{"status":{"RxBytes":123,"TxBytes":456}}`), nil
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
	foundRx := false
	for _, mf := range mfs {
		if mf.GetName() == "experia_v10_netdev_rx_bytes_total" {
			for _, m := range mf.GetMetric() {
				if m.GetGauge().GetValue() > 0 {
					foundRx = true
				}
			}
		}
	}
	if !foundRx {
		t.Fatalf("expected netdev_rx_bytes metric > 0 when stats under status map")
	}
}

func TestCollect_SetPortExtractionFromLLIntfMap(t *testing.T) {
	c := NewCollector(net.ParseIP("127.0.0.1"), "u", "p", 1*time.Second)
	// Build a MIBs response where base.ETH0.LLIntf is a map with key "WAN1"
	mib := `{"status":{"base":{"ETH0":{"LLIntf":{"WAN1":{}}},"ETH1":{}},"alias":{"ETH0":{"Alias":"wan0"}}}}`
	c.client.Transport = testutil.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		b, _ := io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewReader(b))
		body := string(b)
		if strings.Contains(body, "createContext") {
			return testutil.MakeResp(`{"data":{"contextID":"CTX"}}`), nil
		}
		if strings.Contains(body, "getMIBs") {
			// Return our crafted MIB for ETH0
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
	found := false
	for _, mf := range mfs {
		if mf.GetName() == "experia_v10_netdev_port_set_port_info" {
			for _, m := range mf.GetMetric() {
				for _, lp := range m.GetLabel() {
					if lp.GetName() == "set_port" && lp.GetValue() == "WAN1" {
						found = true
					}
				}
			}
		}
	}
	if !found {
		t.Fatalf("expected netdev_port_set_port_info with set_port=WAN1")
	}
}
