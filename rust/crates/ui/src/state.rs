//! Screen and mode enums (ports of the Go `iota` constants in `model.go`).

/// The current UI screen. 1:1 with the Go `Screen` constants.
#[derive(Clone, Copy, PartialEq, Eq, Debug)]
pub enum Screen {
    MainMenu,
    GameTypeSelect,
    BotSelect,
    ColorSelect,
    FenInput,
    GamePlay,
    GameOver,
    Settings,
    SavePrompt,
    DrawPrompt,
    BvBBotSelect,
    BvBGameMode,
    BvBGridConfig,
    BvBGamePlay,
    BvBStats,
    BvBViewModeSelect,
    BvBConcurrencySelect,
}

/// The type of game being played.
#[derive(Clone, Copy, PartialEq, Eq, Debug)]
pub enum GameType {
    PvP,
    PvBot,
    BvB,
}

/// Bot difficulty as chosen in the UI.
#[derive(Clone, Copy, PartialEq, Eq, Debug)]
pub enum BotDifficulty {
    Easy,
    Medium,
    Hard,
}

impl BotDifficulty {
    /// Display name, mirroring Go's `botDifficultyName`.
    pub fn name(self) -> &'static str {
        match self {
            BotDifficulty::Easy => "Easy",
            BotDifficulty::Medium => "Medium",
            BotDifficulty::Hard => "Hard",
        }
    }
}

/// Maps the UI difficulty to the bot crate difficulty (Go `uiBotDiffToBvB`).
impl From<BotDifficulty> for bot::Difficulty {
    fn from(d: BotDifficulty) -> bot::Difficulty {
        match d {
            BotDifficulty::Easy => bot::Difficulty::Easy,
            BotDifficulty::Medium => bot::Difficulty::Medium,
            BotDifficulty::Hard => bot::Difficulty::Hard,
        }
    }
}

/// Display mode for BvB gameplay.
#[derive(Clone, Copy, PartialEq, Eq, Debug)]
pub enum BvBViewMode {
    Grid,
    Single,
    StatsOnly,
}

/// Which action follows a save-prompt decision (replaces the "exit"/"menu" strings).
#[derive(Clone, Copy, PartialEq, Eq, Debug)]
pub enum SaveAction {
    Exit,
    Menu,
}
