package connect4FMI

import (
	"bytes"
	"fmt"
	resources "github.com/TeodorDyakov/spooky-connect4/client/resources"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"net"
	"os"
	"strconv"
	"time"
)

var backgroundImage *ebiten.Image
var owl *ebiten.Image
var redBallImage *ebiten.Image
var dot *ebiten.Image
var ghost *ebiten.Image
var greenBallImage *ebiten.Image
var boardImage *ebiten.Image
var bats *ebiten.Image
var batsX, batsY float64

func loadImageFromFile(relativePath string) *ebiten.Image {
	var err error
	image, _, err := ebitenutil.NewImageFromFile(relativePath)
	if err != nil {
		log.Fatal(err)
	}
	return image
}

func init() {
	img, _, _ := image.Decode(bytes.NewReader(resources.Ghost_png))
	ghost = ebiten.NewImageFromImage(img)

	img, _, _ = image.Decode(bytes.NewReader(resources.Background_png))
	backgroundImage = ebiten.NewImageFromImage(img)

	img, _, _ = image.Decode(bytes.NewReader(resources.Red_png))
	redBallImage = ebiten.NewImageFromImage(img)

	img, _, _ = image.Decode(bytes.NewReader(resources.Green_png))
	greenBallImage = ebiten.NewImageFromImage(img)

	img, _, _ = image.Decode(bytes.NewReader(resources.Owl_png))
	owl = ebiten.NewImageFromImage(img)

	img, _, _ = image.Decode(bytes.NewReader(resources.Dot_png))
	dot = ebiten.NewImageFromImage(img)

	img, _, _ = image.Decode(bytes.NewReader(resources.Bats_png))
	bats = ebiten.NewImageFromImage(img)

	img, _, _ = image.Decode(bytes.NewReader(resources.Board_png))
	boardImage = ebiten.NewImageFromImage(img)

	batsX = 440
	batsY = 200
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
	cantConnectToServer
	enterAIdifficulty
)

