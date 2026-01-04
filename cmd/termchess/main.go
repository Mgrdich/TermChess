// Package main is the entry point for the TermChess application.
package main

import (
	"fmt"

	"github.com/Mgrdich/TermChess/internal/engine"
)

func main() {
	board := engine.NewBoard()

	fmt.Println("TermChess - Terminal Chess Game")
	fmt.Println("================================")
	fmt.Println()
	fmt.Println(board.String())
	fmt.Println()
	fmt.Printf("Active color: %s\n", colorName(board.ActiveColor))
	fmt.Printf("Full move number: %d\n", board.FullMoveNum)
	fmt.Println("Ready to play!")
}

func colorName(c engine.Color) string {
	if c == engine.White {
		return "White"
	}
	return "Black"
}
