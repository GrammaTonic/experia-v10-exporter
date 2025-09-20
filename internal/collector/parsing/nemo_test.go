package parsing_test

import (
	"encoding/json"
	"testing"

	nemo "github.com/GrammaTonic/experia-v10-exporter/internal/collector/parser/nemo"
)

func TestParseMIBs_basic(t *testing.T) {
	payload := map[string]any{
		"status": map[string]any{
			"base": map[string]any{
				"ETH0": map[string]any{
					"Name":       "eth0",
					"OperStatus": "up",
				},
			},
			"netdev": map[string]any{
				"ETH0": map[string]any{
					"RxBytes": 12345,
				},
			},
		},
	}
	b, _ := json.Marshal(payload)
	norm, _, err := nemo.ParseMIBs(b, "ETH0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if norm == nil {
		t.Fatalf("expected non-nil norm map")
	}
	if v, ok := norm["name"]; !ok || v != "eth0" {
		t.Fatalf("expected name=eth0, got %v", v)
	}
	if v, ok := norm["rxbytes"]; !ok || v.(float64) != 12345 {
		t.Fatalf("expected rxbytes=12345, got %v", v)
	}
}

func TestParseNetDevStats_basic(t *testing.T) {
	payload := map[string]any{
		"data": map[string]any{
			"ifstats": map[string]any{
				"ETH0": map[string]any{"tx": 10},
			},
		},
	}
	b, _ := json.Marshal(payload)
	m, err := nemo.ParseNetDevStats(b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m == nil {
		t.Fatalf("expected non-nil map")
	}
	if d, ok := m["ifstats"].(map[string]any); !ok {
		t.Fatalf("expected ifstats map, got %T", m["ifstats"])
	} else {
		if v, ok := d["ETH0"].(map[string]any); !ok || v["tx"].(float64) != 10 {
			t.Fatalf("unexpected ETH0.tx: %v", v)
		}
	}
}
