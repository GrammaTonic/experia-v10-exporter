package collector

import (
	"testing"
)

func TestHelloWorld(t *testing.T) {
	if 1+1 != 2 {
		t.Errorf("Expected %d, but got %d", 2, 1+1)
	}
}
