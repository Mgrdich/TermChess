//! Standard Algebraic Notation parsing and formatting (port of Go `san.go`).

use engine::{Board, Color, Move, PieceType, Square, NO_SQUARE};

/// Parses SAN into a `Move` against the given board.
///
/// Supports pawn moves, piece moves with disambiguation, captures, castling,
/// and promotions. Returns `Err(message)` on failure.
pub fn parse_san(b: &Board, san: &str) -> Result<Move, String> {
    if san.is_empty() {
        return Err("empty move notation".to_string());
    }

    // Strip trailing check/checkmate symbols.
    let mut s = san;
    if let Some(stripped) = s.strip_suffix('+') {
        s = stripped;
    }
    if let Some(stripped) = s.strip_suffix('#') {
        s = stripped;
    }

    if s.is_empty() {
        return Err("invalid move notation".to_string());
    }

    if s == "O-O" || s == "0-0" {
        return parse_castling(b, true);
    }
    if s == "O-O-O" || s == "0-0-0" {
        return parse_castling(b, false);
    }

    let first = s.as_bytes()[0] as char;
    if matches!(first, 'K' | 'Q' | 'R' | 'B' | 'N') {
        return parse_piece_move(b, s);
    }

    parse_pawn_move(b, s)
}

fn parse_pawn_move(b: &Board, san: &str) -> Result<Move, String> {
    let mut promotion = PieceType::Empty;
    let mut move_str = san.to_string();

    if san.contains('=') {
        let parts: Vec<&str> = san.split('=').collect();
        if parts.len() != 2 {
            return Err(format!("invalid promotion format: {}", san));
        }
        move_str = parts[0].to_string();
        promotion = parse_promotion(parts[1])?;
    }

    let is_capture = move_str.contains('x');
    let mut source_file: i32 = -1;
    let dest_square: Square;

    if is_capture {
        let parts: Vec<&str> = move_str.split('x').collect();
        if parts.len() != 2 {
            return Err(format!("invalid capture format: {}", san));
        }
        if parts[0].len() != 1 {
            return Err(format!("invalid source file in capture: {}", san));
        }
        source_file = parse_file(parts[0].as_bytes()[0] as char)
            .map_err(|e| format!("invalid source file: {}", e))?;
        dest_square =
            parse_square(parts[1]).map_err(|e| format!("invalid destination square: {}", e))?;
    } else {
        dest_square =
            parse_square(&move_str).map_err(|e| format!("invalid destination square: {}", e))?;
    }

    let mut candidates: Vec<Move> = Vec::new();
    for mv in b.legal_moves() {
        let piece = b.piece_at(mv.from);
        if piece.piece_type() != PieceType::Pawn {
            continue;
        }
        if mv.to != dest_square {
            continue;
        }
        if is_capture && mv.from.file() != source_file {
            continue;
        }
        if promotion != PieceType::Empty && mv.promotion != promotion {
            continue;
        }
        candidates.push(mv);
    }

    if candidates.is_empty() {
        return Err(format!("no legal pawn move matches: {}", san));
    }
    if candidates.len() > 1 {
        return Err(format!(
            "ambiguous pawn move: {} (multiple candidates)",
            san
        ));
    }
    Ok(candidates[0])
}

fn parse_square(s: &str) -> Result<Square, String> {
    if s.len() != 2 {
        return Err(format!(
            "invalid square notation: {} (expected 2 characters)",
            s
        ));
    }
    let bytes = s.as_bytes();
    let file = bytes[0] as i32 - b'a' as i32;
    let rank = bytes[1] as i32 - b'1' as i32;
    if !(0..=7).contains(&file) {
        return Err(format!("invalid file: {} (expected a-h)", bytes[0] as char));
    }
    if !(0..=7).contains(&rank) {
        return Err(format!("invalid rank: {} (expected 1-8)", bytes[1] as char));
    }
    Ok(Square::new(file, rank))
}

