package modes

import (
	"regexp"
	"strings"
)

var schemeRe = regexp.MustCompile(`(?i)scheme comparison|方案对比|^\| .* \| .* \| .* \|`)
var planContentRe = regexp.MustCompile(`(?im)task dag|risk register|^### task \d+:|^depends on:`)

// DetectMode returns the detected mode and an optional notice string.
// Priority: path-based plan detection → scheme content signals → plan content signals → default plan.
func DetectMode(content, path string) (string, string) {
	if strings.Contains(strings.ToLower(path), "docs/plans/") {
		return "plan", ""
	}
	if schemeRe.MatchString(content) {
		return "scheme", ""
	}
	if planContentRe.MatchString(content) {
		return "plan", ""
	}
	return "plan", "Used default plan mode; pass --scheme or --execution if intended otherwise"
}
