package nmc

import "testing"

func TestParseWANStatusEmpty(t *testing.T) {
	var empty []byte
	s, err := ParseWANStatus(empty)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if s.Data.IPAddress != "" {
		t.Fatalf("expected empty data, got %#v", s)
	}
}

func TestParseWANStatusValid(t *testing.T) {
	b := []byte(`{"status":true,"data":{"ConnectionState":"Connected","IPAddress":"1.2.3.4"}}`)
	s, err := ParseWANStatus(b)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !s.Status || s.Data.ConnectionState != "Connected" || s.Data.IPAddress != "1.2.3.4" {
		t.Fatalf("unexpected parse result: %#v", s)
	}
}
