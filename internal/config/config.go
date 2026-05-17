package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// ResolveAPIKey returns (key, sourceLabel, error). Priority:
//  1. $DEEPSEEK_API_KEY env var
//  2. $XDG_CONFIG_HOME/ds-rescue/key (Linux/macOS) or %APPDATA%\ds-rescue\key (Windows)
//  3. $HOME/.ds-rescue/key fallback
func ResolveAPIKey() (string, string, error) {
	if v := strings.TrimSpace(os.Getenv("DEEPSEEK_API_KEY")); v != "" {
		return v, "env", nil
	}

	paths := candidateKeyPaths()
	for _, p := range paths {
		if data, err := os.ReadFile(p); err == nil {
			return strings.TrimSpace(string(data)), "file:" + p, nil
		}
	}

	return "", "", fmt.Errorf("no API key found; set $DEEPSEEK_API_KEY env var or write key to one of: %v", paths)
}

func candidateKeyPaths() []string {
	var paths []string
	if runtime.GOOS == "windows" {
		if appdata := os.Getenv("APPDATA"); appdata != "" {
			paths = append(paths, filepath.Join(appdata, "ds-rescue", "key"))
		}
	} else {
		cfgHome := os.Getenv("XDG_CONFIG_HOME")
		if cfgHome == "" {
			if home := homeDir(); home != "" {
				cfgHome = filepath.Join(home, ".config")
			}
		}
		if cfgHome != "" {
			paths = append(paths, filepath.Join(cfgHome, "ds-rescue", "key"))
		}
	}
	if home := homeDir(); home != "" {
		paths = append(paths, filepath.Join(home, ".ds-rescue", "key"))
	}
	return paths
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	h, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return h
}

const (
	ModePlan      = "plan"
	ModeScheme    = "scheme"
	ModeExecution = "execution"
)

type Flags struct {
	Mode        string
	SkillPath   string
	Model       string
	APITimeout  int
	ExecTimeout int
	MaxTokens   int
	NoTools     bool
	Check       bool
	Version     bool
	InputFile   string
}
