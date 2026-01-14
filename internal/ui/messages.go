package ui

import (
	"math/rand"
	"time"
)

// thinkingMessages contains humorous chess-themed messages displayed while the bot is thinking.
// These messages are randomly selected to entertain the user during bot computation.
var thinkingMessages = []string{
	"Consulting the ancient chess masters...",
	"Calculating infinite possibilities...",
	"Pondering the meaning of chess...",
	"Summoning the spirit of Bobby Fischer...",
	"Analyzing 42 dimensions of chess space...",
	"Teaching my neural networks a lesson...",
	"Asking my rubber duck for advice...",
	"Flipping through my opening book...",
	"Sacrificing pawns to the chess gods...",
	"Pretending to think really hard...",
	"Counting squares intensely...",
	"Channeling my inner Stockfish...",
}

// getRandomThinkingMessage returns a random thinking message from the predefined list.
// Uses a local random number generator seeded with the current time for variety.
func getRandomThinkingMessage() string {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	index := rng.Intn(len(thinkingMessages))
	return thinkingMessages[index]
}
