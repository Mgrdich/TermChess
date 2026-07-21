//! Board encoder for the RL agent (ported from `rl_encoder.go`).
//!
//! Produces a flat `[channel, rank, file]` float32 tensor that must stay
//! byte-identical to the Python training encoder (`training/board_encoder.py`).

use engine::{Board, Color};

const PIECE_CHANNELS: usize = 12;
const CURRENT_POS_CHANNELS: usize = 18;
const NUM_HISTORY_POSITIONS: usize = 4;
/// Total channels: 18 current-position channels + 4 history positions x 12 = 66.
pub(crate) const NUM_CHANNELS: usize =
    CURRENT_POS_CHANNELS + NUM_HISTORY_POSITIONS * PIECE_CHANNELS;
const BOARD_SIZE: usize = 8;
/// Flat encoding length: 66 * 8 * 8 = 4224.
pub(crate) const ENCODING_SIZE: usize = NUM_CHANNELS * BOARD_SIZE * BOARD_SIZE;
/// Policy vector size: 64 * 64 from-to squares = 4096.
pub(crate) const POLICY_SIZE: usize = 4096;

/// Converts a board position and its history into a 66-channel float32 tensor.
///
/// Returns a flat `Vec<f32>` of length 4224 in `[channel, rank, file]` order.
///
/// Channel layout:
/// - 0-5:   White pieces (Pawn..King) — current position
/// - 6-11:  Black pieces — current position
/// - 12:    Side to move (1.0 if White to move)
/// - 13-16: Castling rights (WK, WQ, BK, BQ)
/// - 17:    En passant file (column filled with 1.0 if available)
/// - 18-65: History piece planes (12 channels each), most recent first
pub fn encode_board(board: &Board, history: &[Board]) -> Vec<f32> {
    let mut encoding = vec![0.0f32; ENCODING_SIZE];

    // Channels 0-11: current position piece placement.
    encode_pieces(&mut encoding, board, 0);

    // Channel 12: side to move.
    if board.active_color == Color::White {
        fill_plane(&mut encoding, 12);
    }

    // Channels 13-16: castling rights.
    if board.castling_rights & engine::CASTLE_WHITE_KING != 0 {
        fill_plane(&mut encoding, 13);
    }
    if board.castling_rights & engine::CASTLE_WHITE_QUEEN != 0 {
        fill_plane(&mut encoding, 14);
    }
    if board.castling_rights & engine::CASTLE_BLACK_KING != 0 {
        fill_plane(&mut encoding, 15);
    }
    if board.castling_rights & engine::CASTLE_BLACK_QUEEN != 0 {
        fill_plane(&mut encoding, 16);
    }

    // Channel 17: en passant.
    if board.en_passant_sq >= 0 {
        let ep_file = (board.en_passant_sq as usize) % 8;
        for rank in 0..8 {
            encoding[17 * 64 + rank * 8 + ep_file] = 1.0;
        }
    }

    // Channels 18+: history position piece planes (12 channels each), most
    // recent history position first.
    if !history.is_empty() {
        let start = history.len().saturating_sub(NUM_HISTORY_POSITIONS);
        let hist_slice = &history[start..];

        for i in 0..hist_slice.len() {
            // i = 0 (reversed order) is the most recent history position.
            let hist_idx = hist_slice.len() - 1 - i;
            let channel_offset = CURRENT_POS_CHANNELS + i * PIECE_CHANNELS;
            encode_pieces(&mut encoding, &hist_slice[hist_idx], channel_offset);
        }
    }

    encoding
}

/// Writes piece placement data into 12 channels starting at `channel_offset`.
fn encode_pieces(encoding: &mut [f32], board: &Board, channel_offset: usize) {
    for sq in 0..64usize {
        let piece = board.squares[sq];
        if piece.is_empty() {
            continue;
        }

        let pt = piece.piece_type(); // 1=Pawn ... 6=King
        let piece_idx = (pt.as_u8() as usize) - 1;

        let channel = if piece.color() == Color::White {
            channel_offset + piece_idx
        } else {
            channel_offset + 6 + piece_idx
        };

        let rank = sq / 8;
        let file = sq % 8;
        encoding[channel * 64 + rank * 8 + file] = 1.0;
    }
}

/// Sets all 64 values in the given channel to 1.0.
fn fill_plane(encoding: &mut [f32], channel: usize) {
    let start = channel * 64;
    for v in encoding.iter_mut().skip(start).take(64) {
        *v = 1.0;
    }
}