fn parse_promotion(s: &str) -> Result<PieceType, String> {
    if s.len() != 1 {
        return Err(format!("invalid promotion piece: {}", s));
    }
    match s.as_bytes()[0].to_ascii_uppercase() {
        b'Q' => Ok(PieceType::Queen),
        b'R' => Ok(PieceType::Rook),
        b'B' => Ok(PieceType::Bishop),
        b'N' => Ok(PieceType::Knight),
        _ => Err(format!(
            "invalid promotion piece: {} (expected Q, R, B, or N)",
            s
        )),
    }
}

fn parse_file(r: char) -> Result<i32, String> {
    let file = r as i32 - 'a' as i32;
    if !(0..=7).contains(&file) {
        return Err(format!("invalid file: {} (expected a-h)", r));
    }
    Ok(file)
}

fn parse_piece_move(b: &Board, san: &str) -> Result<Move, String> {
    if san.len() < 2 {
        return Err(format!("invalid piece move format: {}", san));
    }

    let piece_type = parse_piece_type(san.as_bytes()[0] as char)?;
    let move_str = &san[1..];

    let mut from_file: i32 = -1;
    let mut from_rank: i32 = -1;

    let disambiguation_part: &str;
    let remaining_part: &str;

    if let Some(capture_idx) = move_str.find('x') {
        disambiguation_part = &move_str[..capture_idx];
        remaining_part = &move_str[capture_idx + 1..];
    } else if move_str.len() > 2 {
        disambiguation_part = &move_str[..move_str.len() - 2];
        remaining_part = &move_str[move_str.len() - 2..];
    } else {
        disambiguation_part = "";
        remaining_part = move_str;
    }

    for ch in disambiguation_part.bytes() {
        if (b'a'..=b'h').contains(&ch) {
            from_file = (ch - b'a') as i32;
        } else if (b'1'..=b'8').contains(&ch) {
            from_rank = (ch - b'1') as i32;
        }
    }

    if remaining_part.len() != 2 {
        return Err(format!("invalid piece move format: {}", san));
    }

    let dest_square =
        parse_square(remaining_part).map_err(|e| format!("invalid destination square: {}", e))?;

    let mut candidates: Vec<Move> = Vec::new();
    for mv in b.legal_moves() {
        let piece = b.piece_at(mv.from);
        if piece.piece_type() != piece_type {
            continue;
        }
        if mv.to != dest_square {
            continue;
        }
        if from_file >= 0 && mv.from.file() != from_file {
            continue;
        }
        if from_rank >= 0 && mv.from.rank() != from_rank {
            continue;
        }
        candidates.push(mv);
    }

    if candidates.is_empty() {
        return Err(format!("no legal move matches: {}", san));
    }
    if candidates.len() > 1 {
        return Err(format!(
            "move is still ambiguous: {} (multiple candidates)",
            san
        ));
    }
    Ok(candidates[0])
}

fn parse_castling(b: &Board, kingside: bool) -> Result<Move, String> {
    let (king_from, king_to) = if b.active_color == Color::White {
        (
            Square::new(4, 0),
            if kingside {
                Square::new(6, 0)
            } else {
                Square::new(2, 0)
            },
        )
    } else {
        (
            Square::new(4, 7),
            if kingside {
                Square::new(6, 7)
            } else {
                Square::new(2, 7)
            },
        )
    };

    for mv in b.legal_moves() {
        if mv.from == king_from && mv.to == king_to {
            return Ok(mv);
        }
    }

    if kingside {
        Err("kingside castling is not legal".to_string())
    } else {
        Err("queenside castling is not legal".to_string())
    }
}

fn parse_piece_type(r: char) -> Result<PieceType, String> {
    match r {
        'K' => Ok(PieceType::King),
        'Q' => Ok(PieceType::Queen),
        'R' => Ok(PieceType::Rook),
        'B' => Ok(PieceType::Bishop),
        'N' => Ok(PieceType::Knight),
        _ => Err(format!(
            "invalid piece type: {} (expected K, Q, R, B, or N)",
            r
        )),
    }
}

