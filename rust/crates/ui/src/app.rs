//! The `App` struct (port of Go `Model`) and the central update/effect logic.

use std::cell::Cell;
use std::sync::mpsc::Sender;
use std::sync::Arc;
use std::thread;
use std::time::{Duration, Instant};

use bvb::{PlaybackSpeed, SessionManager};
use config::Config;
use engine::{Board, Color, Move, Square};
use rand::Rng;

use crate::event::{AppEvent, Key};
use crate::state::{BotDifficulty, BvBViewMode, GameType, SaveAction, Screen};
use crate::text_field::TextField;
use crate::theme::{get_theme, parse_theme_name, Theme};

/// Minimum terminal dimensions for the UI to render properly.
pub const MIN_TERMINAL_WIDTH: u16 = 40;
pub const MIN_TERMINAL_HEIGHT: u16 = 20;
/// Fixed BvB grid cell dimensions.
pub const BVB_CELL_HEIGHT: usize = 12;
pub const BVB_CELL_WIDTH: usize = 22;

/// Humorous "bot is thinking" messages (Go `thinkingMessages`).
const THINKING_MESSAGES: &[&str] = &[
    "Consulting the ancient chess masters...",
    "Calculating infinite possibilities...",
    "Pondering the meaning of chess...",
    "Summoning the spirit of Bobby Fischer...",
    "Analyzing 42 dimensions of chess space...",
    "Teaching my neural networks a lesson...",
    "Asking my rubber duck for advice...",
    "Flipping through my opening book...",
    "Sacrificing pawns to the chess gods...",
    "Pretending to think really hard...",
    "Counting squares intensely...",
    "Channeling my inner Stockfish...",
];

/// The whole application state.
pub struct App {
    // Game state.
    pub board: Option<Board>,
    pub move_history: Vec<Move>,

    // UI / navigation.
    pub screen: Screen,
    pub nav_stack: Vec<Screen>,
    pub config: Config,
    pub theme: Theme,
    pub term_width: u16,
    pub term_height: u16,
    pub should_quit: bool,

    // Input / messages.
    pub input: String,
    pub fen_input: TextField,
    pub error_msg: String,
    pub status_msg: String,

    // Menu selections.
    pub menu_selection: usize,
    pub menu_options: Vec<String>,
    pub settings_selection: usize,
    pub save_prompt_selection: usize,
    pub save_prompt_action: SaveAction,
    pub draw_prompt_selection: usize,

    // Game metadata.
    pub game_type: GameType,
    pub bot_difficulty: BotDifficulty,
    pub user_color: Color,
    pub resigned_by: Option<Color>,
    pub draw_offered_by: Option<Color>,
    pub draw_offered_by_white: bool,
    pub draw_offered_by_black: bool,
    pub draw_by_agreement: bool,

    // Bot vs Bot.
    pub bvb_white_diff: BotDifficulty,
    pub bvb_black_diff: BotDifficulty,
    pub bvb_selecting_white: bool,
    pub bvb_game_count: i32,
    pub bvb_count_input: String,
    pub bvb_inputting_count: bool,
    pub bvb_grid_rows: i32,
    pub bvb_grid_cols: i32,
    pub bvb_custom_grid_input: String,
    pub bvb_inputting_grid: bool,
    pub bvb_manager: Option<Arc<SessionManager>>,
    pub bvb_speed: PlaybackSpeed,
    pub bvb_selected_game: usize,
    pub bvb_view_mode: BvBViewMode,
    pub bvb_paused: bool,
    pub bvb_page_index: usize,
    pub bvb_stats_selection: usize,
    pub bvb_stats_results_page: usize,
    pub bvb_jump_input: String,
    pub bvb_show_jump_prompt: bool,
    pub bvb_view_mode_selection: usize,
    pub bvb_recent_completions: Vec<String>,
    pub bvb_concurrency_selection: usize,
    pub bvb_custom_concurrency: String,
    pub bvb_inputting_concurrency: bool,
    pub bvb_concurrency: i32,
    pub bvb_show_abort_confirm: bool,
    pub bvb_abort_selection: usize,

    // Overlays.
    pub show_shortcuts_overlay: bool,

