package connect4FMI

import (
	"fmt"
	"os"
	"log"
	"net"
	"time"
	"strconv"
	"image/color"
	_ "image/png"
	_ "image/jpeg"
	"golang.org/x/image/font"
	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/font/opentype"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
)

var bg *ebiten.Image
var red *ebiten.Image
var yellow *ebiten.Image
var boardImage *ebiten.Image

func init() {
	var err error
	boardImage, _, err = ebitenutil.NewImageFromFile("images/conn4trans.png")
	if err != nil {
		log.Fatal(err)
	}
	bg, _, err = ebitenutil.NewImageFromFile("images/bg2.jpeg")
	if err != nil {
		log.Fatal(err)
	}
	red, _, err = ebitenutil.NewImageFromFile("images/red.png")
	if err != nil {
		log.Fatal(err)
	}
	yellow, _, err = ebitenutil.NewImageFromFile("images/yellow.png")
	if err != nil {
		log.Fatal(err)
	}
	tt, _ := opentype.Parse(fonts.MPlus1pRegular_ttf)
	mplusNormalFont, _ = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    24,
		DPI:     72,
		Hinting: font.HintingFull,
	})
}

type Game struct{}

type GameState int

const (
	yourTurn GameState   = 0
	waiting              = 1
	win                  = 2
	lose                 = 3
	tie                  = 4
)

const (
	MIN_DIFFICULTY       = 1
	MAX_DIFFICULTY       = 12
	SECONDS_TO_MAKE_TURN = 59
	fps                  = 60
	tileHeight           = 65
	tileOffset           = 10
	boardX               = 84
	boardY               = 100
	gravity              = 0.5
	PLAYER_ONE_COLOR     = "◯"
	PLAYER_TWO_COLOR     = "⬤"
	CONN_TYPE            = "tcp"
	CONN_PORT            = "12345"
	CONN_HOST            = "localhost"
)

var frameCount int
var aiDifficulty int
var gameState GameState
var animated [7][6]bool
var fallSpeed float64
var b *Board = NewBoard()
var again chan bool = make(chan bool)
var playingAgainstAi bool
var mplusNormalFont font.Face
var wonLast bool
var fallY float64 = -tileHeight
var readyToStartGui chan int = make(chan int)
var mouseClickBuffer chan int = make(chan int)
var messages [5]string = [5]string{"your turn", "other's turn", "you win :D", "you lose :(", "tie"}

func (g *Game) Update() error {
	press := inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft)

	if press {
		mouseX, _ := ebiten.CursorPosition()
		/*
			only send click event to buffer if someone is waiting for it
		*/
		select {
		case mouseClickBuffer <- col(mouseX):
		default:
		}
	}
	if isGameOver() && press {
		mouseX, mouseY := ebiten.CursorPosition()
		if mouseX >= 230 && mouseX <= 600 && mouseY >= 500 {
			resetGameState()
			if playingAgainstAi{
				go playAgainstAi()
			}else{ 
				select {
				case again <- true:
				default:
				}
			}
		}
	}
	return nil
}

func isGameOver() bool {
	return gameState == tie || gameState == win || gameState == lose
}

func (g *Game) Draw(screen *ebiten.Image) {
	var msg string = messages[gameState]

	if gameState == yourTurn || gameState == waiting{
		frameCount++
	}

	if frameCount == fps * SECONDS_TO_MAKE_TURN {
		os.Exit(1)
	}
	screen.DrawImage(bg, nil)
	op := &ebiten.DrawImageOptions{}
	text.Draw(screen, msg, mplusNormalFont, boardX, 580, color.White)
	text.Draw(screen, "00:" + strconv.Itoa(SECONDS_TO_MAKE_TURN - frameCount / fps), mplusNormalFont, 490, 580, color.White)

	for i := 0; i < len(b.board); i++ {
		for j := 0; j < len(b.board[0]); j++ {
			if b.board[i][j] == PLAYER_TWO_COLOR {
				drawTile(j, i, PLAYER_TWO_COLOR, screen)
			} else if b.board[i][j] == PLAYER_ONE_COLOR {
				drawTile(j, i, PLAYER_ONE_COLOR, screen)
			}
		}
	}
	op.GeoM.Translate(boardX, boardY)
	screen.DrawImage(boardImage, op)
	if isGameOver() {
		text.Draw(screen, "Click here\nto play again", mplusNormalFont, 230, 580, color.White)
	}
}