const (
	secondsToMakeTurn = 59
	fps               = 60
	tileHeight        = 65
	tileOffset        = 10
	boardX            = 84
	boardY            = 130
	gravity           = 0.3
	playerOneColor    = "◯"
	playerTwoColor    = "⬤"
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
var board *Board = NewBoard()
var playingAgainstAi bool
var mplusNormalFont font.Face
var fallY float64 = -tileHeight
var again chan bool = make(chan bool)
var mouseClickBuffer chan int = make(chan int)
var messages [5]string = [5]string{"Your turn", "Other's turn", "You win!", "You lost.", "Tie."}
var opponentAnimation bool
var difficulty int
var serverCommunicationChannel chan serverMessage = make(chan serverMessage)
var token string

func (g *Game) Update() error {
	press := inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft)

	if gameState == yourTurn || gameState == opponentTurn {
		frameCount++
	}

	if frameCount == fps*secondsToMakeTurn {
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
		gameState = enterAIdifficulty
	}

	if gameState == enterAIdifficulty {
		diff := string(ebiten.InputChars())
		if len(diff) == 1 {
			var err error
			difficulty, err = strconv.Atoi(diff)
			if err == nil {
				difficulty += 3
				gameState = yourTurn
				go playAgainstAi()
			}
		}

	}
	if gameState == menu && ebiten.IsKeyPressed(ebiten.KeyO) {
		gameState = waitingForConnect
		go quickplayLobby(serverCommunicationChannel)
	}

	if gameState == menu && ebiten.IsKeyPressed(ebiten.KeyR) {
		gameState = waitingForToken
		go createRoom(serverCommunicationChannel)
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
		case gameInfo := <-serverCommunicationChannel:
			if gameInfo.conn == nil {
				gameState = cantConnectToServer
			} else {
				token = gameInfo.token
				gameState = waitingForConnect
			}
		default:
		}
	}

	if gameState == waitingForConnect {
		select {
		case gameInfo := <-serverCommunicationChannel:
			if gameInfo.conn == nil {
				token = ""
				gameState = cantConnectToServer
			} else {
				if gameInfo.waiting {
					gameState = opponentTurn
				} else {
					gameState = yourTurn
				}
				go playMultiplayer(gameInfo.waiting, gameInfo.conn)
			}
		default:
		}
	}

	if gameState == cantConnectToServer {
		frameCount++
		if frameCount == 2*fps {
			frameCount = 0
			gameState = menu
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
	screen.DrawImage(backgroundImage, nil)
	op := &ebiten.DrawImageOptions{}

	op.GeoM.Translate(batsX, batsY)
	screen.DrawImage(bats, op)
	op.GeoM.Reset()

	op.GeoM.Translate(boardX, boardY)
	if gameState == menu || gameState == cantConnectToServer {
		screen.DrawImage(boardImage, op)
		text.Draw(screen, "[A] - play against AI", mplusNormalFont, boardX, boardY-30, color.White)
		text.Draw(screen, "[R] - create a room", mplusNormalFont, boardX, 570, color.White)
		text.Draw(screen, "[C] - connect to a room", mplusNormalFont, boardX+250, 570, color.White)
		text.Draw(screen, "[O] - play online (quick play)", mplusNormalFont, boardX+250, boardY-30, color.White)
		if gameState == cantConnectToServer {
			text.Draw(screen, "Can't connect to server!", mplusNormalFont, 200, 200, color.White)
		}
		return
	}

	if gameState == connectToRoomWithToken {
		screen.DrawImage(boardImage, op)
		text.Draw(screen, "Enter the code for room:\n"+token, mplusNormalFont, 200, 50, color.White)
		return
	}

	if gameState == enterAIdifficulty {
		screen.DrawImage(boardImage, op)
		text.Draw(screen, "Enter difficulty (1-9)\n"+token, mplusNormalFont, 200, 50, color.White)
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
	text.Draw(screen, "00:"+strconv.Itoa(secondsToMakeTurn-frameCount/fps), mplusNormalFont, 500, 580, color.White)

	drawOwl(screen)
	if opponentAnimation {
		drawGhost(screen)
	}

	drawBalls(screen)
	screen.DrawImage(boardImage, op)

	if isGameOver() {
		text.Draw(screen, "Click here\nto play again", mplusNormalFont, 250, 580, color.White)
		if gameState != tie {
			drawWinnerDots(screen)
		}
	}
}

func drawBalls(screen *ebiten.Image) {
	for i := 0; i < len(board.board); i++ {
		for j := 0; j < len(board.board[0]); j++ {
			if board.board[i][j] == playerTwoColor {
				drawBall(j, i, playerTwoColor, screen)
			} else if board.board[i][j] == playerOneColor {
				drawBall(j, i, playerOneColor, screen)
			}
		}
	}
}

func drawWinnerDots(screen *ebiten.Image) {
	playerOneWin, dotsY, dotsX := board.WhereConnected(playerOneColor)
	if !playerOneWin {
		_, dotsY, dotsX = board.WhereConnected(playerTwoColor)
	}
	for i := 0; i < 4; i++ {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(boardX+tileOffset, boardY+tileOffset)
		op.GeoM.Translate(float64(dotsX[i])*tileHeight+25, float64(dotsY[i])*tileHeight+25)
		screen.DrawImage(dot, op)
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
	owlX := xcoordToColumn(mouseX)*tileHeight + boardX
	op.GeoM.Translate(float64(owlX), boardY-75)
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
	if player == playerTwoColor {
		screen.DrawImage(redBallImage, op)
	} else {
		screen.DrawImage(greenBallImage, op)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 640, 640
}

/*
on which column to Drop based on x coordinate of click
*/
func xcoordToColumn(x int) int {
	return int(float64(x-tileOffset-boardX) / tileHeight)
}

/*
choose difficulty and start AI game loop
*/

func playAgainstAi() {
	playingAgainstAi = true
	gameLogic(playerOneColor, playerTwoColor, nil)
}

/*
show menu to choose game type - quick or with friend. After user chooses from console
starts the game loop.
*/
func playMultiplayer(wait bool, conn net.Conn) {
	playerColor := playerOneColor
	opponentColor := playerTwoColor
	if wait {
		playerColor = playerTwoColor
		opponentColor = playerOneColor
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
			column = getAiMove(board, difficulty)
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
		board.Drop(column, opponentColor)
		/*
			wait for the animation of falling circle to finish
		*/
		time.Sleep(1 * time.Second)
		opponentAnimation = false
		frameCount = 0
		gameState = yourTurn
	} else if gameState == yourTurn {
		column := <-mouseClickBuffer
		if board.Drop(column, playerColor) {
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
		for !board.gameOver() {
			playTurn(playerColor, opponentColor, conn)
		}
		var won bool
		if board.areFourConnected(playerColor) {
			gameState = win
			won = true
			wonGames++
		} else if board.areFourConnected(opponentColor) {
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
		board = NewBoard()
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
