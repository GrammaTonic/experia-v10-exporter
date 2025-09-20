package nemo_test

import (
	"fmt"
	"testing"

	nemo "github.com/GrammaTonic/experia-v10-exporter/internal/collector/services/nemo"
	"github.com/GrammaTonic/experia-v10-exporter/internal/testutil"
)

func TestGetMIBsTyped_Sample(t *testing.T) {
	b := []byte(testutil.SampleMibJSON)
	mi, s, err := nemo.GetMIBsTyped(b, "ETH2")
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
	ns, err := nemo.GetNetDevStatsTyped([]byte(resp))
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
