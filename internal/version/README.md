# internal/version

Build-time version metadata exposed at runtime (`termchess --version`).

## Injected via LDFLAGS

The root `Makefile` injects:

- `Version` — `git describe --tags --always --dirty`
- `BuildDate` — ISO 8601 UTC timestamp
- `GitCommit` — short SHA

Default value is `"dev"` when built outside the Makefile.
