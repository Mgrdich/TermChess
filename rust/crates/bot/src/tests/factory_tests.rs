//! Ported from `factory_test.go`: functional options and engine factories.

use std::collections::HashMap;
use std::time::Duration;

use crate::factory::{
    new_minimax_engine, new_random_engine, with_options, with_search_depth, with_time_limit,
    EngineConfig, OptionValue,
};
use crate::interfaces::{Difficulty, Engine};

#[test]
fn with_time_limit_valid() {
    let mut cfg = EngineConfig::default();
    let opt = with_time_limit(Duration::from_secs(5));
    opt(&mut cfg).expect("apply");
    assert_eq!(cfg.time_limit, Duration::from_secs(5));
}

#[test]
fn with_time_limit_zero() {
    let mut cfg = EngineConfig::default();
    let opt = with_time_limit(Duration::ZERO);
    let err = opt(&mut cfg).unwrap_err();
    assert!(err.to_string().contains("positive"));
}

#[test]
fn with_search_depth_valid() {
    for depth in [1, 5, 10, 15, 20] {
        let mut cfg = EngineConfig::default();
        let opt = with_search_depth(depth);
        opt(&mut cfg).expect("apply");
        assert_eq!(cfg.search_depth, depth);
    }
}

#[test]
fn with_search_depth_zero() {
    let mut cfg = EngineConfig::default();
    let err = with_search_depth(0)(&mut cfg).unwrap_err();
    assert!(err.to_string().contains("1-20"));
}

#[test]
fn with_search_depth_too_high() {
    let mut cfg = EngineConfig::default();
    let err = with_search_depth(21)(&mut cfg).unwrap_err();
    assert!(err.to_string().contains("1-20"));
}

#[test]
fn with_search_depth_negative() {
    let mut cfg = EngineConfig::default();
    let err = with_search_depth(-5)(&mut cfg).unwrap_err();
    assert!(err.to_string().contains("1-20"));
}

#[test]
fn with_options_valid() {
    let mut cfg = EngineConfig::default();
    let mut opts = HashMap::new();
    opts.insert("threads".to_string(), OptionValue::Int(4));
    opts.insert("hash".to_string(), OptionValue::Int(256));
    opts.insert("opening_book".to_string(), OptionValue::Bool(true));

    with_options(opts)(&mut cfg).expect("apply");
    let stored = cfg.options.expect("options set");
    assert_eq!(stored["threads"], OptionValue::Int(4));
    assert_eq!(stored["hash"], OptionValue::Int(256));
    assert_eq!(stored["opening_book"], OptionValue::Bool(true));
}

#[test]
fn with_options_empty() {
    let mut cfg = EngineConfig::default();
    with_options(HashMap::new())(&mut cfg).expect("apply");
    assert!(cfg.options.is_some());
}

#[test]
fn new_random_engine_default_config() {
    let eng = new_random_engine(&[]).expect("create");
    assert_eq!(eng.name(), "Easy Bot");
    eng.close().ok();
}

#[test]
fn new_random_engine_custom_time_limit() {
    let eng = new_random_engine(&[with_time_limit(Duration::from_secs(3))]).expect("create");
    assert_eq!(eng.time_limit, Duration::from_secs(3));
    eng.close().ok();
}

#[test]
fn new_random_engine_invalid_time_limit() {
    let err = new_random_engine(&[with_time_limit(Duration::ZERO)]).unwrap_err();
    assert!(err.to_string().contains("positive"));
}

#[test]
fn new_random_engine_custom_search_depth_ignored() {
    // Random engine doesn't use search depth, but should accept it.
    let eng = new_random_engine(&[with_search_depth(5)]).expect("create");
    eng.close().ok();
}

#[test]
fn new_minimax_engine_medium() {
    let eng = new_minimax_engine(Difficulty::Medium, &[]).expect("create");
    assert_eq!(eng.name(), "Medium Bot");
    assert_eq!(eng.difficulty, Difficulty::Medium);
    assert_eq!(eng.max_depth, 4);
    assert_eq!(eng.time_limit, Duration::from_secs(4));
    eng.close().ok();
}

#[test]
fn new_minimax_engine_hard() {
    let eng = new_minimax_engine(Difficulty::Hard, &[]).expect("create");
    assert_eq!(eng.name(), "Hard Bot");
    assert_eq!(eng.difficulty, Difficulty::Hard);
    assert_eq!(eng.max_depth, 7);
    assert_eq!(eng.time_limit, Duration::from_secs(8));
    eng.close().ok();
}

#[test]
fn new_minimax_engine_easy_invalid() {
    let err = new_minimax_engine(Difficulty::Easy, &[]).unwrap_err();
    assert!(err.to_string().contains("invalid difficulty"));
}

#[test]
fn new_minimax_engine_custom_search_depth() {
    let eng = new_minimax_engine(Difficulty::Medium, &[with_search_depth(8)]).expect("create");
    assert_eq!(eng.max_depth, 8);
    eng.close().ok();
}

#[test]
fn new_minimax_engine_custom_time_limit() {
    let eng = new_minimax_engine(
        Difficulty::Hard,
        &[with_time_limit(Duration::from_secs(10))],
    )
    .expect("create");
    assert_eq!(eng.time_limit, Duration::from_secs(10));
    eng.close().ok();
}

#[test]
fn new_minimax_engine_invalid_search_depth() {
    let err = new_minimax_engine(Difficulty::Medium, &[with_search_depth(0)]).unwrap_err();
    assert!(err.to_string().contains("1-20"));
}

#[test]
fn new_minimax_engine_invalid_time_limit() {
    let err = new_minimax_engine(Difficulty::Hard, &[with_time_limit(Duration::ZERO)]).unwrap_err();
    assert!(err.to_string().contains("positive"));
}

#[test]
fn new_minimax_engine_multiple_options() {
    let mut opts = HashMap::new();
    opts.insert("transposition_table".to_string(), OptionValue::Bool(true));
    let eng = new_minimax_engine(
        Difficulty::Hard,
        &[
            with_time_limit(Duration::from_secs(5)),
            with_search_depth(10),
            with_options(opts),
        ],
    )
    .expect("create");
    assert_eq!(eng.time_limit, Duration::from_secs(5));
    assert_eq!(eng.max_depth, 10);
    eng.close().ok();
}

#[test]
fn engine_option_chaining() {
    let mut cfg = EngineConfig::default();
    let mut opts_map = HashMap::new();
    opts_map.insert("key".to_string(), OptionValue::Str("value".to_string()));
    let options = [
        with_time_limit(Duration::from_secs(5)),
        with_search_depth(10),
        with_options(opts_map),
    ];
    for opt in &options {
        opt(&mut cfg).expect("apply");
    }
    assert_eq!(cfg.time_limit, Duration::from_secs(5));
    assert_eq!(cfg.search_depth, 10);
    assert_eq!(
        cfg.options.unwrap()["key"],
        OptionValue::Str("value".to_string())
    );
}

#[test]
fn engine_option_overrides() {
    let mut cfg = EngineConfig::default();
    with_time_limit(Duration::from_secs(3))(&mut cfg).expect("apply");
    with_time_limit(Duration::from_secs(7))(&mut cfg).expect("apply");
    assert_eq!(cfg.time_limit, Duration::from_secs(7));
}
