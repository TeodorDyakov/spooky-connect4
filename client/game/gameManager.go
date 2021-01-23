package gameLogic

import (
	"fmt"
	"net"
	"strconv"
)

type GameManager struct {
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
	PlayerOneColor    = "◯"
	PlayerTwoColor    = "⬤"
)

func NewGameManager(con net.Conn, ai bool, aiDiff int) GameManager {
	b := *NewBoard()
	return GameManager{board: b, conn: con, ai: ai, aiDiff: aiDiff}
}

func (gm *GameManager) GetHoleColor(i, j int) string {
	return gm.board.board[i][j]
}

func (gm *GameManager) color() string {
	if gm.turn%2 == 0 {
		return PlayerOneColor
	}
	return PlayerTwoColor
}

func (gm *GameManager) GetState() int {
	return gm.state
}

func (gm *GameManager) makePlayerTurn(column int) bool {
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

func (gm *GameManager) makeOpponentTurn() int {
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

func (gm *GameManager) WhereConnected() (bool, [4]int, [4]int) {
	connected, x, y := gm.board.WhereConnected(gm.winner)
	return connected, x, y
}

func (gm *GameManager) resetGame() {
	gm.board = *NewBoard()
	gm.turn = 0
	gm.state = 0
}

func (gm* GameManager) GetWonGames()int{
	return gm.wonGames
}

func (gm* GameManager) GetLostGames()int{
	return gm.lostGames
}
