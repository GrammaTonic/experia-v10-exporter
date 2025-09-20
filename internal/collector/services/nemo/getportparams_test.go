package nemo

import (
	"testing"

	"github.com/GrammaTonic/experia-v10-exporter/internal/testutil"
)

func TestGetPortParamsFromMIBs_ExtractsSetPort(t *testing.T) {
	b := []byte(testutil.SampleMibJSON)
	pp, err := GetPortParamsFromMIBs(b, "ETH2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pp.CurrentBitRate == 0 {
		t.Fatalf("expected CurrentBitRate non-zero")
	}
}
