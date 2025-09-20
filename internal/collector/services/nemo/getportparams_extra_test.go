package nemo

import (
	"encoding/json"
	"testing"
)

func TestGetPortParamsFromMIBs_SetPortAndDuplex(t *testing.T) {
	// Build a status with base.<CAND>.LLIntf as a map to test SetPort extraction
	j := map[string]any{
		"status": map[string]any{
			"base": map[string]any{
				"ETHX": map[string]any{
					"LLIntf":              map[string]any{"WAN1": map[string]any{}},
					"CurrentBitRate":      2500,
					"MaxBitRateSupported": 10000,
					"MaxBitRateEnabled":   5000,
					"DuplexModeEnabled":   "1",
				},
			},
		},
	}
	b, _ := json.Marshal(j)
	pp, err := GetPortParamsFromMIBs(b, "ETHX")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if pp.SetPort != "WAN1" {
		t.Fatalf("expected SetPort WAN1, got %v", pp.SetPort)
	}
	if int(pp.CurrentBitRate) != 2500 {
		t.Fatalf("expected CurrentBitRate 2500, got %v", pp.CurrentBitRate)
	}
	if !pp.DuplexModeEnabled {
		t.Fatalf("expected DuplexModeEnabled true for string '1'")
	}
}

func TestParseNetDevStats_StatusAndTopLevel(t *testing.T) {
	// Case: data map present
	b := []byte(`{"status":true,"data":{"RxBytes":12345}}`)
	m, err := ParseNetDevStats(b)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if v, ok := m["RxBytes"].(float64); !ok || int(v) != 12345 {
		t.Fatalf("expected RxBytes 12345, got %v", m["RxBytes"])
	}

	// Case: top-level map with direct keys
	b2 := []byte(`{"RxBytes":200}`)
	m2, err := ParseNetDevStats(b2)
	if err != nil {
		t.Fatalf("unexpected err2: %v", err)
	}
	if v, ok := m2["RxBytes"].(float64); !ok || int(v) != 200 {
		t.Fatalf("expected RxBytes 200, got %v", m2["RxBytes"])
	}

	// malformed JSON returns error
	if _, err := ParseNetDevStats([]byte("not-json")); err == nil {
		t.Fatalf("expected error for malformed JSON")
	}
}
