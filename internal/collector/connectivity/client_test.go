package connectivity

import (
	"testing"
	"time"
)

func TestNewHTTPClientHasJarAndTimeout(t *testing.T) {
	c := NewHTTPClient(250 * time.Millisecond)
	if c == nil {
		t.Fatalf("expected client, got nil")
	}
	if c.Jar == nil {
		t.Fatalf("expected cookie jar, got nil")
	}
	if c.Timeout != 250*time.Millisecond {
		t.Fatalf("unexpected timeout: %v", c.Timeout)
	}
}
