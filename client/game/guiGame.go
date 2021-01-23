package gameLogic

import (
	"bytes"
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
	tt, _ := opentype.Parse(fonts.MPlus1pRegular_ttf)
	mplusNormalFont, _ = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    20,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	initBallYCoords()
}

type Game struct{}

type GameState int

func initBallYCoords() {
	for i := 0; i < 7; i++ {
		for j := 0; j < 6; j++ {
			ballYcoords[i][j] = -tileHeight
		}
	}
}

const (
	yourTurn GameState = iota
	opponentTurn
	win
	lose
	tie
	animation
	opponentAnimation
	menu
	waitingForConnect
	waitingForToken
	connectToRoomWithToken
	cantConnectToServer
	enterAIdifficulty
)

const (
	batsX             = 440
	batsY             = 200
	secondsToMakeTurn = 59
	fps               = 60
	tileHeight        = 65
	tileOffset        = 10
	boardX            = 84
	boardY            = 130
	gravity           = 0.5
)

//the column the opponent has chosen last
var opponentLastCol int
var frameCount int
var gameState GameState = menu

var ballYcoords [7][6]float64
var ballFallSpeed [7][6]float64

var mplusNormalFont font.Face

//this is used to receive information for setting up an online game
var serverCommunicationChannel chan ServerMessage = make(chan ServerMessage)

//messages shown during a match of the game
var messages [7]string = [7]string{"Your turn", "Other's turn", "You win!", "You lost.", "Tie.", "...", "..."}

//the token with which a user connects or the token received by server
var token string

var gm GameManager

func changeGameStateBasedOnGameManagerState(gmState int) {
	if gmState != 0 {
		if gmState == Win {
			gameState = win
		}
		if gmState == Lose {
			gameState = lose
		}
		if gmState == Tie {
			gameState = tie
		}
	}
}

//the main logic of the game, changing game state moving between menus and starting a match of the game
func (g *Game) Update() error {
	press := inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft)

	if gameState == yourTurn || gameState == opponentTurn {
		frameCount++
	}

	if gameState == animation || gameState == opponentAnimation {
		frameCount = 0
	}

	if frameCount == fps*secondsToMakeTurn {
		os.Exit(1)
	}

	if gameState == yourTurn && press {
		mouseX, _ := ebiten.CursorPosition()
		if gm.makePlayerTurn(xcoordToColumn(mouseX)) {
			gameState = animation
			go func() {
				gmState := gm.GetState()
				time.Sleep(1 * time.Second)
				if gmState == Running {
					gameState = opponentTurn
				} else {
					changeGameStateBasedOnGameManagerState(gmState)
				}
			}()
		}
	}

	if gameState == opponentTurn {
		gameState = animation
		go func() {
			opponentLastCol = gm.makeOpponentTurn()
			gameState = opponentAnimation
			time.Sleep(1 * time.Second)
			gmState := gm.GetState()
			if gmState == Running {
				gameState = yourTurn
			} else {
				changeGameStateBasedOnGameManagerState(gmState)
			}
		}()
	}

	if gameState == menu && ebiten.IsKeyPressed(ebiten.KeyA) {
		gameState = enterAIdifficulty
	}

	if gameState == enterAIdifficulty {
		diff := string(ebiten.InputChars())
		if len(diff) == 1 {
			difficulty, err := strconv.Atoi(diff)
			if err == nil {
				gameState = yourTurn
				gm = NewGameManager(nil, true, difficulty+3)
			}
		}
	}

	if gameState == menu && ebiten.IsKeyPressed(ebiten.KeyO) {
		gameState = waitingForConnect
		go QuickplayLobby(serverCommunicationChannel)
	}

	if gameState == menu && ebiten.IsKeyPressed(ebiten.KeyR) {
		gameState = waitingForToken
		go CreateRoom(serverCommunicationChannel)
	}

	if gameState == menu && inpututil.IsKeyJustReleased(ebiten.KeyC) {
		gameState = connectToRoomWithToken
	}

	if gameState == connectToRoomWithToken {
		token += string(ebiten.InputChars())
		if len(token) == 5 {
			gameState = waitingForConnect
			go ConnectToRoom(token, serverCommunicationChannel)
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
				gm = NewGameManager(gameInfo.conn, false, 0)
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

			gmState := gm.GetState()
			gm.resetGame()
			var s [7][6]float64
			ballFallSpeed = s
			initBallYCoords()
			if gmState == Win{
				gameState = opponentTurn
			}else{
				gameState = yourTurn
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
	text.Draw(screen, "W  "+strconv.Itoa(gm.GetWonGames())+":"+strconv.Itoa(gm.GetLostGames())+"  L", mplusNormalFont, boardX, 50, color.White)
	text.Draw(screen, msg, mplusNormalFont, boardX, 580, color.White)
	text.Draw(screen, "00:"+strconv.Itoa(secondsToMakeTurn-frameCount/fps), mplusNormalFont, 500, 580, color.White)

	drawOwl(screen)
	if gameState == opponentAnimation {
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
	for i := 0; i < 6; i++ {
		for j := 0; j < 7; j++ {
			if gm.GetHoleColor(i, j) == PlayerTwoColor {
				drawBall(j, i, PlayerTwoColor, screen)
			} else if gm.GetHoleColor(i, j) == PlayerOneColor {
				drawBall(j, i, PlayerOneColor, screen)
			}
		}
	}
}

//drawWinnerDors draws the dots indicating where the winner has four connected balls
func drawWinnerDots(screen *ebiten.Image) {
	win, dotsY, dotsX := gm.WhereConnected()
	if !win {
		return
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

	destY := float64(y) * tileHeight
	fallY := &ballYcoords[x][y]
	fallSpeed := &ballFallSpeed[x][y]

	*fallY += *fallSpeed
	*fallSpeed += gravity
	if *fallY > destY {
		*fallY = destY
		*fallSpeed = 0
	}
	op.GeoM.Translate(float64(x)*tileHeight, *fallY)

	if player == PlayerTwoColor {
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

//StartGuiGame initializes the game and the gui, this is the entry point for the whole game
func StartGuiGame() {
	ebiten.SetWindowSize(640, 640)
	ebiten.SetWindowTitle("Connect four")
	ebiten.RunGame(&Game{})
}
