package nemo

import "testing"

func TestAnyToFloatEdgeCases(t *testing.T) {
	if _, ok := anyToFloat(""); ok {
		t.Fatalf("expected empty string to not parse")
	}
	if _, ok := anyToFloat("not-a-number"); ok {
		t.Fatalf("expected non-numeric string to not parse")
	}
	if f, ok := anyToFloat("42"); !ok || f != 42 {
		t.Fatalf("expected "+"42"+" -> 42, got %v %v", f, ok)
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
