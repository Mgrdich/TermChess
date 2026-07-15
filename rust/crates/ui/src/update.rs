//! Per-screen key handlers (port of Go `update.go`).

use std::sync::Arc;

use bvb::{PlaybackSpeed, SessionManager};
use engine::{Board, Color};

use crate::app::{build_main_menu_options, App};
use crate::event::Key;
use crate::san::parse_san;
use crate::state::{BotDifficulty, BvBViewMode, GameType, SaveAction, Screen};
use crate::theme::{cycle_theme, get_theme, parse_theme_name};

/// Parses a string as a positive integer (>= 1) (Go `parsePositiveInt`).
fn parse_positive_int(s: &str) -> Result<i32, String> {
    if s.is_empty() {
        return Err("empty input".to_string());
    }
    let mut n: i32 = 0;
    for r in s.chars() {
        if !r.is_ascii_digit() {
            return Err("not a number".to_string());
        }
        n = n * 10 + (r as i32 - '0' as i32);
    }
    if n < 1 {
        return Err("must be at least 1".to_string());
    }
    Ok(n)
}

/// Parses "RxC" into (rows, cols), max 8 boards (Go `parseGridDimensions`).
fn parse_grid_dimensions(s: &str) -> Result<(i32, i32), String> {
    let sep = s.find(['x', 'X']);
    let sep = match sep {
        Some(i) => i,
        None => return Err("use format RxC (e.g., 2x3)".to_string()),
    };
    let rows = parse_positive_int(&s[..sep]).map_err(|e| format!("invalid rows: {}", e))?;
    let cols = parse_positive_int(&s[sep + 1..]).map_err(|e| format!("invalid cols: {}", e))?;
    if rows * cols > 8 {
        return Err(format!(
            "max 8 boards (got {}x{} = {})",
            rows,
            cols,
            rows * cols
        ));
    }
    Ok((rows, cols))
}

/// Navigates a wrapping menu selection.
fn nav_up(sel: usize, len: usize) -> usize {
    if len == 0 {
        return 0;
    }
    if sel > 0 {
        sel - 1
    } else {
        len - 1
    }
}

fn nav_down(sel: usize, len: usize) -> usize {
    if len == 0 {
        return 0;
    }
    if sel < len - 1 {
        sel + 1
    } else {
        0
    }
}

impl App {
    // ---- Main menu ----

    pub(crate) fn handle_main_menu_keys(&mut self, key: Key) {
        self.error_msg.clear();
        self.status_msg.clear();
        match key.token().as_str() {
            "up" | "k" => {
                self.menu_selection = nav_up(self.menu_selection, self.menu_options.len())
            }
            "down" | "j" => {
                self.menu_selection = nav_down(self.menu_selection, self.menu_options.len())
            }
            "enter" => self.handle_main_menu_selection(),
            _ => {}
        }
    }

    fn handle_main_menu_selection(&mut self) {
        let selected = self.menu_options[self.menu_selection].clone();
        match selected.as_str() {
            "Resume Game" => match config::load_game() {
                Ok(board) => {
                    self.board = Some(board);
                    self.move_history.clear();
                    self.clear_nav_stack();
                    self.new_generation();
                    self.screen = Screen::GamePlay;
                    self.input.clear();
                    self.error_msg.clear();
                    self.status_msg = "Game resumed".to_string();
                    self.reset_game_flags();
                }
                Err(e) => {
                    self.error_msg = format!("Failed to load saved game: {}", e);
                }
            },
            "Exit" => self.should_quit = true,
            "New Game" => {
                self.push_screen(Screen::GameTypeSelect);
                self.menu_options = vec![
                    "Player vs Player".into(),
                    "Player vs Bot".into(),
                    "Bot vs Bot".into(),
                ];
                self.menu_selection = 0;
                self.status_msg.clear();
                self.error_msg.clear();
                self.input.clear();
            }
            "Load Game" => {
                self.push_screen(Screen::FenInput);
                self.fen_input.set_value("");
                self.status_msg.clear();
                self.error_msg.clear();
            }
            "Settings" => {
                self.push_screen(Screen::Settings);
                self.settings_selection = 0;
                self.status_msg.clear();
                self.error_msg.clear();
            }
            _ => {}
        }
    }

    fn reset_game_flags(&mut self) {
        self.resigned_by = None;
        self.draw_offered_by = None;
        self.draw_offered_by_white = false;
        self.draw_offered_by_black = false;
        self.draw_by_agreement = false;
    }

    // ---- Game type select ----

    pub(crate) fn handle_game_type_select_keys(&mut self, key: Key) {
        self.error_msg.clear();
        self.status_msg.clear();
        match key.token().as_str() {
            "up" | "k" => {
                self.menu_selection = nav_up(self.menu_selection, self.menu_options.len())
            }
            "down" | "j" => {
                self.menu_selection = nav_down(self.menu_selection, self.menu_options.len())
            }
            "enter" => self.handle_game_type_selection(),
            "esc" => {
                self.pop_screen();
                self.status_msg.clear();
            }
            _ => {}
        }
    }

