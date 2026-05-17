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

## What is ds-rescue?

ds-rescue is a standalone Go binary with three review modes — plan, scheme, and execution — each backed by the DeepSeek v4-pro model family. It is not a chat assistant; it does not generate code or answer questions. It is a single-purpose adversarial reviewer: give it a document or a diff, and it returns structured findings from a model that had no part in producing the artifact under review.

The tool is designed to be habit-forming. At ~$0.01 per review call, there is no meaningful cost barrier to running it on every plan, every architectural decision, every commit. The review behavior is defined in a plain Markdown file (`SKILL.md`) that you can read, fork, and customize without touching any Go code.

ds-rescue works as a standalone CLI on Linux, macOS, and Windows, and optionally as a Claude Code plugin that surfaces findings inline in your coding session.

---

## When should I use ds-rescue?

### Use ds-rescue when

- You finish a plan and want a second opinion before starting implementation
- You are choosing between two or more architectural approaches and want adversarial pressure on each option
- You finish a unit of work and want a pre-push review of your diff
- You are using Claude Code and want to add a non-Anthropic reviewer to the workflow
- You work in Chinese or on problems where Chinese-language technical depth matters (logistics, manufacturing, policy analysis)
- You want a reviewer with different training priors than the model that wrote the artifact

### Do not use ds-rescue when

- You need a chat assistant or code generator — use Claude / GPT / Gemini directly
- Your organization policy prohibits sending code to external LLM APIs (this is a universal LLM constraint, not specific to ds-rescue)
- You want local/offline review — ds-rescue always calls the DeepSeek API
- You want the review to be from the same model family as the author (by design, it never is)

---

## Three review modes

### Mode: plan (default)

Reviews implementation plan documents: task DAGs, architecture decisions, risk registers, sprint plans. Surfaces missing assumptions, ordering bugs, rollback gaps, and unhandled failure branches.

```bash
ds-rescue docs/plans/my-feature.md          # auto-detected as plan mode
ds-rescue --mode plan docs/plans/my-feature.md
ds-rescue --plan docs/plans/my-feature.md   # alias
```

Use before starting a feature or sprint.

### Mode: scheme

Reviews scheme comparison tables: A vs B vs C analysis, option trade-offs, technology selection matrices. Applies adversarial pressure to each option, flags missing criteria, and challenges unstated assumptions.

```bash
ds-rescue --mode scheme architecture-choices.md
ds-rescue --scheme architecture-choices.md  # alias
```

Use when choosing between architectural alternatives.

### Mode: execution

Reviews the git diff of the last commit or staged changes. The binary runs `git diff HEAD~1 HEAD` internally — no manual diff piping required.

```bash
ds-rescue --mode execution
ds-rescue --execution                        # alias
```

Use after finishing a unit of work, before push.

---

## Why cross-family review?

You finish a plan. You run it through Claude for review. Claude says it looks great — clean architecture, sensible approach, no obvious gaps. You ship it.

Three days later you discover a race condition in the retry logic that Claude never flagged. Why? Because Claude wrote the plan in the first place. Asking the same model family to review its own output is, structurally, an **echo chamber**.

This is not a criticism of Claude. It is a property of how LLMs work: models trained on similar data corpora share systematic blind spots. A model that confidently generates a plan with a subtle ordering bug will, with similarly high confidence, fail to notice that bug on re-read. The failure mode is correlated, not independent.

Concrete example from real dogfooding:

```
Claude self-review of a 3-service deployment plan: 5/5 PASS — no issues found.
DeepSeek review of the same plan:
  - Missing idempotency guard on the DB migration step (data corruption risk on retry)
  - Rollback order reverses dependency graph (service B torn down before service A)
  - Health-check timeout too short for cold-start on constrained VMs
  - No mention of what happens if Step 4 succeeds but Step 5 times out
```

Four findings. Zero from Claude. Same plan.

### The engineering value

Software engineering has a well-studied analog: code review effectiveness scales with reviewer diversity. Studies on pair review and inspection programs consistently show that 1.7–2.3× more defects are found when reviewers have different backgrounds and prior exposure than when they share the same mental model.

