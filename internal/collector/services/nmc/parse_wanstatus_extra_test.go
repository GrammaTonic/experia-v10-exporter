package nmc

import "testing"

func TestParseWANStatus_EmptyAndError(t *testing.T) {
	// empty input
	s, err := ParseWANStatus([]byte(nil))
	if err != nil {
		t.Fatalf("unexpected err for empty: %v", err)
	}
	if s.Status {
		t.Fatalf("expected default Status false for empty input")
	}

	// malformed JSON should return error
	if _, err := ParseWANStatus([]byte("not-json")); err == nil {
		t.Fatalf("expected error for malformed JSON")
	}
}