    fn handle_game_type_selection(&mut self) {
        let selected = self.menu_options[self.menu_selection].clone();
        match selected.as_str() {
            "Player vs Player" => {
                self.game_type = GameType::PvP;
                self.board = Some(Board::new());
                self.clear_nav_stack();
                self.new_generation();
                self.screen = Screen::GamePlay;
                self.status_msg.clear();
                self.error_msg.clear();
                self.input.clear();
                self.reset_game_flags();
            }
            "Player vs Bot" => {
                self.game_type = GameType::PvBot;
                self.push_screen(Screen::BotSelect);
                self.menu_options = vec!["Easy".into(), "Medium".into(), "Hard".into()];
                self.menu_selection = 0;
                self.status_msg.clear();
                self.error_msg.clear();
            }
            "Bot vs Bot" => {
                self.game_type = GameType::BvB;
                self.bvb_selecting_white = true;
                self.push_screen(Screen::BvBBotSelect);
                self.menu_options = vec!["Easy".into(), "Medium".into(), "Hard".into()];
                self.menu_selection = 0;
                self.status_msg.clear();
                self.error_msg.clear();
            }
            _ => {}
        }
    }

    // ---- Gameplay ----

    pub(crate) fn handle_gameplay_keys(&mut self, key: Key) {
        let token = key.token();
        if token == "q" || token == "Q" {
            self.screen = Screen::SavePrompt;
            self.save_prompt_selection = 0;
            self.save_prompt_action = SaveAction::Exit;
            self.error_msg.clear();
            self.status_msg.clear();
            return;
        }
        if token == "esc" {
            self.screen = Screen::SavePrompt;
            self.save_prompt_selection = 0;
            self.save_prompt_action = SaveAction::Menu;
            self.error_msg.clear();
            self.status_msg.clear();
            return;
        }

        match key {
            Key::Backspace => {
                self.input.pop();
                self.error_msg.clear();
            }
            Key::Enter => {
                if !self.input.is_empty() {
                    self.handle_gameplay_input();
                }
            }
            _ => {
                if let Some(c) = key.as_char() {
                    self.error_msg.clear();
                    self.input.push(c);
                }
            }
        }
    }

    fn handle_gameplay_input(&mut self) {
        let input = self.input.trim().to_lowercase();
        match input.as_str() {
            "resign" => self.handle_resign_command(),
            "showfen" => self.handle_show_fen_command(),
            "menu" => self.handle_menu_command(),
            "offerdraw" => self.handle_offer_draw_command(),
            _ => self.handle_move_input(),
        }
    }

    fn handle_resign_command(&mut self) {
        if let Some(board) = &self.board {
            self.resigned_by = Some(board.active_color);
        }
        self.screen = Screen::GameOver;
        self.input.clear();
        self.error_msg.clear();
        self.status_msg.clear();
        let _ = config::delete_save_game();
    }

    fn handle_show_fen_command(&mut self) {
        let fen = match &self.board {
            Some(b) => b.to_fen(),
            None => return,
        };
        match util::copy_to_clipboard(&fen) {
            Ok(()) => self.status_msg = format!("FEN: {} (Copied to clipboard)", fen),
            Err(e) => {
                self.status_msg = format!("FEN: {} (Failed to copy to clipboard: {})", fen, e)
            }
        }
        self.input.clear();
        self.error_msg.clear();
    }

    fn handle_menu_command(&mut self) {
        self.screen = Screen::SavePrompt;
        self.save_prompt_selection = 0;
        self.save_prompt_action = SaveAction::Menu;
        self.input.clear();
        self.error_msg.clear();
        self.status_msg.clear();
    }

    fn handle_move_input(&mut self) {
        let mv = {
            let board = match &self.board {
                Some(b) => b,
                None => return,
            };
            match parse_san(board, &self.input) {
                Ok(m) => Some(m),
                Err(_) => match engine::Move::parse(&self.input) {
                    Ok(m) => Some(m),
                    Err(e) => {
                        self.error_msg = format!("Invalid move: {}", e);
                        return;
                    }
                },
            }
        };
        let mv = match mv {
            Some(m) => m,
            None => return,
        };

        if let Some(board) = self.board.as_mut() {
            if let Err(e) = board.make_move(mv) {
                self.error_msg = e.to_string();
                return;
            }
        }

        self.input.clear();
        self.error_msg.clear();
        self.status_msg.clear();
        self.move_history.push(mv);

        if self
            .board
            .as_ref()
            .map(|b| b.is_game_over())
            .unwrap_or(false)
        {
            self.screen = Screen::GameOver;
            let _ = config::delete_save_game();
            return;
        }

        if self.game_type == GameType::PvBot {
            self.make_bot_move();
        }
    }

