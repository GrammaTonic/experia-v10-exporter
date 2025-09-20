package nmc

import "testing"

func TestRequestBody(t *testing.T) {
	body := RequestBody()
	if body == "" {
		t.Fatalf("expected non-empty request body")
	}
}
