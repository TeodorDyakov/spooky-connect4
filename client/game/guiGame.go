package connect4FMI

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"net"
	"os"
	"strconv"
	"time"
)

var bg *ebiten.Image
var owl *ebiten.Image
var red *ebiten.Image
var ghost *ebiten.Image
var yellow *ebiten.Image
var boardImage *ebiten.Image

func init() {
	var err error
	boardImage, _, err = ebitenutil.NewImageFromFile("images/conn4trans2.png")
	if err != nil {
		log.Fatal(err)
	}
	bg, _, err = ebitenutil.NewImageFromFile("images/bg2.jpeg")
	if err != nil {
		log.Fatal(err)
	}
	red, _, err = ebitenutil.NewImageFromFile("images/redzwei.png")
	if err != nil {
		log.Fatal(err)
	}
	yellow, _, err = ebitenutil.NewImageFromFile("images/green.png")
	if err != nil {
		log.Fatal(err)
	}
	owl, _, err = ebitenutil.NewImageFromFile("images/owl2.png")
	if err != nil {
		log.Fatal(err)
	}
	ghost, _, err = ebitenutil.NewImageFromFile("images/ghost.png")
	if err != nil {
		log.Fatal(err)
	}
	tt, _ := opentype.Parse(fonts.MPlus1pRegular_ttf)
	mplusNormalFont, _ = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    20,
		DPI:     72,
		Hinting: font.HintingFull,
	})
}

type Game struct{}

type GameState int

const (
	yourTurn GameState = iota
	opponentTurn
	win
	lose
	tie
	menu
	waitingForConnect
	waitingForToken
	connectToRoomWithToken
)

const (
	MIN_DIFFICULTY       = 1
	MAX_DIFFICULTY       = 12
	SECONDS_TO_MAKE_TURN = 59
	fps                  = 60
	tileHeight           = 65
	tileOffset           = 10
	boardX               = 84
	boardY               = 130
	gravity              = 0.3
	PLAYER_ONE_COLOR     = "◯"
	PLAYER_TWO_COLOR     = "⬤"
)

var opponentLastCol int
var lostGames int
var wonGames int
var frameCount int
var gameState GameState = menu

/*
whether the fall animation for the given circle was done already
*/
var animated [7][6]bool
var fallSpeed float64
var b *Board = NewBoard()
var playingAgainstAi bool
var mplusNormalFont font.Face
var fallY float64 = -tileHeight
var again chan bool = make(chan bool)
var mouseClickBuffer chan int = make(chan int)
var messages [7]string = [7]string{"Your turn", "Other's turn", "You win!", "You lost.", "Tie.", "", ""}
var opponentAnimation bool
var difficulty int
var serverCommunicationChannel chan gameInfo = make(chan gameInfo)
var token string
var tokenChan chan string = make(chan string)

func (g *Game) Update() error {
	press := inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft)

	if gameState == yourTurn || gameState == opponentTurn {
		frameCount++
	}

	if frameCount == fps*SECONDS_TO_MAKE_TURN {
		os.Exit(1)
	}

	if gameState == yourTurn && press {
		mouseX, _ := ebiten.CursorPosition()
		/*
			only send click event to buffer if someone is waiting for it
		*/
		select {
		case mouseClickBuffer <- xcoordToColumn(mouseX):
		default:
		}
	}

	if gameState == menu && ebiten.IsKeyPressed(ebiten.KeyA) {
		gameState = yourTurn
		go playAgainstAi()
	}

	if gameState == menu && ebiten.IsKeyPressed(ebiten.KeyO) {
		gameState = waitingForConnect
		go quickplayLobby(serverCommunicationChannel)
	}

	if gameState == menu && ebiten.IsKeyPressed(ebiten.KeyR) {
		gameState = waitingForToken
		go createRoom(serverCommunicationChannel, tokenChan)
	}

	if gameState == menu && inpututil.IsKeyJustReleased(ebiten.KeyC) {
		gameState = connectToRoomWithToken
	}

	if gameState == connectToRoomWithToken {
		token += string(ebiten.InputChars())
		if len(token) == 5 {
			gameState = waitingForConnect
			go connectToRoom(token, serverCommunicationChannel)
		}
	}

	if gameState == waitingForToken {
		select {
		case token = <-tokenChan:
			gameState = waitingForConnect
		default:
		}
	}

	if gameState == waitingForConnect {
		select {
		case gameInfo := <-serverCommunicationChannel:
			if gameInfo.waiting {
				gameState = opponentTurn
			} else {
				gameState = yourTurn
			}
			go playMultiplayer(gameInfo.waiting, gameInfo.conn)
		default:
		}
	}

	if isGameOver() && press {
		mouseX, mouseY := ebiten.CursorPosition()
		/*check if mouse is in play again area
		 */
		if mouseX >= 230 && mouseX <= 600 && mouseY >= 500 {
			select {
			case again <- true:
			default:
			}
		}
	}
	return nil
}

