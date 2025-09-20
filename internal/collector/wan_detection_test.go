package collector

import (
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/GrammaTonic/experia-v10-exporter/internal/testutil"
	"github.com/prometheus/client_golang/prometheus"
)

// TestWanDetection ensures that when getWANStatus reports a MAC that matches
// an interface's LLAddress in the MIBs response, the collector sets the
// alias to "wan" for that interface's NetdevInfo metric.
func TestWanDetection(t *testing.T) {
	// Build transport that returns getWANStatus for NMC call and sample MIBs
	wanJSON := `{"status":true,"data":{"MACAddress":"AA:BB:CC:DD:EE:FF","ConnectionState":"Connected","LinkType":"ethernet","Protocol":"ppp","IPAddress":"1.2.3.4"}}`
	// Build a sample MIBs response for ETH0 where LLAddress matches WAN MAC
	mibETH0 := `{"status":true,"data":{"status":{"base":{"ETH0":{"LLAddress":"AA:BB:CC:DD:EE:FF","NetDevState":"up","MTU":1500}}}}}`
	// Other candidates return empty

	rt := testutil.MakeRoundTripper(func(req *http.Request) (*http.Response, error) {
		// read body for simple string checks
		buf := ""
		if req != nil && req.Body != nil {
			b, _ := io.ReadAll(req.Body)
			buf = string(b)
		}
		// authentication createContext payload contains sah.Device.Information
		if strings.Contains(buf, "sah.Device.Information") {
			return testutil.MakeResp(`{"data":{"contextID":"CTX-TEST"}}`), nil
		}
		// follow-up GET after auth may be an empty body; return empty JSON
		if req.Method == "GET" {
			return testutil.MakeResp(`{}`), nil
		}
		if strings.Contains(buf, "getWANStatus") || strings.Contains(buf, "NMC") {
			return testutil.MakeResp(wanJSON), nil
		}
		if strings.Contains(buf, "getMIBs") || strings.Contains(buf, "NeMo.Intf.ETH0") {
			return testutil.MakeResp(mibETH0), nil
		}
		return testutil.MakeResp(""), nil
	})

	c := NewCollector(net.ParseIP("192.0.2.1"), "u", "p", 1*time.Second)
	// inject the test transport into the client's transport
	c.client.Transport = rt

	// perform Login so the collector has a session token for subsequent POSTs
	if err := c.Login(); err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	// create a registry and register the collector
	reg := prometheus.NewRegistry()
	reg.MustRegister(c)

	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("Gather() error: %v", err)
	}

	// find netdev_info for eth1 (first candidate -> eth1). Newer behavior
	// may not overwrite the alias label unless EXPERIA_FORCE_WAN_ALIAS is set.
	// Accept either the alias being "wan" on netdev_info OR the presence of
	// the dedicated experia_v10_wan_ifname metric for eth1.
	foundAlias := false
	foundWanIfname := false
	for _, mf := range mfs {
		if mf.GetName() == "experia_v10_netdev_info" {
			for _, m := range mf.GetMetric() {
				hasIfname := false
				for _, lp := range m.GetLabel() {
					if lp.GetName() == "ifname" && lp.GetValue() == "eth1" {
						hasIfname = true
						break
					}
				}
				if !hasIfname {
					continue
				}
				// find alias label
				for _, lp2 := range m.GetLabel() {
					if lp2.GetName() == "alias" {
						if lp2.GetValue() == "wan" {
							foundAlias = true
							break
						}
					}
				}
			}
		}
		if mf.GetName() == "experia_v10_wan_ifname" {
			for _, m := range mf.GetMetric() {
				for _, lp := range m.GetLabel() {
					if lp.GetName() == "ifname" && lp.GetValue() == "eth1" {
						foundWanIfname = true
						break
					}
				}
				if foundWanIfname {
					break
				}
			}
		}
		if foundAlias || foundWanIfname {
			break
		}
	}
	if !foundAlias && !foundWanIfname {
		t.Fatalf("expected netdev_info alias to contain 'wan' for eth1 or a wan_ifname metric, but didn't find either; metrics: %v", len(mfs))
	}
}
