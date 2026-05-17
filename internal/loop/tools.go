package loop

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

const (
	MaxToolResultChars = 40000
	MaxCmdLength       = 4096
	DefaultExecTimeout = 30.0
)

var denyPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\bsudo\b`),
	regexp.MustCompile(`(?i)\brm\s+-rf\s+/(\s|$)`),
	regexp.MustCompile(`(?i):\s*\(\s*\)\s*\{.*\|\s*:\s*&\s*\}\s*;\s*:`),
	regexp.MustCompile(`(?i)\bcurl\b.*(\.deepseek-key|\.ds-rescue/key|/etc/shadow|/etc/passwd)`),
	regexp.MustCompile(`(?i)\bnc\b.*\b-l\b`),
	regexp.MustCompile(`(?i)>\s*/dev/(sd|nvme|disk)`),
	regexp.MustCompile(`(?i)\bmkfs\b|\bdd\s+if=`),
}

var denyPathPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\.deepseek-key$`),
	regexp.MustCompile(`(?i)/\.ds-rescue/(.+/)?key$`),
	regexp.MustCompile(`(?i)/\.config/ds-rescue/(.+/)?key$`),
	regexp.MustCompile(`(?i)/ds-rescue/key$`),
	regexp.MustCompile(`(?i)/\.env(\..+)?$`),
	regexp.MustCompile(`(?i)^\.env(\..+)?$`),
	regexp.MustCompile(`(?i)/\.aws/credentials$`),
	regexp.MustCompile(`(?i)/\.ssh/id_`),
	regexp.MustCompile(`(?i)/secrets?/`),
}

var secretEnvPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)_API_KEY$`),
	regexp.MustCompile(`(?i)_TOKEN$`),
	regexp.MustCompile(`(?i)_SECRET$`),
	regexp.MustCompile(`(?i)_PASSWORD$`),
	regexp.MustCompile(`(?i)_CREDENTIAL`),
	regexp.MustCompile(`(?i)^DEEPSEEK_`),
	regexp.MustCompile(`(?i)^DS_RESCUE_`),
	regexp.MustCompile(`(?i)^OPENAI_`),
	regexp.MustCompile(`(?i)^ANTHROPIC_`),
	regexp.MustCompile(`(?i)^AWS_(ACCESS|SECRET)_`),
	regexp.MustCompile(`(?i)^GH_TOKEN$`),
	regexp.MustCompile(`(?i)^GITHUB_TOKEN$`),
}

func isDeniedPath(p string) bool {
	norm := filepath.ToSlash(p)
	for _, r := range denyPathPatterns {
		if r.MatchString(norm) {
			return true
		}
	}
	return false
}

func sanitizedEnv() []string {
	parent := os.Environ()
	out := make([]string, 0, len(parent))
	for _, e := range parent {
		name := strings.SplitN(e, "=", 2)[0]
		blocked := false
		for _, p := range secretEnvPatterns {
			if p.MatchString(name) {
				blocked = true
				break
			}
		}
		if !blocked {
			out = append(out, e)
		}
	}
	return out
}

func ExecReadFile(args map[string]interface{}) (string, error) {
	p, ok := args["path"].(string)
	if !ok || p == "" {
		return "", fmt.Errorf("read_file: missing 'path' argument")
	}
	if isDeniedPath(p) {
		return "", fmt.Errorf("read_file: path denied by safety policy: %s", p)
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
	if isDeniedPath(p) {
		return "", fmt.Errorf("write_file: path denied by safety policy: %s", p)
	}
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("write_file %s: %w", p, err)
	}
	return fmt.Sprintf("wrote %d bytes to %s", len(content), p), nil
}

func ExecBash(args map[string]interface{}, maxTimeoutSec int) (string, error) {
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
	timeout := DefaultExecTimeout
	if t, ok := args["timeout_sec"].(float64); ok && t > 0 {
		timeout = t
	}
	if maxTimeoutSec > 0 && timeout > float64(maxTimeoutSec) {
		timeout = float64(maxTimeoutSec)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout*float64(time.Second)))
	defer cancel()

	var c *exec.Cmd
	if runtime.GOOS == "windows" {
		c = exec.CommandContext(ctx, "cmd.exe", "/C", cmd)
	} else {
		c = exec.CommandContext(ctx, "sh", "-c", cmd)
	}
	c.Env = sanitizedEnv()
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
	return fmt.Sprintf("%s\n\n[truncated %d bytes; output exceeds %d-char limit]", s[:max], len(s)-max, max)
}

func DefaultTools() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"type": "function",
			"function": map[string]interface{}{
				"name":        "read_file",
				"description": "Read a UTF-8 text file from local disk. Returns truncated content (<=40000 chars). Secret-file paths are denied.",
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
				"description": "Write content to a file, creating it if missing. Overwrites existing content. Secret-file paths are denied.",
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
				"description": "Execute a shell command (sh on Unix, cmd.exe on Windows). Secret environment variables are stripped from the child process. Returns combined stdout+stderr, truncated.",
				"parameters": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"cmd":         map[string]interface{}{"type": "string"},
						"timeout_sec": map[string]interface{}{"type": "number", "description": "Per-call timeout; capped by --exec-timeout CLI flag"},
					},
					"required": []string{"cmd"},
				},
			},
		},
	}
}

func DispatchTool(name string, args map[string]interface{}, maxExecTimeoutSec int) (string, error) {
	switch name {
	case "read_file":
		return ExecReadFile(args)
	case "write_file":
		return ExecWriteFile(args)
	case "bash_exec":
		return ExecBash(args, maxExecTimeoutSec)
	default:
		return "", fmt.Errorf("unknown tool: %s", name)
	}
}
