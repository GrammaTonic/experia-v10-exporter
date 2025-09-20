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

func TestCollect_EmitsPortAndStats(t *testing.T) {
	c := NewCollector(net.ParseIP("127.0.0.1"), "u", "p", 1*time.Second)
	// Transport that returns auth, WAN, MIBs and NetDevStats
	c.client.Transport = testutil.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		b, _ := io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewReader(b))
		body := string(b)
		var respBody string
		if strings.Contains(body, "createContext") {
			respBody = `{"data":{"contextID":"CTX-TEST"}}`
		} else if strings.Contains(body, "getWANStatus") {
			respBody = `{"status":true,"data":{"ConnectionState":"Connected","LinkType":"ETH","Protocol":"DHCP","IPAddress":"1.2.3.4","MACAddress":"aa:bb"}}`
		} else if strings.Contains(body, "getMIBs") {
			respBody = testutil.SampleMibJSON
		} else if strings.Contains(body, "getNetDevStats") {
			// Use SampleStatsFmt with RxBytes=300 and TxBytes=400
			respBody = strings.ReplaceAll(testutil.SampleStatsFmt, "%d", "1")
		} else {
			respBody = `{"status":true}`
		}
		return testutil.MakeResp(respBody), nil
	})

	reg := prometheus.NewRegistry()
	if err := reg.Register(c); err != nil {
		t.Fatalf("register failed: %v", err)
	}

	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("gather failed: %v", err)
	}

	// Verify that at least one NetdevPortCurrentBitrate metric is present
	foundPortBitrate := false
	foundRxBytes := false
	for _, mf := range mfs {
		if mf.GetName() == "experia_v10_netdev_port_current_bitrate_mbps" {
			for _, m := range mf.GetMetric() {
				if m.GetGauge().GetValue() > 0 {
					foundPortBitrate = true
				}
			}
		}
		if mf.GetName() == "experia_v10_netdev_rx_bytes_total" {
			for _, m := range mf.GetMetric() {
				if m.GetGauge().GetValue() > 0 {
					foundRxBytes = true
				}
			}
		}
	}
	if !foundPortBitrate {
		t.Fatalf("expected at least one NetdevPortCurrentBitrate > 0")
	}
	if !foundRxBytes {
		t.Fatalf("expected at least one NetdevRxBytes > 0")
	}
}
