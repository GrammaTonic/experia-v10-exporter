package collector

import "testing"

func TestBuildMIBMetrics_Basic(t *testing.T) {
	norm := map[string]any{
		"status":         true,
		"netdevstate":    "up",
		"mtu":            "1500",
		"txqueuelen":     10,
		"currentbitrate": "1000",
		"lastchangetime": 7,
		"lladdress":      "aa:bb",
		"netdevtype":     "eth",
		"netdevflags":    "flag1",
	}
	status := map[string]any{"alias": map[string]any{"Alias": "wan0"}}

	m := BuildMIBMetrics(norm, status)
	if m.Up != 1.0 {
		t.Fatalf("expected Up=1.0, got %v", m.Up)
	}
	if m.Mtu != 1500 {
		t.Fatalf("expected Mtu=1500, got %v", m.Mtu)
	}
	if m.SpeedMbps != 1000 {
		t.Fatalf("expected SpeedMbps=1000, got %v", m.SpeedMbps)
	}
	if m.Lladdr != "aa:bb" {
		t.Fatalf("expected Lladdr=aa:bb, got %v", m.Lladdr)
	}
	if m.Alias != "wan0" {
		t.Fatalf("expected Alias=wan0, got %v", m.Alias)
	}
}

func TestBuildNetDevStatsMetrics_StringAndNumeric(t *testing.T) {
	data := map[string]any{"RxBytes": "300", "TxBytes": 400}
	res := BuildNetDevStatsMetrics(data)
	if res.Values["RxBytes"] != 300 {
		t.Fatalf("expected RxBytes 300, got %v", res.Values["RxBytes"])
	}
	if res.Values["TxBytes"] != 400 {
		t.Fatalf("expected TxBytes 400, got %v", res.Values["TxBytes"])
	}
}