    // Mouse interaction.
    pub selected_square: Option<Square>,
    pub valid_moves: Vec<Square>,
    pub blink_on: bool,
    /// Screen (column, row) of the board's top-left piece cell (file a, rank 8),
    /// recorded during `draw` from the actual render area so mouse clicks map to
    /// squares relative to where ratatui really draws the board. `None` until the
    /// gameplay board has been drawn at least once.
    pub board_origin: Cell<Option<(u16, u16)>>,

    // Update notification.
    pub update_available: String,

    // Runtime plumbing.
    pub tx: Sender<AppEvent>,
    pub bot_move_generation: u64,
}

impl App {
    /// Creates a new app (Go `NewModel`), always starting at the main menu.
    pub fn new(config: Config, tx: Sender<AppEvent>) -> App {
        let theme = get_theme(parse_theme_name(&config.theme));
        let menu_options = build_main_menu_options();
        App {
            board: None,
            move_history: Vec::new(),
            screen: Screen::MainMenu,
            nav_stack: Vec::new(),
            config,
            theme,
            term_width: 0,
            term_height: 0,
            should_quit: false,
            input: String::new(),
            fen_input: TextField::new("Enter FEN string...", 100),
            error_msg: String::new(),
            status_msg: String::new(),
            menu_selection: 0,
            menu_options,
            settings_selection: 0,
            save_prompt_selection: 0,
            save_prompt_action: SaveAction::Exit,
            draw_prompt_selection: 0,
            game_type: GameType::PvP,
            bot_difficulty: BotDifficulty::Easy,
            user_color: Color::White,
            resigned_by: None,
            draw_offered_by: None,
            draw_offered_by_white: false,
            draw_offered_by_black: false,
            draw_by_agreement: false,
            bvb_white_diff: BotDifficulty::Easy,
            bvb_black_diff: BotDifficulty::Easy,
            bvb_selecting_white: true,
            bvb_game_count: 0,
            bvb_count_input: String::new(),
            bvb_inputting_count: false,
            bvb_grid_rows: 1,
            bvb_grid_cols: 1,
            bvb_custom_grid_input: String::new(),
            bvb_inputting_grid: false,
            bvb_manager: None,
            bvb_speed: PlaybackSpeed::Normal,
            bvb_selected_game: 0,
            bvb_view_mode: BvBViewMode::Grid,
            bvb_paused: false,
            bvb_page_index: 0,
            bvb_stats_selection: 0,
            bvb_stats_results_page: 0,
            bvb_jump_input: String::new(),
            bvb_show_jump_prompt: false,
            bvb_view_mode_selection: 0,
            bvb_recent_completions: Vec::new(),
            bvb_concurrency_selection: 0,
            bvb_custom_concurrency: String::new(),
            bvb_inputting_concurrency: false,
            bvb_concurrency: 0,
            bvb_show_abort_confirm: false,
            bvb_abort_selection: 0,
            show_shortcuts_overlay: false,
            selected_square: None,
            valid_moves: Vec::new(),
            blink_on: false,
            board_origin: Cell::new(None),
            update_available: String::new(),
            tx,
            bot_move_generation: 0,
        }
    }

    /// Bumps the bot-move generation, invalidating any in-flight bot replies.
    pub fn new_generation(&mut self) {
        self.bot_move_generation = self.bot_move_generation.wrapping_add(1);
    }

    /// Handles a terminal resize.
    pub fn on_resize(&mut self, w: u16, h: u16) {
        self.term_width = w;
        self.term_height = h;
        if self.screen == Screen::BvBGamePlay && self.bvb_view_mode == BvBViewMode::Grid {
            self.adjust_bvb_grid_for_width();
        }
    }

    /// Central update dispatcher (Go `Model.Update`).
    pub fn update(&mut self, ev: AppEvent) {
        match ev {
            AppEvent::Key(k) => self.handle_key(k),
            AppEvent::Resize(w, h) => self.on_resize(w, h),
            AppEvent::Mouse(m) => {
                if self.screen == Screen::GamePlay && self.game_type != GameType::BvB {
                    self.handle_mouse_event(m);
                }
            }
            AppEvent::BvBTick => self.handle_bvb_tick(),
            AppEvent::BotMove { generation, mv } => self.handle_bot_move(generation, mv),
            AppEvent::BotMoveError { generation, err } => {
                self.handle_bot_move_error(generation, err)
            }
            AppEvent::BlinkTick => {
                if self.selected_square.is_some() {
                    self.blink_on = !self.blink_on;
                    self.spawn_blink_tick();
                } else {
                    self.blink_on = false;
                }
            }
            AppEvent::UpdateAvailable(v) => {
                self.update_available = v;
            }
        }
    }

