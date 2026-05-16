package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveKeyPath_EnvVarPriority(t *testing.T) {
	t.Setenv("DEEPSEEK_API_KEY", "sk-test-env-value")
	key, src, err := ResolveAPIKey()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key != "sk-test-env-value" {
		t.Errorf("expected env value, got: %s", key)
	}
	if src != "env" {
		t.Errorf("expected source 'env', got: %s", src)
	}
}

func TestResolveKeyPath_HomeFileFallback(t *testing.T) {
	t.Setenv("DEEPSEEK_API_KEY", "")
	t.Setenv("XDG_CONFIG_HOME", "")
	home := t.TempDir()
	t.Setenv("HOME", home)
	keyDir := filepath.Join(home, ".ds-rescue")
	os.MkdirAll(keyDir, 0700)
	keyFile := filepath.Join(keyDir, "key")
	os.WriteFile(keyFile, []byte("sk-test-file-value\n"), 0600)

	key, src, err := ResolveAPIKey()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key != "sk-test-file-value" {
		t.Errorf("expected file value (trimmed), got: %q", key)
	}
	if src != "file:"+keyFile {
		t.Errorf("expected source 'file:%s', got: %s", keyFile, src)
	}
}

func TestResolveKeyPath_NotFound(t *testing.T) {
	t.Setenv("DEEPSEEK_API_KEY", "")
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("HOME", t.TempDir())
	_, _, err := ResolveAPIKey()
	if err == nil {
		t.Error("expected error when no key found, got nil")
	}
}
