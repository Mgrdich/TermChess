//! Screen rendering (port of Go `view.go`).
//!
//! Each screen builds a `Vec<Line>` of styled spans (mirroring the Go
//! string-builder views) and is drawn with a `Paragraph`.

use std::sync::Arc;
use std::time::Duration;

use bvb::GameSession;
use engine::{Board, Color, GameStatus, Move, PieceType, Square};
use ratatui::layout::Rect;
use ratatui::style::{Color as TColor, Modifier, Style};
use ratatui::text::{Line, Span, Text};
use ratatui::widgets::Paragraph;
use ratatui::Frame;

use crate::app::{App, MIN_TERMINAL_HEIGHT, MIN_TERMINAL_WIDTH};
use crate::board::BoardRenderer;
use crate::san::format_move_history_san;
use crate::state::{BvBViewMode, Screen};
use crate::theme::theme_display_name;

/// Number of header lines rendered before the board on the gameplay screen:
/// the "TermChess" title line and a single blank line (see `render_gameplay`).
/// The board's first rank (rank 8) is drawn at this row offset from the screen top.
const GAMEPLAY_HEADER_LINES: u16 = 2;

impl App {
    // ---- style helpers ----

    /// Divider line used between grouped menu/settings sections
    /// (Go `renderMenuSeparator`, view.go:156-159).
    fn settings_separator_line(&self) -> Line<'static> {
        Line::styled(
            "  ────────────────".to_string(),
            Style::default().fg(self.theme.menu_separator),
        )
    }

    fn title_style(&self) -> Style {
        Style::default()
            .fg(self.theme.title_text)
            .add_modifier(Modifier::BOLD)
    }
    fn help_style(&self) -> Style {
        Style::default().fg(self.theme.help_text)
    }
    fn error_style(&self) -> Style {
        Style::default()
            .fg(self.theme.error_text)
            .add_modifier(Modifier::BOLD)
    }
    fn status_style(&self) -> Style {
        Style::default().fg(self.theme.status_text)
    }
    fn cursor_style(&self) -> Style {
        Style::default()
            .fg(self.theme.menu_selected)
            .add_modifier(Modifier::BOLD)
    }
    fn menu_primary_style(&self) -> Style {
        Style::default()
            .fg(self.theme.menu_primary)
            .add_modifier(Modifier::BOLD)
    }
    fn menu_secondary_style(&self) -> Style {
        Style::default().fg(self.theme.menu_secondary)
    }
    fn selected_style(&self) -> Style {
        Style::default()
            .fg(self.theme.menu_selected)
            .add_modifier(Modifier::BOLD)
    }
    fn menu_normal_style(&self) -> Style {
        Style::default().fg(self.theme.menu_normal)
    }

    fn help_text_line(&self, text: &str) -> Option<Line<'static>> {
        if !self.config.show_help_text {
            return None;
        }
        Some(Line::styled(text.to_string(), self.help_style()))
    }

    fn error_status_lines(&self, lines: &mut Vec<Line<'static>>) {
        if !self.error_msg.is_empty() {
            lines.push(Line::from(""));
            lines.push(Line::styled(
                format!("Error: {}", self.error_msg),
                self.error_style(),
            ));
        }
        if !self.status_msg.is_empty() {
            lines.push(Line::from(""));
            lines.push(Line::styled(self.status_msg.clone(), self.status_style()));
        }
    }

    // ---- top-level draw ----

    /// Renders the current screen (Go `Model.View`).
    pub fn draw(&self, frame: &mut Frame) {
        let area = frame.area();

        // Reset the recorded board origin each frame; only the gameplay screen
        // (below) records a live value. This keeps stale coordinates from a prior
        // screen out of the mouse mapping.
        self.board_origin.set(None);

        if self.term_width > 0
            && self.term_height > 0
            && (self.term_width < MIN_TERMINAL_WIDTH || self.term_height < MIN_TERMINAL_HEIGHT)
        {
            self.render_paragraph(frame, area, self.render_min_size_warning());
            return;
        }

        if self.show_shortcuts_overlay {
            self.render_paragraph(frame, area, self.render_shortcuts_overlay());
            return;
        }

        let lines = match self.screen {
            Screen::MainMenu => self.render_main_menu(),
            Screen::GameTypeSelect => self.render_menu_screen(
                "Select Game Type:",
                "ESC: back to menu | arrows/jk: navigate | enter: select",
                true,
            ),
            Screen::BotSelect => self.render_menu_screen(
                "Select Bot Difficulty:",
                "ESC: back to game type | arrows/jk: navigate | enter: select",
                true,
            ),
            Screen::ColorSelect => self.render_menu_screen(
                "Select Your Color:",
                "ESC: back to difficulty | arrows/jk: navigate | enter: select",
                true,
            ),
            Screen::FenInput => self.render_fen_input(),
            Screen::GamePlay => {
                // Record where the board's top-left piece cell (a8) actually lands
                // so mouse clicks map relative to the real render, not a hardcoded
                // lipgloss-derived constant. The paragraph is drawn at `area`; the
                // gameplay view emits GAMEPLAY_HEADER_LINES header lines (title +
                // blank) before the board, and each rank line is prefixed by a
                // 2-char rank label when coordinates are shown.
                if self.board.is_some() {
                    let coords_offset: u16 = if self.config.show_coords { 2 } else { 0 };
                    self.board_origin.set(Some((
                        area.x + coords_offset,
                        area.y + GAMEPLAY_HEADER_LINES,
                    )));
                }
                self.render_gameplay()
            }
            Screen::GameOver => self.render_game_over(),
            Screen::Settings => self.render_settings(),
            Screen::SavePrompt => self.render_save_prompt(),
            Screen::DrawPrompt => self.render_draw_prompt(),
            Screen::BvBBotSelect => self.render_bvb_bot_select(),
            Screen::BvBGameMode => self.render_bvb_game_mode(),
            Screen::BvBGridConfig => self.render_bvb_grid_config(),
            Screen::BvBGamePlay => self.render_bvb_gameplay(),
            Screen::BvBStats => self.render_bvb_stats(),
            Screen::BvBViewModeSelect => self.render_bvb_view_mode_select(),
            Screen::BvBConcurrencySelect => self.render_bvb_concurrency_select(),
        };

        self.render_paragraph(frame, area, lines);
    }

    fn render_paragraph(&self, frame: &mut Frame, area: Rect, lines: Vec<Line<'static>>) {
        let para = Paragraph::new(Text::from(lines));
        frame.render_widget(para, area);
    }

    fn title_line(&self, text: &str) -> Line<'static> {
        Line::styled(text.to_string(), self.title_style())
    }

    fn render_min_size_warning(&self) -> Vec<Line<'static>> {
        let mut lines = Vec::new();
        lines.push(Line::styled(
            "Terminal too small".to_string(),
            self.error_style(),
        ));
        lines.push(Line::from(""));
        lines.push(Line::styled(
            format!("Current: {}x{}", self.term_width, self.term_height),
            self.help_style(),
        ));
        lines.push(Line::styled(
            format!("Minimum: {}x{}", MIN_TERMINAL_WIDTH, MIN_TERMINAL_HEIGHT),
            self.help_style(),
        ));
        lines.push(Line::from(""));
        lines.push(Line::styled(
            "Please resize your terminal.".to_string(),
            self.help_style(),
        ));
        lines
    }

    // ---- menu screens ----

    fn menu_option_line(&self, i: usize, option: &str, primary: bool) -> Line<'static> {
        let selected = i == self.menu_selection;
        let cursor = if selected { ">> " } else { "  " };
        let style = if selected {
            self.selected_style()
        } else if primary {
            self.menu_primary_style()
        } else {
            self.menu_secondary_style()
        };
        Line::from(vec![
            Span::styled(cursor.to_string(), self.cursor_style()),
            Span::styled(option.to_string(), style),
        ])
    }

    fn render_breadcrumb(&self, lines: &mut Vec<Line<'static>>) {
        let bc = self.breadcrumb();
        if !bc.is_empty() {
            lines.push(Line::styled(bc, self.help_style()));
            lines.push(Line::from(""));
        }
    }

    /// Generic menu screen with title, breadcrumb, header, options, and help.
    fn render_menu_screen(&self, header: &str, help: &str, primary: bool) -> Vec<Line<'static>> {
        let mut lines = Vec::new();
        lines.push(self.title_line("TermChess"));
        self.render_breadcrumb(&mut lines);
        lines.push(Line::styled(header.to_string(), self.title_style()));
        lines.push(Line::from(""));
        for (i, opt) in self.menu_options.iter().enumerate() {
            lines.push(self.menu_option_line(i, opt, primary));
        }
        if let Some(h) = self.help_text_line(help) {
            lines.push(Line::from(""));
            lines.push(h);
        }
        self.error_status_lines(&mut lines);
        lines
    }

    fn render_main_menu(&self) -> Vec<Line<'static>> {
        let mut lines = Vec::new();
        lines.push(self.title_line("TermChess"));
        lines.push(Line::from(""));

        let mut separator_inserted = false;
        for (i, option) in self.menu_options.iter().enumerate() {
            if option == "Settings" && !separator_inserted {
                lines.push(Line::styled(
                    "  ────────────────".to_string(),
                    Style::default().fg(self.theme.menu_separator),
                ));
                separator_inserted = true;
            }
            let selected = i == self.menu_selection;
            let is_resume = option == "Resume Game";
            let is_primary = is_primary_action(option);
            let cursor = if selected {
                if is_resume || is_primary {
                    ">> "
                } else {
                    " > "
                }
            } else {
                "  "
            };
            let style = if is_resume {
                Style::default().fg(self.theme.status_text)
            } else if selected {
                self.selected_style()
            } else if is_primary {
                self.menu_primary_style()
            } else {
                self.menu_secondary_style()
            };
            lines.push(Line::from(vec![
                Span::styled(cursor.to_string(), self.cursor_style()),
                Span::styled(option.to_string(), style),
            ]));
        }

        if let Some(h) = self.help_text_line("arrows/jk: navigate | enter: select | q: quit") {
            lines.push(Line::from(""));
            lines.push(h);
        }
        self.error_status_lines(&mut lines);

        if !self.update_available.is_empty() {
            lines.push(Line::from(""));
            let install_method = updater::detect_install_method();
            let text = if install_method == updater::InstallMethod::GoInstall {
                format!(
                    "Update available: {} (current: {}). Run 'go install github.com/Mgrdich/TermChess/cmd/termchess@latest' to update.",
                    self.update_available, version::VERSION
                )
            } else {
                format!(
                    "Update available: {} (current: {}). Run 'termchess --upgrade' to update.",
                    self.update_available,
                    version::VERSION
                )
            };
            lines.push(Line::styled(
                text,
                Style::default()
                    .fg(TColor::Indexed(208))
                    .add_modifier(Modifier::BOLD),
            ));
        }
        lines
    }

    // ---- FEN input ----

    fn render_fen_input(&self) -> Vec<Line<'static>> {
        let mut lines = Vec::new();
        lines.push(self.title_line("TermChess"));
        self.render_breadcrumb(&mut lines);
        lines.push(Line::styled(
            "Load Game from FEN".to_string(),
            self.title_style(),
        ));
        lines.push(Line::from(""));
        lines.push(Line::from(
            "Enter a FEN string to load a chess position:".to_string(),
        ));
        lines.push(Line::from(""));
        lines.push(Line::styled(
            self.fen_input.display(),
            self.menu_normal_style(),
        ));
        lines.push(Line::from(""));
        lines.push(Line::styled(
            "Example: rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1".to_string(),
            self.help_style(),
        ));
        lines.push(Line::from(""));
        if let Some(h) = self.help_text_line("ESC: back to menu | enter: load position") {
            lines.push(h);
        }
        self.error_status_lines(&mut lines);
        lines
    }

    // ---- Gameplay ----

    fn render_gameplay(&self) -> Vec<Line<'static>> {
        let mut lines = Vec::new();
        lines.push(self.title_line("TermChess"));
        lines.push(Line::from(""));

        if let Some(board) = &self.board {
            let renderer = BoardRenderer::with_theme(&self.config, self.theme);
            let text = renderer.render(
                board,
                self.selected_square,
                &self.valid_moves,
                self.blink_on,
            );
            lines.extend(text.lines);

            // Move history in SAN near the board (Go renderGamePlay:590-598).
            if self.config.show_move_history && !self.move_history.is_empty() {
                lines.push(Line::from(""));
                lines.push(Line::styled(
                    format_move_history_san(&self.move_history),
                    self.menu_normal_style(),
                ));
            }

            lines.push(Line::from(""));
            let (turn_text, turn_style) = if board.active_color == Color::Black {
                (
                    "Black to move",
                    Style::default()
                        .fg(self.theme.black_turn_text)
                        .add_modifier(Modifier::BOLD),
                )
            } else {
                (
                    "White to move",
                    Style::default()
                        .fg(self.theme.white_turn_text)
                        .add_modifier(Modifier::BOLD),
                )
            };
            lines.push(Line::styled(turn_text.to_string(), turn_style));
            lines.push(Line::from(""));
            lines.push(Line::from(vec![
                Span::styled("Enter move: ".to_string(), self.menu_normal_style()),
                Span::styled(self.input.clone(), turn_style),
            ]));
        }

        if let Some(h) = self.help_text_line(
            "ESC: menu (with save) | type move (e.g. e4, Nf3) | Commands: resign, offerdraw, showfen, menu",
        ) {
            lines.push(Line::from(""));
            lines.push(h);
        }
        self.error_status_lines(&mut lines);

        if self.config.show_move_history && !self.move_history.is_empty() {
            lines.push(Line::from(""));
            lines.push(Line::styled(
                "Move History:".to_string(),
                self.title_style(),
            ));
            lines.push(Line::styled(
                crate::san::format_move_history(&self.move_history),
                self.selected_style(),
            ));
        }
        lines
    }

    // ---- Game over ----

    fn render_game_over(&self) -> Vec<Line<'static>> {
        let mut lines = Vec::new();
        lines.push(self.title_line("TermChess"));
        lines.push(Line::from(""));

        if let Some(board) = &self.board {
            let msg = game_result_message(board, self.resigned_by, self.draw_by_agreement);
            lines.push(Line::styled(
                msg,
                Style::default()
                    .fg(TColor::Rgb(0xFF, 0xD7, 0x00))
                    .add_modifier(Modifier::BOLD),
            ));
            lines.push(Line::from(""));

            let renderer = BoardRenderer::new(&self.config);
            lines.extend(renderer.render(board, None, &[], false).lines);
            lines.push(Line::from(""));
            lines.push(Line::styled(
                format!("Game ended after {} moves", board.full_move_num),
                self.menu_normal_style(),
            ));
        }

        lines.push(Line::from(""));
        lines.push(Line::styled(
            "Press 'n' for New Game  |  Press 'm' for Main Menu  |  Press 'q' to Quit".to_string(),
            self.selected_style(),
        ));
        if let Some(h) = self.help_text_line("ESC/m: menu | n: new game | q: quit") {
            lines.push(Line::from(""));
            lines.push(h);
        }
        lines
    }

    // ---- Settings ----

    fn render_settings(&self) -> Vec<Line<'static>> {
        let mut lines = Vec::new();
        lines.push(self.title_line("TermChess"));
        self.render_breadcrumb(&mut lines);
        lines.push(Line::styled("Settings".to_string(), self.title_style()));
        lines.push(Line::from(""));

        let toggles = [
            ("Use Unicode Pieces", self.config.use_unicode),
            ("Show Coordinates", self.config.show_coords),
            ("Use Colors", self.config.use_colors),
            ("Show Move History", self.config.show_move_history),
            ("Show Help Text", self.config.show_help_text),
        ];
        for (i, (label, enabled)) in toggles.iter().enumerate() {
            let checkbox = if *enabled { "[X]" } else { "[ ]" };
            let text = format!("{} {}", label, checkbox);
            let selected = i == self.settings_selection;
            let cursor = if selected { ">> " } else { "  " };
            let style = if selected {
                self.selected_style()
            } else {
                self.menu_normal_style()
            };
            lines.push(Line::from(vec![
                Span::styled(cursor.to_string(), self.cursor_style()),
                Span::styled(text, style),
            ]));
            // Group separators mirror Go renderSettings (view.go:806-811, 836-838):
            // Display group (0-2) | Info group (3-4) | Theme (5).
            if i == 2 || i == 4 {
                lines.push(self.settings_separator_line());
            }
        }

        // Theme (index 5)
        let theme_text = format!("Theme: {}", theme_display_name(&self.config.theme));
        let selected = self.settings_selection == 5;
        let cursor = if selected { ">> " } else { "  " };
        let style = if selected {
            self.selected_style()
        } else {
            self.menu_normal_style()
        };
        lines.push(Line::from(vec![
            Span::styled(cursor.to_string(), self.cursor_style()),
            Span::styled(theme_text, style),
        ]));

        if let Some(h) =
            self.help_text_line("ESC: back | arrows/jk: navigate | enter/space: toggle/cycle")
        {
            lines.push(Line::from(""));
            lines.push(h);
        }
        self.error_status_lines(&mut lines);
        lines
    }

    // ---- Save prompt ----

    fn render_save_prompt(&self) -> Vec<Line<'static>> {
        let mut lines = Vec::new();
        lines.push(self.title_line("TermChess"));
        lines.push(Line::from(""));
        if let Some(board) = &self.board {
            let renderer = BoardRenderer::new(&self.config);
            lines.extend(renderer.render(board, None, &[], false).lines);
            lines.push(Line::from(""));
        }
        lines.push(Line::styled(
            "Save current game before exiting?".to_string(),
            Style::default()
                .fg(TColor::Rgb(0xFF, 0xD7, 0x00))
                .add_modifier(Modifier::BOLD),
        ));
        lines.push(Line::from(""));
        lines.push(Line::styled(
            "y: Save & Exit  |  n: Exit without saving  |  ESC: Cancel".to_string(),
            self.selected_style(),
        ));
        if let Some(h) =
            self.help_text_line("y: save & exit | n: exit without saving | ESC: cancel")
        {
            lines.push(Line::from(""));
            lines.push(h);
        }
        if !self.error_msg.is_empty() {
            lines.push(Line::from(""));
            lines.push(Line::styled(
                format!("Error: {}", self.error_msg),
                self.error_style(),
            ));
        }
        lines
    }

    // ---- Draw prompt ----

    fn render_draw_prompt(&self) -> Vec<Line<'static>> {
        let mut lines = vec![
            self.title_line("TermChess"),
            Line::from(""),
            Line::styled(
                "Draw Offer".to_string(),
                Style::default()
                    .fg(TColor::Rgb(0xFF, 0xD7, 0x00))
                    .add_modifier(Modifier::BOLD),
            ),
            Line::from(""),
        ];
        let offer = if self.draw_offered_by == Some(Color::Black) {
            "Black offers a draw. Accept?"
        } else {
            "White offers a draw. Accept?"
        };
        lines.push(Line::styled(offer.to_string(), self.menu_normal_style()));
        lines.push(Line::from(""));
        for (i, option) in ["Accept", "Decline"].iter().enumerate() {
            let selected = i == self.draw_prompt_selection;
            let cursor = if selected { ">> " } else { "  " };
            let style = if selected {
                self.selected_style()
            } else {
                self.menu_primary_style()
            };
            lines.push(Line::from(vec![
                Span::styled(cursor.to_string(), self.cursor_style()),
                Span::styled(option.to_string(), style),
            ]));
        }
        lines.push(Line::from(""));
        lines.push(Line::styled(
            "Use arrow keys to select, Enter to confirm, ESC to cancel".to_string(),
            self.help_style(),
        ));
        self.error_status_lines(&mut lines);
        lines
    }

    // ---- BvB bot select ----

    fn render_bvb_bot_select(&self) -> Vec<Line<'static>> {
        let mut lines = Vec::new();
        lines.push(self.title_line("TermChess"));
        lines.push(Line::from(""));
        let header = if self.bvb_selecting_white {
            "Select White Bot Difficulty:"
        } else {
            "Select Black Bot Difficulty:"
        };
        lines.push(Line::styled(header.to_string(), self.title_style()));
        lines.push(Line::from(""));
        for (i, opt) in self.menu_options.iter().enumerate() {
            lines.push(self.menu_option_line(i, opt, true));
        }
        if !self.bvb_selecting_white {
            lines.push(Line::from(""));
            lines.push(Line::styled(
                format!("White: {} Bot", self.bvb_white_diff.name()),
                self.status_style(),
            ));
        }
        if let Some(h) = self.help_text_line("ESC: back | arrows/jk: navigate | enter: select") {
            lines.push(Line::from(""));
            lines.push(h);
        }
        self.error_status_lines(&mut lines);
        lines
    }

    // ---- BvB game mode ----

    fn render_bvb_game_mode(&self) -> Vec<Line<'static>> {
        let mut lines = Vec::new();
        lines.push(self.title_line("TermChess"));
        lines.push(Line::from(""));
        lines.push(Line::styled(
            "Select Game Mode:".to_string(),
            self.title_style(),
        ));
        lines.push(Line::from(""));
        lines.push(Line::styled(
            format!(
                "{} Bot (White) vs {} Bot (Black)",
                self.bvb_white_diff.name(),
                self.bvb_black_diff.name()
            ),
            self.status_style(),
        ));
        lines.push(Line::from(""));

        if self.bvb_inputting_count {
            lines.push(Line::styled(
                "Number of games:".to_string(),
                self.menu_normal_style(),
            ));
            lines.push(Line::from(""));
            let display = if self.bvb_count_input.is_empty() {
                "_".to_string()
            } else {
                self.bvb_count_input.clone()
            };
            lines.push(Line::styled(
                format!(">> {}", display),
                self.selected_style(),
            ));
            if let Some(h) = self.help_text_line("ESC: back | enter: confirm | type number") {
                lines.push(Line::from(""));
                lines.push(h);
            }
        } else {
            for (i, opt) in self.menu_options.iter().enumerate() {
                lines.push(self.menu_option_line(i, opt, true));
            }
            if let Some(h) = self.help_text_line("ESC: back | arrows/jk: navigate | enter: select")
            {
                lines.push(Line::from(""));
                lines.push(h);
            }
        }
        self.error_status_lines(&mut lines);
        lines
    }

    // ---- BvB grid config ----

    fn render_bvb_grid_config(&self) -> Vec<Line<'static>> {
        let mut lines = Vec::new();
        lines.push(self.title_line("TermChess"));
        lines.push(Line::from(""));
        lines.push(Line::styled(
            "Select Grid Layout:".to_string(),
            self.title_style(),
        ));
        lines.push(Line::from(""));
        lines.push(Line::styled(
            format!(
                "{} game(s) | {} Bot (White) vs {} Bot (Black)",
                self.bvb_game_count,
                self.bvb_white_diff.name(),
                self.bvb_black_diff.name()
            ),
            self.status_style(),
        ));
        lines.push(Line::from(""));

        if self.bvb_inputting_grid {
            lines.push(Line::styled(
                "Enter grid dimensions (RxC, max 8 total):".to_string(),
                self.menu_normal_style(),
            ));
            lines.push(Line::from(""));
            let display = if self.bvb_custom_grid_input.is_empty() {
                "_".to_string()
            } else {
                self.bvb_custom_grid_input.clone()
            };
            lines.push(Line::styled(
                format!(">> {}", display),
                self.selected_style(),
            ));
            if let Some(h) = self.help_text_line("ESC: back | enter: confirm | e.g. 2x3") {
                lines.push(Line::from(""));
                lines.push(h);
            }
        } else {
            for (i, opt) in self.menu_options.iter().enumerate() {
                lines.push(self.menu_option_line(i, opt, true));
            }
            if let Some(h) = self.help_text_line("ESC: back | arrows/jk: navigate | enter: select")
            {
                lines.push(Line::from(""));
                lines.push(h);
            }
        }
        if !self.error_msg.is_empty() {
            lines.push(Line::from(""));
            lines.push(Line::styled(
                format!("Error: {}", self.error_msg),
                self.error_style(),
            ));
        }
        lines
    }

    // ---- BvB concurrency select ----

    fn render_bvb_concurrency_select(&self) -> Vec<Line<'static>> {
        let mut lines = Vec::new();
        lines.push(self.title_line("TermChess"));
        lines.push(Line::from(""));
        lines.push(Line::styled(
            "Select Concurrency:".to_string(),
            self.title_style(),
        ));
        lines.push(Line::from(""));
        lines.push(Line::styled(
            format!(
                "{} game(s) | {} Bot (White) vs {} Bot (Black)",
                self.bvb_game_count,
                self.bvb_white_diff.name(),
                self.bvb_black_diff.name()
            ),
            self.status_style(),
        ));
        lines.push(Line::from(""));

        let recommended = bvb::calculate_default_concurrency();
        let num_cpu = std::thread::available_parallelism()
            .map(|n| n.get())
            .unwrap_or(1);

        if self.bvb_inputting_concurrency {
            lines.push(Line::from(vec![
                Span::styled("Enter concurrency: ".to_string(), self.title_style()),
                Span::styled(
                    format!("{}_", self.bvb_custom_concurrency),
                    self.selected_style(),
                ),
            ]));
            if parse_concurrency_value(&self.bvb_custom_concurrency) > 50 {
                lines.push(Line::from(""));
                lines.push(Line::styled(
                    "Warning: High concurrency may cause lag. Consider using Stats Only view mode."
                        .to_string(),
                    self.error_style(),
                ));
            }
            if let Some(h) = self.help_text_line("enter: confirm | esc: cancel") {
                lines.push(Line::from(""));
                lines.push(h);
            }
        } else {
            let options = [
                (
                    format!("Recommended ({} concurrent games)", recommended),
                    format!("Based on your CPU ({} cores)", num_cpu),
                ),
                (
                    "Custom".to_string(),
                    "Enter your own value (may cause lag)".to_string(),
                ),
            ];
            for (i, (name, desc)) in options.iter().enumerate() {
                let selected = i == self.bvb_concurrency_selection;
                let cursor = if selected { ">> " } else { "  " };
                let style = if selected {
                    self.selected_style()
                } else {
                    self.menu_primary_style()
                };
                lines.push(Line::from(vec![
                    Span::styled(cursor.to_string(), self.cursor_style()),
                    Span::styled(name.clone(), style),
                ]));
                lines.push(Line::styled(format!("    {}", desc), self.help_style()));
            }
            if let Some(h) = self.help_text_line("arrows/jk: navigate | enter: select | esc: back")
            {
                lines.push(Line::from(""));
                lines.push(h);
            }
        }
        if !self.error_msg.is_empty() {
            lines.push(Line::from(""));
            lines.push(Line::styled(
                format!("Error: {}", self.error_msg),
                self.error_style(),
            ));
        }
        lines
    }

    // ---- BvB view mode select ----

    fn render_bvb_view_mode_select(&self) -> Vec<Line<'static>> {
        let mut lines = Vec::new();
        lines.push(self.title_line("TermChess"));
        lines.push(Line::from(""));
        lines.push(Line::styled(
            "Select View Mode:".to_string(),
            self.title_style(),
        ));
        lines.push(Line::from(""));
        lines.push(Line::styled(
            format!(
                "{} game(s) | {} Bot (White) vs {} Bot (Black) | Grid: {}x{}",
                self.bvb_game_count,
                self.bvb_white_diff.name(),
                self.bvb_black_diff.name(),
                self.bvb_grid_rows,
                self.bvb_grid_cols
            ),
            self.status_style(),
        ));
        lines.push(Line::from(""));

        let options = [
            ("Grid View", "Watch multiple games in a grid layout", ""),
            ("Single Board", "Focus on one game at a time", ""),
            (
                "Stats Only",
                "No boards, just statistics",
                "(Recommended for 50+ games)",
            ),
        ];
        for (i, (name, desc, hint)) in options.iter().enumerate() {
            let selected = i == self.bvb_view_mode_selection;
            let cursor = if selected { ">> " } else { "  " };
            let style = if selected {
                self.selected_style()
            } else {
                self.menu_primary_style()
            };
            lines.push(Line::from(vec![
                Span::styled(cursor.to_string(), self.cursor_style()),
                Span::styled(name.to_string(), style),
            ]));
            let desc_text = if hint.is_empty() {
                format!("    {}", desc)
            } else {
                format!("    {} {}", desc, hint)
            };
            lines.push(Line::styled(desc_text, self.help_style()));
        }
        if let Some(h) = self.help_text_line("ESC: back | arrows/jk: navigate | enter: select") {
            lines.push(Line::from(""));
            lines.push(h);
        }
        if !self.error_msg.is_empty() {
            lines.push(Line::from(""));
            lines.push(Line::styled(
                format!("Error: {}", self.error_msg),
                self.error_style(),
            ));
        }
        lines
    }

    // ---- BvB gameplay ----

    fn render_bvb_gameplay(&self) -> Vec<Line<'static>> {
        if self.bvb_manager.is_none() {
            return vec![Line::from("No session running.".to_string())];
        }
        if self.bvb_show_abort_confirm {
            return self.render_bvb_abort_confirm();
        }
        match self.bvb_view_mode {
            BvBViewMode::Single => self.render_bvb_single_view(),
            BvBViewMode::StatsOnly => self.render_bvb_stats_only(),
            BvBViewMode::Grid => self.render_bvb_grid_view(),
        }
    }

    fn render_bvb_abort_confirm(&self) -> Vec<Line<'static>> {
        let mut lines = vec![
            self.title_line("TermChess - Bot vs Bot"),
            Line::from(""),
            Line::styled("Abort Session?".to_string(), self.title_style()),
            Line::from(""),
            Line::from("Games in progress will be lost.".to_string()),
            Line::from(""),
        ];
        for (i, opt) in ["Cancel", "Abort Session"].iter().enumerate() {
            let selected = i == self.bvb_abort_selection;
            let cursor = if selected { "  > " } else { "    " };
            let style = if selected {
                self.selected_style()
            } else {
                self.menu_normal_style()
            };
            lines.push(Line::styled(format!("{}{}", cursor, opt), style));
        }
        lines.push(Line::from(""));
        lines.push(Line::styled(
            "esc: cancel | enter: select".to_string(),
            self.help_style(),
        ));
        lines
    }

    fn render_bvb_grid_view(&self) -> Vec<Line<'static>> {
        let mut lines = Vec::new();
        lines.push(self.title_line("TermChess - Bot vs Bot"));
        lines.push(Line::from(""));

        let cols = self.bvb_grid_cols.max(1) as usize;
        let rows = self.bvb_grid_rows.max(1) as usize;

        // Terminal-too-small check (Go view.go:1231-1244): each cell needs ~14
        // width and ~11 height, plus 8 lines for header/footer.
        let min_width = cols * 14;
        let min_height = rows * 11 + 8;
        if self.term_width > 0
            && self.term_height > 0
            && ((self.term_width as usize) < min_width || (self.term_height as usize) < min_height)
        {
            lines.push(Line::styled(
                format!(
                    "Terminal too small for {}x{} grid (need {}x{}, have {}x{})",
                    rows, cols, min_width, min_height, self.term_width, self.term_height
                ),
                self.error_style(),
            ));
            lines.push(Line::styled(
                "Press Tab to switch to single-board view".to_string(),
                self.error_style(),
            ));
            return lines;
        }

        let mgr = self.bvb_manager.as_ref().unwrap();
        let sessions = mgr.sessions();
        if sessions.is_empty() {
            lines.push(Line::from("No games available.".to_string()));
            return lines;
        }

        let boards_per_page = (rows * cols).max(1);
        let total_pages = sessions.len().div_ceil(boards_per_page);
        let page_idx = self.bvb_page_index.min(total_pages.saturating_sub(1));
        let start_idx = page_idx * boards_per_page;
        let end_idx = (start_idx + boards_per_page).min(sessions.len());

        let finished = sessions.iter().filter(|s| s.is_finished()).count();
        lines.push(Line::styled(
            format!(
                "{} Bot (White) vs {} Bot (Black) | Completed: {}/{} | Running: {} | Queued: {} | Concurrency: {}",
                self.bvb_white_diff.name(),
                self.bvb_black_diff.name(),
                finished,
                sessions.len(),
                mgr.running_count(),
                mgr.queued_count(),
                mgr.concurrency()
            ),
            self.status_style(),
        ));
        lines.push(Line::from(""));

        self.append_live_stats(&mut lines);

        // Render the visible boards as a fixed-size R×C grid so the layout does
        // not shift as games finish (Go renderBoardGrid, view.go:1364-1390).
        let cells: Vec<(Vec<String>, bool)> = sessions[start_idx..end_idx]
            .iter()
            .map(|session| self.build_compact_cell(session))
            .collect();
        let cell_height = crate::app::BVB_CELL_HEIGHT;
        for row_cells in cells.chunks(cols) {
            for row in 0..cell_height {
                let mut spans: Vec<Span<'static>> = Vec::new();
                for (cell, finished) in row_cells {
                    // Finished games get a dimmed foreground (Go view.go:1482-1485).
                    let style = if *finished {
                        Style::default().fg(self.theme.help_text)
                    } else {
                        Style::default()
                    };
                    // Mirror Go's per-cell horizontal Margin(0,1).
                    spans.push(Span::styled(format!(" {} ", cell[row]), style));
                }
                lines.push(Line::from(spans));
            }
        }

        if total_pages > 1 {
            lines.push(Line::styled(
                format!("Page {}/{}", page_idx + 1, total_pages),
                self.selected_style(),
            ));
        }
        lines.push(Line::styled(self.speed_status(), self.menu_normal_style()));

        if self.bvb_show_jump_prompt {
            self.append_jump_prompt(&mut lines);
        }
        if !self.error_msg.is_empty() {
            lines.push(Line::styled(
                format!("Error: {}", self.error_msg),
                self.error_style(),
            ));
        }
        if let Some(h) = self.help_text_line(
            "Space: pause/resume | t: toggle speed | ←/→: pages | g: jump to game | Tab: single view | f: FEN | ESC: abort",
        ) {
            lines.push(h);
        }
        lines
    }

    fn render_bvb_single_view(&self) -> Vec<Line<'static>> {
        let mut lines = Vec::new();
        lines.push(self.title_line("TermChess - Bot vs Bot"));
        lines.push(Line::from(""));

        let mgr = self.bvb_manager.as_ref().unwrap();
        let sessions = mgr.sessions();
        if sessions.is_empty() {
            lines.push(Line::from("No games available.".to_string()));
            return lines;
        }
        let selected_idx = self.bvb_selected_game.min(sessions.len() - 1);
        let session = &sessions[selected_idx];

        lines.push(Line::styled(
            format!(
                "{} Bot (White) vs {} Bot (Black)",
                self.bvb_white_diff.name(),
                self.bvb_black_diff.name()
            ),
            self.status_style(),
        ));
        if sessions.len() > 1 {
            lines.push(Line::styled(
                format!(">>> Game {} of {} <<<", selected_idx + 1, sessions.len()),
                self.selected_style(),
            ));
        }
        lines.push(Line::from(""));

        self.append_live_stats(&mut lines);

        let board = session.current_board();
        let renderer = BoardRenderer::new(&self.config);
        lines.extend(renderer.render(&board, None, &[], false).lines);
        lines.push(Line::from(""));

        let moves = session.current_move_history();
        let mut status_line = format!("Moves: {}", moves.len());
        if session.is_finished() {
            if let Some(result) = session.result() {
                status_line += &format!(" | Result: {} ({})", result.winner, result.end_reason);
            }
        } else if board.active_color == Color::White {
            status_line += " | White to move";
        } else {
            status_line += " | Black to move";
        }
        lines.push(Line::styled(status_line, self.selected_style()));
        lines.push(Line::styled(self.speed_status(), self.menu_normal_style()));

        if self.config.show_move_history && !moves.is_empty() {
            lines.push(Line::from(""));
            lines.push(Line::styled(
                "Move History:".to_string(),
                self.title_style(),
            ));
            lines.push(Line::styled(
                crate::san::format_move_history(&moves),
                self.selected_style(),
            ));
        }

        if self.bvb_show_jump_prompt {
            self.append_jump_prompt(&mut lines);
        }
        if !self.error_msg.is_empty() {
            lines.push(Line::styled(
                format!("Error: {}", self.error_msg),
                self.error_style(),
            ));
        }
        let mut help = String::from("Space: pause/resume | t: toggle speed | ");
        if self.bvb_game_count > 1 {
            help += "left/right: games | g: jump to game | ";
        }
        help += "Tab: view | f: FEN | ESC: abort";
        if let Some(h) = self.help_text_line(&help) {
            lines.push(h);
        }
        lines
    }

    fn render_bvb_stats_only(&self) -> Vec<Line<'static>> {
        let mut lines = Vec::new();
        lines.push(self.title_line("TermChess - Bot vs Bot (Stats Only)"));
        lines.push(Line::from(""));

        let mgr = match &self.bvb_manager {
            Some(m) => m,
            None => {
                lines.push(Line::from("No session running.".to_string()));
                return lines;
            }
        };
        let sessions = mgr.sessions();
        let total_games = sessions.len();
        if total_games == 0 {
            lines.push(Line::from("No games available.".to_string()));
            return lines;
        }

        let stats = mgr.stats();
        let completed = sessions.iter().filter(|s| s.is_finished()).count();
        let in_progress = total_games - completed;

        lines.push(Line::styled(
            format!(
                "{} Bot (White) vs {} Bot (Black)",
                self.bvb_white_diff.name(),
                self.bvb_black_diff.name()
            ),
            self.status_style(),
        ));
        lines.push(Line::from(""));
        lines.push(Line::styled(
            render_progress_bar(completed as i32, total_games as i32, 40),
            self.selected_style(),
        ));
        lines.push(Line::from(""));
        lines.push(Line::styled(
            format!(
                "Score:  White: {}  |  Black: {}  |  Draws: {}",
                stats.white_wins, stats.black_wins, stats.draws
            ),
            self.title_style(),
        ));
        lines.push(Line::from(""));
        lines.push(Line::styled(
            format!("Average moves per game: {:.1}", stats.avg_move_count),
            self.menu_normal_style(),
        ));
        lines.push(Line::styled(
            format!("{} game(s) in progress", in_progress),
            self.status_style(),
        ));
        lines.push(Line::from(""));
        lines.push(Line::styled(
            format!(
                "Running: {} | Queued: {} | Concurrency: {}",
                mgr.running_count(),
                mgr.queued_count(),
                mgr.concurrency()
            ),
            self.menu_normal_style(),
        ));
        lines.push(Line::from(""));
        lines.push(Line::styled(
            "Recent Completions:".to_string(),
            self.title_style(),
        ));
        if self.bvb_recent_completions.is_empty() {
            lines.push(Line::styled(
                "    (none yet)".to_string(),
                self.help_style(),
            ));
        } else {
            for entry in &self.bvb_recent_completions {
                lines.push(Line::styled(format!("    {}", entry), self.help_style()));
            }
        }
        lines.push(Line::from(""));
        lines.push(Line::styled(self.speed_status(), self.menu_normal_style()));
        if !self.error_msg.is_empty() {
            lines.push(Line::styled(
                format!("Error: {}", self.error_msg),
                self.error_style(),
            ));
        }
        if let Some(h) =
            self.help_text_line("[Space] Pause/Resume | [v] Change view | [t] Speed | [q/ESC] Quit")
        {
            lines.push(h);
        }
        lines
    }

    // ---- BvB stats screen ----

    fn render_bvb_stats(&self) -> Vec<Line<'static>> {
        let mut lines = Vec::new();
        lines.push(self.title_line("TermChess - Bot vs Bot Results"));
        lines.push(Line::from(""));

        let mgr = match &self.bvb_manager {
            Some(m) => m,
            None => {
                lines.push(Line::from("No session data available.".to_string()));
                return lines;
            }
        };
        let stats = mgr.stats();
        if stats.total_games == 0 {
            lines.push(Line::from("No games completed.".to_string()));
            return lines;
        }

        if stats.total_games == 1 {
            let r = &stats.individual_results[0];
            lines.push(Line::styled(
                format!(
                    "{} (White) vs {} (Black)",
                    stats.white_bot_name, stats.black_bot_name
                ),
                self.status_style(),
            ));
            lines.push(Line::from(""));
            if r.winner == "Draw" {
                lines.push(Line::from(format!("Result: Draw ({})", r.end_reason)));
            } else {
                lines.push(Line::from(format!(
                    "Winner: {} ({})",
                    r.winner, r.end_reason
                )));
            }
            lines.push(Line::from(format!("Total moves: {}", r.move_count)));
            lines.push(Line::from(format!("Duration: {:?}", r.duration)));
        } else {
            lines.push(Line::styled(
                format!(
                    "{} (White) vs {} (Black) — {} games",
                    stats.white_bot_name, stats.black_bot_name, stats.total_games
                ),
                self.status_style(),
            ));
            lines.push(Line::from(""));
            lines.push(Line::from(format!(
                "{} wins: {} ({:.1}%)",
                stats.white_bot_name, stats.white_wins, stats.white_win_pct
            )));
            lines.push(Line::from(format!(
                "{} wins: {} ({:.1}%)",
                stats.black_bot_name, stats.black_wins, stats.black_win_pct
            )));
            lines.push(Line::from(format!("Draws: {}", stats.draws)));
            lines.push(Line::from(""));
            lines.push(Line::from(format!(
                "Avg moves: {:.1} | Avg duration: {:?}",
                stats.avg_move_count, stats.avg_duration
            )));
            lines.push(Line::from(format!(
                "Shortest game: #{} ({} moves) | Longest game: #{} ({} moves)",
                stats.shortest_game.game_number,
                stats.shortest_game.move_count,
                stats.longest_game.game_number,
                stats.longest_game.move_count
            )));
            lines.push(Line::from(""));

            let results_per_page = 15;
            let total = stats.individual_results.len();
            let total_pages = total.div_ceil(results_per_page);
            let current_page = self
                .bvb_stats_results_page
                .min(total_pages.saturating_sub(1));
            let start_idx = current_page * results_per_page;
            let end_idx = (start_idx + results_per_page).min(total);

            lines.push(Line::styled(
                format!(
                    "Individual Results (Page {}/{}):",
                    current_page + 1,
                    total_pages
                ),
                self.help_style(),
            ));
            for r in &stats.individual_results[start_idx..end_idx] {
                let text = if r.winner == "Draw" {
                    format!(
                        "  Game {}: Draw ({}) — {} moves",
                        r.game_number, r.end_reason, r.move_count
                    )
                } else {
                    format!(
                        "  Game {}: {} wins ({}) — {} moves",
                        r.game_number, r.winner, r.end_reason, r.move_count
                    )
                };
                lines.push(Line::styled(text, self.help_style()));
            }
        }

        lines.push(Line::from(""));
        for (i, opt) in self.menu_options.iter().enumerate() {
            let selected = i == self.bvb_stats_selection;
            let cursor = if selected { ">> " } else { "  " };
            let primary = is_primary_action(opt);
            let style = if selected {
                self.selected_style()
            } else if primary {
                self.menu_primary_style()
            } else {
                self.menu_secondary_style()
            };
            lines.push(Line::from(vec![
                Span::styled(cursor.to_string(), self.cursor_style()),
                Span::styled(opt.clone(), style),
            ]));
        }

        if !self.status_msg.is_empty() {
            lines.push(Line::styled(self.status_msg.clone(), self.status_style()));
        }
        if !self.error_msg.is_empty() {
            lines.push(Line::styled(self.error_msg.clone(), self.error_style()));
        }

        let mut help = "up/down: navigate | s: export | Enter: select | ESC: menu".to_string();
        if stats.total_games > 1 {
            let total_pages = stats.individual_results.len().div_ceil(15);
            if total_pages > 1 {
                help =
                    "up/down: navigate | left/right: page | s: export | Enter: select | ESC: menu"
                        .to_string();
            }
        }
        if let Some(h) = self.help_text_line(&help) {
            lines.push(h);
        }
        lines
    }

    // ---- BvB helpers ----

    fn speed_status(&self) -> String {
        let speed = match self.bvb_speed {
            bvb::PlaybackSpeed::Instant => "Instant",
            bvb::PlaybackSpeed::Normal => "Normal",
        };
        let mut s = format!("Speed: {}", speed);
        if self.bvb_paused {
            s += " | PAUSED";
        }
        s
    }

    fn append_jump_prompt(&self, lines: &mut Vec<Line<'static>>) {
        let display = if self.bvb_jump_input.is_empty() {
            "_".to_string()
        } else {
            self.bvb_jump_input.clone()
        };
        lines.push(Line::styled(
            format!("Jump to game (1-{}): {}", self.bvb_game_count, display),
            self.selected_style(),
        ));
        lines.push(Line::styled(
            "Enter: jump | Esc: cancel".to_string(),
            self.help_style(),
        ));
    }

    /// Builds one compact board cell as fixed-size plain-text lines plus a
    /// finished flag, mirroring Go `renderCompactBoardCell` (view.go:1411-1488).
    /// The result is exactly `BVB_CELL_HEIGHT` lines, each `BVB_CELL_WIDTH` wide.
    fn build_compact_cell(&self, session: &Arc<GameSession>) -> (Vec<String>, bool) {
        let board = session.current_board();
        let move_count = session.current_move_history().len();
        let is_finished = session.is_finished();

        let mut cell: Vec<String> = Vec::new();
        // Line 1: game header.
        cell.push(format!("Game {}", session.game_number()));

        // Lines 2-9: compact board (no coords/colors).
        let compact_cfg = config::Config {
            use_unicode: self.config.use_unicode,
            show_coords: false,
            use_colors: false,
            show_move_history: false,
            show_help_text: false,
            theme: self.config.theme.clone(),
        };
        let renderer = BoardRenderer::new(&compact_cfg);
        for line in renderer.render(&board, None, &[], false).lines {
            cell.push(line_plain_text(&line));
        }

        // Status line (always shows move count).
        cell.push(format!("Moves: {}", move_count));

        // Result line (winner for finished, empty placeholder otherwise).
        if is_finished {
            match session.result() {
                Some(result) => cell.push(result.winner),
                None => cell.push(String::new()),
            }
        } else {
            cell.push(String::new());
        }

        // Spacing line.
        cell.push(String::new());

        // Pad or truncate to exactly BVB_CELL_HEIGHT lines.
        while cell.len() < crate::app::BVB_CELL_HEIGHT {
            cell.push(String::new());
        }
        cell.truncate(crate::app::BVB_CELL_HEIGHT);

        // Normalize each line to exactly BVB_CELL_WIDTH display columns.
        for line in cell.iter_mut() {
            *line = fit_display_width(line, crate::app::BVB_CELL_WIDTH);
        }

        (cell, is_finished)
    }

    fn append_live_stats(&self, lines: &mut Vec<Line<'static>>) {
        let mgr = match &self.bvb_manager {
            Some(m) => m,
            None => return,
        };
        let sessions = mgr.sessions();
        if sessions.is_empty() {
            return;
        }
        let stats = mgr.stats();

        lines.push(Line::styled(
            "══════ Statistics ══════".to_string(),
            self.title_style(),
        ));
        lines.push(Line::styled(
            format!(
                "Score: White {} | Black {} | Draws {}",
                stats.white_wins, stats.black_wins, stats.draws
            ),
            self.selected_style(),
        ));
        lines.push(Line::styled(
            format!("Progress: {} / {} games", stats.total_games, sessions.len()),
            self.status_style(),
        ));
        if stats.total_games > 0 {
            lines.push(Line::styled(
                format!("Avg Moves: {:.1}", stats.avg_move_count),
                self.menu_normal_style(),
            ));
            lines.push(Line::styled(
                format!(
                    "Longest: {} moves | Shortest: {} moves",
                    stats.longest_game.move_count, stats.shortest_game.move_count
                ),
                self.menu_normal_style(),
            ));
        }

        if let Some(current) = self.current_bvb_session() {
            lines.push(Line::styled(
                "─── Current Game ───".to_string(),
                self.title_style(),
            ));
            lines.push(Line::styled(
                format!("Duration: {}", format_bvb_duration(current.duration())),
                self.menu_normal_style(),
            ));
            let moves = current.current_move_history();
            if !moves.is_empty() {
                lines.push(Line::styled(
                    format!("Last moves: {}", format_last_moves(&moves, 10)),
                    self.menu_normal_style(),
                ));
            }
            let board = current.current_board();
            let (cw, cb) = compute_captured_pieces(&board);
            if !cw.is_empty() || !cb.is_empty() {
                lines.push(Line::styled(
                    format!("Captured: {} | {}", cw, cb),
                    self.menu_normal_style(),
                ));
            }
        }
        lines.push(Line::styled(
            "═════════════════════════".to_string(),
            self.title_style(),
        ));
        lines.push(Line::from(""));
    }

    fn current_bvb_session(&self) -> Option<Arc<GameSession>> {
        let mgr = self.bvb_manager.as_ref()?;
        if self.bvb_view_mode == BvBViewMode::Single {
            return mgr.get_session(self.bvb_selected_game as i32);
        }
        mgr.sessions().into_iter().find(|s| !s.is_finished())
    }

    // ---- Shortcuts overlay ----

    fn render_shortcuts_overlay(&self) -> Vec<Line<'static>> {
        let mut lines = Vec::new();
        let section = |t: &str, out: &mut Vec<Line<'static>>, style: Style| {
            out.push(Line::styled(t.to_string(), style));
        };
        let key_style = self.selected_style();
        let desc_style = self.menu_normal_style();
        let shortcut = |k: &str, d: &str, out: &mut Vec<Line<'static>>| {
            out.push(Line::from(vec![
                Span::styled(format!("{:<15}", k), key_style),
                Span::styled(d.to_string(), desc_style),
            ]));
        };

        lines.push(self.title_line("Keyboard Shortcuts"));

        section("Global", &mut lines, self.selected_style());
        shortcut("?", "Show this help overlay", &mut lines);
        shortcut("n", "Start new game", &mut lines);
        shortcut("s", "Open settings", &mut lines);
        shortcut("Ctrl+C", "Quit application", &mut lines);
        shortcut("q", "Quit (or show save prompt in game)", &mut lines);
        shortcut("Esc", "Go back / Cancel", &mut lines);

        section("Menu Navigation", &mut lines, self.selected_style());
        shortcut("Up / k", "Move selection up", &mut lines);
        shortcut("Down / j", "Move selection down", &mut lines);
        shortcut("Enter", "Select / Confirm", &mut lines);

        section("Settings", &mut lines, self.selected_style());
        shortcut("Up / k", "Previous setting", &mut lines);
        shortcut("Down / j", "Next setting", &mut lines);
        shortcut("Enter/Space", "Toggle / Cycle setting", &mut lines);

        section("Gameplay", &mut lines, self.selected_style());
        shortcut("Type move", "Enter move (e.g., e4, Nf3, O-O)", &mut lines);
        shortcut("Enter", "Submit move", &mut lines);
        shortcut("resign", "Resign the game", &mut lines);
        shortcut("offerdraw", "Offer a draw", &mut lines);
        shortcut("showfen", "Show/copy FEN position", &mut lines);
        shortcut("menu", "Return to menu (with save)", &mut lines);

        section("Bot vs Bot", &mut lines, self.selected_style());
        shortcut("Space", "Pause / Resume", &mut lines);
        shortcut("Left / h", "Previous game / page", &mut lines);
        shortcut("Right / l", "Next game / page", &mut lines);
        shortcut("g", "Jump to game (enter game number)", &mut lines);
        shortcut("Tab", "Toggle grid / single view", &mut lines);
        shortcut("t", "Toggle speed (Normal / Instant)", &mut lines);
        shortcut("f", "Copy FEN of current game", &mut lines);

        lines.push(Line::from(""));
        lines.push(Line::styled(
            "Press any key to close".to_string(),
            self.help_style(),
        ));
        lines
    }
}

