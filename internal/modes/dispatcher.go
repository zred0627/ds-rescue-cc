package modes

import (
	"fmt"
	"os"
	"strings"

	"github.com/zred0627/ds-rescue-cc/internal/deepseek"
	"github.com/zred0627/ds-rescue-cc/internal/git"
	"github.com/zred0627/ds-rescue-cc/internal/loop"
)

// LoadSection extracts the section matching mode from SKILL.md.
// Sections are denoted by H2 headers: "## Mode: Plan", "## Mode: Scheme", "## Mode: Execution"
func LoadSection(skillPath, mode string) (string, error) {
	data, err := os.ReadFile(skillPath)
	if err != nil {
		return "", fmt.Errorf("load SKILL.md %s: %w", skillPath, err)
	}
	text := string(data)
	header := "## Mode: " + capitalize(mode)
	idx := strings.Index(text, header)
	if idx < 0 {
		return "", fmt.Errorf("section %q not found in %s", header, skillPath)
	}
	after := text[idx:]
	nextHeaderIdx := strings.Index(after[len(header):], "\n## ")
	if nextHeaderIdx < 0 {
		return after, nil
	}
	return after[:len(header)+nextHeaderIdx], nil
}

// capitalize returns the string with the first rune uppercased (stdlib replacement for deprecated strings.Title)
func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

type RunInput struct {
	Mode      string
	SkillPath string
	InputFile string
	Model     string
	APIKey    string
	NoTools   bool
}

func (in RunInput) Run() (string, error) {
	section, err := LoadSection(in.SkillPath, in.Mode)
	if err != nil {
		return "", err
	}

	var userPrompt string
	if in.Mode == "execution" {
		diff, derr := git.CurrentDiff()
		if derr != nil {
			return "", fmt.Errorf("execution mode requires git repo: %w", derr)
		}
		userPrompt = "## Diff to review\n\n```diff\n" + diff + "\n```\n"
	} else {
		if in.InputFile == "" {
			return "", fmt.Errorf("%s mode requires input file (positional arg)", in.Mode)
		}
		data, ferr := os.ReadFile(in.InputFile)
		if ferr != nil {
			return "", fmt.Errorf("read input: %w", ferr)
		}
		userPrompt = string(data)
	}

	req := deepseek.ChatRequest{
		Model: deepseek.ResolveModel(in.Model),
		Messages: []deepseek.Message{
			{Role: "system", Content: section},
			{Role: "user", Content: userPrompt},
		},
		Temperature: 0.3,
	}
	if !in.NoTools {
		toolsSlice, _ := toolsAsTypedSlice()
		req.Tools = toolsSlice
	}

	c := &deepseek.Client{APIKey: in.APIKey}
	if in.NoTools {
		resp, err := c.Call(req)
		if err != nil {
			return "", err
		}
		return resp.Choices[0].Message.Content, nil
	}
	content, _, err := loop.Run(c, req)
	return content, err
}

func toolsAsTypedSlice() ([]deepseek.Tool, error) {
	raw := loop.DefaultTools()
	var out []deepseek.Tool
	for _, t := range raw {
		fn := t["function"].(map[string]interface{})
		out = append(out, deepseek.Tool{
			Type: "function",
			Function: deepseek.ToolFunction{
				Name:        fn["name"].(string),
				Description: fn["description"].(string),
				Parameters:  fn["parameters"].(map[string]interface{}),
			},
		})
	}
	return out, nil
}
