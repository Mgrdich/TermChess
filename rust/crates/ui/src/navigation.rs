//! Navigation stack management (port of Go `navigation.go`).

use crate::app::{build_main_menu_options, App};
use crate::state::Screen;

/// Human-readable name for a screen (Go `screenName`).
pub fn screen_name(s: Screen) -> &'static str {
    match s {
        Screen::MainMenu => "Main Menu",
        Screen::GameTypeSelect => "New Game",
        Screen::BotSelect => "Bot Difficulty",
        Screen::ColorSelect => "Choose Color",
        Screen::FenInput => "Load Game",
        Screen::GamePlay => "Game",
        Screen::GameOver => "Game Over",
        Screen::Settings => "Settings",
        Screen::SavePrompt => "Save Game",
        Screen::DrawPrompt => "Draw Offer",
        Screen::BvBBotSelect => "Bot vs Bot Setup",
        Screen::BvBGameMode => "Game Mode",
        Screen::BvBGridConfig => "Grid Layout",
        Screen::BvBGamePlay => "Bot vs Bot",
        Screen::BvBStats => "Statistics",
        Screen::BvBViewModeSelect => "View Mode",
        Screen::BvBConcurrencySelect => "Concurrency Select",
    }
}

impl App {
    /// Navigates forward, pushing the current screen (Go `pushScreen`).
    pub fn push_screen(&mut self, new_screen: Screen) {
        if self.screen == new_screen {
            return;
        }
        self.nav_stack.push(self.screen);
        self.screen = new_screen;
    }

    /// Returns to the previous screen (Go `popScreen`).
    pub fn pop_screen(&mut self) -> Screen {
        match self.nav_stack.pop() {
            Some(prev) => {
                self.screen = prev;
                self.restore_menu_state();
                prev
            }
            None => {
                self.screen = Screen::MainMenu;
                self.restore_menu_state();
                Screen::MainMenu
            }
        }
    }

    /// Restores menu options/selection for the current screen (Go `restoreMenuState`).
    pub fn restore_menu_state(&mut self) {
        match self.screen {
            Screen::GameTypeSelect => {
                self.menu_options = vec![
                    "Player vs Player".into(),
                    "Player vs Bot".into(),
                    "Bot vs Bot".into(),
                ];
            }
            Screen::BvBBotSelect => {
                self.menu_options = vec!["Easy".into(), "Medium".into(), "Hard".into()];
            }
            Screen::BvBGameMode => {
                self.menu_options = vec!["Single Game".into(), "Multi-Game".into()];
            }
            Screen::BvBGridConfig => {
                self.menu_options = vec![
                    "1x1".into(),
                    "2x2".into(),
                    "2x3".into(),
                    "2x4".into(),
                    "Custom".into(),
                ];
            }
            Screen::BvBConcurrencySelect => {}
            Screen::BvBViewModeSelect => {
                self.menu_options = vec![
                    "Grid View".into(),
                    "Single Board".into(),
                    "Stats Only".into(),
                ];
            }
            Screen::MainMenu => {
                self.menu_options = build_main_menu_options();
            }
            Screen::BotSelect => {
                self.menu_options = vec!["Easy".into(), "Medium".into(), "Hard".into()];
            }
            Screen::ColorSelect => {
                self.menu_options = vec!["Play as White".into(), "Play as Black".into()];
            }
            Screen::Settings => {
                self.menu_options = vec![format!("Theme: {}", self.theme.name)];
            }
            _ => {}
        }
        self.menu_selection = 0;
        self.error_msg.clear();
    }

    /// Clears the navigation stack (Go `clearNavStack`).
    pub fn clear_nav_stack(&mut self) {
        self.nav_stack.clear();
    }

    /// Generates a breadcrumb string (Go `breadcrumb`).
    pub fn breadcrumb(&self) -> String {
        if self.screen == Screen::MainMenu || self.nav_stack.is_empty() {
            return String::new();
        }
        let parent = *self.nav_stack.last().unwrap();
        format!("{} > {}", screen_name(parent), screen_name(self.screen))
    }

    /// Whether there is a screen to go back to (Go `canGoBack`).
    pub fn can_go_back(&self) -> bool {
        !self.nav_stack.is_empty()
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::app::App;
    use std::sync::mpsc::channel;

    fn app() -> App {
        let (tx, _rx) = channel();
        App::new(config::Config::default(), tx)
    }

    #[test]
    fn screen_names() {
        assert_eq!(screen_name(Screen::MainMenu), "Main Menu");
        assert_eq!(screen_name(Screen::GameTypeSelect), "New Game");
        assert_eq!(screen_name(Screen::BvBGamePlay), "Bot vs Bot");
    }

    #[test]
    fn push_and_pop() {
        let mut a = app();
        assert_eq!(a.screen, Screen::MainMenu);
        assert!(!a.can_go_back());
        a.push_screen(Screen::GameTypeSelect);
        assert_eq!(a.screen, Screen::GameTypeSelect);
        assert!(a.can_go_back());
        a.push_screen(Screen::BotSelect);
        assert_eq!(a.screen, Screen::BotSelect);
        let prev = a.pop_screen();
        assert_eq!(prev, Screen::GameTypeSelect);
        assert_eq!(a.screen, Screen::GameTypeSelect);
    }

    #[test]
    fn push_same_screen_is_noop() {
        let mut a = app();
        a.push_screen(Screen::MainMenu);
        assert!(a.nav_stack.is_empty());
    }

    #[test]
    fn pop_empty_returns_main_menu() {
        let mut a = app();
        a.screen = Screen::Settings;
        let s = a.pop_screen();
        assert_eq!(s, Screen::MainMenu);
        assert_eq!(a.screen, Screen::MainMenu);
    }

    #[test]
    fn breadcrumb_string() {
        let mut a = app();
        assert_eq!(a.breadcrumb(), "");
        a.push_screen(Screen::GameTypeSelect);
        assert_eq!(a.breadcrumb(), "Main Menu > New Game");
    }

    #[test]
    fn clear_nav_stack_works() {
        let mut a = app();
        a.push_screen(Screen::GameTypeSelect);
        a.push_screen(Screen::BotSelect);
        a.clear_nav_stack();
        assert!(!a.can_go_back());
    }
}
