# internal/util

Cross-cutting utilities. Currently contains only a cross-platform clipboard helper.

- `clipboard.go` — copy text via `pbcopy` (macOS) / `xclip` / `wl-copy` (Linux). Returns an error if no clipboard tool is available; callers should degrade gracefully.

**Naming note:** `util/` is a generic name. If this package accumulates more unrelated helpers, split them into purpose-named packages (e.g., `internal/clipboard/`).
