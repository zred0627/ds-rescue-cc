package loop

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"time"
)

const MaxToolResultChars = 40000

func ExecReadFile(args map[string]interface{}) (string, error) {
	p, ok := args["path"].(string)
	if !ok || p == "" {
		return "", fmt.Errorf("read_file: missing 'path' argument")
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return "", fmt.Errorf("read_file %s: %w", p, err)
	}
	return TruncateOutput(string(data), MaxToolResultChars), nil
}

func ExecWriteFile(args map[string]interface{}) (string, error) {
	p, _ := args["path"].(string)
	content, _ := args["content"].(string)
	if p == "" {
		return "", fmt.Errorf("write_file: missing 'path'")
	}
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("write_file %s: %w", p, err)
	}
	return fmt.Sprintf("wrote %d bytes to %s", len(content), p), nil
}

// MaxCmdLength enforces a hard ceiling on bash_exec command-string length (Risk Register T5)
const MaxCmdLength = 4096

// denyPatterns: case-insensitive regex patterns blocked from bash_exec (zero-trust on tool_call source)
var denyPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\bsudo\b`),
	regexp.MustCompile(`(?i)\brm\s+-rf\s+/(\s|$)`),
	regexp.MustCompile(`(?i):\s*\(\s*\)\s*\{.*\|\s*:\s*&\s*\}\s*;\s*:`), // fork bomb
	regexp.MustCompile(`(?i)\bcurl\b.*(\.deepseek-key|\.ds-rescue/key|/etc/shadow|/etc/passwd)`),
	regexp.MustCompile(`(?i)\bnc\b.*\b-l\b`),
	regexp.MustCompile(`(?i)>\s*/dev/(sd|nvme|disk)`),
	regexp.MustCompile(`(?i)\bmkfs\b|\bdd\s+if=`),
}

func ExecBash(args map[string]interface{}) (string, error) {
	cmd, _ := args["cmd"].(string)
	if cmd == "" {
		return "", fmt.Errorf("bash_exec: missing 'cmd'")
	}
	if len(cmd) > MaxCmdLength {
		return "", fmt.Errorf("bash_exec: command length %d exceeds max %d", len(cmd), MaxCmdLength)
	}
	for _, pat := range denyPatterns {
		if pat.MatchString(cmd) {
			return "", fmt.Errorf("bash_exec: command denied by safety policy (pattern: %s)", pat.String())
		}
	}
	timeout := 30.0
	if t, ok := args["timeout_sec"].(float64); ok && t > 0 {
		timeout = t
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout*float64(time.Second)))
	defer cancel()

	var c *exec.Cmd
	if runtime.GOOS == "windows" {
		c = exec.CommandContext(ctx, "cmd.exe", "/C", cmd)
	} else {
		c = exec.CommandContext(ctx, "sh", "-c", cmd)
	}
	out, err := c.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return "", fmt.Errorf("bash_exec timeout after %.1fs (killed)", timeout)
	}
	if err != nil {
		return TruncateOutput(string(out), MaxToolResultChars), fmt.Errorf("bash_exec exit error: %w", err)
	}
	return TruncateOutput(string(out), MaxToolResultChars), nil
}

func TruncateOutput(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "\n\n[truncated " + fmt.Sprintf("%d", len(s)-max) + " bytes; output exceeds " + fmt.Sprintf("%d", max) + "-char limit]"
}

func DefaultTools() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"type": "function",
			"function": map[string]interface{}{
				"name":        "read_file",
				"description": "Read a UTF-8 text file from local disk. Returns truncated content (<=40000 chars).",
				"parameters": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"path": map[string]interface{}{"type": "string", "description": "Absolute or repo-relative file path"},
					},
					"required": []string{"path"},
				},
			},
		},
		{
			"type": "function",
			"function": map[string]interface{}{
				"name":        "write_file",
				"description": "Write content to a file, creating it if missing. Overwrites existing content.",
				"parameters": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"path":    map[string]interface{}{"type": "string"},
						"content": map[string]interface{}{"type": "string"},
					},
					"required": []string{"path", "content"},
				},
			},
		},
		{
			"type": "function",
			"function": map[string]interface{}{
				"name":        "bash_exec",
				"description": "Execute a shell command (sh on Unix, cmd.exe on Windows) with timeout. Returns combined stdout+stderr, truncated.",
				"parameters": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"cmd":         map[string]interface{}{"type": "string"},
						"timeout_sec": map[string]interface{}{"type": "number", "description": "Default 30s, max 300s"},
					},
					"required": []string{"cmd"},
				},
			},
		},
	}
}

func DispatchTool(name string, args map[string]interface{}) (string, error) {
	switch name {
	case "read_file":
		return ExecReadFile(args)
	case "write_file":
		return ExecWriteFile(args)
	case "bash_exec":
		return ExecBash(args)
	default:
		return "", fmt.Errorf("unknown tool: %s", name)
	}
}