    fn handle_offer_draw_command(&mut self) {
        let active = match &self.board {
            Some(b) => b.active_color,
            None => return,
        };
        if (active == Color::White && self.draw_offered_by_white)
            || (active == Color::Black && self.draw_offered_by_black)
        {
            self.error_msg = "You have already offered a draw this game".to_string();
            self.input.clear();
            return;
        }

        self.draw_offered_by = Some(active);
        if active == Color::White {
            self.draw_offered_by_white = true;
        } else {
            self.draw_offered_by_black = true;
        }

        self.screen = Screen::DrawPrompt;
        self.draw_prompt_selection = 0;
        self.input.clear();
        self.error_msg.clear();
        self.status_msg.clear();
    }

    // ---- Game over ----

    pub(crate) fn handle_game_over_keys(&mut self, key: Key) {
        match key.token().as_str() {
            "n" | "N" => {
                self.new_generation();
                self.board = None;
                self.move_history.clear();
                self.screen = Screen::GameTypeSelect;
                self.input.clear();
                self.error_msg.clear();
                self.status_msg.clear();
                self.menu_options = vec![
                    "Player vs Player".into(),
                    "Player vs Bot".into(),
                    "Bot vs Bot".into(),
                ];
                self.menu_selection = 0;
                self.reset_game_flags();
            }
            "m" | "M" | "esc" => {
                self.new_generation();
                self.screen = Screen::MainMenu;
                self.board = None;
                self.move_history.clear();
                self.input.clear();
                self.error_msg.clear();
                self.status_msg.clear();
                self.menu_options = vec![
                    "New Game".into(),
                    "Load Game".into(),
                    "Settings".into(),
                    "Exit".into(),
                ];
                self.menu_selection = 0;
            }
            "q" | "Q" => {
                self.should_quit = true;
            }
            _ => {}
        }
    }

    // ---- Settings ----

    pub(crate) fn handle_settings_keys(&mut self, key: Key) {
        self.error_msg.clear();
        self.status_msg.clear();
        let num_settings = 6usize;
        match key.token().as_str() {
            "up" | "k" => self.settings_selection = nav_up(self.settings_selection, num_settings),
            "down" | "j" => {
                self.settings_selection = nav_down(self.settings_selection, num_settings)
            }
            "enter" | " " => self.toggle_selected_setting(),
            "esc" | "q" | "b" | "backspace" => {
                self.pop_screen();
                self.status_msg.clear();
            }
            _ => {}
        }
    }

    fn toggle_selected_setting(&mut self) {
        match self.settings_selection {
            0 => self.config.use_unicode = !self.config.use_unicode,
            1 => self.config.show_coords = !self.config.show_coords,
            2 => self.config.use_colors = !self.config.use_colors,
            3 => self.config.show_move_history = !self.config.show_move_history,
            4 => self.config.show_help_text = !self.config.show_help_text,
            5 => {
                self.config.theme = cycle_theme(&self.config.theme);
                self.theme = get_theme(parse_theme_name(&self.config.theme));
            }
            _ => {}
        }
        match config::save_config(&self.config) {
            Ok(()) => self.status_msg = "Setting saved successfully".to_string(),
            Err(e) => self.error_msg = format!("Failed to save settings: {}", e),
        }
    }

    // ---- Save prompt ----

    pub(crate) fn handle_save_prompt_keys(&mut self, key: Key) {
        self.error_msg.clear();
        self.status_msg.clear();
        match key.token().as_str() {
            "up" | "k" => {
                self.save_prompt_selection = if self.save_prompt_selection > 0 { 0 } else { 1 }
            }
            "down" | "j" => {
                self.save_prompt_selection = if self.save_prompt_selection < 1 { 1 } else { 0 }
            }
            "y" | "Y" => {
                if let Some(board) = &self.board {
                    if let Err(e) = config::save_game(board) {
                        self.error_msg = format!("Failed to save game: {}", e);
                        return;
                    }
                }
                self.status_msg = "Game saved!".to_string();
                self.finish_save_prompt();
            }
            "n" | "N" => {
                self.finish_save_prompt();
            }
            "enter" => {
                if self.save_prompt_selection == 0 {
                    if let Some(board) = &self.board {
                        if let Err(e) = config::save_game(board) {
                            self.error_msg = format!("Failed to save: {}", e);
                            return;
                        }
                    }
                    self.status_msg = "Game saved!".to_string();
                }
                self.finish_save_prompt();
            }
            "esc" => {
                self.screen = Screen::GamePlay;
                self.error_msg.clear();
            }
            _ => {}
        }
    }

    fn finish_save_prompt(&mut self) {
        self.cleanup_game();
        self.screen = Screen::MainMenu;
        self.menu_options = build_main_menu_options();
        self.menu_selection = 0;
        self.nav_stack.clear();
    }

    // ---- FEN input ----

