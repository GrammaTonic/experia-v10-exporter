package nemo

import (
	"testing"

	"github.com/GrammaTonic/experia-v10-exporter/internal/testutil"
)

func TestAnyToFloatAndReadHelpers(t *testing.T) {
	if f, ok := anyToFloat(float64(3.14)); !ok || f != 3.14 {
		t.Fatalf("anyToFloat float64 failed")
	}
	if f, ok := anyToFloat("42"); !ok || f != 42 {
		t.Fatalf("anyToFloat string failed")
	}

	norm := map[string]any{"mtu": "1500", "duplexmodeenabled": "true"}
	if v, ok := readFloat(norm, "MTU"); !ok || v != 1500 {
		t.Fatalf("readFloat failed")
	}
	if b, ok := readBool(norm, "duplexmodeenabled"); !ok || !b {
		t.Fatalf("readBool failed")
	}
	// use fixture to exercise ParseMIBs path in typed helpers
	mi, _, err := GetMIBsTyped([]byte(testutil.SampleMibJSON), "ETH2")
	if err != nil {
		t.Fatalf("GetMIBsTyped failed: %v", err)
	}
	if mi.Candidate != "ETH2" {
		t.Fatalf("unexpected candidate")
	}

	// basic smoke tests to increase coverage of small helpers
	_ = RequestBody("ETH0")
	_ = RequestBodyStats("ETH0")
	_ = RequestBodyWanMibs()
	_ = RequestBodyWanStats()
}

func TestAnyToFloatEdgeCases(t *testing.T) {
	if _, ok := anyToFloat(""); ok {
		t.Fatalf("expected empty string to not parse")
	}
	if _, ok := anyToFloat("not-a-number"); ok {
		t.Fatalf("expected non-numeric string to not parse")
	}
	if f, ok := anyToFloat("42"); !ok || f != 42 {
		t.Fatalf("expected 42 -> 42, got %v %v", f, ok)
	}
}

func TestReadBoolVariants(t *testing.T) {
	norm := map[string]any{
		"duplexmodeenabled": true,
		"duplexstr":         "true",
		"duplexnum":         "0",
	}

	if b, ok := readBool(norm, "duplexmodeenabled"); !ok || !b {
		t.Fatalf("expected duplexmodeenabled==true, got %v %v", b, ok)
	}
	if b, ok := readBool(norm, "duplexstr"); !ok || !b {
		t.Fatalf("expected duplexstr==true, got %v %v", b, ok)
	}
	if b, ok := readBool(norm, "duplexnum"); !ok || b {
		t.Fatalf("expected duplexnum==false, got %v %v", b, ok)
	}
	if _, ok := readBool(nil, "x"); ok {
		t.Fatalf("expected readBool(nil) to return not-ok")
	}
}
