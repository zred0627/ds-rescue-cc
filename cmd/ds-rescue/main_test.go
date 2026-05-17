package main

import (
	"strings"
	"testing"
)

func TestVersionString(t *testing.T) {
	got := versionString()
	if !strings.HasPrefix(got, "ds-rescue v") {
		t.Errorf("expected prefix 'ds-rescue v', got: %q", got)
	}
	if !strings.Contains(got, Version) {
		t.Errorf("expected version %q embedded, got: %q", Version, got)
	}
}