    pub(crate) fn handle_fen_input_keys(&mut self, key: Key) {
        match key {
            Key::Esc => {
                self.pop_screen();
                self.status_msg.clear();
                self.fen_input.set_value("");
            }
            Key::Enter => {
                let fen = self.fen_input.value().to_string();
                if fen.is_empty() {
                    self.error_msg = "Please enter a FEN string".to_string();
                    return;
                }
                match Board::from_fen(&fen) {
                    Ok(board) => {
                        self.board = Some(board);
                        self.move_history.clear();
                        self.clear_nav_stack();
                        self.new_generation();
                        self.screen = Screen::GamePlay;
                        self.game_type = GameType::PvP;
                        self.input.clear();
                        self.error_msg.clear();
                        self.status_msg.clear();
                        self.fen_input.set_value("");
                        self.reset_game_flags();
                    }
                    Err(e) => {
                        self.error_msg = format!("Invalid FEN: {}", e);
                    }
                }
            }
            Key::Backspace => {
                self.fen_input.backspace();
                self.error_msg.clear();
            }
            Key::Left => self.fen_input.left(),
            Key::Right => self.fen_input.right(),
            _ => {
                if let Some(c) = key.as_char() {
                    self.fen_input.insert(c);
                    self.error_msg.clear();
                }
            }
        }
    }

    // ---- Draw prompt ----

    pub(crate) fn handle_draw_prompt_keys(&mut self, key: Key) {
        self.error_msg.clear();
        self.status_msg.clear();
        match key.token().as_str() {
            "up" | "k" => {
                self.draw_prompt_selection = if self.draw_prompt_selection > 0 { 0 } else { 1 }
            }
            "down" | "j" => {
                self.draw_prompt_selection = if self.draw_prompt_selection < 1 { 1 } else { 0 }
            }
            "enter" => {
                if self.draw_prompt_selection == 0 {
                    self.draw_by_agreement = true;
                    self.screen = Screen::GameOver;
                    self.input.clear();
                    self.error_msg.clear();
                    self.status_msg.clear();
                    let _ = config::delete_save_game();
                } else {
                    self.screen = Screen::GamePlay;
                    self.status_msg = "Draw offer declined".to_string();
                    self.input.clear();
                    self.error_msg.clear();
                    self.draw_offered_by = None;
                }
            }
            "esc" => {
                self.screen = Screen::GamePlay;
                self.status_msg = "Draw offer cancelled".to_string();
                self.input.clear();
                self.error_msg.clear();
                match self.draw_offered_by {
                    Some(Color::White) => self.draw_offered_by_white = false,
                    Some(Color::Black) => self.draw_offered_by_black = false,
                    None => {}
                }
                self.draw_offered_by = None;
            }
            _ => {}
        }
    }

    // ---- Bot select ----

    pub(crate) fn handle_bot_select_keys(&mut self, key: Key) {
        self.error_msg.clear();
        self.status_msg.clear();
        match key.token().as_str() {
            "up" | "k" => {
                self.menu_selection = nav_up(self.menu_selection, self.menu_options.len())
            }
            "down" | "j" => {
                self.menu_selection = nav_down(self.menu_selection, self.menu_options.len())
            }
            "enter" => self.handle_bot_difficulty_selection(),
            "esc" => {
                self.pop_screen();
                self.status_msg.clear();
            }
            _ => {}
        }
    }

    fn handle_bot_difficulty_selection(&mut self) {
        self.bot_difficulty = match self.menu_options[self.menu_selection].as_str() {
            "Medium" => BotDifficulty::Medium,
            "Hard" => BotDifficulty::Hard,
            _ => BotDifficulty::Easy,
        };
        self.push_screen(Screen::ColorSelect);
        self.menu_options = vec!["Play as White".into(), "Play as Black".into()];
        self.menu_selection = 0;
        self.status_msg.clear();
        self.error_msg.clear();
    }

    // ---- Color select ----

    pub(crate) fn handle_color_select_keys(&mut self, key: Key) {
        self.error_msg.clear();
        self.status_msg.clear();
        match key.token().as_str() {
            "up" | "k" => {
                self.menu_selection = nav_up(self.menu_selection, self.menu_options.len())
            }
            "down" | "j" => {
                self.menu_selection = nav_down(self.menu_selection, self.menu_options.len())
            }
            "enter" => self.handle_color_selection(),
            "esc" => {
                self.pop_screen();
                self.status_msg.clear();
            }
            _ => {}
        }
    }

    fn handle_color_selection(&mut self) {
        self.user_color = match self.menu_options[self.menu_selection].as_str() {
            "Play as Black" => Color::Black,
            _ => Color::White,
        };
        self.board = Some(Board::new());
        self.clear_nav_stack();
        self.new_generation();
        self.screen = Screen::GamePlay;
        self.status_msg.clear();
        self.error_msg.clear();
        self.input.clear();
        self.reset_game_flags();

        if self.user_color == Color::Black {
            self.make_bot_move();
        }
    }

    // ---- BvB bot select ----

    pub(crate) fn handle_bvb_bot_select_keys(&mut self, key: Key) {
        self.error_msg.clear();
        self.status_msg.clear();
        match key.token().as_str() {
            "up" | "k" => {
                self.menu_selection = nav_up(self.menu_selection, self.menu_options.len())
            }
            "down" | "j" => {
                self.menu_selection = nav_down(self.menu_selection, self.menu_options.len())
            }
            "enter" => self.handle_bvb_bot_difficulty_selection(),
            "esc" => {
                if self.bvb_selecting_white {
                    self.pop_screen();
                } else {
                    self.bvb_selecting_white = true;
                    self.menu_selection = 0;
                }
            }
            _ => {}
        }
    }