The same principle applies to LLM review. A model from a different training family has different biases about what "normal" looks like, different priors about which edge cases are worth mentioning, and different failure modes that are less likely to overlap with the author's. DeepSeek v4-pro is trained on a materially different data distribution than Claude. That complementarity is the engineering value being unlocked here.

The goal is not to replace Claude. It is to add an independent second reviewer from a different family — the same way you would not let the author of a PR be its only reviewer.

---

## Why DeepSeek?

[codex-plugin-cc](https://github.com/anthropics/codex-plugin-cc) pioneered the concept of bringing a non-Claude reviewer into the Claude Code workflow. It deserves full credit for establishing that pattern. ds-rescue-cc exists because of that work, not in spite of it.

ds-rescue uses DeepSeek because the cross-family value requires a model from a materially different training distribution than the one that produced the artifact under review. DeepSeek's pricing makes habitual use practical — the marginal cost of one more review is low enough that you stop counting and just run it.

### Why DeepSeek specifically

- **Cross-family complementarity**: trained on a different data distribution than Anthropic models, so its blind spots do not overlap with Claude's
- **Cost-efficient**: low enough per-review cost that running it on every plan and commit is a habit, not a budget decision
- **Chinese-domain reasoning depth**: for teams working in Chinese or on manufacturing / logistics / policy problems, DeepSeek's Chinese-language reasoning quality is noticeably stronger than most Western-origin models
- **Strong coding benchmark performance**: good enough to catch real bugs, as the example above demonstrates

---

## Quickstart (60 seconds)

### Recommended for all OSes: install via Go

```bash
go install github.com/zred0627/ds-rescue-cc/cmd/ds-rescue@latest
```

Why this is the universal first choice:

- Works identically on Linux, macOS, and Windows
- **Bypasses Windows Smart App Control** — your binary, compiled locally, no Mark-of-the-Web flag
- Always builds from your pinned Go toolchain
- Requires Go 1.22+ (install from [go.dev/dl](https://go.dev/dl/) if needed)

### Linux / macOS: curl-pipe pre-built binary

```bash
curl -fsSL https://raw.githubusercontent.com/zred0627/ds-rescue-cc/main/scripts/install.sh | sh
```

### Windows: PowerShell installer (with SAC detection)

```powershell
iwr https://raw.githubusercontent.com/zred0627/ds-rescue-cc/main/scripts/install.ps1 | iex
```

> **Note**: if Windows Smart App Control is ON, the installer detects it and falls back to `go install` instructions automatically. See the [Smart App Control section](#why-smart-app-control-blocks-our-binary-technical-deep-dive) for details.

### Set your API key

```bash
# Option A: environment variable (recommended for CI)
export DEEPSEEK_API_KEY=sk-your-key-here

# Option B: key file (survives shell restarts)
mkdir -p ~/.ds-rescue && echo "sk-your-key-here" > ~/.ds-rescue/key && chmod 600 ~/.ds-rescue/key
```

Get a DeepSeek API key at [platform.deepseek.com](https://platform.deepseek.com).

### First review

```bash
ds-rescue --check                              # self-diagnose: key, binary, SKILL.md, model
ds-rescue docs/plans/my-feature.md            # plan mode (auto-detected by file content)
ds-rescue --scheme architecture-choices.md    # scheme mode
ds-rescue --execution                         # review last commit
```

---

## Claude Code plugin install

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

The plugin registers a `SessionStart` hook that checks for the ds-rescue binary and recommends `go install` if it is missing. It also sets `$CLAUDE_PLUGIN_ROOT` so the binary finds `SKILL.md` automatically, and surfaces findings inline in the conversation.

The underlying binary is identical to the standalone CLI — the plugin is a thin integration layer, not a separate product.

---

## How it works

```
  Your plan / scheme / git diff
          │
          ▼
    ds-rescue CLI
    ─────────────────────────────────────────────────────
    │  1. Loads SKILL.md (defines review persona + schema)
    │  2. Composes prompt for selected mode
    │  3. Calls DeepSeek API (v4-pro by default)
    │  4. Optional agentic tool-use loop (read_file / write_file / bash_exec)
    │     up to 20 rounds, deny-list protected
    │  5. Returns structured findings to stdout
    ─────────────────────────────────────────────────────
          │
          ▼
    Adversarial findings (stdout)
    Diagnostics (stderr)
```

The three-mode review behavior is fully defined in `plugins/ds-rescue/skills/ds-plan-challenger/SKILL.md`. That file is loaded from disk at runtime — it is not embedded in the binary. You can customize it without touching Go code.

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

Both tools belong in the ecosystem. They are complements, not competitors.

### Does ds-rescue support OpenAI / Claude / Gemini?

No, by design. The cross-family value comes specifically from using a different model family than the one that wrote the artifact. If you want a DeepSeek review of a Claude-authored plan, ds-rescue delivers that. If you want a GPT-5 review, use codex-plugin-cc.

### Why does Smart App Control block the Windows download?

Windows 11 (25H2 and later) ships with Smart App Control (SAC) turned ON by default. SAC blocks execution of downloaded binaries that carry the Mark of the Web (MOTW) zone flag — which any file downloaded from the internet via a browser or `Invoke-WebRequest` receives automatically.

Our pre-built binary is unsigned. Extended Validation (EV) code signing costs $300–500/year and requires organizational verification. That is not feasible for a free open-source project maintained by a single developer.

Three workarounds, in order of preference:

1. **Recommended — compile locally** (no MOTW, SAC trusts by default):
   ```powershell
   go install github.com/zred0627/ds-rescue-cc/cmd/ds-rescue@latest
   ```

2. **Unblock the downloaded file** (right-click → Properties → Unblock, or PowerShell):
   ```powershell
   Unblock-File "$env:USERPROFILE\bin\ds-rescue.exe"
   ```

3. **Add an exclusion** (requires Administrator):
   ```powershell
   Add-MpPreference -ExclusionPath "$env:USERPROFILE\bin\ds-rescue.exe"
   ```

The install.ps1 script detects your SAC state and prints the appropriate guidance automatically.

### How much does daily usage cost?

v4-pro runs at ~$0.01 per review — low enough that daily habitual use costs a few dollars a month. v4-flash (~$0.002/review) is even cheaper for high-frequency iteration.

### Can I customize the review persona?

Yes. Fork or edit `SKILL.md` at `plugins/ds-rescue/skills/ds-plan-challenger/SKILL.md`. The file has three sections (`## Plan Mode`, `## Scheme Mode`, `## Execution Mode`), each containing the system prompt and output schema for that review type.

Point ds-rescue at your custom version:

```bash
export DS_RESCUE_SKILL_PATH=/path/to/your/SKILL.md
```

Or per-invocation:

```bash
ds-rescue --skill /path/to/your/SKILL.md --plan my-plan.md
```

Ideas for custom lenses: finance / investment memo audit, security audit (injection surfaces, auth boundaries), infrastructure-as-code review (rollback paths, blast radius), Chinese-language reasoning depth mode.

### Is my code sent to DeepSeek's servers?

Yes, as with any LLM API call. The plan document, scheme comparison, or git diff is sent to DeepSeek's API servers for processing. This is identical to the data sharing that occurs when using Codex, Claude API, or Gemini API. If your organization's policy prohibits sending proprietary code to external LLM APIs, that constraint applies to ds-rescue as it does to all LLM-powered tools.

### Do I need an internet connection?

Yes. ds-rescue always calls the DeepSeek API. There is no offline or local-model mode.

### Which DeepSeek models are supported?

| Alias | Model | Characteristics |
|---|---|---|
| `pro` (default) | deepseek-chat (v4-pro) | Best quality, ~$0.01/review |
| `flash` | deepseek-chat-fast | Faster, ~$0.002/review |

Select with `--model pro` or `--model flash`.

### Can I use this without Claude Code?

Yes — ds-rescue is a universal CLI. Examples for other LLM coding tools:

```bash
# opencode
ds-rescue --mode plan docs/plans/current-plan.md

# Aider
ds-rescue --mode execution && aider --message "fix the issues found by ds-rescue"

# Plain bash / pre-push hook
ds-rescue --mode execution --no-tools || echo "Review found issues — check stderr"

# stdin pipe
git diff HEAD~1 HEAD | ds-rescue --mode execution --stdin
```

### How do I update?

```bash
# Re-run the same install command — always fetches latest
go install github.com/zred0627/ds-rescue-cc/cmd/ds-rescue@latest

# If using the Claude Code plugin
/plugin update ds-rescue
```

---

## Universal CLI usage (non-Claude-Code)

ds-rescue is a standard CLI binary. It works in any environment where you can run a shell command.

**opencode:**

```bash
ds-rescue --mode plan docs/plans/current-plan.md
```

**Continue (VS Code / JetBrains) — add a custom command:**

```json
{
  "name": "ds-rescue review",
  "command": "ds-rescue --mode execution",
  "description": "Run DeepSeek adversarial review of current changes"
}
```

**Aider:**

```bash
ds-rescue --mode execution && aider --message "fix the issues found by ds-rescue"
```

**Bare shell / CI pre-push hook:**

```bash
ds-rescue --mode execution --no-tools || echo "Review found issues — check stderr"
```

The binary reads from `stdin` or a file argument, writes findings to `stdout`, and logs diagnostics to `stderr`. It is composable with any toolchain.

---

## Customize the review persona

The three-mode review behavior is defined in a single Markdown file:

```
plugins/ds-rescue/skills/ds-plan-challenger/SKILL.md
```

This file has three sections — `## Plan Mode`, `## Scheme Mode`, `## Execution Mode` — each containing the system prompt and output schema for that review type. The binary reads this file at runtime; it is not embedded in the binary.

```bash
# Fork the repo, edit SKILL.md to your taste
vim plugins/ds-rescue/skills/ds-plan-challenger/SKILL.md

# Point ds-rescue at your custom version
export DS_RESCUE_SKILL_PATH=/path/to/your/SKILL.md
ds-rescue --mode plan my-plan.md

# Or pass it per-invocation
ds-rescue --skill /path/to/your/SKILL.md --mode plan my-plan.md
```

Example fork ideas:

- **Finance / investment memo review**: flag missing assumptions, check DCF inputs, challenge revenue projections
- **Security audit lens**: flag injection surfaces, auth boundaries, secret handling
- **Infrastructure-as-code review**: flag missing rollback paths, blast radius, dependency ordering
- **Chinese-language reasoning depth**: deeper analysis for CN-market context (manufacturing, logistics, policy)

A PR that improves `SKILL.md` is as valuable as a PR that improves the Go code. No Go knowledge required.

---

## Configuration

### API key resolution priority

1. `$DEEPSEEK_API_KEY` environment variable (highest priority)
2. `$XDG_CONFIG_HOME/ds-rescue/key` (Linux/macOS XDG standard)
3. `%APPDATA%\ds-rescue\key` (Windows)
4. `~/.ds-rescue/key` (fallback for all platforms)

### SKILL.md path resolution priority

1. `--skill <path>` CLI flag (highest priority)
2. `$DS_RESCUE_SKILL_PATH` environment variable
3. `$CLAUDE_PLUGIN_ROOT/skills/ds-plan-challenger/SKILL.md` (set automatically by Claude Code plugin)
4. `~/.ds-rescue/skills/ds-plan-challenger/SKILL.md` (user-installed standalone)
5. If none found: binary exits with an error listing all paths attempted

### CLI flags reference

| Flag | Default | Description |
|---|---|---|
| `--mode {plan\|scheme\|execution}` | auto-detect | Review mode |
| `--plan` | — | Alias for `--mode plan` |
| `--scheme` | — | Alias for `--mode scheme` |
| `--execution` | — | Alias for `--mode execution` |
| `--skill <path>` | (see above) | Path to SKILL.md |
| `--model {pro\|flash}` | `pro` | DeepSeek model alias |
| `--api-timeout <seconds>` | `90` | API call timeout |
| `--exec-timeout <seconds>` | `30` | Per-tool-call timeout |
| `--no-tools` | false | Disable agentic tool-use loop |
| `--verbose` | false | Log every tool call to stderr |
| `--version` | — | Print version and git commit |
| `--check` | — | Self-check: key, binary, SKILL.md, model response |

---

## Why Smart App Control blocks our binary (technical deep-dive)

Windows 11 25H2 introduced Smart App Control (SAC) as a default-ON security feature. SAC evaluates every executable before it runs. For executables it cannot verify via Microsoft's reputation service or a trusted code-signing certificate, SAC checks for the Mark of the Web (MOTW) zone flag.

Any file downloaded from the internet — via a browser, `curl`, `Invoke-WebRequest`, or a GitHub Releases download — automatically receives `ZoneId=3` (Internet zone) in an NTFS alternate data stream. SAC blocks execution of such files unless:

- The file is signed with an Extended Validation (EV) certificate trusted by Microsoft, or
- The file has established a Microsoft reputation score (typically requires millions of downloads), or
- The MOTW flag is removed (via Unblock-File or the Properties dialog).

Our binary is unsigned because EV code signing costs $300–500/year and requires organizational verification — not feasible for a free open-source tool.

**The cleanest solution is `go install`**: when Go compiles the binary locally, the resulting executable is written to your local `$GOPATH/bin` without any MOTW flag. SAC has no basis to block it. This is why `go install` is the recommended installation method for Windows users.

Three workarounds in full:

```powershell
# 1. Compile locally — no MOTW, SAC-safe (RECOMMENDED)
go install github.com/zred0627/ds-rescue-cc/cmd/ds-rescue@latest

# 2. Unblock after download (removes MOTW flag from a specific file)
Unblock-File "$env:USERPROFILE\bin\ds-rescue.exe"

# 3. Add an exclusion (requires Administrator, affects Defender + SAC)
Add-MpPreference -ExclusionPath "$env:USERPROFILE\bin\ds-rescue.exe"
```

The `install.ps1` script checks `(Get-MpComputerStatus).SmartAppControlState` at install time and prints the appropriate guidance if SAC is ON.

---

## Agentic tool use

In the default mode, ds-rescue runs DeepSeek in an agentic tool-use loop (up to 20 rounds). The model can invoke three tools during a review session:

| Tool | What it does |
|---|---|
| `read_file` | Read a file at a given path (for inspecting referenced code or configs) |
| `write_file` | Write a file (for generating structured output or fix suggestions) |
| `bash_exec` | Execute a shell command (for running `git log`, `grep`, or similar) |

The tool-use loop is protected by a deny-pattern list. Commands matching patterns for `sudo`, `rm -rf /`, fork bombs, or credential exfiltration via curl are rejected before execution. Use `--no-tools` to disable agentic mode entirely.

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for contribution guidelines, commit message format, and security rules. PRs on `SKILL.md` prompt definitions are especially welcome — you can meaningfully improve ds-rescue without writing any Go.

---

## License

MIT — see [LICENSE](./LICENSE)

Copyright (c) 2026 Daniel

---

## Related work / Citations

- [codex-plugin-cc](https://github.com/anthropics/codex-plugin-cc) — the pioneer of bringing cross-family LLM review into Claude Code
- [DeepSeek API documentation](https://api-docs.deepseek.com/) — official API reference for supported models and pricing
- [Smart App Control overview](https://learn.microsoft.com/en-us/windows/apps/develop/smart-app-control/overview) — Microsoft's official SAC documentation
- [Generative Engine Optimization](https://en.wikipedia.org/wiki/Generative_engine_optimization) — README structure principles for AI search visibility

---

Maintained by [Daniel (zred0627)](https://github.com/zred0627) · v0.1.0 (2026-05)
