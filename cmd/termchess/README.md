# cmd/termchess

Main binary entry point for TermChess. Parses top-level CLI flags (`--version`, `--upgrade`, `--uninstall`) and hands off to the Bubbletea program in `internal/ui`.

## Build

```bash
make build             # → bin/termchess
make build-all         # cross-compile to dist/
```

Version metadata is injected at build time via LDFLAGS — see the root `Makefile`.