    fn handle_bvb_bot_difficulty_selection(&mut self) {
        let diff = match self.menu_options[self.menu_selection].as_str() {
            "Medium" => BotDifficulty::Medium,
            "Hard" => BotDifficulty::Hard,
            _ => BotDifficulty::Easy,
        };
        if self.bvb_selecting_white {
            self.bvb_white_diff = diff;
            self.bvb_selecting_white = false;
            self.menu_selection = 0;
            self.status_msg.clear();
            self.error_msg.clear();
        } else {
            self.bvb_black_diff = diff;
            self.push_screen(Screen::BvBGameMode);
            self.menu_options = vec!["Single Game".into(), "Multi-Game".into()];
            self.menu_selection = 0;
            self.bvb_inputting_count = false;
            self.bvb_count_input.clear();
            self.status_msg.clear();
            self.error_msg.clear();
        }
    }

    // ---- BvB game mode ----

    pub(crate) fn handle_bvb_game_mode_keys(&mut self, key: Key) {
        self.error_msg.clear();
        if self.bvb_inputting_count {
            self.handle_bvb_count_input(key);
            return;
        }
        match key.token().as_str() {
            "up" | "k" => {
                self.menu_selection = nav_up(self.menu_selection, self.menu_options.len())
            }
            "down" | "j" => {
                self.menu_selection = nav_down(self.menu_selection, self.menu_options.len())
            }
            "enter" => self.handle_bvb_game_mode_selection(),
            "esc" => {
                self.pop_screen();
                self.bvb_selecting_white = false;
            }
            _ => {}
        }
    }

    fn handle_bvb_game_mode_selection(&mut self) {
        match self.menu_options[self.menu_selection].as_str() {
            "Single Game" => {
                self.bvb_game_count = 1;
                self.bvb_grid_rows = 1;
                self.bvb_grid_cols = 1;
                self.bvb_view_mode = BvBViewMode::Single;
                self.start_bvb_session();
            }
            "Multi-Game" => {
                self.bvb_inputting_count = true;
                self.bvb_count_input.clear();
                self.status_msg.clear();
                self.error_msg.clear();
            }
            _ => {}
        }
    }

    fn handle_bvb_count_input(&mut self, key: Key) {
        match key {
            Key::Esc => {
                self.bvb_inputting_count = false;
                self.bvb_count_input.clear();
                self.error_msg.clear();
            }
            Key::Backspace => {
                self.bvb_count_input.pop();
            }
            Key::Enter => match parse_positive_int(&self.bvb_count_input) {
                Ok(count) => {
                    self.bvb_game_count = count;
                    self.bvb_inputting_count = false;
                    self.push_screen(Screen::BvBGridConfig);
                    self.menu_options = vec![
                        "1x1".into(),
                        "2x2".into(),
                        "2x3".into(),
                        "2x4".into(),
                        "Custom".into(),
                    ];
                    self.menu_selection = 0;
                    self.bvb_inputting_grid = false;
                    self.bvb_custom_grid_input.clear();
                    self.status_msg.clear();
                    self.error_msg.clear();
                }
                Err(_) => {
                    self.error_msg = "Please enter a positive integer".to_string();
                }
            },
            _ => {
                if let Some(c) = key.as_char() {
                    if c.is_ascii_digit() {
                        self.bvb_count_input.push(c);
                    }
                }
            }
        }
    }

    // ---- BvB grid config ----

    pub(crate) fn handle_bvb_grid_config_keys(&mut self, key: Key) {
        self.error_msg.clear();
        if self.bvb_inputting_grid {
            self.handle_bvb_grid_input(key);
            return;
        }
        match key.token().as_str() {
            "up" | "k" => {
                self.menu_selection = nav_up(self.menu_selection, self.menu_options.len())
            }
            "down" | "j" => {
                self.menu_selection = nav_down(self.menu_selection, self.menu_options.len())
            }
            "enter" => self.handle_bvb_grid_selection(),
            "esc" => {
                self.pop_screen();
                self.bvb_inputting_grid = false;
            }
            _ => {}
        }
    }

    fn handle_bvb_grid_selection(&mut self) {
        match self.menu_options[self.menu_selection].as_str() {
            "1x1" => {
                self.bvb_grid_rows = 1;
                self.bvb_grid_cols = 1;
            }
            "2x2" => {
                self.bvb_grid_rows = 2;
                self.bvb_grid_cols = 2;
            }
            "2x3" => {
                self.bvb_grid_rows = 2;
                self.bvb_grid_cols = 3;
            }
            "2x4" => {
                self.bvb_grid_rows = 2;
                self.bvb_grid_cols = 4;
            }
            "Custom" => {
                self.bvb_inputting_grid = true;
                self.bvb_custom_grid_input.clear();
                return;
            }
            _ => {}
        }
        self.navigate_to_concurrency_select();
    }

