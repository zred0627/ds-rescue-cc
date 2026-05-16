package main

import (
	"strings"
	"testing"
)

// TestVersionString verifies the version line is well-formed without spawning
// a subprocess (avoids Windows Smart App Control blocking go-build temp binaries
// and avoids cwd-sensitivity of `go run .`).
func TestVersionString(t *testing.T) {
	got := versionString()
	if !strings.HasPrefix(got, "ds-rescue v") {
		t.Errorf("expected prefix 'ds-rescue v', got: %q", got)
	}
	if !strings.Contains(got, Version) {
		t.Errorf("expected version %q embedded, got: %q", Version, got)
	}
}
