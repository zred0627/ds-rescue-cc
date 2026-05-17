# Contributing to ds-rescue-cc

Thank you for your interest in contributing. This document covers the essentials.

## What contributions are most welcome

**Skill prompt PRs are the highest-value contribution.** The three-mode prompt definitions
live in `plugins/ds-rescue/skills/ds-plan-challenger/SKILL.md` — not in the binary.
If you have a specialized review lens (finance, security, Chinese reasoning depth,
infrastructure-as-code), a focused PR on SKILL.md is exactly the right place.

Other welcome contributions:
- Bug fixes in Go code (especially edge cases on Windows or non-UTF-8 paths)
- Additional deny-pattern entries in the agentic tool-use security boundary
- Improvements to the cross-platform install scripts
- Translations or corrections to `README.zh.md`

## Go code style

- Run `gofmt -w .` before committing — CI will reject unformatted code.
- Run `go vet ./...` locally; PRs must pass `go vet` in CI.
- Keep the core module (`cmd/`, `internal/`) dependency-free (stdlib only).
- Use `filepath.Join` exclusively for path construction — never string concatenation.

## Commit message format

This project uses **Conventional Commits**:

```
<type>(<scope>): <short imperative summary>

[optional body]

[optional footer: Signed-off-by: Your Name <email>]
```

Types: `feat`, `fix`, `docs`, `chore`, `refactor`, `test`, `ci`

Examples:
- `feat(loop): add write_file tool to agentic loop`
- `fix(config): handle missing APPDATA on Windows Server`
- `docs: add finance lens section to SKILL.md`

## Security rules — hard requirements

- **No hardcoded API keys** in tests or examples. Use `t.Setenv("DEEPSEEK_API_KEY", "sk-test-...")` in tests.
- **No secrets in filenames or paths** that could be accidentally staged.
- If you're unsure whether something might leak credentials, run `gitleaks detect --no-git --source .` before opening a PR.

## Sign-off (recommended)

A `Signed-off-by` trailer is not required but is appreciated:

```
git commit -s -m "feat(modes): improve scheme auto-detection heuristic"
```

This signals that you have the rights to submit the contribution under the MIT license.

## Opening a PR

1. Fork the repository and create a feature branch.
2. Make your changes; include tests where applicable.
3. Run `go test ./...` and confirm all tests pass.
4. Open a PR with a clear description of the change and why it improves ds-rescue.

Questions? Open an issue — we're happy to help scope out a contribution before you invest time.
