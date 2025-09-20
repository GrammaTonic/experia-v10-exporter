package nemo

import "testing"

func TestRequestBodyStats(t *testing.T) {
	b := RequestBodyStats("ETH0")
	if b == "" {
		t.Fatalf("expected non-empty request body")
	}
}