func isGameOver() bool {
	return gameState == tie || gameState == win || gameState == lose
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.DrawImage(bg, nil)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(boardX, boardY)
	if gameState == menu {
		screen.DrawImage(boardImage, op)
		// topTextX := 200
		text.Draw(screen, "[A] - play against AI", mplusNormalFont, boardX, boardY -30, color.White)
		text.Draw(screen, "[R] - create a room", mplusNormalFont, boardX, 570, color.White)
		text.Draw(screen, "[C] - connect to a room", mplusNormalFont, boardX+250, 570, color.White)
		text.Draw(screen, "[O] - play online (quick play)", mplusNormalFont, boardX + 250, boardY-30, color.White)
		return
	}

	if gameState == connectToRoomWithToken {
		screen.DrawImage(boardImage, op)
		text.Draw(screen, "Enter the code for room:\n"+token, mplusNormalFont, 200, 50, color.White)
		return
	}

	if gameState == waitingForConnect || gameState == waitingForToken {
		screen.DrawImage(boardImage, op)
		text.Draw(screen, "waiting for opponent...", mplusNormalFont, 200, 50, color.White)
		if token != "" {
			text.Draw(screen, "Your token is: "+token, mplusNormalFont, 200, 80, color.White)
		}
		return
	}

	var msg string = messages[gameState]
	text.Draw(screen, "W  "+strconv.Itoa(wonGames)+":"+strconv.Itoa(lostGames)+"  L", mplusNormalFont, boardX, 50, color.White)
	text.Draw(screen, msg, mplusNormalFont, boardX, 580, color.White)
	text.Draw(screen, "00:"+strconv.Itoa(SECONDS_TO_MAKE_TURN-frameCount/fps), mplusNormalFont, 500, 580, color.White)

	screen.DrawImage(boardImage, op)
	if opponentAnimation {
		drawGhost(screen)
	}
	drawOwl(screen)
	if isGameOver() {
		text.Draw(screen, "Click here\nto play again", mplusNormalFont, 250, 580, color.White)
	}
	drawBalls(screen)

}

func drawBalls(screen *ebiten.Image) {
	for i := 0; i < len(b.board); i++ {
		for j := 0; j < len(b.board[0]); j++ {
			if b.board[i][j] == PLAYER_TWO_COLOR {
				drawBall(j, i, PLAYER_TWO_COLOR, screen)
			} else if b.board[i][j] == PLAYER_ONE_COLOR {
				drawBall(j, i, PLAYER_ONE_COLOR, screen)
			}
		}
	}
}

func drawGhost(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(opponentLastCol)*tileHeight+boardX+10, boardY-75)
	screen.DrawImage(ghost, op)
}

func drawOwl(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	mouseX, _ := ebiten.CursorPosition()
	if mouseX < boardX {
		mouseX = boardX
	}
	if mouseX > boardX+7*tileHeight {
		mouseX = boardX + 7*tileHeight
	}
	op.GeoM.Translate(float64(mouseX)-30, boardY-75)
	screen.DrawImage(owl, op)
}

