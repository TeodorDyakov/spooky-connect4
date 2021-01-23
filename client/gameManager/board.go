package gameManager

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
	boardWidth  = 7
	boardHeight = 6
	emptySpot   = "∟"
)

func (b *Board) gameOver() bool {
	return b.movesMade == 42 || b.areFourConnected(playerOneColor) || b.areFourConnected(playerTwoColor)
}

func (b *Board) copyOfBoard() *Board {
	boardCopy := NewBoard()
	for i := 0; i < boardHeight; i++ {
		for j := 0; j < boardWidth; j++ {
			boardCopy.board[i][j] = b.board[i][j]
		}
	}
	boardCopy.col = make([]int, boardWidth)
	copy(boardCopy.col, b.col)
	return boardCopy
}

func NewBoard() *Board {
	var b *Board
	b = new(Board)
	b.movesMade = 0
	b.col = make([]int, boardWidth)
	//initialize the connect 4 b.board
	for i := 0; i < boardHeight; i++ {
		row := make([]string, boardWidth)

		for i := 0; i < len(row); i++ {
			row[i] = emptySpot
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
	b.board[5-b.col[column]][column] = emptySpot
	b.movesMade--
}

func (b *Board) Drop(column int, player string) bool {
	if column < len(b.board[0]) && (column >= 0) && b.col[column] < len(b.board) {
		b.board[5-b.col[column]][column] = player
		b.col[column]++
		b.movesMade++
		return true
	}
	return false
}

func (b *Board) WhereConnected(player string) (bool, [4]int, [4]int) {
	for j := 0; j < len(b.board[0])-3; j++ {
		for i := 0; i < len(b.board); i++ {
			if b.board[i][j] == player &&
				b.board[i][j+1] == player &&
				b.board[i][j+2] == player &&
				b.board[i][j+3] == player {
				return true, [4]int{i, i, i, i}, [4]int{j, j + 1, j + 2, j + 3}
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
				return true, [4]int{i, i + 1, i + 2, i + 3}, [4]int{j, j, j, j}
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
				return true, [4]int{i, i - 1, i - 2, i - 3}, [4]int{j, j + 1, j + 2, j + 3}
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
				return true, [4]int{i, i - 1, i - 2, i - 3}, [4]int{j, j - 1, j - 2, j - 3}
			}
		}
	}
	return false, [4]int{-1, -1, -1, -1}, [4]int{-1, -1, -1, -1}
}

func (b *Board) areFourConnected(player string) bool {
	connected, _, _ := b.WhereConnected(player)
	return connected
}
