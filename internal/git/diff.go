package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// CurrentDiff returns the diff of HEAD~1..HEAD (last commit) by default.
// Falls back to staged-vs-HEAD diff if HEAD is the initial commit.
func CurrentDiff() (string, error) {
	out, err := exec.Command("git", "diff", "HEAD~1", "HEAD").CombinedOutput()
	if err == nil {
		return string(out), nil
	}
	if strings.Contains(string(out), "unknown revision") {
		out, err = exec.Command("git", "diff", "--cached").CombinedOutput()
		if err == nil && len(strings.TrimSpace(string(out))) > 0 {
			return string(out), nil
		}
	}
	return "", fmt.Errorf("git diff failed: %s", out)
}
