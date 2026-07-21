//! Ported from `engine_test.go`: Engine trait plumbing, optional traits,
//! and the `EngineType` / `Difficulty` / `Info` metadata types.

use std::cell::Cell;
use std::collections::HashMap;

use engine::{Board, Move, PieceType, Square};

use crate::context::Context;
use crate::error::EngineError;
use crate::interfaces::{
    Configurable, Difficulty, Engine, EngineType, Info, Inspectable, MinimaxConfig, Stateful,
};

/// A minimal mock implementation of the [`Engine`] trait.
struct MockEngine {
    name: String,
    closed: Cell<bool>,
}

impl MockEngine {
    fn new(name: &str) -> MockEngine {
        MockEngine {
            name: name.to_string(),
            closed: Cell::new(false),
        }
    }
}

impl Engine for MockEngine {
    fn select_move(&self, _ctx: &Context, _board: &Board) -> Result<Move, EngineError> {
        // Return a simple mock move (e2e4).
        Ok(Move {
            from: Square::new(4, 1),
            to: Square::new(4, 3),
            promotion: PieceType::Empty,
        })
    }

    fn name(&self) -> &str {
        &self.name
    }

    fn close(&self) -> Result<(), EngineError> {
        self.closed.set(true);
        Ok(())
    }
}

struct MockConfigurableEngine {
    base: MockEngine,
    configured: Cell<bool>,
    config: std::cell::RefCell<MinimaxConfig>,
}

impl Engine for MockConfigurableEngine {
    fn select_move(&self, ctx: &Context, board: &Board) -> Result<Move, EngineError> {
        self.base.select_move(ctx, board)
    }
    fn name(&self) -> &str {
        self.base.name()
    }
    fn close(&self) -> Result<(), EngineError> {
        self.base.close()
    }
}

impl Configurable for MockConfigurableEngine {
    fn configure(&mut self, config: MinimaxConfig) -> Result<(), EngineError> {
        self.configured.set(true);
        *self.config.borrow_mut() = config;
        Ok(())
    }
}

struct MockStatefulEngine {
    base: MockEngine,
    history: Vec<Board>,
}

impl Engine for MockStatefulEngine {
    fn select_move(&self, ctx: &Context, board: &Board) -> Result<Move, EngineError> {
        self.base.select_move(ctx, board)
    }
    fn name(&self) -> &str {
        self.base.name()
    }
    fn close(&self) -> Result<(), EngineError> {
        self.base.close()
    }
}

impl Stateful for MockStatefulEngine {
    fn set_position_history(&mut self, history: Vec<Board>) -> Result<(), EngineError> {
        self.history = history;
        Ok(())
    }
}

struct MockInspectableEngine {
    base: MockEngine,
    info: Info,
}

impl Engine for MockInspectableEngine {
    fn select_move(&self, ctx: &Context, board: &Board) -> Result<Move, EngineError> {
        self.base.select_move(ctx, board)
    }
    fn name(&self) -> &str {
        self.base.name()
    }
    fn close(&self) -> Result<(), EngineError> {
        self.base.close()
    }
}

impl Inspectable for MockInspectableEngine {
    fn info(&self) -> Info {
        self.info.clone()
    }
}

#[test]
fn engine_interface() {
    let mock = MockEngine::new("TestEngine");

    assert_eq!(mock.name(), "TestEngine");

    let board = Board::new();
    let ctx = Context::background();
    let mv = mock.select_move(&ctx, &board).expect("select_move");
    assert!(mv.from.is_valid() && mv.to.is_valid());

    assert!(!mock.closed.get());
    mock.close().expect("close");
    assert!(mock.closed.get());
}

#[test]
fn optional_interfaces_configurable() {
    let mut eng = MockConfigurableEngine {
        base: MockEngine::new("ConfigurableEngine"),
        configured: Cell::new(false),
        config: std::cell::RefCell::new(MinimaxConfig::default()),
    };

    let config = MinimaxConfig {
        search_depth: Some(5),
        ..MinimaxConfig::default()
    };
    eng.configure(config).expect("configure");

    assert!(eng.configured.get());
    assert_eq!(eng.config.borrow().search_depth, Some(5));
}

#[test]
fn optional_interfaces_stateful() {
    let mut eng = MockStatefulEngine {
        base: MockEngine::new("StatefulEngine"),
        history: Vec::new(),
    };

    let history = vec![Board::new()];
    eng.set_position_history(history).expect("set history");
    assert_eq!(eng.history.len(), 1);
}

#[test]
fn optional_interfaces_inspectable() {
    let mut features = HashMap::new();
    features.insert("analysis".to_string(), true);
    let expected = Info {
        name: "InspectableEngine".to_string(),
        author: "Test Author".to_string(),
        version: "1.0.0".to_string(),
        engine_type: EngineType::Internal,
        difficulty: Difficulty::Medium,
        features,
    };

    let eng = MockInspectableEngine {
        base: MockEngine::new("InspectableEngine"),
        info: expected.clone(),
    };

    let info = eng.info();
    assert_eq!(info.name, expected.name);
    assert_eq!(info.author, expected.author);
    assert_eq!(info.version, expected.version);
    assert_eq!(info.engine_type, expected.engine_type);
    assert_eq!(info.difficulty, expected.difficulty);
}

#[test]
fn engine_type_constants() {
    assert_eq!(EngineType::Internal.to_string(), "Internal");
    assert_eq!(EngineType::Uci.to_string(), "UCI");
    assert_eq!(EngineType::Rl.to_string(), "RL");

    // Discriminant ordering (Go iota: Internal=0, UCI=1, RL=2).
    assert_eq!(EngineType::Internal as i32, 0);
    assert_eq!(EngineType::Uci as i32, 1);
    assert_eq!(EngineType::Rl as i32, 2);
}

#[test]
fn difficulty_constants() {
    assert_eq!(Difficulty::Easy.to_string(), "Easy");
    assert_eq!(Difficulty::Medium.to_string(), "Medium");
    assert_eq!(Difficulty::Hard.to_string(), "Hard");

    assert_eq!(Difficulty::Easy as i32, 0);
    assert_eq!(Difficulty::Medium as i32, 1);
    assert_eq!(Difficulty::Hard as i32, 2);
}

#[test]
fn info_struct() {
    let mut features = HashMap::new();
    features.insert("opening_book".to_string(), true);
    features.insert("endgame_tb".to_string(), false);
    let info = Info {
        name: "TestBot".to_string(),
        author: "Test Author".to_string(),
        version: "1.0.0".to_string(),
        engine_type: EngineType::Internal,
        difficulty: Difficulty::Hard,
        features,
    };

    assert_eq!(info.name, "TestBot");
    assert_eq!(info.author, "Test Author");
    assert_eq!(info.version, "1.0.0");
    assert_eq!(info.engine_type, EngineType::Internal);
    assert_eq!(info.difficulty, Difficulty::Hard);
    assert!(info.features["opening_book"]);
    assert!(!info.features["endgame_tb"]);
}

#[test]
fn engine_close_idempotency() {
    let mock = MockEngine::new("TestEngine");
    assert!(mock.close().is_ok());
    assert!(mock.close().is_ok());
    assert!(mock.close().is_ok());
}

#[test]
fn engine_with_context() {
    let mock = MockEngine::new("TestEngine");
    let board = Board::new();

    // Valid context.
    assert!(mock.select_move(&Context::background(), &board).is_ok());

    // Cancelled context: the mock does not check it, mirroring the Go mock.
    let ctx = Context::background();
    ctx.cancel();
    let _ = mock.select_move(&ctx, &board);
}
