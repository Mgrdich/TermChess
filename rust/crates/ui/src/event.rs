//! Application events and key normalization.

use crossterm::event::{KeyCode, KeyEvent, KeyModifiers, MouseEvent};
use engine::Move;

/// Events fed into `App::update`, from the input thread, timers, or workers.
pub enum AppEvent {
    /// A normalized key press.
    Key(Key),
    /// A raw mouse event.
    Mouse(MouseEvent),
    /// Terminal resized to (width, height).
    Resize(u16, u16),
    /// The bot selected a move (guarded by generation).
    BotMove { generation: u64, mv: Move },
    /// The bot failed to select a move.
    BotMoveError { generation: u64, err: String },
    /// A newer version is available.
    UpdateAvailable(String),
    /// The blink timer fired.
    BlinkTick,
    /// The BvB playback timer fired.
    BvBTick,
}

/// A normalized key, mapping crossterm `KeyEvent` to the tokens the Go
/// `switch msg.String()` handlers expect.
#[derive(Clone, Copy, PartialEq, Eq, Debug)]
pub enum Key {
    Up,
    Down,
    Left,
    Right,
    Enter,
    Esc,
    Backspace,
    Tab,
    Space,
    Char(char),
    CtrlC,
    Other,
}

impl Key {
    /// Normalizes a crossterm key event.
    pub fn from_event(ev: KeyEvent) -> Key {
        match ev.code {
            KeyCode::Up => Key::Up,
            KeyCode::Down => Key::Down,
            KeyCode::Left => Key::Left,
            KeyCode::Right => Key::Right,
            KeyCode::Enter => Key::Enter,
            KeyCode::Esc => Key::Esc,
            KeyCode::Backspace => Key::Backspace,
            KeyCode::Tab => Key::Tab,
            KeyCode::Char(' ') => Key::Space,
            KeyCode::Char(c) => {
                if ev.modifiers.contains(KeyModifiers::CONTROL) && (c == 'c' || c == 'C') {
                    Key::CtrlC
                } else {
                    Key::Char(c)
                }
            }
            _ => Key::Other,
        }
    }

    /// Returns the string token used by the per-screen `match` handlers.
    pub fn token(self) -> String {
        match self {
            Key::Up => "up".to_string(),
            Key::Down => "down".to_string(),
            Key::Left => "left".to_string(),
            Key::Right => "right".to_string(),
            Key::Enter => "enter".to_string(),
            Key::Esc => "esc".to_string(),
            Key::Backspace => "backspace".to_string(),
            Key::Tab => "tab".to_string(),
            Key::Space => " ".to_string(),
            Key::Char(c) => c.to_string(),
            Key::CtrlC => "ctrl+c".to_string(),
            Key::Other => String::new(),
        }
    }

    /// Returns the typed character, if this is a printable key (incl. space).
    pub fn as_char(self) -> Option<char> {
        match self {
            Key::Char(c) => Some(c),
            Key::Space => Some(' '),
            _ => None,
        }
    }
}
