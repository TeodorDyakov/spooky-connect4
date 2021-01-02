package connect4FMI

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

const (
	big   = 100000
	small = -big
)

//getAiMove returns the best move for the given board position based on the strength of the AI
func getAiMove(b *Board, strength int) int {
	copy := b.copyOfBoard()
	_, move := alphabeta(copy, true, 0, small, big, strength)
	return move
}

//alphabeta implenets the alphabeta algorithm and returns the score of the given board position
//and the best move for the given board position
func alphabeta(b *Board, maximizer bool, depth, alpha, beta, max_depth int) (int, int) {
	if depth == max_depth {
		return 0, -1
	}
	if b.areFourConnected(playerTwoColor) {
		return big - depth, -1
	} else if b.areFourConnected(playerOneColor) {
		return small + depth, -1
	}
	var value int
	var bestMove int
	shuffledColumns := rand.Perm(7)

	if maximizer {
		value = small
		for _, column := range shuffledColumns {
			if b.Drop(column, playerTwoColor) {
				new_score, _ := alphabeta(b, false, depth+1, alpha, beta, max_depth)
				b.undoDrop(column)

				if value < new_score {
					bestMove = column
					value = new_score
				}
				alpha = max(alpha, value)
				if alpha >= beta {
					break
				}
			}
		}
	} else {
		value = big
		for _, column := range shuffledColumns {
			if b.Drop(column, playerOneColor) {
				new_score, _ := alphabeta(b, true, depth+1, alpha, beta, max_depth)
				b.undoDrop(column)

				if value > new_score {
					bestMove = column
					value = new_score
				}
				beta = min(beta, value)
				if alpha >= beta {
					break
				}
			}
		}
	}
	return value, bestMove
}

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}

func max(a, b int) int {
	if a < b {
		return b
	}
	return a
}
