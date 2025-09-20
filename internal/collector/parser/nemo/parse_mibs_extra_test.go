package nemo

import (
	"encoding/json"
	"testing"
)

func TestFindStatusNil(t *testing.T) {
	if got := findStatus(nil); got != nil {
		t.Fatalf("expected nil when input nil, got %#v", got)
	}
}

func TestParseMIBs_VariousShapes(t *testing.T) {
	// Case: top-level data->status->base with lowercase candidate
	j := map[string]any{
		"status": map[string]any{
			"base": map[string]any{
				"eth1": map[string]any{"MTU": 1400},
			},
		},
	}
	b, _ := json.Marshal(j)
	norm, s, err := ParseMIBs(b, "ETH1")
	if err != nil || norm == nil || s == nil {
		t.Fatalf("unexpected nil/err: %v %v %v", err, norm, s)
	}
	if mtu, ok := norm["mtu"].(float64); !ok || int(mtu) != 1400 {
		t.Fatalf("expected mtu 1400, got %v", norm["mtu"])
	}

	// Case: netdev nested under another key
	j2 := map[string]any{
		"outer": map[string]any{
			"netdev": map[string]any{
				"ETH2": map[string]any{"Alias": "wan"},
			},
		},
	}
	b2, _ := json.Marshal(j2)
	norm2, s2, err := ParseMIBs(b2, "ETH2")
	if err != nil || norm2 == nil || s2 == nil {
		t.Fatalf("unexpected nil/err for nested netdev: %v %v %v", err, norm2, s2)
	}
	if a, ok := norm2["alias"].(string); !ok || a != "wan" {
		t.Fatalf("expected alias 'wan', got %v", norm2["alias"])
	}
}