    fn handle_bvb_grid_input(&mut self, key: Key) {
        match key {
            Key::Esc => {
                self.bvb_inputting_grid = false;
                self.bvb_custom_grid_input.clear();
                self.error_msg.clear();
            }
            Key::Backspace => {
                self.bvb_custom_grid_input.pop();
            }
            Key::Enter => match parse_grid_dimensions(&self.bvb_custom_grid_input) {
                Ok((rows, cols)) => {
                    self.bvb_grid_rows = rows;
                    self.bvb_grid_cols = cols;
                    self.bvb_inputting_grid = false;
                    self.navigate_to_concurrency_select();
                }
                Err(e) => self.error_msg = e,
            },
            _ => {
                if let Some(c) = key.as_char() {
                    if c.is_ascii_digit() || c == 'x' || c == 'X' {
                        self.bvb_custom_grid_input.push(c);
                    }
                }
            }
        }
    }

    fn navigate_to_concurrency_select(&mut self) {
        self.push_screen(Screen::BvBConcurrencySelect);
        self.bvb_concurrency_selection = 0;
        self.bvb_inputting_concurrency = false;
        self.bvb_custom_concurrency.clear();
        self.status_msg.clear();
        self.error_msg.clear();
    }

    fn navigate_to_view_mode_select(&mut self) {
        self.push_screen(Screen::BvBViewModeSelect);
        self.bvb_view_mode_selection = 0;
        self.status_msg.clear();
        self.error_msg.clear();
    }

    // ---- BvB view mode select ----

    pub(crate) fn handle_bvb_view_mode_select_keys(&mut self, key: Key) {
        self.error_msg.clear();
        let num_options = 3usize;
        match key.token().as_str() {
            "up" | "k" => {
                self.bvb_view_mode_selection = nav_up(self.bvb_view_mode_selection, num_options)
            }
            "down" | "j" => {
                self.bvb_view_mode_selection = nav_down(self.bvb_view_mode_selection, num_options)
            }
            "enter" => {
                self.bvb_view_mode = match self.bvb_view_mode_selection {
                    1 => BvBViewMode::Single,
                    2 => BvBViewMode::StatsOnly,
                    _ => BvBViewMode::Grid,
                };
                self.start_bvb_session();
            }
            "esc" => {
                self.pop_screen();
                self.bvb_concurrency_selection = 0;
                self.bvb_inputting_concurrency = false;
                self.bvb_custom_concurrency.clear();
                self.status_msg.clear();
            }
            _ => {}
        }
    }

    // ---- BvB concurrency select ----

    pub(crate) fn handle_bvb_concurrency_select_keys(&mut self, key: Key) {
        self.error_msg.clear();
        if self.bvb_inputting_concurrency {
            self.handle_bvb_concurrency_input(key);
            return;
        }
        let num_options = 2usize;
        match key.token().as_str() {
            "up" | "k" => {
                self.bvb_concurrency_selection = nav_up(self.bvb_concurrency_selection, num_options)
            }
            "down" | "j" => {
                self.bvb_concurrency_selection =
                    nav_down(self.bvb_concurrency_selection, num_options)
            }
            "enter" => match self.bvb_concurrency_selection {
                0 => {
                    self.bvb_concurrency = bvb::calculate_default_concurrency();
                    self.navigate_to_view_mode_select();
                }
                1 => {
                    self.bvb_inputting_concurrency = true;
                    self.bvb_custom_concurrency.clear();
                }
                _ => {}
            },
            "esc" => {
                self.pop_screen();
                self.bvb_inputting_grid = false;
                self.status_msg.clear();
            }
            _ => {}
        }
    }

    fn handle_bvb_concurrency_input(&mut self, key: Key) {
        match key {
            Key::Esc => {
                self.bvb_inputting_concurrency = false;
                self.bvb_custom_concurrency.clear();
                self.error_msg.clear();
            }
            Key::Backspace => {
                self.bvb_custom_concurrency.pop();
            }
            Key::Enter => match parse_positive_int(&self.bvb_custom_concurrency) {
                Ok(concurrency) => {
                    self.bvb_concurrency = concurrency;
                    self.bvb_inputting_concurrency = false;
                    self.navigate_to_view_mode_select();
                }
                Err(_) => {
                    self.error_msg = "Must be a positive integer (minimum 1)".to_string();
                }
            },
            _ => {
                if let Some(c) = key.as_char() {
                    if c.is_ascii_digit() {
                        self.bvb_custom_concurrency.push(c);
                    }
                }
            }
        }
    }

    // ---- Start BvB session ----

