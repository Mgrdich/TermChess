//! Event loop, terminal setup/teardown, and input thread (Go `main.go` bootstrap
//! + the Bubbletea program loop).

use std::io::{self, Stdout};
use std::sync::mpsc::{channel, Sender};
use std::thread;

use crossterm::event::{self, DisableMouseCapture, EnableMouseCapture, Event, KeyEventKind};
use crossterm::execute;
use crossterm::terminal::{
    disable_raw_mode, enable_raw_mode, EnterAlternateScreen, LeaveAlternateScreen,
};
use ratatui::backend::CrosstermBackend;
use ratatui::Terminal;

use crate::app::App;
use crate::event::{AppEvent, Key};

/// Runs the TUI to completion (mirrors `tea.NewProgram(...).Run()`).
pub fn run(config: config::Config) -> anyhow::Result<()> {
    enable_raw_mode()?;
    let mut stdout = io::stdout();
    execute!(stdout, EnterAlternateScreen, EnableMouseCapture)?;
    let backend = CrosstermBackend::new(stdout);
    let mut terminal = Terminal::new(backend)?;

    let result = run_loop(&mut terminal, config);

    // Teardown, regardless of loop outcome.
    disable_raw_mode()?;
    execute!(
        terminal.backend_mut(),
        LeaveAlternateScreen,
        DisableMouseCapture
    )?;
    terminal.show_cursor()?;

    result
}

fn run_loop(
    terminal: &mut Terminal<CrosstermBackend<Stdout>>,
    config: config::Config,
) -> anyhow::Result<()> {
    let (tx, rx) = channel::<AppEvent>();

    spawn_input_thread(tx.clone());

    let mut app = App::new(config, tx.clone());
    app.spawn_update_check();

    let size = terminal.size()?;
    app.on_resize(size.width, size.height);

    terminal.draw(|f| app.draw(f))?;

    while !app.should_quit {
        let ev = match rx.recv() {
            Ok(ev) => ev,
            Err(_) => break,
        };
        app.update(ev);
        terminal.draw(|f| app.draw(f))?;
    }

    Ok(())
}

/// Reads terminal events and forwards normalized `AppEvent`s.
fn spawn_input_thread(tx: Sender<AppEvent>) {
    thread::spawn(move || loop {
        match event::read() {
            Ok(Event::Key(key)) => {
                // Ignore key releases; forward presses and repeats.
                if key.kind == KeyEventKind::Release {
                    continue;
                }
                if tx.send(AppEvent::Key(Key::from_event(key))).is_err() {
                    break;
                }
            }
            Ok(Event::Mouse(m)) => {
                if tx.send(AppEvent::Mouse(m)).is_err() {
                    break;
                }
            }
            Ok(Event::Resize(w, h)) => {
                if tx.send(AppEvent::Resize(w, h)).is_err() {
                    break;
                }
            }
            Ok(_) => {}
            Err(_) => break,
        }
    });
}
