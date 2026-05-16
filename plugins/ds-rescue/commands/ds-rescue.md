---
description: "Cross-family DeepSeek adversarial review of plans, schemes, or commits"
argument-hint: "[--plan|--scheme|--execution] [--model pro|flash] [--check] [input-file]"
allowed-tools: Bash, Read
---

Invoke the ds-rescue Go binary to get an independent second-opinion review from DeepSeek.

## Argument parsing

The command accepts these user-facing flags and forwards them to the binary. Note: the binary's CLI parser registers `--plan`, `--scheme`, `--execution` as flag aliases (see Data Model > CLI Contract), so `$ARGUMENTS` is passed verbatim — no translation needed.

- `--check` → run `ds-rescue --check` for self-diagnosis
- `--version` → run `ds-rescue --version`
- `--plan <file>` → binary alias for `--mode plan`
- `--scheme <file>` → binary alias for `--mode scheme`
- `--execution` → binary alias for `--mode execution` (binary self-runs `git diff HEAD~1 HEAD`)
- `<file>` (no flag) → let binary auto-detect mode from content

## Execution

```bash
BINARY="${CLAUDE_PLUGIN_ROOT}/bin/ds-rescue"
if [ ! -x "$BINARY" ]; then
  BINARY="$(command -v ds-rescue 2>/dev/null)"
fi
if [ -z "$BINARY" ]; then
  echo "ds-rescue binary not found. Run /ds-rescue:install or:"
  echo "  curl -L https://github.com/zred0627/ds-rescue-cc/releases/latest/download/ds-rescue-\$(uname -s | tr A-Z a-z)-\$(uname -m) -o /usr/local/bin/ds-rescue && chmod +x /usr/local/bin/ds-rescue"
  exit 1
fi

export DS_RESCUE_SKILL_PATH="${CLAUDE_PLUGIN_ROOT}/skills/ds-plan-challenger/SKILL.md"
"$BINARY" $ARGUMENTS
```

## Examples

```
/ds-rescue docs/plans/2026-05-17-foo.md
  → auto-detects plan mode (path contains docs/plans/)

/ds-rescue --scheme scheme-A-vs-B.md
  → forces scheme mode

/ds-rescue --execution
  → binary self-runs git diff HEAD~1 HEAD and reviews the change

/ds-rescue --check
  → self-diagnostic
```

## Notes

- Default model is `pro` (deepseek-chat v4-pro). Use `--model flash` for faster/cheaper but less thorough review.
- For execution mode, you must be inside a git repo with at least one previous commit on HEAD.
- Findings are raw DeepSeek output — don't expect Claude-style politeness; that's the point.
