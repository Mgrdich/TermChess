// Package engine implements the chess engine for TermChess.
package engine

import (
	"math/rand"
)

// Zobrist hash tables - initialized at package init time with deterministic values.
// These are used to compute and incrementally update position hashes for
// repetition detection and transposition tables.
var (
	// zobristPieces[pieceIndex][square] - random value for each piece type on each square.
	// pieceIndex = color * 6 + (pieceType - 1), where pieceType is 1-6 (Pawn-King).
	// This gives us 12 piece indices (0-5 for White, 6-11 for Black) x 64 squares.
	zobristPieces [12][64]uint64

	// zobristSideToMove - XORed when it's Black's turn.
	zobristSideToMove uint64

	// zobristCastling[rights] - random value for each combination of castling rights (0-15).
	zobristCastling [16]uint64

	// zobristEnPassant[file] - random value for en passant on each file (0-7).
	// Only hashed when there is an en passant square available.
	zobristEnPassant [8]uint64
)

// init initializes the Zobrist hash tables with deterministic pseudo-random values.
// Using a fixed seed ensures that the same position always produces the same hash,
// even across different runs of the program.
func init() {
	// Use a fixed seed for deterministic hashes
	rng := rand.New(rand.NewSource(0x5D4E3C2B1A))

	// Initialize piece-square values
	for pieceIndex := 0; pieceIndex < 12; pieceIndex++ {
		for square := 0; square < 64; square++ {
			zobristPieces[pieceIndex][square] = rng.Uint64()
		}
	}

	// Initialize side to move
	zobristSideToMove = rng.Uint64()

	// Initialize castling rights
	for rights := 0; rights < 16; rights++ {
		zobristCastling[rights] = rng.Uint64()
	}

	// Initialize en passant files
	for file := 0; file < 8; file++ {
		zobristEnPassant[file] = rng.Uint64()
	}
}

// pieceZobristIndex returns the Zobrist table index for a piece.
// Returns -1 for empty squares.
func pieceZobristIndex(p Piece) int {
	if p.IsEmpty() {
		return -1
	}
	// pieceType is 1-6 (Pawn to King), so subtract 1 to get 0-5
	// Color is 0 (White) or 1 (Black), multiply by 6 to offset
	return int(p.Color())*6 + int(p.Type()) - 1
}

// ComputeHash computes the full Zobrist hash for the current board position.
// This is called once when creating a new board and can be used to verify
// incremental hash updates.
func (b *Board) ComputeHash() uint64 {
	var hash uint64

	// Hash all pieces on the board
	for sq := Square(0); sq < 64; sq++ {
		piece := b.Squares[sq]
		if !piece.IsEmpty() {
			pieceIdx := pieceZobristIndex(piece)
			hash ^= zobristPieces[pieceIdx][sq]
		}
	}

	// Hash side to move (XOR if Black)
	if b.ActiveColor == Black {
		hash ^= zobristSideToMove
	}

	// Hash castling rights
	hash ^= zobristCastling[b.CastlingRights]

	// Hash en passant file (only if there is an en passant square)
	if b.EnPassantSq >= 0 {
		epFile := Square(b.EnPassantSq).File()
		hash ^= zobristEnPassant[epFile]
	}

	return hash
}

// hashPiece returns the Zobrist hash contribution for a piece on a square.
// Used for incremental hash updates - XOR to add or remove a piece.
func hashPiece(p Piece, sq Square) uint64 {
	if p.IsEmpty() {
		return 0
	}
	pieceIdx := pieceZobristIndex(p)
	return zobristPieces[pieceIdx][sq]
}
