package main

import (
	"fmt"
	"strings"
)

type Board struct {
	board     [][]string
	col       []int
	movesMade int
}

const (
	BOARD_WIDTH  = 7
	BOARD_HEIGHT = 6
	EMPTY_SPOT   = "∟"
)

func (b *Board) gameOver() bool {
	return b.movesMade == 42 || b.areFourConnected(PLAYER_ONE_COLOR) || b.areFourConnected(PLAYER_TWO_COLOR)
}

func NewBoard() *Board {
	var b *Board
	b = new(Board)
	b.movesMade = 0
	b.col = make([]int, BOARD_WIDTH)
	//initialize the connect 4 b.board
	for i := 0; i < BOARD_HEIGHT; i++ {
		row := make([]string, BOARD_WIDTH)

		for i := 0; i < len(row); i++ {
			row[i] = EMPTY_SPOT
		}
		b.board = append(b.board, row)
	}
	return b
}

func (b *Board) printBoard() {
	space := strings.Repeat(" ", 20)
	fmt.Print(space)
	for i := 0; i < len(b.board[0]); i++ {
		fmt.Printf("%d ", i)
	}
	fmt.Println()
	for i := 0; i < len(b.board); i++ {
		fmt.Print(space)
		for j := 0; j < len(b.board[0]); j++ {
			fmt.Print(b.board[i][j] + " ")
		}
		fmt.Println()
	}
}

func (b *Board) undoDrop(column int) {
	b.col[column]--
	b.board[5-b.col[column]][column] = EMPTY_SPOT
	b.movesMade--
}

func (b *Board) drop(column int, player string) bool {
	if column < len(b.board[0]) && (column >= 0) && b.col[column] < len(b.board) {
		b.board[5-b.col[column]][column] = player
		b.col[column]++
		b.movesMade++
		return true
	}
	return false
}
	
func (b *Board) whereConnected(player string)(bool, int, int, int){
	for j := 0; j < len(b.board[0])-3; j++ {
		for i := 0; i < len(b.board); i++ {
			if b.board[i][j] == player &&
				b.board[i][j+1] == player &&
				b.board[i][j+2] == player &&
				b.board[i][j+3] == player {
				return true, i, j, 0
			}
		}
	}
	// verticalCheck
	for i := 0; i < len(b.board)-3; i++ {
		for j := 0; j < len(b.board[0]); j++ {
			if b.board[i][j] == player &&
				b.board[i+1][j] == player &&
				b.board[i+2][j] == player &&
				b.board[i+3][j] == player {
				return true, i, j, 1
			}
		}
	}
	// ascendingDiagonalCheck
	for i := 3; i < len(b.board); i++ {
		for j := 0; j < len(b.board[0])-3; j++ {
			if b.board[i][j] == player &&
				b.board[i-1][j+1] == player &&
				b.board[i-2][j+2] == player &&
				b.board[i-3][j+3] == player {
				return true, i, j, 2
			}
		}
	}
	// descendingDiagonalCheck
	for i := 3; i < len(b.board); i++ {
		for j := 3; j < len(b.board[0]); j++ {
			if b.board[i][j] == player &&
				b.board[i-1][j-1] == player &&
				b.board[i-2][j-2] == player &&
				b.board[i-3][j-3] == player {
				return true, i, j, 3
			}
		}
	}
	return false, -1, -1, -1
}

func (b *Board) areFourConnected(player string) bool {
	connected, _, _, _ := b.whereConnected(player)
	return connected
}
