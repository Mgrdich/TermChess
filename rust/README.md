# TermChess (Rust)

A Rust port of the TermChess terminal chess TUI, living alongside the original Go
implementation in the same repository. This is a Cargo workspace (`rust/Cargo.toml`);
library crates live under `rust/crates/` and the binary crate is `rust/app/`
(package/binary name `termchess`).

## Documentation

The authoritative migration guide — build/run instructions, the full crate ↔ Go-package
mapping, the Bubbletea-MVU → ratatui architectural change, the tri-language board-encoder
invariant, parity notes, and follow-ups — lives at:

**[../docs/MIGRATION.md](../docs/MIGRATION.md)**

## Quick reference

From the repo root:

```sh
make build-rust   # cd rust && cargo build --release -p termchess
make run-rust     # cd rust && cargo run -p termchess
make test-rust    # cd rust && cargo test
```

Or drive Cargo directly from inside `rust/`:

```sh
cd rust
cargo build                 # build the whole workspace
cargo test                  # run all crate test suites
cargo run -p termchess -- --help
cargo doc --no-deps         # browse crate-level docs
```
