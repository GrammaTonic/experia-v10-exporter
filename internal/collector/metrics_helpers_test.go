package collector

import "testing"

func TestBuildMIBMetrics(t *testing.T) {
	norm := map[string]any{
		"status":         true,
		"mtu":            1500,
		"txqueuelen":     1000,
		"currentbitrate": 1000,
		"lastchangetime": 42,
		"lladdress":      "aa:bb:cc",
		"netdevflags":    "enabled",
		"netdevtype":     "ethernet",
	}
	status := map[string]any{"alias": map[string]any{"Alias": "cpe-eth0"}}

	m := BuildMIBMetrics(norm, status)
	if m.Up != 1.0 {
		t.Fatalf("expected Up==1.0, got %v", m.Up)
	}
	if int(m.Mtu) != 1500 {
		t.Fatalf("expected Mtu==1500, got %v", m.Mtu)
	}
	if int(m.TxQueueLen) != 1000 {
		t.Fatalf("expected TxQueueLen==1000, got %v", m.TxQueueLen)
	}
	if int(m.SpeedMbps) != 1000 {
		t.Fatalf("expected SpeedMbps==1000, got %v", m.SpeedMbps)
	}
	if int(m.LastChange) != 42 {
		t.Fatalf("expected LastChange==42, got %v", m.LastChange)
	}
	if m.Lladdr != "aa:bb:cc" {
		t.Fatalf("expected Lladdr, got %v", m.Lladdr)
	}
	if m.Flags != "enabled" {
		t.Fatalf("expected Flags, got %v", m.Flags)
	}
	if m.Type != "ethernet" {
		t.Fatalf("expected Type, got %v", m.Type)
	}
	if m.Alias != "cpe-eth0" {
		t.Fatalf("expected Alias, got %v", m.Alias)
	}
}

func TestBuildNetDevStatsMetrics(t *testing.T) {
	data := map[string]any{
		"RxPackets":    "123",
		"TxBytes":      456,
		"CustomMetric": 7.5,
	}
	res := BuildNetDevStatsMetrics(data)
	if v, ok := res.Values["RxPackets"]; !ok || int(v) != 123 {
		t.Fatalf("expected RxPackets==123 got %v present=%v", v, ok)
	}
	if v, ok := res.Values["TxBytes"]; !ok || int(v) != 456 {
		t.Fatalf("expected TxBytes==456 got %v present=%v", v, ok)
	}
	if v, ok := res.Values["CustomMetric"]; !ok || v != 7.5 {
		t.Fatalf("expected CustomMetric==7.5 got %v present=%v", v, ok)
	}
}
