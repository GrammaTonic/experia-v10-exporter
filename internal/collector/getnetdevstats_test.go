package collector

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	metrics "github.com/GrammaTonic/experia-v10-exporter/internal/collector/metrics"
	"github.com/GrammaTonic/experia-v10-exporter/internal/testutil"
	"github.com/prometheus/client_golang/prometheus"
)

// sampleStatsFmt provided by testhelpers_test.go

func TestCollect_ParsesGetNetDevStats(t *testing.T) {
	// Create collector and override client transport to return canned responses
	c := NewCollector(net.ParseIP("127.0.0.1"), "u", "p", 1*time.Second)
	c.client.Transport = testutil.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		b, _ := io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewReader(b))

		var respBody string
		switch {
		case bytes.Contains(b, []byte("createContext")):
			respBody = `{"data":{"contextID":"CTX-TEST"}}`
		case bytes.Contains(b, []byte("getWANStatus")):
			respBody = `{"status":true,"data":{"ConnectionState":"Connected","LinkType":"ETH","Protocol":"DHCP","IPAddress":"1.2.3.4","MACAddress":"aa:bb:cc:dd:ee:ff"}}`
		case bytes.Contains(b, []byte("getMIBs")):
			// Reuse the existing sample that provides a usable 'base' section so
			// the collector proceeds to call getNetDevStats for each candidate.
			respBody = testutil.SampleMibJSON
		case bytes.Contains(b, []byte("getNetDevStats")):
			// Determine which interface was requested and return distinct
			// numeric values per candidate so we can assert label mapping.
			if bytes.Contains(b, []byte("ETH0")) {
				respBody = fmt.Sprintf(testutil.SampleStatsFmt, 101, 201, 1001, 2001, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17)
			} else if bytes.Contains(b, []byte("ETH1")) {
				respBody = fmt.Sprintf(testutil.SampleStatsFmt, 102, 202, 1002, 2002, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18)
			} else if bytes.Contains(b, []byte("ETH2")) {
				respBody = fmt.Sprintf(testutil.SampleStatsFmt, 103, 203, 1003, 2003, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19)
			} else if bytes.Contains(b, []byte("ETH3")) {
				respBody = fmt.Sprintf(testutil.SampleStatsFmt, 104, 204, 1004, 2004, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20)
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

	// Register collector with a fresh registry and call Gather
	reg := prometheus.NewRegistry()
	if err := reg.Register(c); err != nil {
		t.Fatalf("failed to register collector: %v", err)
	}

	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("gather failed: %v", err)
	}

	// Helper to find metrics by family name and build map[ifname]value
	extract := func(name string) map[string]float64 {
		out := map[string]float64{}
		for _, mf := range mfs {
			if mf.GetName() != metrics.MetricPrefix+name {
				continue
			}
			for _, m := range mf.GetMetric() {
				var ifn string
				val := m.GetGauge().GetValue()
				for _, lp := range m.GetLabel() {
					if lp.GetName() == "ifname" {
						ifn = lp.GetValue()
					}
				}
				out[ifn] = val
			}
		}
		return out
	}

	rxPackets := extract("netdev_rx_packets_total")
	txPackets := extract("netdev_tx_packets_total")
	rxBytes := extract("netdev_rx_bytes_total")
	txBytes := extract("netdev_tx_bytes_total")

	// Expect candidate ETH0 -> label eth1 with RxPackets 101 etc.
	expected := map[string]struct{ rx, txPkt, rxb, txb float64 }{
		"eth1": {rx: 101, txPkt: 201, rxb: 1001, txb: 2001},
		"eth2": {rx: 102, txPkt: 202, rxb: 1002, txb: 2002},
		"eth3": {rx: 103, txPkt: 203, rxb: 1003, txb: 2003},
		"eth4": {rx: 104, txPkt: 204, rxb: 1004, txb: 2004},
	}

	for ifn, exp := range expected {
		if v, ok := rxPackets[ifn]; !ok || v != exp.rx {
			t.Fatalf("expected %s RxPackets=%v present=%v got=%v", ifn, exp.rx, ok, v)
		}
		if v, ok := txPackets[ifn]; !ok || v != exp.txPkt {
			t.Fatalf("expected %s TxPackets=%v present=%v got=%v", ifn, exp.txPkt, ok, v)
		}
		if v, ok := rxBytes[ifn]; !ok || v != exp.rxb {
			t.Fatalf("expected %s RxBytes=%v present=%v got=%v", ifn, exp.rxb, ok, v)
		}
		if v, ok := txBytes[ifn]; !ok || v != exp.txb {
			t.Fatalf("expected %s TxBytes=%v present=%v got=%v", ifn, exp.txb, ok, v)
		}
	}
}
