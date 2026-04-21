# Security Guardrails — Audit Results

**Date:** 2026-04-21
**Score:** 100% — Grade **A**

## Results

| # | Check | Severity | Status | Evidence |
| - | ----- | -------- | ------ | -------- |
| 1 | SEC-01 .env files gitignored | critical | PASS | `.gitignore` lines 19-23 cover `.env`, `.env.local`, `.env.*.local`, `.env.production`; `git ls-files '*.env*'` returns 0 tracked files |
| 2 | SEC-02 AI hooks restrict sensitive files | critical | PASS | `.claude/settings.json` defines `hooks.PreToolUse` with two entries — `Read\|Edit\|Write\|NotebookEdit` matcher (lines 12-20) and `Bash` matcher (lines 21-29) — blocking `.env`, `*.pem`, `*.key`, `*.p12`, `*.pfx`, `credentials*`, `secrets*`, `id_rsa`, `id_ed25519`, `.aws/credentials`, `.kube/config`. Hook verified active during audit (blocked a `ls .env.*` probe). |
| 3 | SEC-03 .env template exists | high | SKIP | No runtime env-var usage: Go has only `os.Getenv("CI")` in `internal/util/clipboard_test.go:12` (CI build var, not app config); zero matches for `os.environ`/`os.getenv`/`dotenv`/`load_dotenv` under `training/*.py`. Offline terminal app with no env-driven config. |
| 4 | SEC-04 No secrets in committed files | critical | PASS | 0 matches for `api[_-]?key\s*[:=]`, `apikey\s*[:=]`, `secret\s*[:=]\s*"..."`, `password\s*[:=]\s*"..."`, `token\s*[:=]\s*"[A-Za-z0-9+/=]{20,}"`, `-----BEGIN ... PRIVATE KEY-----`, `AKIA[0-9A-Z]{16}` across the full tree |
| 5 | SEC-05 Sensitive file coverage in .gitignore | high | PASS | `.gitignore` covers stack-relevant patterns: `.env*` (20-23), `*.pem`/`*.key`/`*.p12`/`*.pfx` (30-33), `.DS_Store`/`Thumbs.db` (26-27), `.venv/`/`venv/`/`__pycache__/`/`*.pyc` (6-11), `training/checkpoints/*` (17), `.idea/` (1), `bin/termchess`/`/termchess` binaries (2-3). No missing patterns for Go+Python+ONNX+CI stack. |

## Security Summary

- **Secrets management:** N/A — offline terminal chess app; no API keys, passwords, tokens, or cloud credentials in scope. Only secret-adjacent surface is the GitHub release token, which is a workflow secret (not in repo).
- **AI hooks:** configured — `.claude/settings.json` enforces SEC-02 via PreToolUse hooks on both file-access tools and Bash, with exit-code-2 blocking and informative stderr messages.
- **.gitignore coverage:** complete for the Go + Python + ONNX + local-config stack; no gaps relative to stack-relevant secret/artifact patterns.
- **Committed secrets:** 0/5 pattern families matched.

## Scoring

- Max (with SEC-03 SKIP): 13 − 2 = **11**
- Deductions: 0
- Final: 11/11 = **100%** → Grade **A**
