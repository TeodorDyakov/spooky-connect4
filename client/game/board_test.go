package gameLogic

import (
	"testing"
)

var _ = func() bool {
	testing.Init()
	return true
}()

func TestBoardWhereConnectedHorizontal(t *testing.T) {
	board := NewBoard()
	board.Drop(5, "*")
	board.Drop(4, "*")
	board.Drop(3, "*")
	board.Drop(2, "*")

	areConnected, row, col := board.WhereConnected("*")
	expectedCol := [4]int{2, 3, 4, 5}
	if col != expectedCol {
		t.Errorf("columns are incorrect, expected %v got %v", expectedCol, col)
	}
	expectedRow := [4]int{5, 5, 5, 5}
	if row != expectedRow {
		t.Errorf("rows are incorrect, expected %v got %v", expectedRow, row)
	}
	if !areConnected {
		t.Errorf("expected four to be connected but return false")
	}
}

func TestBoardWhereConnectedVertical(t *testing.T) {
	board := NewBoard()
	board.Drop(5, "*")
	board.Drop(5, "*")
	board.Drop(5, "*")
	board.Drop(5, "*")

	areConnected, row, col := board.WhereConnected("*")
	expectedCol := [4]int{5, 5, 5, 5}
	if col != expectedCol {
		t.Errorf("columns are incorrect, expected %v got %v", expectedCol, col)
	}
	expectedRow := [4]int{2, 3, 4, 5}
	if row != expectedRow {
		t.Errorf("rows are incorrect, expected %v got %v", expectedRow, row)
	}
	if !areConnected {
		t.Errorf("expected four to be connected but return false")
	}
}

func TestBoardWhereConnectedAscendingDiagonal(t *testing.T) {
	board := NewBoard()
	board.board[5][0] = "*"
	board.board[4][1] = "*"
	board.board[3][2] = "*"
	board.board[2][3] = "*"

	areConnected, row, col := board.WhereConnected("*")
	expectedCol := [4]int{0, 1, 2, 3}
	if col != expectedCol {
		t.Errorf("columns are incorrect, expected %v got %v", expectedCol, col)
	}
	expectedRow := [4]int{5, 4, 3, 2}
	if row != expectedRow {
		t.Errorf("rows are incorrect, expected %v got %v", expectedRow, row)
	}
	if !areConnected {
		t.Errorf("expected four to be connected but return false")
	}
}
