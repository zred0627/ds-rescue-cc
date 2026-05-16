# ds-rescue-cc

> **Cross-family adversarial review for your plans and commits — powered by DeepSeek, 50x cheaper than running Codex on every review.**

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go 1.22+](https://img.shields.io/badge/Go-1.22%2B-00ADD8.svg)](https://go.dev)
[![Releases](https://img.shields.io/github/v/release/zred0627/ds-rescue-cc)](https://github.com/zred0627/ds-rescue-cc/releases)

---

## The Engineering Pain: Same-Family Review Is an Echo Chamber

You finish a plan. You run it through Claude for review. Claude says it looks great — clean
architecture, sensible approach, no obvious gaps. You ship it.

Three days later you discover a race condition in the retry logic that Claude never flagged.
Why? Because Claude wrote the plan in the first place. Asking the same model family to
review its own output is, structurally, an **echo chamber**.

This is not a criticism of Claude. It is a property of how LLMs work: models trained on
similar data corpora share systematic blind spots. A model that confidently generates
a plan with a subtle ordering bug will, with similarly high confidence, fail to notice
that bug on re-read. The failure mode is correlated, not independent.

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

---

## Why Cross-Family Review Catches More Bugs

Software engineering has a well-studied analog: **code review effectiveness scales with
reviewer diversity**. Studies on pair review and inspection programs consistently show that
1.7–2.3× more defects are found when reviewers have different backgrounds and prior
exposure than when they share the same mental model.

The same principle applies to LLM review. A model from a different training family:

- Has different biases about what "normal" looks like
- Has different priors about which edge cases are worth mentioning
- Has different failure modes that are less likely to overlap with the author's

DeepSeek v4-pro is trained on a materially different data distribution than Claude.
When Claude misses something, DeepSeek often catches it — and vice versa. That
complementarity is the engineering value being unlocked here.

The goal is not to replace Claude. It is to add an independent second reviewer from a
different family to the workflow — the same way you would not let the author of a PR
be its only reviewer.

---

## Why Not Just Use Codex?

[codex-plugin-cc](https://github.com/anthropics/codex-plugin-cc) pioneered the concept
of bringing a non-Claude reviewer into the Claude Code workflow. It deserves full credit
for establishing that pattern and making it a first-class part of the plugin ecosystem.
ds-rescue-cc exists because of that work, not in spite of it.

The practical limitation is cost. A single Codex review call costs $0.30–0.60 USD in
API spend at typical plan/commit sizes. That is fine for occasional use — a weekly
architecture review, a pre-launch audit. It is expensive enough that most developers
will not run it on every plan, every commit, every scheme comparison. The habit does
not form.

Habit formation in developer tooling requires that the marginal cost of one more use
be low enough to feel free. At $0.30–0.60 per review, that threshold is not met.

ds-rescue-cc is designed around a different cost curve.

---

## Why DeepSeek: Engineering Economics + Cross-Family Complementarity

### Cost

DeepSeek v4-pro costs approximately **$0.01 per review call** at typical input/output
sizes. That is 50x cheaper than an equivalent Codex review call.

At $0.01 per review, the economic barrier to habitual use effectively disappears.
Running a review on every meaningful commit feels free — because, at that price point,
it is. This is the primary engineering argument for DeepSeek as the review engine.

### Quality

DeepSeek v4-pro approaches GPT-4-class performance on SWE-Bench and coding benchmarks.
For the review use case — analyzing a plan document or git diff and surfacing issues —
the quality difference versus the most expensive alternatives is not material. It is
good enough to catch real bugs, as the example above demonstrates.

### Cross-Family Complementarity

DeepSeek is trained on a different data distribution with different architectural choices
than OpenAI or Anthropic models. This is the property that makes it valuable as a
reviewer: its blind spots are different. When combined with Claude-authored plans,
you get genuine independent review, not a paraphrase of the author's own analysis.

### Chinese-Domain Reasoning Depth

For teams working in Chinese or on problems where Chinese-language technical literature
is relevant (manufacturing systems, logistics infrastructure, policy analysis), DeepSeek's
Chinese-language reasoning quality is noticeably stronger than most Western-origin models.
This is not a secondary feature — for certain domains it is the primary differentiator.

---

## Three Modes: What ds-rescue Does

ds-rescue operates in three modes, matching the three review contexts in a typical
development workflow:

| Mode | What it reviews | When to use |
|------|----------------|-------------|
| `plan` | Implementation plan documents (task DAGs, architecture decisions, risk registers) | Before starting a feature or sprint |
| `scheme` | Scheme comparison tables (A vs B vs C analysis, option trade-offs) | When choosing between architectural alternatives |
| `execution` | Git diff of the last commit or staged changes | After finishing a unit of work, before push |

ds-rescue is **not** a chat replacement. It does not answer questions or generate code.
It is a single-purpose reviewer: give it a document or a diff, get back structured
findings from a model that was not involved in producing the artifact under review.

The prompt behavior for each mode is defined in
`plugins/ds-rescue/skills/ds-plan-challenger/SKILL.md`. That file is the single
source of truth for what ds-rescue says and how it formats its output — you can read
it, fork it, and customize it without touching any Go code.

---

## Quickstart: 60 Seconds to First Review

### Step 1: Install the binary

**Linux / macOS:**

```bash
curl -L https://github.com/zred0627/ds-rescue-cc/releases/latest/download/ds-rescue-linux-amd64 \
  -o ds-rescue && chmod +x ds-rescue && sudo mv ds-rescue /usr/local/bin/
```

For macOS (Apple Silicon):

```bash
curl -L https://github.com/zred0627/ds-rescue-cc/releases/latest/download/ds-rescue-darwin-arm64 \
  -o ds-rescue && chmod +x ds-rescue && sudo mv ds-rescue /usr/local/bin/
```

**Windows (PowerShell):**

```powershell
Invoke-WebRequest -Uri https://github.com/zred0627/ds-rescue-cc/releases/latest/download/ds-rescue-windows-amd64.exe `
  -OutFile ds-rescue.exe
# Move to a directory in your PATH, e.g.:
Move-Item ds-rescue.exe "$env:LOCALAPPDATA\Programs\ds-rescue\ds-rescue.exe"
```

**Go install (any platform with Go 1.22+):**

```bash
go install github.com/zred0627/ds-rescue-cc/cmd/ds-rescue@latest
```

### Step 2: Set your DeepSeek API key

```bash
# Option A: environment variable (recommended for CI)
export DEEPSEEK_API_KEY=sk-your-key-here

# Option B: key file (recommended for interactive use, survives shell restarts)
mkdir -p ~/.ds-rescue && echo "sk-your-key-here" > ~/.ds-rescue/key && chmod 600 ~/.ds-rescue/key
```

Get a DeepSeek API key at [platform.deepseek.com](https://platform.deepseek.com).

### Step 3: Run your first review

```bash
# Review a plan document
ds-rescue --mode plan path/to/your-plan.md

# Review a scheme comparison
ds-rescue --mode scheme path/to/scheme-comparison.md

# Review your last commit
ds-rescue --mode execution
```

### Self-check

```bash
ds-rescue --check
```

This verifies: key found, binary version OK, SKILL.md loadable, model responds.

---

## Claude Code Plugin: One-Click Install

If you use Claude Code (claude.ai/code), you can install ds-rescue as a plugin
that adds a `/ds-rescue` slash command to your sessions:

```
/plugin marketplace add zred0627/ds-rescue-cc
/plugin install ds-rescue
```

After installation, use it directly from the Claude Code prompt:

```
/ds-rescue --plan path/to/plan.md
/ds-rescue --scheme path/to/scheme-comparison.md
/ds-rescue --execution
```

The plugin also registers a `SessionStart` hook that checks for the ds-rescue binary
on startup and offers to install it if missing. The check is a single `test -f` call
and adds no perceptible latency to session start.

### What the plugin provides on top of the CLI

The plugin wrapper:
- Sets `$CLAUDE_PLUGIN_ROOT` so the binary finds SKILL.md automatically
- Surfaces findings inline in the Claude Code conversation
- Lets you reference findings in follow-up messages without leaving the session

The underlying binary is identical to the standalone CLI — the plugin is a thin
integration layer, not a separate product.

---

## Universal CLI: Works in Any Environment

ds-rescue is a standard CLI binary. It works in any environment where you can run
a shell command — no Claude Code required.

**opencode:**

```bash
# In your opencode tool configuration or terminal panel
ds-rescue --mode plan docs/plans/current-plan.md
```

**Continue (VS Code / JetBrains):**

Add a custom command in your Continue configuration:

```json
{
  "name": "ds-rescue review",
  "command": "ds-rescue --mode execution",
  "description": "Run DeepSeek adversarial review of current changes"
}
```

**Aider:**

```bash
# Run before an aider session to review the current state
ds-rescue --mode execution && aider --message "fix the issues found by ds-rescue"
```

**Bare shell / CI:**

```bash
# In a pre-push hook or CI step
ds-rescue --mode execution --no-tools || echo "Review found issues — check stderr"
```

The binary reads from `stdin` or a file argument, writes findings to `stdout`,
and logs diagnostics to `stderr`. It is composable with any toolchain.

---

## Customizing the Skill Prompts

The three-mode review behavior is defined in a single Markdown file:

```
plugins/ds-rescue/skills/ds-plan-challenger/SKILL.md
```

This file has three sections (`## Plan Mode`, `## Scheme Mode`, `## Execution Mode`),
each containing the system prompt and output schema for that review type. The binary
reads this file at runtime — it is not embedded in the binary.

**To customize:**

```bash
# Fork the repository, edit SKILL.md to your taste
vim plugins/ds-rescue/skills/ds-plan-challenger/SKILL.md

# Point ds-rescue at your custom version
export DS_RESCUE_SKILL_PATH=/path/to/your/SKILL.md
ds-rescue --mode plan my-plan.md

# Or pass it per-invocation
ds-rescue --skill /path/to/your/SKILL.md --mode plan my-plan.md
```

**Ideas for custom lenses:**

- Finance / investment memo review (flag missing assumptions, check DCF inputs)
- Security audit lens (flag injection surfaces, auth boundaries, secret handling)
- Infrastructure-as-code review (flag missing rollback paths, blast radius)
- Chinese-language reasoning mode (deeper analysis for CN-market context)

Because SKILL.md is plain Markdown and the binary loads it from disk, you can
maintain a fleet of specialized review prompts and switch between them with a
single flag. No Go knowledge required to contribute prompt improvements — a PR
that improves the SKILL.md is as valuable as a PR that improves the Go code.

---

## SKILL.md Path Resolution

The binary looks for SKILL.md in this priority order:

1. `--skill <path>` CLI flag (highest priority)
2. `$DS_RESCUE_SKILL_PATH` environment variable
3. `$CLAUDE_PLUGIN_ROOT/skills/ds-plan-challenger/SKILL.md` (set automatically by Claude Code plugin)
4. `~/.ds-rescue/skills/ds-plan-challenger/SKILL.md` (user-installed standalone)
5. If none found: binary exits with an error listing all paths attempted

Run `ds-rescue --check` to see which path was resolved and confirm the file is loadable.

---

## Agentic Tool Use

In the default mode, ds-rescue runs the DeepSeek model in an **agentic tool-use loop**
(up to 20 rounds). The model can invoke three tools during a review session:

| Tool | What it does |
|------|-------------|
| `read_file` | Read a file at a given path (for inspecting referenced code or configs) |
| `write_file` | Write a file (for generating structured output or fix suggestions) |
| `bash_exec` | Execute a shell command (for running `git log`, `grep`, or similar) |

The tool-use loop is protected by a deny-pattern list. Commands matching patterns
for `sudo`, `rm -rf /`, fork bombs, or credential exfiltration via curl are
rejected before execution. Use `--no-tools` to disable agentic mode entirely
and run as a single-shot reviewer.

---

## CLI Reference

```
ds-rescue [flags] [<input-file>]

Flags:
  --mode {plan|scheme|execution}   Review mode (default: auto-detected)
  --plan                           Alias for --mode plan
  --scheme                         Alias for --mode scheme
  --execution                      Alias for --mode execution
  --skill <path>                   Path to SKILL.md (overrides env / plugin / home lookup)
  --model {pro|flash|reasoner}     DeepSeek model alias (default: pro = deepseek-chat)
  --api-timeout <seconds>          API call timeout (default: 90)
  --exec-timeout <seconds>         Per-tool-call timeout (default: 30)
  --no-tools                       Disable agentic tool-use loop
  --verbose                        Log every tool call to stderr
  --version                        Print version and git commit
  --check                          Self-check: key found? binary OK? SKILL.md loadable? model responds?
```

---

## License

MIT — see [LICENSE](LICENSE).

Copyright (c) 2026 Daniel

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for contribution guidelines, commit message
format, and security rules. PRs on SKILL.md prompt definitions are especially welcome —
you can meaningfully improve ds-rescue without writing any Go.

---

## Acknowledgements

[codex-plugin-cc](https://github.com/anthropics/codex-plugin-cc) established the
pattern of bringing a non-Claude reviewer into the Claude Code workflow. ds-rescue-cc
builds on that foundation with a different cost profile and a different model family.
The approach would not exist without that pioneer work.
