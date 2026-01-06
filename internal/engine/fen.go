package engine

import (
	"fmt"
	"strconv"
	"strings"
)

// FromFEN creates a Board from a FEN (Forsyth-Edwards Notation) string.
// FEN format: <pieces> <active> <castling> <ep> <halfmove> <fullmove>
// Example: "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
func FromFEN(fen string) (*Board, error) {
	parts := strings.Fields(fen)
	if len(parts) != 6 {
		return nil, fmt.Errorf("FEN must have 6 parts, got %d", len(parts))
	}

	b := &Board{
		Squares:        [64]Piece{},
		ActiveColor:    White,
		CastlingRights: 0,
		EnPassantSq:    -1,
		HalfMoveClock:  0,
		FullMoveNum:    1,
		Hash:           0,
		History:        []uint64{},
	}

	// Part 1: Piece placement (from rank 8 to rank 1)
	ranks := strings.Split(parts[0], "/")
	if len(ranks) != 8 {
		return nil, fmt.Errorf("FEN piece placement must have 8 ranks, got %d", len(ranks))
	}

	for rankIdx := 0; rankIdx < 8; rankIdx++ {
		rank := 7 - rankIdx // FEN starts from rank 8 (index 7)
		rankStr := ranks[rankIdx]
		file := 0

		for _, ch := range rankStr {
			if ch >= '1' && ch <= '8' {
				// Empty squares
				emptyCount := int(ch - '0')
				file += emptyCount
			} else {
				// Piece
				if file > 7 {
					return nil, fmt.Errorf("too many pieces in rank %d", rank+1)
				}

				var color Color
				var pieceType PieceType

				// Determine color (uppercase = White, lowercase = Black)
				if ch >= 'A' && ch <= 'Z' {
					color = White
				} else {
					color = Black
					ch = ch - 'a' + 'A' // Convert to uppercase for matching
				}

				// Determine piece type
				switch ch {
				case 'P':
					pieceType = Pawn
				case 'N':
					pieceType = Knight
				case 'B':
					pieceType = Bishop
				case 'R':
					pieceType = Rook
				case 'Q':
					pieceType = Queen
				case 'K':
					pieceType = King
				default:
					return nil, fmt.Errorf("invalid piece character: %c", ch)
				}

				sq := NewSquare(file, rank)
				b.Squares[sq] = NewPiece(color, pieceType)
				file++
			}
		}

		if file != 8 {
			return nil, fmt.Errorf("rank %d has %d squares, expected 8", rank+1, file)
		}
	}

	// Part 2: Active color
	switch parts[1] {
	case "w":
		b.ActiveColor = White
	case "b":
		b.ActiveColor = Black
	default:
		return nil, fmt.Errorf("invalid active color: %s (expected 'w' or 'b')", parts[1])
	}

	// Part 3: Castling rights
	if parts[2] != "-" {
		for _, ch := range parts[2] {
			switch ch {
			case 'K':
				b.CastlingRights |= CastleWhiteKing
			case 'Q':
				b.CastlingRights |= CastleWhiteQueen
			case 'k':
				b.CastlingRights |= CastleBlackKing
			case 'q':
				b.CastlingRights |= CastleBlackQueen
			default:
				return nil, fmt.Errorf("invalid castling character: %c", ch)
			}
		}
	}

	// Part 4: En passant square
	if parts[3] != "-" {
		if len(parts[3]) != 2 {
			return nil, fmt.Errorf("invalid en passant square: %s", parts[3])
		}
		file := int(parts[3][0] - 'a')
		rank := int(parts[3][1] - '1')
		if file < 0 || file > 7 || rank < 0 || rank > 7 {
			return nil, fmt.Errorf("invalid en passant square: %s", parts[3])
		}
		b.EnPassantSq = int8(NewSquare(file, rank))
	}

	// Part 5: Half-move clock
	halfMove, err := strconv.Atoi(parts[4])
	if err != nil {
		return nil, fmt.Errorf("invalid half-move clock: %s", parts[4])
	}
	if halfMove < 0 || halfMove > 255 {
		return nil, fmt.Errorf("half-move clock out of range: %d", halfMove)
	}
	b.HalfMoveClock = uint8(halfMove)

	// Part 6: Full move number
	fullMove, err := strconv.Atoi(parts[5])
	if err != nil {
		return nil, fmt.Errorf("invalid full move number: %s", parts[5])
	}
	if fullMove < 1 || fullMove > 65535 {
		return nil, fmt.Errorf("full move number out of range: %d", fullMove)
	}
	b.FullMoveNum = uint16(fullMove)

	// Compute the Zobrist hash and add it to history
	b.Hash = b.ComputeHash()
	b.History = append(b.History, b.Hash)

	return b, nil
}
