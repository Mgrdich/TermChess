# internal/bvb

Bot-vs-Bot session management for spectator mode.

## Responsibilities

- `session.go` — single-game controller: orchestrates two `bot.Engine` implementations taking turns, records moves, detects game end.
- `manager.go` — multi-game queue: runs up to 50 games concurrently, collects results.
- `stats.go` — win-rate, average length, decisive/draw breakdown.
- `export.go` — PGN-like export of completed games.
- `types.go` — `Result`, `GameRecord`, `Session` types.

## UI integration

Consumed by the `bvb` screen in `internal/ui`. The UI subscribes to session state updates via Bubbletea messages.