    fn start_bvb_session(&mut self) {
        let white_diff: bot::Difficulty = self.bvb_white_diff.into();
        let black_diff: bot::Difficulty = self.bvb_black_diff.into();
        let white_name = format!("{} Bot", self.bvb_white_diff.name());
        let black_name = format!("{} Bot", self.bvb_black_diff.name());

        let manager = SessionManager::new(
            white_diff,
            black_diff,
            white_name,
            black_name,
            self.bvb_game_count,
            self.bvb_concurrency,
        );

        if let Err(e) = manager.start() {
            self.error_msg = format!("Failed to start bot session: {}", e);
            self.screen = Screen::BvBGameMode;
            self.bvb_inputting_count = false;
            return;
        }

        self.bvb_manager = Some(Arc::new(manager));
        self.bvb_speed = PlaybackSpeed::Normal;
        self.bvb_selected_game = 0;
        self.bvb_paused = false;
        self.bvb_recent_completions.clear();
        self.screen = Screen::BvBGamePlay;
        self.status_msg.clear();
        self.error_msg.clear();
        self.spawn_bvb_tick();
    }

    // ---- BvB gameplay ----

    pub(crate) fn handle_bvb_gameplay_keys(&mut self, key: Key) {
        if self.bvb_show_abort_confirm {
            self.handle_bvb_abort_confirm_keys(key);
            return;
        }
        if self.bvb_show_jump_prompt {
            self.handle_bvb_jump_input(key);
            return;
        }

        match key.token().as_str() {
            "g" | "G" if self.bvb_manager.is_some() && self.bvb_game_count > 1 => {
                self.bvb_show_jump_prompt = true;
                self.bvb_jump_input.clear();
                self.error_msg.clear();
            }
            "esc" => {
                let all_finished = self
                    .bvb_manager
                    .as_ref()
                    .map(|m| m.all_finished())
                    .unwrap_or(true);
                if self.bvb_manager.is_some() && !all_finished {
                    self.bvb_show_abort_confirm = true;
                    self.bvb_abort_selection = 0;
                    return;
                }
                self.screen = Screen::BvBStats;
                self.bvb_stats_selection = 0;
                self.bvb_stats_results_page = 0;
                self.menu_options = vec!["New Session".into(), "Return to Menu".into()];
            }
            " " => {
                if let Some(mgr) = &self.bvb_manager {
                    if self.bvb_paused {
                        mgr.resume();
                        self.bvb_paused = false;
                    } else {
                        mgr.pause();
                        self.bvb_paused = true;
                    }
                }
            }
            "t" | "T" => {
                self.bvb_speed = if self.bvb_speed == PlaybackSpeed::Normal {
                    PlaybackSpeed::Instant
                } else {
                    PlaybackSpeed::Normal
                };
                if let Some(mgr) = &self.bvb_manager {
                    mgr.set_speed(self.bvb_speed);
                }
            }
            "tab" | "v" | "V" => {
                self.bvb_view_mode = match self.bvb_view_mode {
                    BvBViewMode::Grid => BvBViewMode::Single,
                    BvBViewMode::Single => BvBViewMode::StatsOnly,
                    BvBViewMode::StatsOnly => BvBViewMode::Grid,
                };
            }
            "left" | "h" => {
                if let Some(mgr) = &self.bvb_manager {
                    if self.bvb_view_mode == BvBViewMode::Single {
                        let len = mgr.sessions().len();
                        if self.bvb_selected_game > 0 {
                            self.bvb_selected_game -= 1;
                        } else if len > 0 {
                            self.bvb_selected_game = len - 1;
                        }
                    } else if self.bvb_page_index > 0 {
                        self.bvb_page_index -= 1;
                    }
                }
            }
            "right" | "l" => {
                if let Some(mgr) = &self.bvb_manager {
                    let sessions_len = mgr.sessions().len();
                    if self.bvb_view_mode == BvBViewMode::Single {
                        if self.bvb_selected_game + 1 < sessions_len {
                            self.bvb_selected_game += 1;
                        } else {
                            self.bvb_selected_game = 0;
                        }
                    } else {
                        let boards_per_page =
                            (self.bvb_grid_rows * self.bvb_grid_cols).max(1) as usize;
                        let total_pages = sessions_len.div_ceil(boards_per_page);
                        if self.bvb_page_index + 1 < total_pages {
                            self.bvb_page_index += 1;
                        }
                    }
                }
            }
            "f" => self.handle_bvb_export_fen(),
            _ => {}
        }
    }

    fn handle_bvb_export_fen(&mut self) {
        let mgr = match &self.bvb_manager {
            Some(m) => m.clone(),
            None => return,
        };
        let sessions = mgr.sessions();
        let target = if self.bvb_view_mode == BvBViewMode::Single {
            sessions.get(self.bvb_selected_game).cloned()
        } else {
            let boards_per_page = (self.bvb_grid_rows * self.bvb_grid_cols).max(1) as usize;
            let start_idx = self.bvb_page_index * boards_per_page;
            sessions.get(start_idx).cloned()
        };
        if let Some(session) = target {
            let fen = session.current_board().to_fen();
            match util::copy_to_clipboard(&fen) {
                Ok(()) => self.status_msg = "FEN copied to clipboard".to_string(),
                Err(e) => self.status_msg = format!("FEN: {} (Failed to copy: {})", fen, e),
            }
        }
    }

