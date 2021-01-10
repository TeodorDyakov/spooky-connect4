package connect4FMI

import (
	"bytes"
	"fmt"
	resources "github.com/TeodorDyakov/spooky-connect4/client/resources"
	"github.com/hajimehoshi/ebiten/v2"
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

func byteSliceToEbitenImage(arr []byte) *ebiten.Image {
	img, _, err := image.Decode(bytes.NewReader(arr))
	if err != nil {
		log.Fatal(err)
	}
	return ebiten.NewImageFromImage(img)
}

func init() {
	ghost = byteSliceToEbitenImage(resources.Ghost_png)
	backgroundImage = byteSliceToEbitenImage(resources.Background_png)
	redBallImage = byteSliceToEbitenImage(resources.Red_png)
	greenBallImage = byteSliceToEbitenImage(resources.Green_png)
	owl = byteSliceToEbitenImage(resources.Owl_png)
	dot = byteSliceToEbitenImage(resources.Dot_png)
	bats = byteSliceToEbitenImage(resources.Bats_png)
	boardImage = byteSliceToEbitenImage(resources.Board_png)
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
	gravity           = 0.5
	playerOneColor    = "◯"
	playerTwoColor    = "⬤"
)

//the column the opponent has chosen last
var opponentLastCol int
var lostGames int
var wonGames int
var frameCount int
var gameState GameState = menu

//wasFallAnimated already for the ball which should fall at the given row and column
var wasFallAnimated [7][6]bool

//fallSpeed of a falling ball
var fallSpeed float64
var board *Board = NewBoard()
var playingAgainstAi bool
var mplusNormalFont font.Face

//fallY - the Y coordinate of the falling ball
var fallY float64 = -tileHeight

//again - channel used to wait for user to click playAgain button
var playAgainClick chan struct{} = make(chan struct{})

//channel used to send mouse clicks during game, idicating which column is clicked by user
var columnClicked chan int = make(chan int)

//this is used to receive information for setting up an online game
var serverCommunicationChannel chan serverMessage = make(chan serverMessage)

//messages shown during a match of the game
var messages [5]string = [5]string{"Your turn", "Other's turn", "You win!", "You lost.", "Tie."}

//whether an opponent is running
var opponentAnimation bool

//difficutly of the AI
var difficulty int

//the token with which a user connects or the token received by server
var token string

//the main logic of the game, changing game state moving between menus and starting a match of the game
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
		case columnClicked <- xcoordToColumn(mouseX):
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
				if gameInfo.isSecond {
					gameState = opponentTurn
				} else {
					gameState = yourTurn
				}
				go playMultiplayer(gameInfo.isSecond, gameInfo.conn)
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
			case playAgainClick <- struct{}{}:
			default:
			}
		}
	}
	return nil
}

//isGameOver returns whether the game is over
func isGameOver() bool {
	return gameState == tie || gameState == win || gameState == lose
}

//this fucntion draws the graphic of the game based on the gameState
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

//drawBalls draws all the balls to the screen
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

//drawWinnerDors draws the dots indicating where the winner has four connected balls
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

//drawGhost draws the ghost image to the screen
func drawGhost(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(opponentLastCol)*tileHeight+boardX+10, boardY-75)
	screen.DrawImage(ghost, op)
}

//drawOwl draws the owl image to the screen
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
	op.GeoM.Translate(float64(owlX), boardY-80)
	screen.DrawImage(owl, op)
}

//drawBall draws the ball to the screen
func drawBall(x, y int, player string, screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(boardX+tileOffset, boardY+tileOffset)
	destY := tileOffset + float64(y)*tileHeight

	if wasFallAnimated[x][y] {
		op.GeoM.Translate(float64(x)*tileHeight, float64(y)*tileHeight)
	} else {
		fallY += fallSpeed
		fallSpeed += gravity
		if fallY > destY {
			fallY = destY
			fallSpeed = 0
			wasFallAnimated[x][y] = true
		}
		op.GeoM.Translate(float64(x)*tileHeight, fallY)
		if wasFallAnimated[x][y] {
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

//xcoordToColumn returns the column correspondidng which contains the x coordinate
func xcoordToColumn(x int) int {
	return int(float64(x-tileOffset-boardX) / tileHeight)
}

//playAgainstAI starts a game against an AI
func playAgainstAi() {
	playingAgainstAi = true
	gameLogic(playerOneColor, playerTwoColor, nil)
}

//playMulitplayer initiliazes the color of player and opponent based on whether the player isSecond and
//stars a mutiplayer mathc on the given connection
func playMultiplayer(isSecond bool, conn net.Conn) {
	playerColor := playerOneColor
	opponentColor := playerTwoColor
	if isSecond {
		playerColor = playerTwoColor
		opponentColor = playerOneColor
		gameState = opponentTurn
	} else {
		gameState = yourTurn
	}
	gameLogic(playerColor, opponentColor, conn)
}

//playTurn executes the logic for a move by one player then a move the next player
//conn is the connection needed for multiplayer
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
		column := <-columnClicked
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

//plays a full match of the game than waits for the user to click play Again and starts another game
func gameLogic(playerColor, opponentColor string, conn net.Conn) {
	for {
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
		<-playAgainClick
		/*reset board*/
		var arr [7][6]bool
		wasFallAnimated = arr
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

//StartGuiGame initializes the game and the gui, this is the entry point for the whole game
func StartGuiGame() {
	ebiten.SetWindowSize(640, 640)
	ebiten.SetWindowTitle("Connect four")
	ebiten.RunGame(&Game{})
}
