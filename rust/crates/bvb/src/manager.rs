//! Multi-game orchestration: runs N parallel game sessions (ported from
//! `manager.go`).

use std::sync::atomic::{AtomicI32, Ordering};
use std::sync::{Arc, Condvar, Mutex, MutexGuard};

use bot::{new_minimax_engine, new_random_engine, Difficulty, EngineError};

use crate::session::{GameSession, SharedEngine};
use crate::stats::{compute_stats, AggregateStats};
use crate::types::{PlaybackSpeed, SessionState};

/// Limits how many games run simultaneously to prevent excessive CPU usage.
pub(crate) const MAX_CONCURRENT_GAMES: i32 = 50;

/// Returns the maximum number of concurrent games. Exported for UI display.
pub fn max_concurrent_games() -> i32 {
    MAX_CONCURRENT_GAMES
}

/// Returns the recommended concurrency based on CPU count.
///
/// Tiered formula: `numCPU <= 2` → `numCPU`; `numCPU <= 4` → `numCPU * 1.5`;
/// otherwise `numCPU * 2`. Capped at `MAX_CONCURRENT_GAMES`, minimum 1.
pub fn calculate_default_concurrency() -> i32 {
    let num_cpu = std::thread::available_parallelism()
        .map(|n| n.get() as i32)
        .unwrap_or(1);
    calculate_default_concurrency_with_cpu(num_cpu)
}

/// The internal implementation that accepts the CPU count as a parameter for
/// testing purposes.
pub(crate) fn calculate_default_concurrency_with_cpu(num_cpu: i32) -> i32 {
    let mut concurrency = if num_cpu <= 2 {
        num_cpu
    } else if num_cpu <= 4 {
        (num_cpu as f64 * 1.5) as i32
    } else {
        num_cpu * 2
    };

    concurrency = concurrency.clamp(1, MAX_CONCURRENT_GAMES);
    concurrency
}

/// A counting semaphore that can be aborted to unblock waiters (mirrors Go's
/// buffered channel used as a semaphore, combined with the abort channel).
struct Semaphore {
    mu: Mutex<SemState>,
    cv: Condvar,
}

struct SemState {
    permits: i64,
    aborted: bool,
}

impl Semaphore {
    fn new() -> Semaphore {
        Semaphore {
            mu: Mutex::new(SemState {
                permits: 0,
                aborted: false,
            }),
            cv: Condvar::new(),
        }
    }

    /// Resets the semaphore to `permits` free slots and clears the abort flag.
    fn reset(&self, permits: i64) {
        let mut s = self.mu.lock().expect("semaphore poisoned");
        s.permits = permits;
        s.aborted = false;
    }

    /// Acquires a slot, blocking until one is free. Returns `false` if aborted.
    fn acquire(&self) -> bool {
        let mut s = self.mu.lock().expect("semaphore poisoned");
        loop {
            if s.aborted {
                return false;
            }
            if s.permits > 0 {
                s.permits -= 1;
                return true;
            }
            s = self.cv.wait(s).expect("semaphore poisoned");
        }
    }

    /// Releases a slot.
    fn release(&self) {
        let mut s = self.mu.lock().expect("semaphore poisoned");
        s.permits += 1;
        self.cv.notify_one();
    }

    /// Aborts the semaphore, waking all waiters. Idempotent.
    fn abort(&self) {
        let mut s = self.mu.lock().expect("semaphore poisoned");
        s.aborted = true;
        self.cv.notify_all();
    }
}

/// The lock-protected configuration and session list.
struct Inner {
    sessions: Vec<Arc<GameSession>>,
    state: SessionState,
    speed: PlaybackSpeed,
    white_diff: Difficulty,
    black_diff: Difficulty,
    white_name: String,
    black_name: String,
    game_count: i32,
    concurrency: i32,
}

/// State shared with the coordinator and game-runner threads.
struct Shared {
    inner: Mutex<Inner>,
    sem: Semaphore,
    active_count: AtomicI32,
}

/// Orchestrates N parallel game sessions.
pub struct SessionManager {
    shared: Arc<Shared>,
}

