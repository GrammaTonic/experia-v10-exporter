package nemo

import (
	"testing"
)

func TestParseNetDevStats_PrefersDataMap(t *testing.T) {
	// Build a small getNetDevStats JSON with data fields
	body := `{"status":true,"data":{"RxBytes":1234,"TxBytes":5678}}`
	m, err := ParseNetDevStats([]byte(body))
	if err != nil {
		t.Fatalf("ParseNetDevStats returned error: %v", err)
	}
	if m == nil {
		t.Fatalf("expected non-nil map")
	}
	if v, ok := m["RxBytes"]; !ok || v == nil {
		t.Fatalf("expected RxBytes in parsed map")
	}

	// Also ensure empty input returns nil without error
	mm, err := ParseNetDevStats([]byte(""))
	if err != nil {
		t.Fatalf("empty parse returned error: %v", err)
	}
	if mm != nil {
		t.Fatalf("expected nil map for empty input")
	}
}