    /// Global key handling and per-screen dispatch (Go `handleKeyPress`).
    fn handle_key(&mut self, key: Key) {
        // Shortcuts overlay: any key dismisses it.
        if self.show_shortcuts_overlay {
            self.show_shortcuts_overlay = false;
            return;
        }

        let token = key.token();

        // '?' toggles the shortcuts overlay (not in text input mode).
        if token == "?" && !self.is_in_text_input_mode() {
            self.show_shortcuts_overlay = true;
            return;
        }

        // 'n' global new game.
        if token == "n"
            && !self.is_in_text_input_mode()
            && self.screen != Screen::GameTypeSelect
            && self.screen != Screen::GamePlay
            && self.screen != Screen::GameOver
            && self.screen != Screen::BvBGamePlay
            && self.screen != Screen::BvBStats
        {
            self.push_screen(Screen::GameTypeSelect);
            self.menu_options = vec![
                "Player vs Player".into(),
                "Player vs Bot".into(),
                "Bot vs Bot".into(),
            ];
            self.menu_selection = 0;
            self.status_msg.clear();
            self.error_msg.clear();
            return;
        }

        // 's' global settings.
        if token == "s"
            && !self.is_in_text_input_mode()
            && self.screen != Screen::Settings
            && self.screen != Screen::BvBStats
        {
            self.push_screen(Screen::Settings);
            self.settings_selection = 0;
            self.status_msg.clear();
            self.error_msg.clear();
            return;
        }

        // Global quit keys.
        match token.as_str() {
            "ctrl+c" => {
                self.cleanup_on_quit();
                self.should_quit = true;
                return;
            }
            "q" if self.screen != Screen::GamePlay => {
                self.cleanup_on_quit();
                self.should_quit = true;
                return;
            }
            _ => {}
        }

        match self.screen {
            Screen::MainMenu => self.handle_main_menu_keys(key),
            Screen::GameTypeSelect => self.handle_game_type_select_keys(key),
            Screen::BotSelect => self.handle_bot_select_keys(key),
            Screen::ColorSelect => self.handle_color_select_keys(key),
            Screen::FenInput => self.handle_fen_input_keys(key),
            Screen::GamePlay => self.handle_gameplay_keys(key),
            Screen::GameOver => self.handle_game_over_keys(key),
            Screen::Settings => self.handle_settings_keys(key),
            Screen::SavePrompt => self.handle_save_prompt_keys(key),
            Screen::DrawPrompt => self.handle_draw_prompt_keys(key),
            Screen::BvBBotSelect => self.handle_bvb_bot_select_keys(key),
            Screen::BvBGameMode => self.handle_bvb_game_mode_keys(key),
            Screen::BvBGridConfig => self.handle_bvb_grid_config_keys(key),
            Screen::BvBGamePlay => self.handle_bvb_gameplay_keys(key),
            Screen::BvBStats => self.handle_bvb_stats_keys(key),
            Screen::BvBViewModeSelect => self.handle_bvb_view_mode_select_keys(key),
            Screen::BvBConcurrencySelect => self.handle_bvb_concurrency_select_keys(key),
        }
    }

    fn cleanup_on_quit(&mut self) {
        if let Some(mgr) = self.bvb_manager.take() {
            mgr.abort();
        }
        self.new_generation();
    }

    /// Returns true when text input should capture '?', 'n', 's' (Go `isInTextInputMode`).
    pub fn is_in_text_input_mode(&self) -> bool {
        if self.screen == Screen::FenInput || self.screen == Screen::GamePlay {
            return true;
        }
        if self.screen == Screen::BvBGameMode && self.bvb_inputting_count {
            return true;
        }
        if self.screen == Screen::BvBGridConfig && self.bvb_inputting_grid {
            return true;
        }
        if self.screen == Screen::BvBGamePlay && self.bvb_show_jump_prompt {
            return true;
        }
        if self.screen == Screen::BvBConcurrencySelect && self.bvb_inputting_concurrency {
            return true;
        }
        false
    }

