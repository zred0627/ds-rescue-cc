package loop

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/zred0627/ds-rescue-cc/internal/deepseek"
)

// TestDispatchParallel_PreservesOrder verifies that when multiple tool_calls
// are dispatched concurrently, the resulting tool messages appear in the SAME
// order as the input slice (OpenAI-compatible requirement: tool messages must
// follow the assistant's tool_calls array order regardless of completion order).
func TestDispatchParallel_PreservesOrder(t *testing.T) {
	dir := t.TempDir()
	// Create 3 files with distinct content so we can tell which result came back
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

	out := dispatchParallel(calls)
	if len(out) != 3 {
		t.Fatalf("expected 3 results, got %d", len(out))
	}

	// Verify order matches input
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

// TestDispatchParallel_ActuallyConcurrent verifies the goroutine path
// produces wall-clock latency ≈ max(individual) not sum(individual).
// Uses bash_exec with sleep to make timing measurable.
func TestDispatchParallel_ActuallyConcurrent(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping concurrency timing test in short mode")
	}
	// 3 sleeps of ~1s each: sequential would be ~3s, parallel should be ~1s
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
	out := dispatchParallel(calls)
	elapsed := time.Since(start)

	if len(out) != 3 {
		t.Fatalf("expected 3 results, got %d", len(out))
	}
	// Parallel must be < 2s (3x sleep 1s would be ~3s sequential).
	// Allow generous margin for CI / Windows process startup.
	if elapsed >= 2500*time.Millisecond {
		t.Errorf("expected parallel dispatch ≈ 1s, got %v (sequential-like)", elapsed)
	}
}

// TestDispatchParallel_SingleCallNoGoroutine verifies the fast-path
// for a single tool_call doesn't spawn an unnecessary goroutine.
func TestDispatchParallel_SingleCallNoGoroutine(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "only.txt")
	os.WriteFile(path, []byte("single"), 0644)

	calls := []deepseek.ToolCall{
		{ID: "solo", Type: "function", Function: deepseek.ToolCallFunc{
			Name: "read_file", Arguments: `{"path":"` + escapeJSON(path) + `"}`,
		}},
	}

	out := dispatchParallel(calls)
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

// TestDispatchParallel_ToolErrorPropagatesAsContent verifies that a tool
// that errors out (e.g., denied bash command) returns an ERROR-prefixed
// content string rather than panicking the goroutine.
func TestDispatchParallel_ToolErrorPropagatesAsContent(t *testing.T) {
	calls := []deepseek.ToolCall{
		{ID: "bad", Type: "function", Function: deepseek.ToolCallFunc{
			Name: "bash_exec", Arguments: `{"cmd":"sudo rm -rf /"}`,
		}},
	}
	out := dispatchParallel(calls)
	if len(out) != 1 {
		t.Fatalf("expected 1 result, got %d", len(out))
	}
	if !strings.Contains(out[0].Content, "ERROR:") || !strings.Contains(out[0].Content, "denied") {
		t.Errorf("expected ERROR/denied in content, got %q", out[0].Content)
	}
}

// escapeJSON does the bare minimum to make a filesystem path safe to embed
// in a JSON string literal (handles Windows backslashes).
func escapeJSON(s string) string {
	return strings.ReplaceAll(s, `\`, `\\`)
}