/// Whether a menu option is a primary action (Go `isPrimaryAction`).
fn is_primary_action(option: &str) -> bool {
    matches!(
        option,
        "New Game" | "Resume Game" | "Start" | "Play Again" | "New Session"
    )
}

/// Human-readable game result message (Go `getGameResultMessage`).
fn game_result_message(
    board: &Board,
    resigned_by: Option<Color>,
    draw_by_agreement: bool,
) -> String {
    if draw_by_agreement {
        return "Draw by agreement".to_string();
    }
    if let Some(color) = resigned_by {
        return if color == Color::White {
            "White resigned - Black wins".to_string()
        } else {
            "Black resigned - White wins".to_string()
        };
    }
    match board.status() {
        GameStatus::Checkmate => {
            if board.winner() == Some(Color::White) {
                "Checkmate! White wins".to_string()
            } else {
                "Checkmate! Black wins".to_string()
            }
        }
        GameStatus::Stalemate => "Stalemate - Draw".to_string(),
        GameStatus::DrawThreefoldRepetition | GameStatus::DrawFivefoldRepetition => {
            "Draw by repetition".to_string()
        }
        GameStatus::DrawFiftyMoveRule => "Draw by fifty-move rule".to_string(),
        GameStatus::DrawSeventyFiveMoveRule => "Draw by seventy-five-move rule".to_string(),
        GameStatus::DrawInsufficientMaterial => "Draw by insufficient material".to_string(),
        _ => "Game Over".to_string(),
    }
}

