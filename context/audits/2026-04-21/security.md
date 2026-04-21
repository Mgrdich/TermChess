# Security Guardrails — Audit Results

**Date:** 2026-04-21
**Score:** 30% — Grade **F**

## Results

| #      | Check                                              | Severity | Status | Evidence                                                                                                                                                                 |
| ------ | -------------------------------------------------- | -------- | ------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| SEC-01 | .env files are gitignored                          | critical | FAIL   | `.gitignore` contains no `.env*` patterns (lines 1-18). `git ls-files '*.env*'` returns nothing, so no tracked `.env` files — but `.env` is not gitignored.              |
| SEC-02 | AI agent hooks restrict access to sensitive files  | critical | FAIL   | `.claude/settings.json` defines only `extraKnownMarketplaces`; no `PreToolUse` hooks present. `.claude/settings.local.json` only sets Bash allowlist and plugin enables. |
| SEC-03 | .env.example or template exists                    | high     | SKIP   | Only env var reference is `os.Getenv("CI")` in `internal/util/clipboard_test.go:12`; no real env var usage. No `.env*` files exist in repo.                              |
| SEC-04 | No secrets in committed files                      | critical | PASS   | No matches for api_key/secret/password/token literal patterns, BEGIN PRIVATE KEY, or AKIA AWS key patterns across repo.                                                  |
| SEC-05 | Sensitive files in .gitignore coverage             | high     | WARN   | `.gitignore` covers Python (`__pycache__/`, `*.py[cod]`, `.venv/`, `venv/`) but misses `.env*`, `.DS_Store`, `Thumbs.db`, `*.pem`, `*.key`.                              |

## Summary

**`.gitignore` coverage:** Python artifacts well covered (`__pycache__/`, `*.py[cod]`, `.venv/`, `venv/`, `.pytest_cache/`, `*.egg-info/`, `dist/`, `build/`) and project-specific artifacts (`bin/termchess`, `termchess`, `training/checkpoints/*`, `.idea/`). Missing: `.env` patterns, OS files (`.DS_Store`, `Thumbs.db`), and private-key patterns (`*.pem`, `*.key`, `*.p12`, `*.pfx`).

**`.env` handling:** No `.env*` files exist in the repo and nothing sensitive is tracked, but `.env` is not in `.gitignore`, leaving a gap for future accidental commits. The project currently has no runtime env var dependency (only a `CI` check in tests), so a `.env.example` is not needed today.

**Hook coverage:** `.claude/settings.json` contains only marketplace config — there are no `PreToolUse` deny-hooks blocking AI read/glob/bash access to `.env`, `*.pem`, `*.key`, `credentials*`, or `*secret*` patterns. `.claude/settings.local.json` grants `Bash(go test:*)`, `Bash(awk:*)`, `Bash(go run:*)`, `Bash(gh pr view:*)`, `Bash(uv run:*)` without corresponding deny rules for sensitive paths.

**Secret-scan findings:** Clean. No hardcoded API keys, passwords, tokens, private keys, or AWS access keys found across Go, Python, or shell files. The single `password`/`secret`-style match universe returned zero hits.

**Risk profile:** This is a local TUI app with filesystem-only storage, no network services, no cloud credentials, and no database. Real blast radius of a secrets leak is low. However, the AI-agent-friendly repo structure (extensive `.claude/` configuration, multiple agent definitions) combined with absent PreToolUse guardrails is the main gap: an AI agent is not prevented from reading dotfiles or private keys if the developer ever places them in the workspace. The `.env` gitignore gap should also be closed pre-emptively before any future integrations (e.g. updater GitHub token, future API keys) get introduced.
