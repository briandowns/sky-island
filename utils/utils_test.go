package utils

import (
	"testing"
)

// TestExists_Exits
func TestExists_Exits(t *testing.T) {
	if !Exists("/tmp") {
		t.Error("expected /tmp to exists however reported otherwise")
	}
}

// TestExists_NotExits
func TestExists_NotExits(t *testing.T) {
	if Exists("/tasdfasdfmp") {
		t.Error("expected /tasdfasdfmp to not exist however reported otherwise")
	}
}