/// Text-based progress bar (Go `renderProgressBar`).
fn render_progress_bar(completed: i32, total: i32, width: i32) -> String {
    if total == 0 {
        return format!("[{}] 0% (0/0)", "░".repeat(width as usize));
    }
    let percent = completed as f64 / total as f64;
    let mut filled = (percent * width as f64) as i32;
    if filled > width {
        filled = width;
    }
    let bar = format!(
        "{}{}",
        "█".repeat(filled as usize),
        "░".repeat((width - filled) as usize)
    );
    format!(
        "[{}] {}% ({}/{})",
        bar,
        (percent * 100.0) as i32,
        completed,
        total
    )
}

/// Concatenates a line's span contents into plain text (drops styling).
fn line_plain_text(line: &Line) -> String {
    line.spans.iter().map(|s| s.content.as_ref()).collect()
}

/// Display width of a string in terminal columns (unicode-aware).
fn display_width(s: &str) -> usize {
    Span::raw(s).width()
}

/// Pads with spaces or truncates by characters so `s` occupies exactly `width`
/// display columns, mirroring Go's lipgloss width normalization + `truncateToWidth`.
fn fit_display_width(s: &str, width: usize) -> String {
    let w = display_width(s);
    if w == width {
        return s.to_string();
    }
    if w < width {
        let mut out = s.to_string();
        out.push_str(&" ".repeat(width - w));
        return out;
    }
    let mut out = String::new();
    let mut cur = 0usize;
    for ch in s.chars() {
        let cw = display_width(ch.encode_utf8(&mut [0u8; 4]));
        if cur + cw > width {
            break;
        }
        out.push(ch);
        cur += cw;
    }
    if cur < width {
        out.push_str(&" ".repeat(width - cur));
    }
    out
}

