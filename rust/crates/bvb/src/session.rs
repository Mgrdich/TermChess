//! Single-game Bot vs Bot controller (ported from `session.go`).

use std::sync::{Arc, Condvar, Mutex, MutexGuard};
use std::time::{Duration, Instant};

use bot::{Context, Engine};
use engine::{Board, Color, GameStatus, Move};

use crate::types::{GameResult, PlaybackSpeed, SessionState};

/// The maximum number of moves before a forced draw.
pub(crate) const MAX_MOVE_COUNT: i32 = 500;

/// A thread-safe, shareable engine handle.
///
/// Go stored `bot.Engine` interface values directly; here we require the engine
/// to be `Send + Sync` so a session can be driven from a background thread while
/// accessors read from others.
pub type SharedEngine = Arc<dyn Engine + Send + Sync>;

/// The mutable, lock-protected interior of a [`GameSession`].
pub(crate) struct Inner {
    pub(crate) board: Board,
    pub(crate) white_engine: Option<SharedEngine>,
    pub(crate) black_engine: Option<SharedEngine>,
    pub(crate) white_name: String,
    pub(crate) black_name: String,
    pub(crate) move_history: Vec<Move>,
    pub(crate) state: SessionState,
    pub(crate) paused: bool,
    pub(crate) aborted: bool,
    pub(crate) result: Option<GameResult>,
    pub(crate) start_time: Option<Instant>,
    pub(crate) speed: PlaybackSpeed,
}

/// Manages a single Bot vs Bot chess game.
///
/// [`GameSession::run`] drives the game loop (typically on a background thread)
/// and provides thread-safe access to the current board state and move history.
pub struct GameSession {
    game_number: i32,
    pub(crate) inner: Mutex<Inner>,
    cv: Condvar,
}

impl GameSession {
    /// Creates a new game session ready to be run.
    ///
    /// Unlike the Go original, `speed` is taken by value; [`GameSession::set_speed`]
    /// provides the external mutation that Go achieved via a shared pointer.
    pub fn new(
        game_number: i32,
        white_engine: SharedEngine,
        black_engine: SharedEngine,
        white_name: impl Into<String>,
        black_name: impl Into<String>,
        speed: PlaybackSpeed,
    ) -> Arc<GameSession> {
        Arc::new(GameSession {
            game_number,
            inner: Mutex::new(Inner {
                board: Board::new(),
                white_engine: Some(white_engine),
                black_engine: Some(black_engine),
                white_name: white_name.into(),
                black_name: black_name.into(),
                move_history: Vec::with_capacity(80),
                state: SessionState::Running,
                paused: false,
                aborted: false,
                result: None,
                start_time: None,
                speed,
            }),
            cv: Condvar::new(),
        })
    }