/// Formats a move to SAN given the board state BEFORE the move (Go `FormatSAN`).
pub fn format_san(board: &Board, mv: Move) -> String {
    let piece = board.piece_at(mv.from);
    if piece.is_empty() {
        return mv.to_string();
    }

    if piece.piece_type() == PieceType::King {
        let file_diff = mv.to.file() - mv.from.file();
        if file_diff == 2 {
            return "O-O".to_string();
        } else if file_diff == -2 {
            return "O-O-O".to_string();
        }
    }

    let mut result = String::new();

    let piece_type = piece.piece_type();
    if piece_type != PieceType::Pawn {
        result.push(piece_type_to_char(piece_type));
    }

    let target_piece = board.piece_at(mv.to);
    let mut is_capture = !target_piece.is_empty();

    if piece_type == PieceType::Pawn
        && board.en_passant_sq >= 0
        && mv.to == Square(board.en_passant_sq)
    {
        is_capture = true;
    }

    if piece_type != PieceType::Pawn {
        result.push_str(&disambiguation(board, mv));
    } else if is_capture {
        result.push((b'a' + mv.from.file() as u8) as char);
    }

    if is_capture {
        result.push('x');
    }

    result.push_str(&mv.to.to_string());

    if mv.promotion != PieceType::Empty {
        result.push('=');
        result.push(piece_type_to_char(mv.promotion));
    }

    let mut board_copy = board.copy();
    let _ = board_copy.make_move(mv);

    if board_copy.in_check() {
        if board_copy.legal_moves().is_empty() {
            result.push('#');
        } else {
            result.push('+');
        }
    }

    result
}

fn piece_type_to_char(pt: PieceType) -> char {
    match pt {
        PieceType::King => 'K',
        PieceType::Queen => 'Q',
        PieceType::Rook => 'R',
        PieceType::Bishop => 'B',
        PieceType::Knight => 'N',
        _ => '?',
    }
}

fn disambiguation(board: &Board, mv: Move) -> String {
    let piece = board.piece_at(mv.from);
    let piece_type = piece.piece_type();

    let mut candidates: Vec<Move> = Vec::new();
    for m in board.legal_moves() {
        if m.to == mv.to && m.from != mv.from {
            let candidate_piece = board.piece_at(m.from);
            if candidate_piece.piece_type() == piece_type {
                candidates.push(m);
            }
        }
    }

    if candidates.is_empty() {
        return String::new();
    }

    let from_file = mv.from.file();
    let from_rank = mv.from.rank();

    let file_unique = !candidates.iter().any(|m| m.from.file() == from_file);
    if file_unique {
        return ((b'a' + from_file as u8) as char).to_string();
    }

    let rank_unique = !candidates.iter().any(|m| m.from.rank() == from_rank);
    if rank_unique {
        return ((b'1' + from_rank as u8) as char).to_string();
    }

    format!(
        "{}{}",
        (b'a' + from_file as u8) as char,
        (b'1' + from_rank as u8) as char
    )
}

// Silence unused import warning for NO_SQUARE if the compiler ever prunes usage.
const _: Square = NO_SQUARE;

/// Formats moves into a numbered, paired coordinate list (Go `FormatMoveHistory`).
pub fn format_move_history(moves: &[Move]) -> String {
    if moves.is_empty() {
        return String::new();
    }

    let mut result = String::new();
    let mut i = 0;
    while i < moves.len() {
        let move_num = (i / 2) + 1;
        result.push_str(&format!("{}. {}", move_num, moves[i]));
        if i + 1 < moves.len() {
            result.push_str(&format!(" {} ", moves[i + 1]));
        }
        i += 2;
    }
    result
}

/// Formats the move history as SAN with a header (Go `Model.formatMoveHistory`).
pub fn format_move_history_san(moves: &[Move]) -> String {
    if moves.is_empty() {
        return String::new();
    }

    let mut b = String::from("Move History: ");
    let mut board = Board::new();

    let mut i = 0;
    while i < moves.len() {
        let move_num = (i / 2) + 1;
        let white_san = format_san(&board, moves[i]);
        let _ = board.make_move(moves[i]);

        if i + 1 < moves.len() {
            let black_san = format_san(&board, moves[i + 1]);
            let _ = board.make_move(moves[i + 1]);
            b.push_str(&format!("{}. {} {}", move_num, white_san, black_san));
            if i + 2 < moves.len() {
                b.push(' ');
            }
        } else {
            b.push_str(&format!("{}. {}", move_num, white_san));
        }
        i += 2;
    }
    b
}

