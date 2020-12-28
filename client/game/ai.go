package connect4FMI

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

const (
	BIG   = 100000
	SMALL = -BIG
)

func getAiMove(b *Board, difficulty int) int {
	copy := b.copyOfBoard()
	_, move := alphabeta(copy, true, 0, SMALL, BIG, difficulty)
	return move
}

func alphabeta(b *Board, maximizer bool, depth, alpha, beta, max_depth int) (int, int) {
	if depth == max_depth {
		return 0, -1
	}
	if b.areFourConnected(PLAYER_TWO_COLOR) {
		return BIG - depth, -1
	} else if b.areFourConnected(PLAYER_ONE_COLOR) {
		return SMALL + depth, -1
	}
	var value int
	var bestMove int
	shuffledColumns := rand.Perm(7)

	if maximizer {
		value = SMALL
		for _, column := range shuffledColumns {
			if b.drop(column, PLAYER_TWO_COLOR) {
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
		value = BIG
		for _, column := range shuffledColumns {
			if b.drop(column, PLAYER_ONE_COLOR) {
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
