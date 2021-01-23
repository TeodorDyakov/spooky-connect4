package gameManager

import (
	"testing"
)

var _ = func() bool {
	testing.Init()
	return true
}()

func TestGetAiMoveOneMoveFromLose(t *testing.T) {
	board := NewBoard()
	board.Drop(5, playerOneColor)
	board.Drop(5, playerOneColor)
	board.Drop(5, playerOneColor)

	bestMove := getAiMove(board, 10)

	if bestMove != 5 {
		t.Errorf("AI did not made expected move, expected %d, got %d", 5, bestMove)
	}
}

func TestGetAiMoveTwoMoveFromLose(t *testing.T) {
	board := NewBoard()
	board.Drop(3, playerOneColor)
	board.Drop(4, playerOneColor)

	bestMove := getAiMove(board, 10)

	if bestMove != 2 && bestMove != 5 {
		t.Errorf("AI did not made expected move, expected %d, got %d", 2, bestMove)
	}
}
