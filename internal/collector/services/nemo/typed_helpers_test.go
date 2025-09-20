package nemo

import (
	"fmt"
	"testing"

	"github.com/GrammaTonic/experia-v10-exporter/internal/testutil"
)

func TestAnyToFloatAndReaders(t *testing.T) {
	// numeric string
	if f, ok := anyToFloat("123.5"); !ok || f != 123.5 {
		t.Fatalf("anyToFloat failed for string")
	}
	// int
	if f, ok := anyToFloat(10); !ok || f != 10.0 {
		t.Fatalf("anyToFloat failed for int")
	}
}

func TestReadFloatAndString(t *testing.T) {
	norm := map[string]any{"alias": "foo", "mtu": 1500, "currentbitrate": "1000"}
	if s, ok := readString(norm, "alias"); !ok || s != "foo" {
		t.Fatalf("readString failed: %v %v", s, ok)
	}
	if f, ok := readFloat(norm, "mtu"); !ok || f != 1500 {
		t.Fatalf("readFloat failed: %v %v", f, ok)
	}
	if f, ok := readFloat(norm, "currentbitrate"); !ok || f != 1000 {
		t.Fatalf("readFloat failed for string numeric: %v %v", f, ok)
	}
}

func TestGetMIBsTyped_Sample(t *testing.T) {
	b := []byte(testutil.SampleMibJSON)
	mi, s, err := GetMIBsTyped(b, "ETH2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mi.Candidate != "ETH2" {
		t.Fatalf("expected candidate ETH2, got %s", mi.Candidate)
	}
	if mi.CurrentBitRate != 1000 {
		t.Fatalf("expected CurrentBitRate 1000, got %v", mi.CurrentBitRate)
	}
	if mi.MTU != 1500 {
		t.Fatalf("expected MTU 1500, got %v", mi.MTU)
	}
	// alias is provided under the status.alias map in the sample; ensure it's present
	if s == nil {
		t.Fatalf("expected non-nil status map")
	}
	if am, ok := s["alias"].(map[string]any); !ok {
		t.Fatalf("expected alias map in status")
	} else {
		if v, ok := am["ETH2"].(map[string]any); !ok {
			t.Fatalf("expected ETH2 alias map entry")
		} else {
			if a, ok := v["Alias"].(string); !ok || a == "" {
				t.Fatalf("expected non-empty Alias for ETH2, got %v", a)
			}
		}
	}
}

func TestGetNetDevStatsTyped_Sample(t *testing.T) {
	// build a SampleStatsFmt with known values
	resp := fmt.Sprintf(testutil.SampleStatsFmt,
		1, 2, 300, 400,
		5, 6, 7, 8,
		9, 10,
		11, 12, 13,
		14, 15, 16,
		17, 18, 19, 20, 21,
	)
	ns, err := GetNetDevStatsTyped([]byte(resp))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ns.RxPackets != 1 {
		t.Fatalf("expected RxPackets 1, got %v", ns.RxPackets)
	}
	if ns.TxBytes != 400 {
		t.Fatalf("expected TxBytes 400, got %v", ns.TxBytes)
	}
	if ns.TxCarrierErrors != 18 {
		t.Fatalf("expected TxCarrierErrors 18, got %v", ns.TxCarrierErrors)
	}
}
