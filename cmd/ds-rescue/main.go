package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/zred0627/ds-rescue-cc/internal/config"
	"github.com/zred0627/ds-rescue-cc/internal/modes"
)

var Version = "0.0.1-dev"

func versionString() string {
	return fmt.Sprintf("ds-rescue v%s", Version)
}

func main() {
	var f config.Flags
	flag.StringVar(&f.Mode, "mode", "", "review mode: plan | scheme | execution (default auto-detect)")
	var planAlias, schemeAlias, execAlias bool
	flag.BoolVar(&planAlias, "plan", false, "alias for --mode plan")
	flag.BoolVar(&schemeAlias, "scheme", false, "alias for --mode scheme")
	flag.BoolVar(&execAlias, "execution", false, "alias for --mode execution")
	flag.StringVar(&f.SkillPath, "skill", "", "path to SKILL.md (env DS_RESCUE_SKILL_PATH alt)")
	flag.StringVar(&f.Model, "model", "pro", "deepseek model alias: pro | flash | reasoner | chat")
	flag.IntVar(&f.APITimeout, "api-timeout", 90, "API timeout in seconds")
	flag.IntVar(&f.ExecTimeout, "exec-timeout", 30, "per-tool timeout in seconds (max cap on bash_exec)")
	flag.IntVar(&f.MaxTokens, "max-tokens", 8000, "max output tokens")
	flag.BoolVar(&f.NoTools, "no-tools", false, "disable agentic tool-use loop (prompt-only mode)")
	flag.BoolVar(&f.Check, "check", false, "self-check (key found? skill loadable? model responds?)")
	flag.BoolVar(&f.Version, "version", false, "print version and exit")
	var verbose bool
	flag.BoolVar(&verbose, "verbose", false, "log every tool-call to stderr")
	flag.Parse()

	if f.Mode == "" {
		switch {
		case planAlias:
			f.Mode = "plan"
		case schemeAlias:
			f.Mode = "scheme"
		case execAlias:
			f.Mode = "execution"
		}
	}

	if f.Version {
		fmt.Println(versionString())
		os.Exit(0)
	}

	if f.SkillPath == "" {
		f.SkillPath = os.Getenv("DS_RESCUE_SKILL_PATH")
	}
	if f.SkillPath == "" {
		if root := os.Getenv("CLAUDE_PLUGIN_ROOT"); root != "" {
			f.SkillPath = root + "/skills/ds-plan-challenger/SKILL.md"
		}
	}

	apiKey, src, err := config.ResolveAPIKey()
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR:", err)
		os.Exit(2)
	}

	if f.Check {
		fmt.Println(versionString())
		fmt.Printf("  API key: found via %s\n", src)
		fmt.Printf("  Skill:   %s\n", f.SkillPath)
		if _, err := os.Stat(f.SkillPath); err != nil {
			fmt.Println("  WARNING: skill file not accessible")
			os.Exit(1)
		}
		fmt.Println("  Status: OK")
		os.Exit(0)
	}

	var inputFile string
	if flag.NArg() > 0 {
		inputFile = flag.Arg(0)
	}

	mode := f.Mode
	if mode == "" && inputFile != "" {
		data, _ := os.ReadFile(inputFile)
		mode, _ = modes.DetectMode(string(data), inputFile)
	}
	if mode == "" {
		mode = "plan"
	}

	out, err := modes.RunInput{
		Mode: mode, SkillPath: f.SkillPath, InputFile: inputFile,
		Model: f.Model, APIKey: apiKey, NoTools: f.NoTools,
		APITimeout: f.APITimeout, ExecTimeout: f.ExecTimeout, MaxTokens: f.MaxTokens,
	}.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR:", err)
		os.Exit(1)
	}
	fmt.Println(out)
}
