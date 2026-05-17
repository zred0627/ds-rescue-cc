package loop

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/zred0627/ds-rescue-cc/internal/deepseek"
)

func TestDispatchParallel_PreservesOrder(t *testing.T) {
	dir := t.TempDir()
	for i, content := range []string{"alpha", "bravo", "charlie"} {
		path := filepath.Join(dir, string(rune('a'+i))+".txt")
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("seed file: %v", err)
		}
	}

	calls := []deepseek.ToolCall{
		{ID: "call_1", Type: "function", Function: deepseek.ToolCallFunc{
			Name: "read_file", Arguments: `{"path":"` + escapeJSON(filepath.Join(dir, "a.txt")) + `"}`,
		}},
		{ID: "call_2", Type: "function", Function: deepseek.ToolCallFunc{
			Name: "read_file", Arguments: `{"path":"` + escapeJSON(filepath.Join(dir, "b.txt")) + `"}`,
		}},
		{ID: "call_3", Type: "function", Function: deepseek.ToolCallFunc{
			Name: "read_file", Arguments: `{"path":"` + escapeJSON(filepath.Join(dir, "c.txt")) + `"}`,
		}},
	}

	out := dispatchParallel(calls, Options{})
	if len(out) != 3 {
		t.Fatalf("expected 3 results, got %d", len(out))
	}

	expected := []struct{ id, body string }{
		{"call_1", "alpha"},
		{"call_2", "bravo"},
		{"call_3", "charlie"},
	}
	for i, e := range expected {
		if out[i].ToolCallID != e.id {
			t.Errorf("result %d: expected ToolCallID %q, got %q", i, e.id, out[i].ToolCallID)
		}
		if !strings.Contains(out[i].Content, e.body) {
			t.Errorf("result %d: expected content to contain %q, got %q", i, e.body, out[i].Content)
		}
	}
}

func TestDispatchParallel_ActuallyConcurrent(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping concurrency timing test in short mode")
	}
	calls := []deepseek.ToolCall{
		{ID: "s1", Type: "function", Function: deepseek.ToolCallFunc{
			Name: "bash_exec", Arguments: `{"cmd":"sleep 1","timeout_sec":3.0}`,
		}},
		{ID: "s2", Type: "function", Function: deepseek.ToolCallFunc{
			Name: "bash_exec", Arguments: `{"cmd":"sleep 1","timeout_sec":3.0}`,
		}},
		{ID: "s3", Type: "function", Function: deepseek.ToolCallFunc{
			Name: "bash_exec", Arguments: `{"cmd":"sleep 1","timeout_sec":3.0}`,
		}},
	}

	start := time.Now()
	out := dispatchParallel(calls, Options{})
	elapsed := time.Since(start)

	if len(out) != 3 {
		t.Fatalf("expected 3 results, got %d", len(out))
	}
	if elapsed >= 2500*time.Millisecond {
		t.Errorf("expected parallel dispatch ~1s, got %v (sequential-like)", elapsed)
	}
}

func TestDispatchParallel_SingleCallNoGoroutine(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "only.txt")
	os.WriteFile(path, []byte("single"), 0644)

	calls := []deepseek.ToolCall{
		{ID: "solo", Type: "function", Function: deepseek.ToolCallFunc{
			Name: "read_file", Arguments: `{"path":"` + escapeJSON(path) + `"}`,
		}},
	}

	out := dispatchParallel(calls, Options{})
	if len(out) != 1 {
		t.Fatalf("expected 1 result, got %d", len(out))
	}
	if out[0].ToolCallID != "solo" {
		t.Errorf("expected ToolCallID 'solo', got %q", out[0].ToolCallID)
	}
	if !strings.Contains(out[0].Content, "single") {
		t.Errorf("expected content 'single', got %q", out[0].Content)
	}
}

func TestDispatchParallel_ToolErrorPropagatesAsContent(t *testing.T) {
	calls := []deepseek.ToolCall{
		{ID: "bad", Type: "function", Function: deepseek.ToolCallFunc{
			Name: "bash_exec", Arguments: `{"cmd":"sudo rm -rf /"}`,
		}},
	}
	out := dispatchParallel(calls, Options{})
	if len(out) != 1 {
		t.Fatalf("expected 1 result, got %d", len(out))
	}
	if !strings.Contains(out[0].Content, "ERROR:") || !strings.Contains(out[0].Content, "denied") {
		t.Errorf("expected ERROR/denied in content, got %q", out[0].Content)
	}
}

func escapeJSON(s string) string {
	return strings.ReplaceAll(s, `\`, `\\`)
}