impl SessionManager {
    /// Creates a new manager configured for the given matchup.
    ///
    /// If `concurrency` is 0 it auto-detects based on CPU count (capped at
    /// `MAX_CONCURRENT_GAMES`). An explicit value is not capped, but has a
    /// minimum of 1.
    pub fn new(
        white_diff: Difficulty,
        black_diff: Difficulty,
        white_name: impl Into<String>,
        black_name: impl Into<String>,
        game_count: i32,
        concurrency: i32,
    ) -> SessionManager {
        let mut effective = concurrency;
        if effective == 0 {
            effective = calculate_default_concurrency();
        }
        if effective < 1 {
            effective = 1;
        }

        SessionManager {
            shared: Arc::new(Shared {
                inner: Mutex::new(Inner {
                    sessions: Vec::new(),
                    state: SessionState::Running,
                    speed: PlaybackSpeed::Normal,
                    white_diff,
                    black_diff,
                    white_name: white_name.into(),
                    black_name: black_name.into(),
                    game_count,
                    concurrency: effective,
                }),
                sem: Semaphore::new(),
                active_count: AtomicI32::new(0),
            }),
        }
    }

    fn lock(&self) -> MutexGuard<'_, Inner> {
        self.shared.inner.lock().expect("manager mutex poisoned")
    }

    /// Creates engine instances for each game and launches them via a
    /// coordinator. Games are started in order with up to `concurrency` running
    /// at once.
    pub fn start(&self) -> Result<(), EngineError> {
        let mut inner = self.lock();

        inner.sessions = Vec::with_capacity(inner.game_count.max(0) as usize);

        // Semaphore size is the smaller of concurrency and game count.
        let mut sem_size = inner.concurrency;
        if inner.game_count < sem_size {
            sem_size = inner.game_count;
        }
        if sem_size < 0 {
            sem_size = 0;
        }
        self.shared.sem.reset(sem_size as i64);

        // Pre-create all sessions and their engines.
        for i in 0..inner.game_count {
            let white_engine = match create_engine(inner.white_diff) {
                Ok(e) => e,
                Err(err) => {
                    abort_sessions(&inner.sessions);
                    return Err(err);
                }
            };
            let black_engine = match create_engine(inner.black_diff) {
                Ok(e) => e,
                Err(err) => {
                    let _ = white_engine.close();
                    abort_sessions(&inner.sessions);
                    return Err(err);
                }
            };

            let speed = inner.speed;
            let session = GameSession::new(
                i + 1,
                white_engine,
                black_engine,
                inner.white_name.clone(),
                inner.black_name.clone(),
                speed,
            );
            inner.sessions.push(session);
        }

        drop(inner);

        // Launch coordinator that starts games in order.
        let shared = self.shared.clone();
        std::thread::spawn(move || coordinate_games(shared));

        Ok(())
    }

    /// Pauses all running sessions.
    pub fn pause(&self) {
        let mut inner = self.lock();
        inner.state = SessionState::Paused;
        for s in &inner.sessions {
            if !s.is_finished() {
                s.pause();
            }
        }
    }

    /// Resumes all paused sessions.
    pub fn resume(&self) {
        let mut inner = self.lock();
        inner.state = SessionState::Running;
        for s in &inner.sessions {
            if s.state() == SessionState::Paused {
                s.resume();
            }
        }
    }

    /// Updates the playback speed for all sessions.
    pub fn set_speed(&self, speed: PlaybackSpeed) {
        let mut inner = self.lock();
        inner.speed = speed;
        for s in &inner.sessions {
            s.set_speed(speed);
        }
    }

    /// Stops all sessions and cleans up.
    pub fn abort(&self) {
        let mut inner = self.lock();
        inner.state = SessionState::Finished;
        self.shared.sem.abort();
        abort_sessions(&inner.sessions);
    }

    /// Stops the manager and cleans up all sessions and their resources.
    ///
    /// Preferred for graceful shutdown: ensures all engines are closed and
    /// resources freed. Idempotent.
    pub fn stop(&self) {
        let mut inner = self.lock();
        inner.state = SessionState::Finished;

        self.shared.sem.abort();
        abort_sessions(&inner.sessions);

        for s in &inner.sessions {
            s.cleanup();
        }

        inner.sessions.clear();
    }

    /// Returns the list of game sessions.
    pub fn sessions(&self) -> Vec<Arc<GameSession>> {
        self.lock().sessions.clone()
    }

    /// Returns true if all sessions have completed.
    pub fn all_finished(&self) -> bool {
        let inner = self.lock();
        if inner.sessions.is_empty() {
            return false;
        }
        inner.sessions.iter().all(|s| s.is_finished())
    }

    /// Returns the current manager state.
    pub fn state(&self) -> SessionState {
        self.lock().state
    }

    /// Returns the current playback speed.
    pub fn speed(&self) -> PlaybackSpeed {
        self.lock().speed
    }

    /// Returns the effective concurrency setting.
    pub fn concurrency(&self) -> i32 {
        self.lock().concurrency
    }

    /// Returns the number of games currently executing.
    pub fn running_count(&self) -> i32 {
        self.shared.active_count.load(Ordering::SeqCst)
    }

    /// Returns the number of games waiting to start.
    pub fn queued_count(&self) -> i32 {
        let inner = self.lock();
        let finished = inner.sessions.iter().filter(|s| s.is_finished()).count() as i32;
        let running = self.shared.active_count.load(Ordering::SeqCst);
        let queued = inner.game_count - finished - running;
        queued.max(0)
    }

    /// Computes aggregate statistics from all finished sessions.
    pub fn stats(&self) -> AggregateStats {
        let inner = self.lock();
        let mut results = Vec::new();
        for s in &inner.sessions {
            if s.is_finished() {
                if let Some(r) = s.result() {
                    results.push(r);
                }
            }
        }
        compute_stats(&results, &inner.white_name, &inner.black_name)
    }

    /// Returns the session at the given index (0-indexed), or `None` if out of
    /// bounds or sessions haven't been created yet.
    pub fn get_session(&self, index: i32) -> Option<Arc<GameSession>> {
        let inner = self.lock();
        if index < 0 || index as usize >= inner.sessions.len() {
            return None;
        }
        Some(inner.sessions[index as usize].clone())
    }

    /// Returns the total number of games configured for this session.
    pub fn game_count(&self) -> i32 {
        self.lock().game_count
    }

    /// Snapshots the current sessions and the white bot name for export.
    pub(crate) fn export_snapshot(&self) -> (Vec<Arc<GameSession>>, String) {
        let inner = self.lock();
        (inner.sessions.clone(), inner.white_name.clone())
    }

    /// Test-only helpers for inspecting internal state.
    #[cfg(test)]
    pub(crate) fn set_speed_field(&self, speed: PlaybackSpeed) {
        self.lock().speed = speed;
    }

    #[cfg(test)]
    pub(crate) fn sessions_is_empty(&self) -> bool {
        self.lock().sessions.is_empty()
    }

    #[cfg(test)]
    pub(crate) fn white_name(&self) -> String {
        self.lock().white_name.clone()
    }
}