/// Parses a string into an integer, ignoring non-digits (Go `parseConcurrencyValue`).
fn parse_concurrency_value(s: &str) -> i32 {
    let mut val = 0i32;
    for c in s.chars() {
        if c.is_ascii_digit() {
            val = val * 10 + (c as i32 - '0' as i32);
        }
    }
    val
}

fn format_bvb_duration(d: Duration) -> String {
    let total_seconds = d.as_secs();
    format!("{}:{:02}", total_seconds / 60, total_seconds % 60)
}

fn format_last_moves(moves: &[Move], n: usize) -> String {
    if moves.is_empty() {
        return String::new();
    }
    let start = moves.len().saturating_sub(n);
    moves[start..]
        .iter()
        .map(|m| m.to_string())
        .collect::<Vec<_>>()
        .join(", ")
}

/// Computes captured pieces vs the starting position (Go `computeCapturedPieces`).
fn compute_captured_pieces(board: &Board) -> (String, String) {
    let starting = |pt: PieceType| -> i32 {
        match pt {
            PieceType::Pawn => 8,
            PieceType::Knight | PieceType::Bishop | PieceType::Rook => 2,
            PieceType::Queen | PieceType::King => 1,
            _ => 0,
        }
    };

    let mut white_counts = [0i32; 7];
    let mut black_counts = [0i32; 7];
    for i in 0..64 {
        let piece = board.piece_at(Square(i as i8));
        if piece.is_empty() {
            continue;
        }
        let idx = piece.piece_type() as usize;
        if piece.color() == Color::White {
            white_counts[idx] += 1;
        } else {
            black_counts[idx] += 1;
        }
    }

    let white_sym = |pt: PieceType| match pt {
        PieceType::Pawn => '♙',
        PieceType::Knight => '♘',
        PieceType::Bishop => '♗',
        PieceType::Rook => '♖',
        PieceType::Queen => '♕',
        _ => '?',
    };
    let black_sym = |pt: PieceType| match pt {
        PieceType::Pawn => '♟',
        PieceType::Knight => '♞',
        PieceType::Bishop => '♝',
        PieceType::Rook => '♜',
        PieceType::Queen => '♛',
        _ => '?',
    };

    let order = [
        PieceType::Queen,
        PieceType::Rook,
        PieceType::Bishop,
        PieceType::Knight,
        PieceType::Pawn,
    ];
    let mut white_captured = String::new();
    let mut black_captured = String::new();
    for pt in order {
        let idx = pt as usize;
        let white_missing = starting(pt) - white_counts[idx];
        for _ in 0..white_missing.max(0) {
            white_captured.push(white_sym(pt));
        }
        let black_missing = starting(pt) - black_counts[idx];
        for _ in 0..black_missing.max(0) {
            black_captured.push(black_sym(pt));
        }
    }
    (white_captured, black_captured)
}

#[cfg(test)]
mod grid_cell_tests {
    use super::{display_width, fit_display_width};

    #[test]
    fn pads_short_line_to_exact_width() {
        let out = fit_display_width("Game 1", 22);
        assert_eq!(display_width(&out), 22);
        assert!(out.starts_with("Game 1"));
    }

    #[test]
    fn truncates_long_line_to_exact_width() {
        let long = "x".repeat(40);
        let out = fit_display_width(&long, 22);
        assert_eq!(out.chars().count(), 22);
        assert_eq!(display_width(&out), 22);
    }

    #[test]
    fn exact_width_is_unchanged() {
        let s = "y".repeat(22);
        assert_eq!(fit_display_width(&s, 22), s);
    }

    #[test]
    fn empty_line_pads_to_full_width() {
        assert_eq!(display_width(&fit_display_width("", 22)), 22);
    }
}
