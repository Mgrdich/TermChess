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

// pieceToChar converts a Piece to its FEN character representation.
// Returns uppercase for White pieces, lowercase for Black pieces.
// Returns '?' for Empty pieces (should not occur in valid FEN generation).
func pieceToChar(p Piece) rune {
	// Map piece types to their FEN characters
	pieceChars := map[PieceType]rune{
		Pawn:   'P',
		Knight: 'N',
		Bishop: 'B',
		Rook:   'R',
		Queen:  'Q',
		King:   'K',
	}

	// Get the character for this piece type
	char, ok := pieceChars[p.Type()]
	if !ok {
		return '?' // Should not happen for valid pieces
	}

	// Convert to lowercase for Black pieces
	if p.Color() == Black {
		char = char - 'A' + 'a'
	}

	return char
}

// ToFEN converts the board position to FEN (Forsyth-Edwards Notation) format.
// FEN format consists of 6 space-separated fields:
// 1. Piece placement (from rank 8 to rank 1, separated by '/')
// 2. Active color ('w' or 'b')
// 3. Castling rights (combination of 'KQkq' or '-' if none)
// 4. En passant target square (algebraic notation or '-' if none)
// 5. Halfmove clock (for fifty-move rule)
// 6. Fullmove number (starts at 1, increments after Black's move)
func (b *Board) ToFEN() string {
	var fen strings.Builder

	// Field 1: Piece placement
	// Iterate through ranks from 8 to 1 (rank 7 down to rank 0)
	for rank := 7; rank >= 0; rank-- {
		emptyCount := 0

		// Iterate through files from a to h (file 0 to 7)
		for file := 0; file < 8; file++ {
			sq := NewSquare(file, rank)
			piece := b.Squares[sq]

			if piece.IsEmpty() {
				// Count consecutive empty squares
				emptyCount++
			} else {
				// Write out the count of empty squares if any
				if emptyCount > 0 {
					fen.WriteRune(rune('0' + emptyCount))
					emptyCount = 0
				}
				// Write the piece character
				fen.WriteRune(pieceToChar(piece))
			}
		}

		// Write any remaining empty squares for this rank
		if emptyCount > 0 {
			fen.WriteRune(rune('0' + emptyCount))
		}

		// Add rank separator (except after the last rank)
		if rank > 0 {
			fen.WriteRune('/')
		}
	}

	// Field 2: Active color
	fen.WriteRune(' ')
	if b.ActiveColor == White {
		fen.WriteRune('w')
	} else {
		fen.WriteRune('b')
	}

	// Field 3: Castling rights
	fen.WriteRune(' ')
	castlingStr := ""
	if b.CastlingRights&CastleWhiteKing != 0 {
		castlingStr += "K"
	}
	if b.CastlingRights&CastleWhiteQueen != 0 {
		castlingStr += "Q"
	}
	if b.CastlingRights&CastleBlackKing != 0 {
		castlingStr += "k"
	}
	if b.CastlingRights&CastleBlackQueen != 0 {
		castlingStr += "q"
	}
	if castlingStr == "" {
		castlingStr = "-"
	}
	fen.WriteString(castlingStr)

	// Field 4: En passant target square
	fen.WriteRune(' ')
	if b.EnPassantSq < 0 {
		fen.WriteRune('-')
	} else {
		// Convert en passant square index to algebraic notation
		epSquare := Square(b.EnPassantSq)
		fen.WriteString(epSquare.String())
	}

	// Field 5: Halfmove clock
	fen.WriteRune(' ')
	fen.WriteString(fmt.Sprintf("%d", b.HalfMoveClock))

	// Field 6: Fullmove number
	fen.WriteRune(' ')
	fen.WriteString(fmt.Sprintf("%d", b.FullMoveNum))

	return fen.String()
}
