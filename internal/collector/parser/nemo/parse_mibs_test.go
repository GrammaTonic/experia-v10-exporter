package nemo

import (
	"testing"

	"github.com/GrammaTonic/experia-v10-exporter/internal/testutil"
)

func TestFindStatusNestedAndTopLevel(t *testing.T) {
	// nested status under data
	data := map[string]any{
		"data": map[string]any{
			"foo": map[string]any{
				"base": map[string]any{
					"ETH0": map[string]any{"MTU": 1500},
				},
			},
		},
	}
	if got := findStatus(data); got == nil {
		t.Fatalf("expected to find status in nested map")
	}

	// top-level status map without base/netdev should NOT be considered a
	// 'status' container by findStatus (it looks for base/netdev), so expect nil
	data2 := map[string]any{"status": map[string]any{"Status": true}}
	if got := findStatus(data2); got != nil {
		t.Fatalf("expected findStatus to return nil for status without base/netdev")
	}
}

func TestParseMIBsVariousShapes(t *testing.T) {
	// base and netdev present
	j := `{"status": {"base": {"ETH0": {"MTU": 1400}}, "netdev": {"eth0": {"Flags":"f"}}}}`
	norm, s, err := ParseMIBs([]byte(j), "ETH0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if norm == nil || s == nil {
		t.Fatalf("expected non-nil norm and status")
	}
	if v, ok := norm["mtu"]; !ok || v == nil {
		t.Fatalf("expected mtu in norm")
	}

	// empty data returns nil
	norm, s, err = ParseMIBs([]byte(``), "ETH0")
	if err != nil || norm != nil || s != nil {
		t.Fatalf("expected nil results for empty input")
	}

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
