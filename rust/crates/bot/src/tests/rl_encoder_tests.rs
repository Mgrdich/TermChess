//! Ported from `rl_encoder_test.go`: the 66-channel board encoder.

use engine::{Board, Color, Move, Piece, PieceType, Square};

use crate::rl_encoder::{encode_board, ENCODING_SIZE};

fn at(encoded: &[f32], ch: usize, rank: usize, file: usize) -> f32 {
    encoded[ch * 64 + rank * 8 + file]
}

#[test]
fn shape() {
    let board = Board::new();
    let encoded = encode_board(&board, &[]);
    assert_eq!(encoded.len(), ENCODING_SIZE);
}

#[test]
fn starting_position() {
    let board = Board::new();
    let encoded = encode_board(&board, &[]);

    // Channel 0: white pawns at rank 1.
    for file in 0..8 {
        assert_eq!(at(&encoded, 0, 1, file), 1.0);
        for rank in 0..8 {
            if rank == 1 {
                continue;
            }
            assert_eq!(at(&encoded, 0, rank, file), 0.0);
        }
    }

    // Channel 5: white king at e1.
    assert_eq!(at(&encoded, 5, 0, 4), 1.0);
    // Channel 6: black pawns at rank 6.
    for file in 0..8 {
        assert_eq!(at(&encoded, 6, 6, file), 1.0);
    }
    // Channel 11: black king at e8.
    assert_eq!(at(&encoded, 11, 7, 4), 1.0);

    // Channel 12: side to move all 1.0.
    for i in 0..64 {
        assert_eq!(encoded[12 * 64 + i], 1.0);
    }
    // Channels 13-16: all castling rights.
    for ch in 13..=16 {
        for i in 0..64 {
            assert_eq!(encoded[ch * 64 + i], 1.0);
        }
    }
    // Channel 17: no en passant.
    for i in 0..64 {
        assert_eq!(encoded[17 * 64 + i], 0.0);
    }
}

#[test]
fn after_e4() {
    let mut board = Board::new();
    let e2 = Square::new(4, 1);
    let e4 = Square::new(4, 3);
    board.make_move(Move::new(e2, e4)).expect("make e2e4");

    let encoded = encode_board(&board, &[]);

    assert_eq!(at(&encoded, 0, 1, 4), 0.0);
    assert_eq!(at(&encoded, 0, 3, 4), 1.0);

    // Side to move: black -> all 0.0.
    for i in 0..64 {
        assert_eq!(encoded[12 * 64 + i], 0.0);
    }

    // En passant on file 4 for all ranks.
    for rank in 0..8 {
        assert_eq!(at(&encoded, 17, rank, 4), 1.0);
    }
    for file in 0..8 {
        if file == 4 {
            continue;
        }
        for rank in 0..8 {
            assert_eq!(at(&encoded, 17, rank, file), 0.0);
        }
    }

    // Castling rights unchanged.
    for ch in 13..=16 {
        for i in 0..64 {
            assert_eq!(encoded[ch * 64 + i], 1.0);
        }
    }
}

#[test]
fn python_reference() {
    let board = Board::new();
    let encoded = encode_board(&board, &[]);

    assert_eq!(encoded[192], 1.0); // white rook a1
    assert_eq!(encoded[199], 1.0); // white rook h1
    assert_eq!(encoded[699], 1.0); // black queen d8
    assert_eq!(encoded[259], 1.0); // white queen d1
    assert_eq!(encoded[65], 1.0); // white knight b1
    assert_eq!(encoded[70], 1.0); // white knight g1
    assert_eq!(encoded[505], 1.0); // black knight b8
    assert_eq!(encoded[130], 1.0); // white bishop c1
    assert_eq!(encoded[573], 1.0); // black bishop f8
    assert_eq!(encoded[764], 1.0); // black king e8
    assert_eq!(encoded[768], 1.0); // side-to-move plane start

    // e4 empty in all piece channels.
    for ch in 0..12 {
        assert_eq!(encoded[ch * 64 + 3 * 8 + 4], 0.0);
    }

    // Total 32 pieces.
    let piece_count: f32 = encoded[..12 * 64].iter().sum();
    assert_eq!(piece_count, 32.0);
}

#[test]
fn no_castling_rights() {
    let mut board = Board::new();
    board.castling_rights = 0;
    let encoded = encode_board(&board, &[]);
    for ch in 13..=16 {
        for i in 0..64 {
            assert_eq!(encoded[ch * 64 + i], 0.0);
        }
    }
}

#[test]
fn partial_castling_rights() {
    let mut board = Board::new();
    board.castling_rights = engine::CASTLE_WHITE_KING | engine::CASTLE_BLACK_QUEEN;
    let encoded = encode_board(&board, &[]);

    for i in 0..64 {
        assert_eq!(encoded[13 * 64 + i], 1.0); // WK
        assert_eq!(encoded[14 * 64 + i], 0.0); // WQ
        assert_eq!(encoded[15 * 64 + i], 0.0); // BK
        assert_eq!(encoded[16 * 64 + i], 1.0); // BQ
    }
}

#[test]
fn en_passant() {
    let mut board = Board::new();
    board.en_passant_sq = Square::new(4, 2).0; // e3
    let encoded = encode_board(&board, &[]);

    for rank in 0..8 {
        assert_eq!(encoded[17 * 64 + rank * 8 + 4], 1.0);
    }
    for file in 0..8 {
        if file == 4 {
            continue;
        }
        for rank in 0..8 {
            assert_eq!(encoded[17 * 64 + rank * 8 + file], 0.0);
        }
    }
}

#[test]
fn en_passant_file_a() {
    let mut board = Board::new();
    board.en_passant_sq = Square::new(0, 2).0; // a3
    let encoded = encode_board(&board, &[]);

    for rank in 0..8 {
        assert_eq!(encoded[17 * 64 + rank * 8], 1.0);
        assert_eq!(encoded[17 * 64 + rank * 8 + 1], 0.0);
    }
}

#[test]
fn black_to_move() {
    let mut board = Board::new();
    board.active_color = Color::Black;
    let encoded = encode_board(&board, &[]);
    for i in 0..64 {
        assert_eq!(encoded[12 * 64 + i], 0.0);
    }
}

#[test]
fn empty_board() {
    let mut board = Board {
        active_color: Color::White,
        castling_rights: 0,
        en_passant_sq: -1,
        ..Board::default()
    };
    board.squares[Square::new(4, 0).index()] = Piece::new(Color::White, PieceType::King);
    board.squares[Square::new(4, 7).index()] = Piece::new(Color::Black, PieceType::King);

    let encoded = encode_board(&board, &[]);

    let piece_count: f32 = encoded[..12 * 64].iter().sum();
    assert_eq!(piece_count, 2.0);
    assert_eq!(encoded[5 * 64 + 4], 1.0); // white king e1
    assert_eq!(encoded[11 * 64 + 7 * 8 + 4], 1.0); // black king e8

    for ch in 13..=17 {
        for i in 0..64 {
            assert_eq!(encoded[ch * 64 + i], 0.0);
        }
    }
}
