package connect4FMI

import (
	"fmt"
	"net"
	"strconv"
)

type gameManager struct {
	board     Board
	conn      net.Conn
	ai        bool
	turn      int
	state     int
	winner    string
	aiDiff    int
	lostGames int
	wonGames  int
}

const (
	Running = 0
	Win     = 2
	Lose    = 3
	Tie     = 4
)

func NewGameManager(con net.Conn, ai bool, aiDiff int) gameManager {
	b := *NewBoard()
	return gameManager{board: b, conn: con, ai: ai, aiDiff: aiDiff}
}

func (gm *gameManager) GetHoleColor(i, j int) string {
	return gm.board.board[i][j]
}

func (gm *gameManager) color() string {
	if gm.turn%2 == 0 {
		return playerOneColor
	}
	return playerTwoColor
}

func (gm *gameManager) GetState() int {
	return gm.state
}

func (gm *gameManager) makePlayerTurn(column int) bool {
	if gm.board.Drop(column, gm.color()) {
		if gm.board.areFourConnected(gm.color()) {
			gm.state = Win
			gm.winner = gm.color()
			gm.wonGames++
		}
		gm.turn++
		if gm.turn == 42 {
			gm.state = Tie
		}
		if !gm.ai {
			_, err := fmt.Fprintf(gm.conn, "%d\n", column)
			if err != nil {
				panic(err)
			}
		}
		return true
	}
	return false
}

func (gm *gameManager) makeOpponentTurn() int {
	var column int
	if gm.ai {
		column = getAiMove(&gm.board, gm.aiDiff)
	} else {
		var msg string

		fmt.Println(gm.conn.RemoteAddr().String())

		_, err := fmt.Fscan(gm.conn, &msg)
		if err != nil {
			panic(err)
		}
		if msg == "timeout" || msg == "error" {
			fmt.Println("opponent disconnected!")
			panic(nil)
		}
		column, _ = strconv.Atoi(msg)
	}
	gm.board.Drop(column, gm.color())
	if gm.board.areFourConnected(gm.color()) {
		gm.state = Lose
		gm.lostGames++
		gm.winner = gm.color()
	}
	gm.turn++
	if gm.turn == 42 {
		gm.state = Tie
	}
	return column
}

func (gm *gameManager) WhereConnected() (bool, [4]int, [4]int) {
	connected, x, y := gm.board.WhereConnected(gm.winner)
	return connected, x, y
}

func (gm *gameManager) resetGame() {
	gm.board = *NewBoard()
	gm.turn = 0
	gm.state = 0
}

func (gm* gameManager) GetWonGames()int{
	return gm.wonGames
}

func (gm* gameManager) GetLostGames()int{
	return gm.lostGames
}
