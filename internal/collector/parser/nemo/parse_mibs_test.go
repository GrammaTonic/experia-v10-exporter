package nemo

import (
	"testing"

	"github.com/GrammaTonic/experia-v10-exporter/internal/testutil"
)

func TestParseMIBs_ExtractsBaseAndNetdev(t *testing.T) {
	// Use SampleMibJSON fixture which contains base and netdev sections
	norm, raw, err := ParseMIBs([]byte(testutil.SampleMibJSON), "ETH2")
	if err != nil {
		t.Fatalf("ParseMIBs returned error: %v", err)
	}
	if norm == nil {
		t.Fatalf("expected non-nil norm map")
	}
	// SampleMibJSON contains CurrentBitRate for ETH2 in base
	if v, ok := norm["currentbitrate"]; !ok || v == nil {
		t.Fatalf("expected currentbitrate in norm map")
	}
	// raw status map should include 'base'
	if raw == nil {
		t.Fatalf("expected raw status map present")
	}
	if _, ok := raw["base"]; !ok {
		t.Fatalf("expected raw status to contain base")
	}
}
