package connect4FMI

import (
	"testing"
)

func TestBoardWhereConnected(t *testing.T) {
	board := NewBoard()
	board.Drop(5, "*")
	board.Drop(4, "*")
	board.Drop(3, "*")
	board.Drop(2, "*")
	_, col, row := board.WhereConnected("*")
	expectedCol := [4]int{4,4,3,2}
	if col != expectedCol{
		t.Errorf("GetPieceValue was incorrect, got")
	}
	expectedRow := [4]int{5,5,5,5}
	if row != expectedRow{
		t.Errorf("GetPieceValue was incorrect, got")
	}
}