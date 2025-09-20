package collector

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func TestCollect_UsesOutputJSON_DiscoverAndExportETH0Eth1(t *testing.T) {
	// Read the canned output.json from the package
	data, err := os.ReadFile("output.json")
	if err != nil {
		t.Fatalf("failed to read output.json: %v", err)
	}

	c := NewCollector(net.ParseIP("127.0.0.1"), "u", "p", 1*time.Second)
	c.client.Transport = roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		b, _ := io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewReader(b))
		var respBody string
		if bytes.Contains(b, []byte("createContext")) {
			respBody = `{"data":{"contextID":"CTX-TEST"}}`
		} else if bytes.Contains(b, []byte("getWANStatus")) {
			respBody = `{"status":true,"data":{"ConnectionState":"Connected","LinkType":"ETH","Protocol":"DHCP","IPAddress":"1.2.3.4","MACAddress":"aa:bb:cc:dd:ee:ff"}}`
		} else if bytes.Contains(b, []byte("getMIBs")) {
			respBody = string(data)
		} else {
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

	// (no debug logging)

	// Helper to find metric by family and collect values by ifname
	gatherByName := func(family string) map[string]float64 {
		out := map[string]float64{}
		for _, mf := range mfs {
			if mf.GetName() == metricPrefix+family {
				for _, m := range mf.GetMetric() {
					var ifn string
					var val float64
					for _, lp := range m.GetLabel() {
						if lp.GetName() == "ifname" {
							ifn = strings.ToLower(lp.GetValue())
						}
					}
					if m.GetGauge() != nil {
						val = m.GetGauge().GetValue()
					}
					if ifn != "" {
						out[ifn] = val
					}
				}
			}
		}
		return out
	}

	up := gatherByName("netdev_up")
	mtu := gatherByName("netdev_mtu")
	tx := gatherByName("netdev_tx_queue_len")

	// Collector now emits stable labels eth1..ethN for candidates ETH0..ETHN-1.
	// Ensure eth1..eth4 exist and report expected MTU and up state.
	for i := 1; i <= 4; i++ {
		label := fmt.Sprintf("eth%d", i)
		if v, ok := up[label]; !ok || v != 1 {
			t.Fatalf("expected %s up==1, got %v present=%v", label, v, ok)
		}
		if v, ok := mtu[label]; !ok || int(v) != 1500 {
			t.Fatalf("expected %s mtu 1500 got %v present=%v", label, v, ok)
		}
	}

	// tx queue lengths should include both 0 and 1000 across the labels; the
	// exact mapping may vary so assert the set membership rather than exact
	// per-label values.
	seen := map[int]bool{}
	for i := 1; i <= 4; i++ {
		label := fmt.Sprintf("eth%d", i)
		if v, ok := tx[label]; ok {
			seen[int(v)] = true
		}
	}
	if !seen[0] || !seen[1000] {
		t.Fatalf("expected tx values to include 0 and 1000, saw=%v", seen)
	}
}
