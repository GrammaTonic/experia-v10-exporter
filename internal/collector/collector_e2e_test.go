package collector

import (
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

// This is a local end-to-end test that talks to a real Experia V10 modem.
// It's intentionally disabled by default. To run it locally, set the
// following environment variables and enable the test runner:
//
// EXPERIA_E2E=1 EXPERIA_IP=192.168.2.254 EXPERIA_USER=admin EXPERIA_PASS=secret go test ./internal/collector -run TestE2E
//
// The test will be skipped unless EXPERIA_E2E is set to "1" and the other
// variables are present. It avoids printing secrets and uses a short timeout.
func TestE2E(t *testing.T) {
	if os.Getenv("EXPERIA_E2E") != "1" {
		t.Skip("skipping Experia E2E test; set EXPERIA_E2E=1 to enable")
	}

	ip := os.Getenv("EXPERIA_IP")
	user := os.Getenv("EXPERIA_USER")
	pass := os.Getenv("EXPERIA_PASS")
	if ip == "" || user == "" || pass == "" {
		t.Skip("skipping Experia E2E test; EXPERIA_IP, EXPERIA_USER and EXPERIA_PASS must be set")
	}

	// Parse IP
	parsed := net.ParseIP(ip)
	if parsed == nil {
		t.Fatalf("invalid EXPERIA_IP: %s", ip)
	}

	// Create collector with a modest timeout so test fails fast on network issues.
	c := NewCollector(parsed, user, pass, 8*time.Second)

	// Register the collector with a dedicated registry and Gather metrics.
	// This exercises Describe/Collect and gives us parsed MetricFamily objects
	// we can assert against.
	reg := prometheus.NewRegistry()
	if err := reg.Register(c); err != nil {
		t.Fatalf("failed to register collector: %v", err)
	}

	// Gather in a goroutine and enforce a timeout so test doesn't hang.
	type gatherResult struct {
		mfs []*dto.MetricFamily
		err error
	}
	resCh := make(chan gatherResult, 1)
	go func() {
		mfs, err := reg.Gather()
		resCh <- gatherResult{mfs: mfs, err: err}
	}()

	var res gatherResult
	select {
	case res = <-resCh:
		if res.err != nil {
			t.Fatalf("gather failed: %v", res.err)
		}
	case <-time.After(30 * time.Second):
		t.Fatal("timeout waiting for registry.Gather to finish")
	}

	// Look for the internet_connection metric family and assert expected connection_state.
	famName := metricPrefix + "internet_connection"
	var found bool
	var foundStates []string
	for _, mf := range res.mfs {
		if mf.GetName() == famName {
			for _, m := range mf.Metric {
				// find label pair named "connection_state"
				for _, lp := range m.Label {
					if lp.GetName() == "connection_state" {
						found = true
						foundStates = append(foundStates, lp.GetValue())
					}
				}
			}
		}
	}

	if !found {
		t.Fatalf("metric %s not found in gathered metrics (got %d families)", famName, len(res.mfs))
	}

	expected := os.Getenv("EXPERIA_EXPECT_CONNECTION_STATE")
	if expected == "" {
		expected = "Connected"
	}

	// Verify at least one metric has the expected state.
	var matched bool
	for _, s := range foundStates {
		if s == expected {
			matched = true
			break
		}
	}
	if !matched {
		t.Fatalf("expected connection_state=%q not found; got states=%v", expected, foundStates)
	}

	// Optional: validate netdev_up metrics if the caller set EXPERIA_EXPECT_NETDEV_IFACES
	// Provide a comma-separated list of interface names (e.g. "ETH0,ETH1,ETH2,ETH3").
	// If not set, the test will simply log presence/absence of netdev metrics.
	expectIfaces := os.Getenv("EXPERIA_EXPECT_NETDEV_IFACES")
	var expectedIfaces []string
	if expectIfaces != "" {
		for _, p := range strings.Split(expectIfaces, ",") {
			vv := strings.ToUpper(strings.TrimSpace(p))
			if vv != "" {
				expectedIfaces = append(expectedIfaces, vv)
			}
		}
	}

	// Look for netdev_up and collect ifnames (uppercase) -> values
	netdevName := metricPrefix + "netdev_up"
	netdevFound := false
	netdevIfs := map[string]float64{}
	for _, mf := range res.mfs {
		if mf.GetName() == netdevName {
			netdevFound = true
			for _, m := range mf.Metric {
				var ifn string
				val := m.GetGauge().GetValue()
				for _, lp := range m.Label {
					if lp.GetName() == "ifname" {
						ifn = strings.ToUpper(lp.GetValue())
					}
				}
				if ifn != "" {
					netdevIfs[ifn] = val
				}
			}
		}
	}

	if !netdevFound {
		t.Logf("netdev_up metric family not found; skipping netdev interface checks")
		return
	}

	t.Logf("found netdev_up interfaces: %v", netdevIfs)

	if len(expectedIfaces) == 0 {
		// No explicit expectation set; just return now that we've parsed netdevs.
		return
	}

	// Assert that each expected interface is present and appears up (value==1)
	// If EXPERIA_STRICT_NETDEV is set to "1" then we fail on mismatches. Otherwise
	// we only log warnings so the e2e run doesn't fail if the modem reports interfaces
	// as down (common in some setups).
	strict := os.Getenv("EXPERIA_STRICT_NETDEV") == "1"
	for _, e := range expectedIfaces {
		v, ok := netdevIfs[e]
		if !ok {
			if strict {
				t.Fatalf("expected interface %s not present in netdev_up metrics", e)
			}
			t.Logf("warning: expected interface %s not present in netdev_up metrics", e)
			continue
		}
		if v != 1.0 {
			if strict {
				t.Fatalf("expected interface %s to be up (1.0), got %v", e, v)
			}
			t.Logf("warning: interface %s is present but not up (value=%v)", e, v)
		}
	}
}

// No additional shims required; we import prometheus and use its Metric type.
