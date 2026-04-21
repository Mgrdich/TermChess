# internal/config

User configuration and save-game persistence under `~/.termchess/`.

## Files written

| Path | Purpose |
|------|---------|
| `~/.termchess/config.toml` | User preferences (Unicode pieces, colors, bot delay, etc.) |
| `~/.termchess/savegame.fen` | Auto-saved game state for Resume |

## Modules

- `config.go` — load/save TOML via `BurntSushi/toml`; schema migration hook.
- `paths.go` — resolve platform-appropriate config dir (macOS, Linux).
- `savegame.go` — FEN-based save/restore of the active game.

Test fixtures use `t.TempDir()` to avoid touching the real `~/.termchess/`.
