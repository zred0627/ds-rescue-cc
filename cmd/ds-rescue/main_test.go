package main

import (
	"os/exec"
	"strings"
	"testing"
)

func TestVersionFlag(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "--version")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("exit error: %v, output: %s", err, out)
	}
	if !strings.Contains(string(out), "ds-rescue v") {
		t.Errorf("expected version string with 'ds-rescue v' prefix, got: %s", out)
	}
}
