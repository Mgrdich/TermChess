//! Terminal UI for TermChess. Ported from the Go `internal/ui` package,
//! translating its Bubbletea MVU architecture to ratatui + crossterm.
//!
//! The Bubbletea MVU loop maps onto:
//!   - [`App`] — owns all state (Go `Model`).
//!   - [`App::update`] — mutates state in response to an [`AppEvent`].
//!   - [`App::draw`] — renders the current screen with ratatui.
//!   - worker threads + an `mpsc` channel — replace `tea.Cmd` async effects.
//!
//! Entry point: [`run`].

mod app;
mod board;
mod event;
mod mouse;
mod navigation;
mod runtime;
mod san;
mod state;
mod text_field;
mod theme;
mod update;
mod view;

pub use app::App;
pub use board::BoardRenderer;
pub use event::{AppEvent, Key};
pub use runtime::run;
pub use san::{format_move_history, format_san, parse_san};
pub use state::{BotDifficulty, BvBViewMode, GameType, SaveAction, Screen};
pub use theme::{get_theme, parse_theme_name, Theme, ThemeName};