    /// Clears game state on exit (Go `cleanupGame`).
    pub fn cleanup_game(&mut self) {
        self.board = None;
        self.move_history.clear();
        self.selected_square = None;
        self.valid_moves.clear();
        self.input.clear();
        self.error_msg.clear();
        self.blink_on = false;
        self.new_generation();
    }

    // ---- Bot moves ----

    /// Starts an asynchronous bot move (Go `makeBotMove`).
    pub fn make_bot_move(&mut self) {
        self.status_msg = random_thinking_message();
        let board = match &self.board {
            Some(b) => b.clone(),
            None => return,
        };
        let difficulty = self.bot_difficulty;
        let generation = self.bot_move_generation;
        let tx = self.tx.clone();

        thread::spawn(move || {
            let start = Instant::now();
            let min_delay = minimum_bot_delay(difficulty);

            let engine_result: Result<Move, String> = (|| {
                let ctx = bot::Context::background();
                match difficulty {
                    BotDifficulty::Easy => {
                        let eng = bot::new_random_engine(&[]).map_err(|e| e.to_string())?;
                        bot::Engine::select_move(&eng, &ctx, &board).map_err(|e| e.to_string())
                    }
                    BotDifficulty::Medium => {
                        let eng = bot::new_minimax_engine(bot::Difficulty::Medium, &[])
                            .map_err(|e| e.to_string())?;
                        bot::Engine::select_move(&eng, &ctx, &board).map_err(|e| e.to_string())
                    }
                    BotDifficulty::Hard => {
                        let eng = bot::new_minimax_engine(bot::Difficulty::Hard, &[])
                            .map_err(|e| e.to_string())?;
                        bot::Engine::select_move(&eng, &ctx, &board).map_err(|e| e.to_string())
                    }
                }
            })();

            match engine_result {
                Ok(mv) => {
                    let elapsed = start.elapsed();
                    if elapsed < min_delay {
                        thread::sleep(min_delay - elapsed);
                    }
                    let _ = tx.send(AppEvent::BotMove { generation, mv });
                }
                Err(err) => {
                    let _ = tx.send(AppEvent::BotMoveError { generation, err });
                }
            }
        });
    }

    /// Applies a successful bot move (Go `handleBotMove`).
    fn handle_bot_move(&mut self, generation: u64, mv: Move) {
        if generation != self.bot_move_generation {
            return; // stale reply
        }
        let board = match self.board.as_mut() {
            Some(b) => b,
            None => return,
        };
        if let Err(e) = board.make_move(mv) {
            self.error_msg = format!("Bot generated invalid move: {}", e);
            self.status_msg.clear();
            return;
        }
        self.status_msg.clear();
        self.error_msg.clear();
        self.move_history.push(mv);

        if self
            .board
            .as_ref()
            .map(|b| b.is_game_over())
            .unwrap_or(false)
        {
            self.screen = Screen::GameOver;
            let _ = config::delete_save_game();
        }
    }

    /// Handles a bot move error (Go `handleBotMoveError`).
    fn handle_bot_move_error(&mut self, generation: u64, err: String) {
        if generation != self.bot_move_generation {
            return;
        }
        self.error_msg = format!("Bot error: {}", err);
        self.status_msg.clear();
    }

    // ---- BvB tick ----

    /// Handles a BvB tick, polling the session (Go `handleBvBTick`).
    fn handle_bvb_tick(&mut self) {
        if self.screen != Screen::BvBGamePlay || self.bvb_manager.is_none() {
            return;
        }
        self.update_recent_completions();

        let finished = self
            .bvb_manager
            .as_ref()
            .map(|m| m.all_finished())
            .unwrap_or(true);
        if finished {
            self.screen = Screen::BvBStats;
            self.bvb_stats_selection = 0;
            self.bvb_stats_results_page = 0;
            self.menu_options = vec!["New Session".into(), "Return to Menu".into()];
            return;
        }
        self.spawn_bvb_tick();
    }

