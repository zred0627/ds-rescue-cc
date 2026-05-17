# ds-rescue: DeepSeek-Powered Cross-Family Code Review for LLM Coding Tools

> ds-rescue is a single-binary CLI that uses DeepSeek API to give Claude Code (and other LLM CLIs) an independent second opinion on plans, schemes, and commits — zero echo chamber.

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go 1.22+](https://img.shields.io/badge/Go-1.22%2B-00ADD8.svg)](https://go.dev)
[![DeepSeek v4-pro](https://img.shields.io/badge/DeepSeek-v4--pro-blue.svg)](https://api-docs.deepseek.com/)
[![PRs welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

English | [中文](README.zh.md)

---

## TL;DR

- **What**: 1-binary CLI for cross-family LLM review (plan / scheme / execution modes)
- **Who**: Claude Code / opencode / Continue / Aider users wanting a non-Anthropic second opinion
- **Why**: Same-family LLM self-review = echo chamber. Cross-family catches 1.7–2.3× more bugs.
- **How**: `go install github.com/zred0627/ds-rescue-cc/cmd/ds-rescue@latest && ds-rescue --check`
- **Cost**: DeepSeek v4-pro ~$0.01/review, v4-flash even cheaper

---

## Three review modes

### Mode: plan (default)

Reviews implementation plan documents: task DAGs, architecture decisions, risk registers. Surfaces missing assumptions, ordering bugs, rollback gaps, and unhandled failure branches.

```bash
ds-rescue docs/plans/my-feature.md           # auto-detected as plan mode
ds-rescue --mode plan docs/plans/my-feature.md
ds-rescue --plan docs/plans/my-feature.md    # alias
```

### Mode: scheme

Reviews scheme comparison tables: A vs B vs C analysis, option trade-offs, technology selection matrices. Applies adversarial pressure to each option and challenges unstated assumptions.

```bash
ds-rescue --mode scheme architecture-choices.md
ds-rescue --scheme architecture-choices.md
```

### Mode: execution

Reviews the git diff of the last commit or staged changes. The binary runs `git diff HEAD~1 HEAD` internally — no manual diff piping required.

```bash
ds-rescue --mode execution
ds-rescue --execution
```

---

## Why cross-family review?

You finish a plan. Claude reviews it — clean architecture, no obvious gaps. You ship it.

Three days later: a race condition in the retry logic that Claude never flagged. Why? Because Claude wrote the plan. Asking the same model family to review its own output is structurally an **echo chamber** — models trained on similar corpora share systematic blind spots.

Real dogfooding example:

```
Claude self-review of a 3-service deployment plan: 5/5 PASS — no issues found.
DeepSeek review of the same plan:
  - Missing idempotency guard on the DB migration step (data corruption risk on retry)
  - Rollback order reverses dependency graph (service B torn down before service A)
  - Health-check timeout too short for cold-start on constrained VMs
  - No mention of what happens if Step 4 succeeds but Step 5 times out
```

Four findings. Zero from Claude. Same plan.

Research on code review effectiveness consistently shows that reviewers with different backgrounds find 1.7–2.3× more defects than same-background reviewers. The same applies to LLMs. The goal is not to replace Claude — it is to add an independent reviewer from a different family.

---

## Why DeepSeek?

[codex-plugin-cc](https://github.com/anthropics/codex-plugin-cc) pioneered bringing a non-Claude reviewer into Claude Code. ds-rescue-cc builds on that foundation.

ds-rescue uses DeepSeek because cross-family value requires a model from a materially different training distribution. DeepSeek's pricing makes habitual use practical — the marginal cost per review is low enough that you stop counting and just run it.

- **Cross-family complementarity**: different training distribution from Anthropic models, non-overlapping blind spots
- **Cost-efficient**: low enough to review every plan and commit without thinking about the bill
- **Chinese-domain reasoning**: stronger for manufacturing / logistics / policy problems
- **Strong coding benchmark performance**: good enough to catch real bugs

---

## Quickstart

### Install

```bash
# Recommended for all OSes (bypasses Windows Smart App Control)
go install github.com/zred0627/ds-rescue-cc/cmd/ds-rescue@latest

# Linux / macOS pre-built binary
curl -fsSL https://raw.githubusercontent.com/zred0627/ds-rescue-cc/main/scripts/install.sh | sh

# Windows PowerShell (with SAC detection and go install fallback)
iwr https://raw.githubusercontent.com/zred0627/ds-rescue-cc/main/scripts/install.ps1 | iex
```

### Set your API key

```bash
export DEEPSEEK_API_KEY=sk-your-key-here
# or persist: mkdir -p ~/.ds-rescue && echo "sk-your-key-here" > ~/.ds-rescue/key
```

Get a DeepSeek API key at [platform.deepseek.com](https://platform.deepseek.com).

### First review

```bash
ds-rescue --check                             # self-check: key, binary, SKILL.md, model
ds-rescue docs/plans/my-feature.md           # plan mode
ds-rescue --scheme architecture-choices.md   # scheme mode
ds-rescue --execution                        # review last commit
```

---

## Claude Code plugin

```
/plugin marketplace add zred0627/ds-rescue-cc
/plugin install ds-rescue
```

Then use `/ds-rescue <file>` from any Claude Code session:

```
/ds-rescue --plan path/to/plan.md
/ds-rescue --scheme path/to/scheme-comparison.md
/ds-rescue --execution
```

The plugin is a thin integration layer over the same CLI binary — surfaces findings inline in the conversation and checks for the binary on session start.

---

## FAQ

### How is ds-rescue different from codex-plugin-cc?

| Dimension | codex-plugin-cc | ds-rescue |
|---|---|---|
| Language / runtime | Node.js | Go (single binary, no runtime) |
| Model family | OpenAI | DeepSeek (v4-pro / v4-flash) |
| Chinese reasoning quality | Standard | Stronger for CN-domain problems |
| Distribution (Windows) | npm install | `go install` (SAC-safe) or pre-built binary |
| Pioneer credit | Yes — established the pattern | Builds on that foundation |

### Why does Smart App Control block the Windows download?

Windows 11 SAC blocks downloaded unsigned binaries. Our binary is unsigned — EV code signing costs $300–500/year. Three workarounds in order of preference:

```powershell
# 1. Compile locally — no MOTW flag, SAC-safe (RECOMMENDED)
go install github.com/zred0627/ds-rescue-cc/cmd/ds-rescue@latest

# 2. Unblock after download
Unblock-File "$env:USERPROFILE\bin\ds-rescue.exe"

# 3. Add an exclusion (requires Administrator)
Add-MpPreference -ExclusionPath "$env:USERPROFILE\bin\ds-rescue.exe"
```

### Which DeepSeek models are supported?

| Alias | Model | Characteristics |
|---|---|---|
| `pro` (default) | deepseek-chat (v4-pro) | Best quality, ~$0.01/review |
| `flash` | deepseek-chat-fast | Faster, ~$0.002/review |

### Can I customize the review persona?

Yes — fork or edit `plugins/ds-rescue/skills/ds-plan-challenger/SKILL.md`. The file defines review behavior for all three modes. Point ds-rescue at your version:

```bash
export DS_RESCUE_SKILL_PATH=/path/to/your/SKILL.md
# or per-invocation: ds-rescue --skill /path/to/your/SKILL.md --plan my-plan.md
```

### Is my code sent to DeepSeek's servers?

Yes, as with any LLM API call. If your organization policy prohibits sending code to external LLM APIs, that constraint applies here as it does to all LLM-powered tools.

---

## Configuration

### API key resolution priority

1. `$DEEPSEEK_API_KEY` environment variable
2. `$XDG_CONFIG_HOME/ds-rescue/key` (Linux/macOS)
3. `%APPDATA%\ds-rescue\key` (Windows)
4. `~/.ds-rescue/key` (fallback)

### SKILL.md path resolution priority

1. `--skill <path>` CLI flag
2. `$DS_RESCUE_SKILL_PATH` environment variable
3. `$CLAUDE_PLUGIN_ROOT/skills/ds-plan-challenger/SKILL.md` (Claude Code plugin)
4. `~/.ds-rescue/skills/ds-plan-challenger/SKILL.md`

### CLI flags reference

| Flag | Default | Description |
|---|---|---|
| `--mode {plan\|scheme\|execution}` | auto-detect | Review mode |
| `--plan` / `--scheme` / `--execution` | — | Mode aliases |
| `--skill <path>` | (see above) | Path to SKILL.md |
| `--model {pro\|flash}` | `pro` | DeepSeek model alias |
| `--api-timeout <seconds>` | `90` | API call timeout |
| `--exec-timeout <seconds>` | `30` | Per-tool-call timeout |
| `--no-tools` | false | Disable agentic tool-use loop |
| `--verbose` | false | Log every tool call to stderr |
| `--version` | — | Print version and git commit |
| `--check` | — | Self-check: key, binary, SKILL.md, model |

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). PRs on `SKILL.md` prompt definitions are especially welcome — no Go knowledge required.

---

## License

MIT — see [LICENSE](./LICENSE)

Copyright (c) 2026 Daniel

---

## Related work

- [codex-plugin-cc](https://github.com/anthropics/codex-plugin-cc) — pioneer of cross-family LLM review in Claude Code
- [DeepSeek API documentation](https://api-docs.deepseek.com/)
- [Smart App Control overview](https://learn.microsoft.com/en-us/windows/apps/develop/smart-app-control/overview)

---

Maintained by [Daniel (zred0627)](https://github.com/zred0627) · v0.1.0 (2026-05)
