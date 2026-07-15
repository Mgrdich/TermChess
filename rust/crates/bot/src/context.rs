//! A minimal cancellation/deadline context, mirroring Go's `context.Context`
//! as used by the bot engines (only `Done`/`Err` semantics are needed here).

use std::sync::atomic::{AtomicBool, Ordering};
use std::sync::Arc;
use std::time::{Duration, Instant};

use crate::error::EngineError;

/// A cancellation and deadline handle passed to [`crate::Engine::select_move`].
///
/// Cloning shares the same underlying cancellation flag, so a clone can be used
/// to cancel a context handed to another call (mirroring `context.WithCancel`).
#[derive(Clone)]
pub struct Context {
    inner: Arc<Inner>,
}

struct Inner {
    deadline: Option<Instant>,
    canceled: AtomicBool,
}

impl Context {
    /// Returns a context that is never canceled and has no deadline
    /// (`context.Background()`).
    pub fn background() -> Context {
        Context {
            inner: Arc::new(Inner {
                deadline: None,
                canceled: AtomicBool::new(false),
            }),
        }
    }

    /// Returns a context that is `done` once `timeout` elapses
    /// (`context.WithTimeout`).
    pub fn with_timeout(timeout: Duration) -> Context {
        Context {
            inner: Arc::new(Inner {
                deadline: Some(Instant::now() + timeout),
                canceled: AtomicBool::new(false),
            }),
        }
    }

    /// Marks the context as canceled (the cancel func from
    /// `context.WithCancel`). Idempotent.
    pub fn cancel(&self) {
        self.inner.canceled.store(true, Ordering::SeqCst);
    }

    /// Reports whether the context has been canceled or its deadline passed.
    pub fn is_done(&self) -> bool {
        if self.inner.canceled.load(Ordering::SeqCst) {
            return true;
        }
        matches!(self.inner.deadline, Some(d) if Instant::now() >= d)
    }

    /// Returns the reason the context is done, or `None` if it is not.
    ///
    /// Cancellation takes precedence over deadline expiry, matching Go where an
    /// explicit cancel reports `context.Canceled`.
    pub fn err(&self) -> Option<EngineError> {
        if self.inner.canceled.load(Ordering::SeqCst) {
            return Some(EngineError::ContextCanceled);
        }
        match self.inner.deadline {
            Some(d) if Instant::now() >= d => Some(EngineError::ContextDeadlineExceeded),
            _ => None,
        }
    }
}

impl Default for Context {
    fn default() -> Self {
        Context::background()
    }
}