    /// Updates the last-5 completion log (Go `updateRecentCompletions`).
    fn update_recent_completions(&mut self) {
        let mgr = match &self.bvb_manager {
            Some(m) => m,
            None => return,
        };
        let stats = mgr.stats();
        let results = &stats.individual_results;
        if results.is_empty() {
            return;
        }
        let max_recent = 5;
        let start_idx = results.len().saturating_sub(max_recent);
        let mut recent = Vec::with_capacity(max_recent);
        for r in results[start_idx..].iter().rev() {
            let entry = if r.winner == "Draw" {
                format!("Game {}: Draw ({} moves)", r.game_number, r.move_count)
            } else {
                format!(
                    "Game {}: {} wins ({} moves)",
                    r.game_number, r.winner, r.move_count
                )
            };
            recent.push(entry);
        }
        self.bvb_recent_completions = recent;
    }

    /// Adjusts the BvB grid for the current terminal width (Go `adjustBvBGridForWidth`).
    pub fn adjust_bvb_grid_for_width(&mut self) {
        if self.term_width == 0 {
            return;
        }
        let cell_width_with_margin = BVB_CELL_WIDTH as i32 + 2;
        let mut max_cols = self.term_width as i32 / cell_width_with_margin;
        if max_cols < 1 {
            max_cols = 1;
        }
        if self.bvb_grid_cols > max_cols {
            self.bvb_grid_cols = max_cols;
            if self.bvb_game_count > 0 && self.bvb_grid_cols > 0 {
                let original_boards_per_page = self.bvb_grid_rows * self.bvb_grid_cols;
                self.bvb_grid_rows =
                    (original_boards_per_page + self.bvb_grid_cols - 1) / self.bvb_grid_cols;
                if self.bvb_grid_rows < 1 {
                    self.bvb_grid_rows = 1;
                }
            }
        }
        if (self.term_width as usize) < BVB_CELL_WIDTH {
            self.bvb_view_mode = BvBViewMode::Single;
        }
    }

    // ---- worker/timer spawns ----

    /// Spawns the update-check worker once at startup (Go `Init`/`checkForUpdateCmd`).
    pub fn spawn_update_check(&self) {
        if version::VERSION == "dev" {
            return;
        }
        let tx = self.tx.clone();
        thread::spawn(move || {
            let ctx = updater::Context::with_timeout(Duration::from_secs(5));
            let client = updater::Client::new();
            if let Ok(latest) = client.check_latest_version(&ctx) {
                if updater::compare_versions(&latest, version::VERSION) > 0 {
                    let _ = tx.send(AppEvent::UpdateAvailable(latest));
                }
            }
        });
    }

    /// Spawns a one-shot 500ms blink tick (Go `blinkTickCmd`).
    pub fn spawn_blink_tick(&self) {
        let tx = self.tx.clone();
        thread::spawn(move || {
            thread::sleep(Duration::from_millis(500));
            let _ = tx.send(AppEvent::BlinkTick);
        });
    }

    /// Spawns a one-shot BvB tick based on the current speed (Go `bvbTickCmd`).
    pub fn spawn_bvb_tick(&self) {
        let mut delay = self.bvb_speed.duration();
        if delay.is_zero() {
            delay = Duration::from_millis(100);
        }
        let tx = self.tx.clone();
        thread::spawn(move || {
            thread::sleep(delay);
            let _ = tx.send(AppEvent::BvBTick);
        });
    }
}

/// Builds the main menu options (Go `buildMainMenuOptions`).
pub fn build_main_menu_options() -> Vec<String> {
    if config::save_game_exists() {
        vec![
            "Resume Game".into(),
            "New Game".into(),
            "Load Game".into(),
            "Settings".into(),
            "Exit".into(),
        ]
    } else {
        vec![
            "New Game".into(),
            "Load Game".into(),
            "Settings".into(),
            "Exit".into(),
        ]
    }
}

fn random_thinking_message() -> String {
    let idx = rand::thread_rng().gen_range(0..THINKING_MESSAGES.len());
    THINKING_MESSAGES[idx].to_string()
}

/// Minimum bot move delay (Go `getMinimumBotDelay`).
fn minimum_bot_delay(difficulty: BotDifficulty) -> Duration {
    match difficulty {
        BotDifficulty::Easy | BotDifficulty::Medium => {
            // 1-2s randomized delay, matching Go's rng.Float64() draw.
            let secs = 1.0 + rand::thread_rng().gen::<f64>();
            Duration::from_secs_f64(secs)
        }
        BotDifficulty::Hard => Duration::from_secs(1),
    }
}