func drawBall(x, y int, player string, screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(boardX+tileOffset, boardY+tileOffset)
	destY := tileOffset + float64(y)*tileHeight

	if animated[x][y] {
		op.GeoM.Translate(float64(x)*tileHeight, float64(y)*tileHeight)
	} else {
		fallY += fallSpeed
		fallSpeed += gravity
		if fallY > destY {
			fallY = destY
			fallSpeed = 0
			animated[x][y] = true
		}
		op.GeoM.Translate(float64(x)*tileHeight, fallY)
		if animated[x][y] {
			fallY = -tileHeight
		}
	}
	if player == PLAYER_TWO_COLOR {
		screen.DrawImage(red, op)
	} else {
		screen.DrawImage(yellow, op)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 640, 640
}

/*
on which column to drop based on x coordinate of click
*/
func xcoordToColumn(x int) int {
	return int(float64(x-tileOffset-boardX) / tileHeight)
}

/*
choose difficulty and start AI game loop
*/

func playAgainstAi() {
	difficulty = 10
	playingAgainstAi = true
	gameLogic(PLAYER_ONE_COLOR, PLAYER_TWO_COLOR, nil)
}

/*
show menu to choose game type - quick or with friend. After user chooses from console
starts the game loop.
*/
func playMultiplayer(wait bool, conn net.Conn) {
	playerColor := PLAYER_ONE_COLOR
	opponentColor := PLAYER_TWO_COLOR
	if wait {
		playerColor = PLAYER_TWO_COLOR
		opponentColor = PLAYER_ONE_COLOR
		gameState = opponentTurn
	} else {
		gameState = yourTurn
	}
	gameLogic(playerColor, opponentColor, conn)
}

/*
plays a full turn of the game, meaning you make a turn, and than thhen the opponent makes one
*/
func playTurn(playerColor, opponentColor string, conn net.Conn) {
	if gameState == opponentTurn {
		var column int
		if playingAgainstAi {
			column = getAiMove(b, difficulty)
		} else {
			var msg string
			_, err := fmt.Fscan(conn, &msg)
			if err != nil {
				panic(err)
			}
			if msg == "timeout" || msg == "error" {
				fmt.Println("opponent disconnected!")
				panic(nil)
				return
			}
			column, _ = strconv.Atoi(msg)
		}
		opponentLastCol = column
		opponentAnimation = true
		b.drop(column, opponentColor)
		/*
			wait for the animation of falling circle to finish
		*/
		time.Sleep(1 * time.Second)
		opponentAnimation = false
		frameCount = 0
		gameState = yourTurn
	} else if gameState == yourTurn {
		column := <-mouseClickBuffer
		if b.drop(column, playerColor) {
			frameCount = 0
			gameState = opponentTurn
			if !playingAgainstAi {
				_, err := fmt.Fprintf(conn, "%d\n", column)
				if err != nil {
					panic(err)
				}
			}
			/*
				wait for the animation of falling circle to finish
			*/
			time.Sleep(1 * time.Second)
		}
	}
}

func gameLogic(playerColor, opponentColor string, conn net.Conn) {
	playAgain := true
	for playAgain {
		for !b.gameOver() {
			playTurn(playerColor, opponentColor, conn)
		}
		var won bool
		if b.areFourConnected(playerColor) {
			gameState = win
			won = true
			wonGames++
		} else if b.areFourConnected(opponentColor) {
			gameState = lose
			won = false
			lostGames++
		} else {
			gameState = tie
		}
		/*
			wait for user to click play again
		*/
		// select{
		playAgain = <-again
		/*reset board*/
		var arr [7][6]bool
		animated = arr
		b = NewBoard()
		/*
			if you won the last game you are second in the next
		*/
		if won {
			gameState = opponentTurn
		} else {
			gameState = yourTurn
		}
	}
}

func StartGuiGame() {
	ebiten.SetWindowSize(640, 640)
	ebiten.SetWindowTitle("Connect four")
	ebiten.RunGame(&Game{})
}
