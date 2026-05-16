package modes

import "testing"

func TestDetectMode_SchemeComparison(t *testing.T) {
	content := "# Plan\n## Scheme Comparison\n| A | B | C |\n"
	m, _ := DetectMode(content, "any.md")
	if m != "scheme" {
		t.Errorf("expected scheme, got: %s", m)
	}
}

func TestDetectMode_PlanByPath(t *testing.T) {
	m, _ := DetectMode("any content", "docs/plans/2026-05-17-foo.md")
	if m != "plan" {
		t.Errorf("expected plan by path, got: %s", m)
	}
}

func TestDetectMode_PlanByTaskDAG(t *testing.T) {
	content := "## Tasks\n### Task 1: foo\nDepends on: none\n## Risk Register"
	m, _ := DetectMode(content, "anything.md")
	if m != "plan" {
		t.Errorf("expected plan by DAG content, got: %s", m)
	}
}

func TestDetectMode_DefaultPlanFallback(t *testing.T) {
	m, notice := DetectMode("hello world", "random.md")
	if m != "plan" {
		t.Errorf("expected default plan, got: %s", m)
	}
	if notice == "" {
		t.Errorf("expected default notice, got empty string")
	}
}