    fn lock(&self) -> MutexGuard<'_, Inner> {
        self.inner.lock().expect("session mutex poisoned")
    }

    /// Executes the game loop. Intended to be called on a background thread.
    ///
    /// Plays moves alternately until the game ends, an error occurs, or the
    /// session is aborted. Cleanup always runs when this returns.
    pub fn run(&self) {
        {
            let mut inner = self.lock();
            inner.start_time = Some(Instant::now());
        }

        // Guarantee cleanup runs even on early return (mirrors Go's deferred cleanup).
        let _guard = CleanupGuard(self);

        loop {
            // Check for abort, then handle pause (blocking until resumed or aborted).
            {
                let mut inner = self.lock();
                if inner.aborted {
                    inner.state = SessionState::Finished;
                    return;
                }
                while inner.paused && !inner.aborted {
                    inner = self.cv.wait(inner).expect("session mutex poisoned");
                }
                if inner.aborted {
                    inner.state = SessionState::Finished;
                    return;
                }
            }

            // Determine the current engine and snapshot the board.
            let (engine, name, active_color, board_copy) = {
                let inner = self.lock();
                let active = inner.board.active_color;
                let (eng, nm) = if active == Color::White {
                    (inner.white_engine.clone(), inner.white_name.clone())
                } else {
                    (inner.black_engine.clone(), inner.black_name.clone())
                };
                (eng, nm, active, inner.board.copy())
            };

            let engine = match engine {
                Some(e) => e,
                None => {
                    // Engines were cleaned up out from under us; stop.
                    let mut inner = self.lock();
                    inner.state = SessionState::Finished;
                    return;
                }
            };

            // Ask the engine for a move with a timeout to prevent infinite computation.
            let ctx = Context::with_timeout(Duration::from_secs(30));
            let mv = match engine.select_move(&ctx, &board_copy) {
                Ok(m) => m,
                Err(e) => {
                    self.finish_with_error(&name, active_color, e.to_string());
                    return;
                }
            };
            // Release the caller-facing borrow of the engine explicitly.
            drop(engine);

            // Apply the move and check terminal conditions.
            {
                let mut inner = self.lock();
                if let Err(e) = inner.board.make_move(mv) {
                    let msg = e.to_string();
                    drop(inner);
                    self.finish_with_error(&name, active_color, msg);
                    return;
                }
                inner.move_history.push(mv);
                let move_count = inner.move_history.len() as i32;

                let status = inner.board.status();
                if status != GameStatus::Ongoing {
                    self.finish_with_status(&mut inner, status, move_count);
                    return;
                }

                if move_count >= MAX_MOVE_COUNT {
                    let duration = inner.start_time.map(|t| t.elapsed()).unwrap_or_default();
                    let final_fen = inner.board.to_fen();
                    let history = inner.move_history.clone();
                    inner.result = Some(GameResult {
                        game_number: self.game_number,
                        winner: "Draw".to_string(),
                        winner_color: Color::White,
                        end_reason: "move limit exceeded".to_string(),
                        move_count,
                        duration,
                        final_fen,
                        move_history: history,
                    });
                    inner.state = SessionState::Finished;
                    return;
                }
            }

            // Sleep for the configured playback speed, interruptible by abort.
            let delay = self.lock().speed.duration();
            if !delay.is_zero() {
                let inner = self.lock();
                let (mut inner, _timeout) = self
                    .cv
                    .wait_timeout_while(inner, delay, |i| !i.aborted)
                    .expect("session mutex poisoned");
                if inner.aborted {
                    inner.state = SessionState::Finished;
                    return;
                }
            }
        }
    }

    /// Signals the session to pause. Safe to call multiple times. No-op if the
    /// session is already paused or finished.
    pub fn pause(&self) {
        let mut inner = self.lock();
        if inner.paused || inner.state == SessionState::Finished {
            return;
        }
        inner.paused = true;
        inner.state = SessionState::Paused;
        self.cv.notify_all();
    }

    /// Signals the session to continue after a pause. No-op if not paused.
    pub fn resume(&self) {
        let mut inner = self.lock();
        if !inner.paused {
            return;
        }
        inner.paused = false;
        inner.state = SessionState::Running;
        self.cv.notify_all();
    }

    /// Updates the playback speed for this session. Safe to call concurrently.
    pub fn set_speed(&self, speed: PlaybackSpeed) {
        let mut inner = self.lock();
        inner.speed = speed;
    }

    /// Signals the session to stop immediately. Safe to call multiple times.
    pub fn abort(&self) {
        let mut inner = self.lock();
        inner.aborted = true;
        self.cv.notify_all();
    }

    /// Returns a deep copy of the current board state.
    pub fn current_board(&self) -> Board {
        self.lock().board.copy()
    }

    /// Returns a copy of the move history so far.
    pub fn current_move_history(&self) -> Vec<Move> {
        self.lock().move_history.clone()
    }

    /// Returns true if the game session has completed.
    pub fn is_finished(&self) -> bool {
        self.lock().state == SessionState::Finished
    }

    /// Returns the game result, or `None` if the game is not finished.
    pub fn result(&self) -> Option<GameResult> {
        self.lock().result.clone()
    }

    /// Returns the sequence number of this game.
    pub fn game_number(&self) -> i32 {
        self.game_number
    }

    /// Returns the elapsed time since the game started.
    ///
    /// Returns 0 before the game starts, or the final duration once finished.
    pub fn duration(&self) -> Duration {
        let inner = self.lock();
        match inner.start_time {
            None => Duration::ZERO,
            Some(start) => {
                if inner.state == SessionState::Finished {
                    if let Some(r) = &inner.result {
                        return r.duration;
                    }
                }
                start.elapsed()
            }
        }
    }

    /// Returns the instant the game started, or `None` if it hasn't started.
    pub fn start_time(&self) -> Option<Instant> {
        self.lock().start_time
    }

    /// Returns the current session state.
    pub fn state(&self) -> SessionState {
        self.lock().state
    }

    /// Records the game result based on the board's game status. Called with the
    /// inner lock held.
    fn finish_with_status(&self, inner: &mut Inner, status: GameStatus, move_count: i32) {
        let mut winner = "Draw".to_string();
        let mut winner_color = Color::White;

        if status == GameStatus::Checkmate {
            // The active color is the one checkmated, so the opponent wins.
            if inner.board.active_color == Color::White {
                winner = inner.black_name.clone();
                winner_color = Color::Black;
            } else {
                winner = inner.white_name.clone();
                winner_color = Color::White;
            }
        }

        let duration = inner.start_time.map(|t| t.elapsed()).unwrap_or_default();
        inner.result = Some(GameResult {
            game_number: self.game_number,
            winner,
            winner_color,
            end_reason: status.to_string(),
            move_count,
            duration,
            final_fen: inner.board.to_fen(),
            move_history: inner.move_history.clone(),
        });
        inner.state = SessionState::Finished;
    }

    /// Records the game result when an engine produces an error.
    fn finish_with_error(&self, engine_name: &str, engine_color: Color, err: String) {
        let _ = engine_name;
        let mut inner = self.lock();

        // The engine that errored loses; the opponent wins.
        let (winner, winner_color) = if engine_color == Color::White {
            (inner.black_name.clone(), Color::Black)
        } else {
            (inner.white_name.clone(), Color::White)
        };

        let duration = inner.start_time.map(|t| t.elapsed()).unwrap_or_default();
        inner.result = Some(GameResult {
            game_number: self.game_number,
            winner,
            winner_color,
            end_reason: format!("engine error: {}", err),
            move_count: inner.move_history.len() as i32,
            duration,
            final_fen: inner.board.to_fen(),
            move_history: inner.move_history.clone(),
        });
        inner.state = SessionState::Finished;
    }

    /// Releases resources held by the session: closes both engines and drops the
    /// references. Idempotent and safe to call multiple times.
    pub(crate) fn cleanup(&self) {
        let mut inner = self.lock();
        if let Some(e) = inner.white_engine.take() {
            let _ = e.close();
        }
        if let Some(e) = inner.black_engine.take() {
            let _ = e.close();
        }
    }
}