    fn handle_bvb_jump_input(&mut self, key: Key) {
        match key {
            Key::Esc => {
                self.bvb_show_jump_prompt = false;
                self.bvb_jump_input.clear();
                self.error_msg.clear();
            }
            Key::Backspace => {
                self.bvb_jump_input.pop();
                self.error_msg.clear();
            }
            Key::Enter => self.handle_bvb_jump_submit(),
            _ => {
                if let Some(c) = key.as_char() {
                    if c.is_ascii_digit() {
                        self.bvb_jump_input.push(c);
                    }
                    self.error_msg.clear();
                }
            }
        }
    }

    fn handle_bvb_jump_submit(&mut self) {
        if self.bvb_jump_input.is_empty() {
            self.error_msg = format!("Enter a game number (1-{})", self.bvb_game_count);
            return;
        }
        match parse_positive_int(&self.bvb_jump_input) {
            Ok(game_num) if game_num >= 1 && game_num <= self.bvb_game_count => {
                self.bvb_selected_game = (game_num - 1) as usize;
                self.bvb_show_jump_prompt = false;
                self.bvb_jump_input.clear();
                self.error_msg.clear();
                if self.bvb_view_mode == BvBViewMode::Grid {
                    let boards_per_page = (self.bvb_grid_rows * self.bvb_grid_cols).max(1) as usize;
                    self.bvb_page_index = self.bvb_selected_game / boards_per_page;
                }
            }
            _ => {
                self.error_msg = format!("Invalid game number. Enter 1-{}", self.bvb_game_count);
                self.bvb_show_jump_prompt = false;
                self.bvb_jump_input.clear();
            }
        }
    }

    fn handle_bvb_abort_confirm_keys(&mut self, key: Key) {
        match key.token().as_str() {
            "up" | "down" | "k" | "j" => {
                self.bvb_abort_selection = 1 - self.bvb_abort_selection;
            }
            "enter" => {
                if self.bvb_abort_selection == 1 {
                    if let Some(mgr) = self.bvb_manager.take() {
                        mgr.stop();
                    }
                    self.bvb_show_abort_confirm = false;
                    self.screen = Screen::MainMenu;
                    self.menu_options = build_main_menu_options();
                    self.menu_selection = 0;
                    self.nav_stack.clear();
                } else {
                    self.bvb_show_abort_confirm = false;
                }
            }
            "esc" => {
                self.bvb_show_abort_confirm = false;
            }
            _ => {}
        }
    }

    // ---- BvB stats ----

    pub(crate) fn handle_bvb_stats_keys(&mut self, key: Key) {
        match key.token().as_str() {
            "up" | "k" if self.bvb_stats_selection > 0 => {
                self.bvb_stats_selection -= 1;
            }
            "down" | "j" if self.bvb_stats_selection + 1 < self.menu_options.len() => {
                self.bvb_stats_selection += 1;
            }
            "left" | "h" if self.bvb_stats_results_page > 0 => {
                self.bvb_stats_results_page -= 1;
            }
            "right" | "l" => {
                if let Some(mgr) = &self.bvb_manager {
                    let stats = mgr.stats();
                    if stats.total_games > 1 {
                        let total_pages = stats.individual_results.len().div_ceil(15);
                        if self.bvb_stats_results_page + 1 < total_pages {
                            self.bvb_stats_results_page += 1;
                        }
                    }
                }
            }
            "s" | "S" => self.handle_bvb_stats_export(),
            "enter" => self.handle_bvb_stats_selection(),
            "esc" => {
                self.screen = Screen::MainMenu;
                self.menu_options = build_main_menu_options();
                self.menu_selection = 0;
                self.bvb_manager = None;
            }
            _ => {}
        }
    }

    fn handle_bvb_stats_selection(&mut self) {
        match self.bvb_stats_selection {
            0 => {
                self.screen = Screen::BvBBotSelect;
                self.menu_options = vec!["Easy".into(), "Medium".into(), "Hard".into()];
                self.menu_selection = 0;
                self.bvb_selecting_white = true;
                self.bvb_manager = None;
            }
            1 => {
                self.screen = Screen::MainMenu;
                self.menu_options = build_main_menu_options();
                self.menu_selection = 0;
                self.bvb_manager = None;
            }
            _ => {}
        }
    }

    fn handle_bvb_stats_export(&mut self) {
        let mgr = match &self.bvb_manager {
            Some(m) => m.clone(),
            None => {
                self.error_msg = "No session data to export".to_string();
                return;
            }
        };
        let white_name = self.bvb_white_diff.name();
        let black_name = self.bvb_black_diff.name();
        let export = mgr.export_stats(white_name, black_name);
        match bvb::save_session_export(&export, "") {
            Ok(path) => {
                self.status_msg = format!("Stats exported to: {}", path.display());
                self.error_msg.clear();
            }
            Err(e) => {
                self.error_msg = format!("Failed to export: {}", e);
                self.status_msg.clear();
            }
        }
    }
}
