//! Request context providing deadline and cancellation, mirroring the
//! `context.Context` values threaded through the Go updater's HTTP calls.
//!
//! Rust's HTTP client here is blocking, so cancellation is modelled with an
//! atomic flag that is checked before a request is issued, and deadlines are
//! translated into per-request timeouts.

use std::sync::atomic::{AtomicBool, Ordering};
use std::sync::Arc;
use std::time::{Duration, Instant};

/// A cancellation and deadline carrier, analogous to Go's `context.Context`.
#[derive(Clone)]
pub struct Context {
    deadline: Option<Instant>,
    cancelled: Arc<AtomicBool>,
}

/// Handle returned by [`Context::with_cancel`] used to cancel the context,
/// analogous to Go's `CancelFunc`.
pub struct CancelHandle {
    cancelled: Arc<AtomicBool>,
}

impl CancelHandle {
    /// Cancels the associated context.
    pub fn cancel(&self) {
        self.cancelled.store(true, Ordering::SeqCst);
    }
}

impl Context {
    /// Returns a context with no deadline and no cancellation, equivalent to
    /// `context.Background()`.
    pub fn background() -> Self {
        Self {
            deadline: None,
            cancelled: Arc::new(AtomicBool::new(false)),
        }
    }

    /// Returns a context that expires after `timeout`, equivalent to
    /// `context.WithTimeout`.
    pub fn with_timeout(timeout: Duration) -> Self {
        Self {
            deadline: Some(Instant::now() + timeout),
            cancelled: Arc::new(AtomicBool::new(false)),
        }
    }

    /// Returns a cancellable context together with a handle to cancel it,
    /// equivalent to `context.WithCancel`.
    pub fn with_cancel() -> (Self, CancelHandle) {
        let cancelled = Arc::new(AtomicBool::new(false));
        (
            Self {
                deadline: None,
                cancelled: cancelled.clone(),
            },
            CancelHandle { cancelled },
        )
    }

    /// Reports whether the context has been cancelled.
    pub fn is_cancelled(&self) -> bool {
        self.cancelled.load(Ordering::SeqCst)
    }

    /// Returns the time remaining until the deadline, or `None` if there is no
    /// deadline. A zero duration means the deadline has already passed.
    pub fn remaining(&self) -> Option<Duration> {
        self.deadline
            .map(|d| d.saturating_duration_since(Instant::now()))
    }
}

impl Default for Context {
    fn default() -> Self {
        Self::background()
    }
}