/// Runs [`GameSession::cleanup`] when dropped, mirroring Go's `defer s.cleanup()`.
struct CleanupGuard<'a>(&'a GameSession);

impl Drop for CleanupGuard<'_> {
    fn drop(&mut self) {
        self.0.cleanup();
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use bot::new_random_engine;
    use std::sync::atomic::{AtomicBool, Ordering};
    use std::thread;

    fn random_engine() -> SharedEngine {
        Arc::new(new_random_engine(&[]).expect("failed to create random engine"))
    }

    fn spawn_run(session: Arc<GameSession>) -> (thread::JoinHandle<()>, Arc<AtomicBool>) {
        let done = Arc::new(AtomicBool::new(false));
        let done_clone = done.clone();
        let handle = thread::spawn(move || {
            session.run();
            done_clone.store(true, Ordering::SeqCst);
        });
        (handle, done)
    }

    fn wait_done(handle: thread::JoinHandle<()>, timeout: Duration) -> bool {
        let start = Instant::now();
        loop {
            if handle.is_finished() {
                handle.join().unwrap();
                return true;
            }
            if start.elapsed() > timeout {
                return false;
            }
            thread::sleep(Duration::from_millis(10));
        }
    }

    #[test]
    fn playback_speed_duration() {
        assert_eq!(PlaybackSpeed::Instant.duration(), Duration::ZERO);
        assert_eq!(PlaybackSpeed::Normal.duration(), Duration::from_secs(1));
    }

    #[test]
    fn runs_to_completion() {
        let session = GameSession::new(
            1,
            random_engine(),
            random_engine(),
            "White Bot",
            "Black Bot",
            PlaybackSpeed::Instant,
        );
        let (handle, _done) = spawn_run(session.clone());
        assert!(
            wait_done(handle, Duration::from_secs(60)),
            "game did not complete within timeout"
        );
        assert!(session.is_finished());
    }

    #[test]
    fn result_populated() {
        let session = GameSession::new(
            42,
            random_engine(),
            random_engine(),
            "White Bot",
            "Black Bot",
            PlaybackSpeed::Instant,
        );
        let (handle, _done) = spawn_run(session.clone());
        assert!(
            wait_done(handle, Duration::from_secs(60)),
            "game did not complete within timeout"
        );

        let result = session.result().expect("result should not be nil");
        assert_eq!(result.game_number, 42);
        assert!(!result.winner.is_empty());
        assert!(!result.end_reason.is_empty());
        assert!(result.move_count > 0);
        assert!(result.duration > Duration::ZERO);
        assert!(!result.final_fen.is_empty());
        assert_eq!(result.move_history.len() as i32, result.move_count);

        assert!(matches!(
            result.winner.as_str(),
            "White Bot" | "Black Bot" | "Draw"
        ));
    }

    #[test]
    fn abort_stops_game() {
        let session = GameSession::new(
            1,
            random_engine(),
            random_engine(),
            "White Bot",
            "Black Bot",
            PlaybackSpeed::Normal,
        );
        let (handle, _done) = spawn_run(session.clone());
        thread::sleep(Duration::from_millis(50));
        session.abort();
        assert!(
            wait_done(handle, Duration::from_secs(5)),
            "session did not stop within timeout"
        );
        assert!(session.is_finished());
    }

    #[test]
    fn concurrent_accessors() {
        let session = GameSession::new(
            1,
            random_engine(),
            random_engine(),
            "White Bot",
            "Black Bot",
            PlaybackSpeed::Normal,
        );
        let (handle, _done) = spawn_run(session.clone());

        let mut readers = Vec::new();
        for _ in 0..10 {
            let s = session.clone();
            readers.push(thread::spawn(move || {
                for _ in 0..20 {
                    let _ = s.current_board();
                    let _ = s.current_move_history();
                    let _ = s.is_finished();
                    let _ = s.state();
                    let _ = s.result();
                    let _ = s.game_number();
                    thread::sleep(Duration::from_millis(5));
                }
            }));
        }
        for r in readers {
            r.join().unwrap();
        }

        session.abort();
        assert!(
            wait_done(handle, Duration::from_secs(5)),
            "session did not stop within timeout"
        );
    }

    #[test]
    fn game_number_accessor() {
        let session = GameSession::new(
            7,
            random_engine(),
            random_engine(),
            "A",
            "B",
            PlaybackSpeed::Instant,
        );
        assert_eq!(session.game_number(), 7);
        session.cleanup();
    }

    #[test]
    fn initial_state() {
        let w = random_engine();
        let b = random_engine();
        let session = GameSession::new(1, w.clone(), b.clone(), "W", "B", PlaybackSpeed::Instant);

        assert_eq!(session.state(), SessionState::Running);
        assert!(!session.is_finished());
        assert!(session.result().is_none());

        let board = session.current_board();
        assert_eq!(board.active_color, Color::White);

        assert_eq!(session.current_move_history().len(), 0);

        let _ = w.close();
        let _ = b.close();
    }

    #[test]
    fn max_move_count_constant() {
        assert_eq!(MAX_MOVE_COUNT, 500);
    }

    #[test]
    fn pause_blocks_progress() {
        let session = GameSession::new(
            1,
            random_engine(),
            random_engine(),
            "White Bot",
            "Black Bot",
            PlaybackSpeed::Normal,
        );
        let (handle, _done) = spawn_run(session.clone());

        thread::sleep(Duration::from_millis(100));
        session.pause();
        thread::sleep(Duration::from_millis(50));

        let moves_before = session.current_move_history().len();
        thread::sleep(Duration::from_millis(300));
        let moves_after = session.current_move_history().len();
        assert_eq!(moves_after, moves_before, "moves changed during pause");
        assert_eq!(session.state(), SessionState::Paused);

        session.resume();
        session.abort();
        assert!(
            wait_done(handle, Duration::from_secs(5)),
            "session did not stop within timeout"
        );
    }

    #[test]
    fn resume_after_pause() {
        let session = GameSession::new(
            1,
            random_engine(),
            random_engine(),
            "White Bot",
            "Black Bot",
            PlaybackSpeed::Normal,
        );
        let (handle, _done) = spawn_run(session.clone());

        thread::sleep(Duration::from_millis(50));
        session.pause();
        thread::sleep(Duration::from_millis(100));
        session.resume();
        assert_eq!(session.state(), SessionState::Running);

        session.abort();
        assert!(
            wait_done(handle, Duration::from_secs(5)),
            "session did not stop within timeout after abort"
        );
        assert!(session.is_finished());
    }

    #[test]
    fn abort_during_pause() {
        let session = GameSession::new(
            1,
            random_engine(),
            random_engine(),
            "White Bot",
            "Black Bot",
            PlaybackSpeed::Normal,
        );
        let (handle, _done) = spawn_run(session.clone());

        thread::sleep(Duration::from_millis(50));
        session.pause();
        thread::sleep(Duration::from_millis(50));
        session.abort();

        assert!(
            wait_done(handle, Duration::from_secs(2)),
            "session did not abort during pause within timeout"
        );
        assert!(session.is_finished());
    }

    #[test]
    fn instant_speed_completes() {
        let session = GameSession::new(
            1,
            random_engine(),
            random_engine(),
            "White Bot",
            "Black Bot",
            PlaybackSpeed::Instant,
        );
        let (handle, _done) = spawn_run(session.clone());
        assert!(
            wait_done(handle, Duration::from_secs(5)),
            "instant speed game did not complete within 5 seconds"
        );
        assert!(session.is_finished());
        let result = session.result().expect("result should not be nil");
        assert!(result.move_count > 0);
    }

    #[test]
    fn speed_change_mid_game() {
        let session = GameSession::new(
            1,
            random_engine(),
            random_engine(),
            "White Bot",
            "Black Bot",
            PlaybackSpeed::Normal,
        );
        let (handle, _done) = spawn_run(session.clone());

        thread::sleep(Duration::from_millis(500));
        let moves_before = session.current_move_history().len();

        session.set_speed(PlaybackSpeed::Instant);
        thread::sleep(Duration::from_secs(3));

        let moves_after = session.current_move_history().len();
        if !session.is_finished() {
            assert!(
                moves_after > moves_before + 2,
                "expected significant progress after speed change: before={}, after={}",
                moves_before,
                moves_after
            );
            session.abort();
            assert!(
                wait_done(handle, Duration::from_secs(5)),
                "session did not stop within timeout after abort"
            );
        } else {
            handle.join().unwrap();
        }
    }

    #[test]
    fn cleanup_after_run() {
        let session = GameSession::new(
            1,
            random_engine(),
            random_engine(),
            "White Bot",
            "Black Bot",
            PlaybackSpeed::Instant,
        );
        let (handle, _done) = spawn_run(session.clone());
        assert!(
            wait_done(handle, Duration::from_secs(60)),
            "game did not complete within timeout"
        );
        assert!(session.is_finished());

        let inner = session.inner.lock().unwrap();
        assert!(
            inner.white_engine.is_none(),
            "whiteEngine should be nil after cleanup"
        );
        assert!(
            inner.black_engine.is_none(),
            "blackEngine should be nil after cleanup"
        );
    }

    #[test]
    fn cleanup_is_idempotent() {
        let session = GameSession::new(
            1,
            random_engine(),
            random_engine(),
            "White Bot",
            "Black Bot",
            PlaybackSpeed::Instant,
        );
        let (handle, _done) = spawn_run(session.clone());
        assert!(
            wait_done(handle, Duration::from_secs(60)),
            "game did not complete within timeout"
        );

        session.cleanup();
        session.cleanup();

        let inner = session.inner.lock().unwrap();
        assert!(inner.white_engine.is_none());
        assert!(inner.black_engine.is_none());
    }

    #[test]
    fn cleanup_after_abort() {
        let session = GameSession::new(
            1,
            random_engine(),
            random_engine(),
            "White Bot",
            "Black Bot",
            PlaybackSpeed::Normal,
        );
        let (handle, _done) = spawn_run(session.clone());
        thread::sleep(Duration::from_millis(50));
        session.abort();
        assert!(
            wait_done(handle, Duration::from_secs(5)),
            "session did not stop within timeout"
        );

        let inner = session.inner.lock().unwrap();
        assert!(
            inner.white_engine.is_none(),
            "whiteEngine should be nil after abort"
        );
        assert!(
            inner.black_engine.is_none(),
            "blackEngine should be nil after abort"
        );
    }

    #[test]
    fn duration_before_start() {
        let w = random_engine();
        let b = random_engine();
        let session = GameSession::new(
            1,
            w.clone(),
            b.clone(),
            "White Bot",
            "Black Bot",
            PlaybackSpeed::Instant,
        );
        assert_eq!(session.duration(), Duration::ZERO);
        let _ = w.close();
        let _ = b.close();
    }

    #[test]
    fn duration_during_game() {
        let session = GameSession::new(
            1,
            random_engine(),
            random_engine(),
            "White Bot",
            "Black Bot",
            PlaybackSpeed::Normal,
        );
        let (handle, _done) = spawn_run(session.clone());

        thread::sleep(Duration::from_millis(100));
        let duration = session.duration();
        assert!(duration > Duration::ZERO);

        thread::sleep(Duration::from_millis(200));
        let new_duration = session.duration();
        assert!(new_duration > duration, "duration did not increase");

        session.abort();
        handle.join().unwrap();
    }

    #[test]
    fn duration_after_finish() {
        let session = GameSession::new(
            1,
            random_engine(),
            random_engine(),
            "White Bot",
            "Black Bot",
            PlaybackSpeed::Instant,
        );
        let (handle, _done) = spawn_run(session.clone());
        assert!(
            wait_done(handle, Duration::from_secs(60)),
            "game did not complete within timeout"
        );

        let result = session.result().expect("result should not be nil");
        assert_eq!(session.duration(), result.duration);
    }

    #[test]
    fn start_time_accessor() {
        let session = GameSession::new(
            1,
            random_engine(),
            random_engine(),
            "White Bot",
            "Black Bot",
            PlaybackSpeed::Normal,
        );
        assert!(session.start_time().is_none());

        let before_run = Instant::now();
        let (handle, _done) = spawn_run(session.clone());
        thread::sleep(Duration::from_millis(50));

        let start_time = session.start_time().expect("start time should be set");
        assert!(start_time >= before_run);

        session.abort();
        handle.join().unwrap();
    }
}