/// Aborts all non-finished sessions.
fn abort_sessions(sessions: &[Arc<GameSession>]) {
    for s in sessions {
        if !s.is_finished() {
            s.abort();
        }
    }
}

/// Starts games sequentially as semaphore slots become available, preserving
/// launch order.
fn coordinate_games(shared: Arc<Shared>) {
    let game_count = shared
        .inner
        .lock()
        .expect("manager mutex poisoned")
        .game_count;

    for i in 0..game_count {
        // Wait for a semaphore slot or abort signal.
        if !shared.sem.acquire() {
            return;
        }

        shared.active_count.fetch_add(1, Ordering::SeqCst);
        let sh = shared.clone();
        std::thread::spawn(move || {
            // Safely fetch the session under the lock (races with stop()).
            let session = {
                let inner = sh.inner.lock().expect("manager mutex poisoned");
                inner.sessions.get(i as usize).cloned()
            };

            if let Some(session) = session {
                session.run();
            }

            sh.active_count.fetch_sub(1, Ordering::SeqCst);
            sh.sem.release();
        });
    }
}

/// Creates a bot engine based on difficulty.
fn create_engine(diff: Difficulty) -> Result<SharedEngine, EngineError> {
    match diff {
        Difficulty::Easy => Ok(Arc::new(new_random_engine(&[])?)),
        Difficulty::Medium => Ok(Arc::new(new_minimax_engine(Difficulty::Medium, &[])?)),
        Difficulty::Hard => Ok(Arc::new(new_minimax_engine(Difficulty::Hard, &[])?)),
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::thread;
    use std::time::{Duration, Instant};

    fn wait_all_finished(m: &SessionManager, timeout: Duration) -> bool {
        let start = Instant::now();
        while !m.all_finished() {
            if start.elapsed() > timeout {
                return false;
            }
            thread::sleep(Duration::from_millis(50));
        }
        true
    }

    #[test]
    fn new_session_manager() {
        let m = SessionManager::new(
            Difficulty::Easy,
            Difficulty::Easy,
            "Easy Bot",
            "Easy Bot",
            3,
            0,
        );
        assert_eq!(m.game_count(), 3);
        assert_eq!(m.state(), SessionState::Running);
    }

    #[test]
    fn start_launches_sessions() {
        let m = SessionManager::new(Difficulty::Easy, Difficulty::Easy, "White", "Black", 3, 0);
        m.start().expect("start failed");
        let sessions = m.sessions();
        assert_eq!(sessions.len(), 3);
        m.abort();
    }

    #[test]
    fn all_complete() {
        let m = SessionManager::new(Difficulty::Easy, Difficulty::Easy, "White", "Black", 3, 0);
        m.set_speed_field(PlaybackSpeed::Instant);
        m.start().expect("start failed");

        assert!(
            wait_all_finished(&m, Duration::from_secs(60)),
            "games did not complete within timeout"
        );
        assert!(m.all_finished());

        for s in m.sessions() {
            assert!(s.is_finished());
            assert!(s.result().is_some());
        }
    }

    #[test]
    fn pause_resume() {
        let m = SessionManager::new(Difficulty::Easy, Difficulty::Easy, "White", "Black", 2, 0);
        m.set_speed_field(PlaybackSpeed::Normal);
        m.start().expect("start failed");

        thread::sleep(Duration::from_millis(100));
        m.pause();
        assert_eq!(m.state(), SessionState::Paused);

        thread::sleep(Duration::from_millis(100));
        for s in m.sessions() {
            assert!(s.state() == SessionState::Paused || s.is_finished());
        }

        m.resume();
        assert_eq!(m.state(), SessionState::Running);
        m.abort();
    }

    #[test]
    fn set_speed() {
        let m = SessionManager::new(Difficulty::Easy, Difficulty::Easy, "White", "Black", 2, 0);
        m.set_speed_field(PlaybackSpeed::Normal);
        m.start().expect("start failed");

        m.set_speed(PlaybackSpeed::Instant);
        assert_eq!(m.speed(), PlaybackSpeed::Instant);
        m.abort();
    }

    #[test]
    fn abort_finishes_sessions() {
        let m = SessionManager::new(Difficulty::Easy, Difficulty::Easy, "White", "Black", 3, 0);
        m.set_speed_field(PlaybackSpeed::Normal);
        m.start().expect("start failed");

        thread::sleep(Duration::from_millis(100));
        m.abort();
        thread::sleep(Duration::from_millis(200));

        assert_eq!(m.state(), SessionState::Finished);
        for s in m.sessions() {
            assert!(s.is_finished());
        }
    }

    #[test]
    fn all_finished_false_before_complete() {
        let m = SessionManager::new(Difficulty::Easy, Difficulty::Easy, "White", "Black", 3, 0);
        assert!(!m.all_finished());

        m.set_speed_field(PlaybackSpeed::Normal);
        m.start().expect("start failed");

        thread::sleep(Duration::from_millis(100));
        assert!(!m.all_finished());
        m.abort();
    }

    #[test]
    fn calculate_default_concurrency_with_cpu_tiers() {
        let cases = [
            (1, 1),
            (2, 2),
            (3, 4),
            (4, 6),
            (5, 10),
            (8, 16),
            (16, 32),
            (30, 50),
            (0, 1),
        ];
        for (num_cpu, expected) in cases {
            assert_eq!(
                calculate_default_concurrency_with_cpu(num_cpu),
                expected,
                "cpu={}",
                num_cpu
            );
        }
    }

    #[test]
    fn calculate_default_concurrency_reasonable() {
        let c = calculate_default_concurrency();
        assert!(c >= 1);
        assert!(c <= MAX_CONCURRENT_GAMES);
    }

    #[test]
    fn auto_detect_concurrency() {
        let m = SessionManager::new(Difficulty::Easy, Difficulty::Easy, "White", "Black", 10, 0);
        assert!(m.concurrency() != 0);
        assert!(m.concurrency() >= 1);
        assert!(m.concurrency() <= MAX_CONCURRENT_GAMES);
    }

    #[test]
    fn explicit_concurrency() {
        let m = SessionManager::new(Difficulty::Easy, Difficulty::Easy, "White", "Black", 10, 5);
        assert_eq!(m.concurrency(), 5);
    }

    #[test]
    fn concurrency_no_cap() {
        let m = SessionManager::new(
            Difficulty::Easy,
            Difficulty::Easy,
            "White",
            "Black",
            10,
            100,
        );
        assert_eq!(m.concurrency(), 100);
    }

    #[test]
    fn concurrency_minimum() {
        let m = SessionManager::new(Difficulty::Easy, Difficulty::Easy, "White", "Black", 10, -5);
        assert_eq!(m.concurrency(), 1);
    }

    #[test]
    fn stop_cleans_up() {
        let m = SessionManager::new(Difficulty::Easy, Difficulty::Easy, "White", "Black", 3, 0);
        m.set_speed_field(PlaybackSpeed::Normal);
        m.start().expect("start failed");

        thread::sleep(Duration::from_millis(100));
        m.stop();
        thread::sleep(Duration::from_millis(200));

        assert_eq!(m.state(), SessionState::Finished);
        assert!(m.sessions_is_empty(), "sessions should be empty after stop");
    }

    #[test]
    fn stop_cleans_up_engines() {
        let m = SessionManager::new(Difficulty::Easy, Difficulty::Easy, "White", "Black", 2, 0);
        m.set_speed_field(PlaybackSpeed::Instant);
        m.start().expect("start failed");

        let sessions = m.sessions();

        assert!(
            wait_all_finished(&m, Duration::from_secs(60)),
            "games did not complete within timeout"
        );

        m.stop();

        for s in sessions {
            let inner = s.inner.lock().unwrap();
            assert!(inner.white_engine.is_none());
            assert!(inner.black_engine.is_none());
        }
    }

    #[test]
    fn stop_is_idempotent() {
        let m = SessionManager::new(Difficulty::Easy, Difficulty::Easy, "White", "Black", 2, 0);
        m.set_speed_field(PlaybackSpeed::Normal);
        m.start().expect("start failed");

        thread::sleep(Duration::from_millis(100));
        m.stop();
        m.stop();
        m.stop();

        assert_eq!(m.state(), SessionState::Finished);
    }

    #[test]
    fn get_session() {
        let m = SessionManager::new(Difficulty::Easy, Difficulty::Easy, "White", "Black", 3, 0);
        m.set_speed_field(PlaybackSpeed::Instant);
        m.start().expect("start failed");

        assert!(
            wait_all_finished(&m, Duration::from_secs(60)),
            "games did not complete within timeout"
        );

        for i in 0..3 {
            assert!(m.get_session(i).is_some());
        }
        assert!(m.get_session(-1).is_none());
        assert!(m.get_session(3).is_none());
        assert!(m.get_session(100).is_none());

        m.stop();
    }

    #[test]
    fn get_session_before_start() {
        let m = SessionManager::new(Difficulty::Easy, Difficulty::Easy, "White", "Black", 3, 0);
        assert!(m.get_session(0).is_none());
    }

    #[test]
    fn game_count() {
        for game_count in [1, 3, 10, 50] {
            let m = SessionManager::new(
                Difficulty::Easy,
                Difficulty::Easy,
                "White",
                "Black",
                game_count,
                0,
            );
            assert_eq!(m.game_count(), game_count);
        }
    }

    #[test]
    fn stats_after_completion() {
        let m = SessionManager::new(
            Difficulty::Easy,
            Difficulty::Easy,
            "Easy White",
            "Easy Black",
            3,
            0,
        );
        m.set_speed_field(PlaybackSpeed::Instant);
        m.start().expect("start failed");

        assert!(
            wait_all_finished(&m, Duration::from_secs(60)),
            "games did not complete within timeout"
        );

        let stats = m.stats();
        assert_eq!(stats.total_games, 3);
        assert_eq!(stats.white_bot_name, "Easy White");
        assert_eq!(stats.black_bot_name, "Easy Black");

        assert_eq!(
            stats.white_wins + stats.black_wins + stats.draws,
            stats.total_games
        );
        assert!(stats.avg_move_count > 0.0);
        assert!(stats.avg_duration > Duration::ZERO);
        assert!(stats.shortest_game.move_count > 0);
        assert!(stats.longest_game.move_count > 0);
        assert!(stats.shortest_game.move_count <= stats.longest_game.move_count);
        assert_eq!(stats.individual_results.len(), 3);

        for r in &stats.individual_results {
            assert!(r.move_count > 0);
            assert!(!r.winner.is_empty());
            assert!(!r.end_reason.is_empty());
        }

        // Silence unused warning for the test-only helper.
        let _ = m.white_name();
    }

    #[test]
    fn stats_before_start() {
        let m = SessionManager::new(Difficulty::Easy, Difficulty::Easy, "W", "B", 3, 0);
        let stats = m.stats();
        assert_eq!(stats.total_games, 0);
        assert_eq!(stats.white_bot_name, "W");
        assert_eq!(stats.black_bot_name, "B");
    }
}