func drawTile(x, y int, player string, screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(boardX + tileOffset, boardY + tileOffset)
	destY := tileOffset + float64(y) * tileHeight
	if animated[x][y] {
		op.GeoM.Translate(float64(x) * tileHeight, float64(y) * tileHeight)
	} else {
		fallY += fallSpeed
		fallSpeed += gravity
		if fallY > destY {
			fallY = destY
			fallSpeed = 0
			animated[x][y] = true
		}
		op.GeoM.Translate(float64(x) * tileHeight, fallY)
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

func resetGameState() {
	var arr [7][6]bool
	animated = arr
	gameState = waiting
	b = NewBoard()
}

func col(x int) int {
	return int(float64(x - tileOffset - boardX) / tileHeight)
}

func aiGame(difficulty int) {
	boardCopy := NewBoard()
	for !b.gameOver() {
		if gameState == waiting {
			_, bestMove := alphabeta(boardCopy, true, 0, SMALL, BIG, difficulty)
			b.drop(bestMove, PLAYER_TWO_COLOR)
			boardCopy.drop(bestMove, PLAYER_TWO_COLOR)
			time.Sleep(1 * time.Second)
			gameState = yourTurn
			frameCount = 0
		} else if gameState == yourTurn {
			column := <-mouseClickBuffer
			if b.drop(column, PLAYER_ONE_COLOR) {
				boardCopy.drop(column, PLAYER_ONE_COLOR)
				time.Sleep(1 * time.Second)
				gameState = waiting
				frameCount = 0
			}
		}
	}

	if b.areFourConnected(PLAYER_ONE_COLOR) {
		gameState = win
	} else if b.areFourConnected(PLAYER_TWO_COLOR) {
		gameState = lose
	} else {
		gameState = tie
	}
}

func playAgainstAi() {
	fmt.Printf("Choose difficulty (number between %d and %d)", MIN_DIFFICULTY, MAX_DIFFICULTY)
	var option string
	fmt.Scan(&option)

	difficulty, err := strconv.Atoi(option)

	for err != nil || difficulty < MIN_DIFFICULTY || difficulty > MAX_DIFFICULTY {
		fmt.Println("Invalid input! Try again:")
		fmt.Scan(&option)
		difficulty, err = strconv.Atoi(option)
	}
	readyToStartGui <- 1
	aiDifficulty = difficulty
	playingAgainstAi = true
	aiGame(difficulty)
}

func playMultiplayer() {
	var conn net.Conn
	var color string
	var opponentColor string
	var wait bool
	wait, conn = lobby()

	if wait {
		color = PLAYER_TWO_COLOR
		opponentColor = PLAYER_ONE_COLOR
	} else {
		color = PLAYER_ONE_COLOR
		opponentColor = PLAYER_TWO_COLOR
	}
	if wait {
		gameState = waiting
	} else {
		gameState = yourTurn
	}
	readyToStartGui <- 1

	playAgain := true
	for playAgain {
		for !b.gameOver() {
			if gameState == waiting {
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
				column, _ := strconv.Atoi(msg)
				b.drop(column, opponentColor)
				time.Sleep(1 * time.Second)
				frameCount = 0
				gameState = yourTurn
			} else if gameState == yourTurn {
				column := <-mouseClickBuffer
				if b.drop(column, color) {
					frameCount = 0
					gameState = waiting
					_, err := fmt.Fprintf(conn, "%d\n", column)
					if err != nil {
						panic(err)
					}
					time.Sleep(1 * time.Second)
				}
			}
		}
		won := false
		// fmt.Fprintf(conn, "end")
		if b.areFourConnected(color) {
			gameState = win
			won = true
		} else if b.areFourConnected(opponentColor) {
			won = false
			gameState = lose
		} else {
			gameState = tie
		}
		playAgain = <- again
		
		if won {
			gameState = waiting
		}else{
			gameState = yourTurn
		}
	}
}

func StartGuiGame() {
	ebiten.SetWindowSize(640, 640)
	ebiten.SetWindowTitle("Connect four")

	fmt.Println("Hello! Welcome to connect four CMD!\n" +
		"To enter multiplayer lobby press [1]\n" + "To play against AI press [2]")

	var option string
	fmt.Scan(&option)

	for !(option == "1" || option == "2") {
		fmt.Println("Unknown command! Try again:")
		fmt.Scan(&option)
	}

	if option == "2" {
		go playAgainstAi()
	} else {
		go playMultiplayer()
	}

	<-readyToStartGui
	ebiten.RunGame(&Game{})
}