#[cfg(test)]
mod tests {
    use super::*;
    use engine::Board;

    fn from_fen(fen: &str) -> Board {
        Board::from_fen(fen).expect("valid fen")
    }

    #[test]
    fn simple_pawn_moves() {
        let cases = [
            (
                "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
                "e4",
                "e2e4",
            ),
            (
                "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
                "d4",
                "d2d4",
            ),
            (
                "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
                "a3",
                "a2a3",
            ),
            (
                "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR b KQkq - 0 1",
                "d5",
                "d7d5",
            ),
            (
                "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
                "e3",
                "e2e3",
            ),
        ];
        for (fen, san, want) in cases {
            let b = from_fen(fen);
            let mv = parse_san(&b, san).expect(san);
            assert_eq!(mv.to_string(), want, "{}", san);
        }
    }

    #[test]
    fn piece_moves() {
        let b = from_fen("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1");
        assert_eq!(parse_san(&b, "Nf3").unwrap().to_string(), "g1f3");
        assert_eq!(parse_san(&b, "Nc3").unwrap().to_string(), "b1c3");
    }

    #[test]
    fn castling() {
        let b = from_fen("r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1");
        assert_eq!(parse_san(&b, "O-O").unwrap().to_string(), "e1g1");
        assert_eq!(parse_san(&b, "O-O-O").unwrap().to_string(), "e1c1");
    }

    #[test]
    fn strips_check_symbols() {
        let b = from_fen("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1");
        assert_eq!(parse_san(&b, "e4+").unwrap().to_string(), "e2e4");
        assert_eq!(parse_san(&b, "e4#").unwrap().to_string(), "e2e4");
    }

    #[test]
    fn invalid_moves_error() {
        let b = from_fen("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1");
        assert!(parse_san(&b, "").is_err());
        assert!(parse_san(&b, "e5").is_err());
        assert!(parse_san(&b, "Zf3").is_err());
    }

    #[test]
    fn format_san_basics() {
        let b = Board::new();
        assert_eq!(format_san(&b, Move::parse("e2e4").unwrap()), "e4");
        assert_eq!(format_san(&b, Move::parse("g1f3").unwrap()), "Nf3");
        assert_eq!(format_san(&b, Move::parse("b1c3").unwrap()), "Nc3");
    }

    #[test]
    fn format_san_castling() {
        let b = from_fen("r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1");
        assert_eq!(format_san(&b, Move::parse("e1g1").unwrap()), "O-O");
        assert_eq!(format_san(&b, Move::parse("e1c1").unwrap()), "O-O-O");
    }

    #[test]
    fn format_san_capture() {
        // white pawn on e4, black pawn on d5: exd5
        let b = from_fen("rnbqkbnr/ppp1pppp/8/3p4/4P3/8/PPPP1PPP/RNBQKBNR w KQkq d6 0 2");
        assert_eq!(format_san(&b, Move::parse("e4d5").unwrap()), "exd5");
    }

    #[test]
    fn move_history_pairs() {
        let moves = vec![
            Move::parse("e2e4").unwrap(),
            Move::parse("e7e5").unwrap(),
            Move::parse("g1f3").unwrap(),
        ];
        assert_eq!(format_move_history(&moves), "1. e2e4 e7e5 2. g1f3");
        assert_eq!(format_move_history(&[]), "");
    }

    #[test]
    fn move_history_san_header() {
        let moves = vec![
            Move::parse("e2e4").unwrap(),
            Move::parse("e7e5").unwrap(),
            Move::parse("g1f3").unwrap(),
        ];
        assert_eq!(
            format_move_history_san(&moves),
            "Move History: 1. e4 e5 2. Nf3"
        );
    }
}
