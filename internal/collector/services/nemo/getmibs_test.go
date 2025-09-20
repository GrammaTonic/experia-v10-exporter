package nemo

import "testing"

func TestRequestBody(t *testing.T) {
	b := RequestBody("ETH0")
	if b == "" {
		t.Fatalf("expected non-empty request body")
	}
}
