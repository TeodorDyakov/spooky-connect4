package main

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"image/color"
	_ "image/png"
	_ "image/jpeg"
	"log"
	"net"
	"os"
	"strconv"
	"time"
)

var boardImage *ebiten.Image
var red *ebiten.Image
var yellow *ebiten.Image
var lineV *ebiten.Image
var lineH *ebiten.Image
var lineU *ebiten.Image
var lineD *ebiten.Image

func init() {
	var err error
	boardImage, _, err = ebitenutil.NewImageFromFile("images/conn4.jpeg")
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
	// lineH, _, err = ebitenutil.NewImageFromFile("images/lineH.png")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// lineV, _, err = ebitenutil.NewImageFromFile("images/lineV.png")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// lineD, _, err = ebitenutil.NewImageFromFile("images/lineD.png")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// lineU, _, err = ebitenutil.NewImageFromFile("images/lineU.png")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	tt, err := opentype.Parse(fonts.MPlus1pRegular_ttf)
	if err != nil {
		log.Fatal(err)
	}
	const dpi = 72
	mplusNormalFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    24,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
}

type Game struct{}

const (
	tileHeight           = 65
	tileOffset           = 10
	CONN_HOST            = "localhost"
	CONN_PORT            = "12345"
	CONN_TYPE            = "tcp"
	PLAYER_ONE_COLOR     = "◯"
	PLAYER_TWO_COLOR     = "⬤"
	MIN_DIFFICULTY       = 1
	MAX_DIFFICULTY       = 12
	SECONDS_TO_MAKE_TURN = 60
	fps                  = 60
	gravity              = 1
)

var fallY float64 = 0
var fallSpeed float64 = 5
var mplusNormalFont font.Face
var gameOver bool = false
var waiting bool = false
var playerOneWin bool = false
var gameStarted bool = false
var frameCount int = 0
var lastFrameClicked bool

var mouseClickBuffer chan int = make(chan int)
var b *Board = NewBoard()

func (g *Game) Update() error {
	press := inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft)

	if !press && !gameOver && !waiting && lastFrameClicked {
		mouseX, _ := ebiten.CursorPosition()
		select {
		case mouseClickBuffer <- col(mouseX):
		default:
		}
	}
	if press {
		lastFrameClicked = true
	} else {
		lastFrameClicked = false
	}
	return nil
}

func col(x int) int {
	return int(float64(x - tileOffset) / tileHeight)
}

func (g *Game) Draw(screen *ebiten.Image) {
	var msg string
	if !gameStarted {
		msg = "waiting"
	} else if gameOver {
		if playerOneWin {
			msg = "you win :D"
		} else {
			msg = "you lose :("
		}
	} else if !waiting {
		msg = "your turn"
	} else {
		msg = "other's turn"
	}
	if !gameOver {
		frameCount++
	}
	if frameCount == fps * SECONDS_TO_MAKE_TURN {
		os.Exit(1)
	}
	screen.DrawImage(boardImage, nil)
	text.Draw(screen, msg, mplusNormalFont, 470, 350, color.White)
	text.Draw(screen, strconv.Itoa(SECONDS_TO_MAKE_TURN - frameCount/fps)+"s", mplusNormalFont, 470, 380, color.White)

	for i := 0; i < len(b.board); i++ {
		for j := 0; j < len(b.board[0]); j++ {
			if b.board[i][j] == PLAYER_TWO_COLOR {
				drawTile(j, i, PLAYER_TWO_COLOR, screen)
			} else if b.board[i][j] == PLAYER_ONE_COLOR {
				drawTile(j, i, PLAYER_ONE_COLOR, screen)
			}
		}
	}
	// if(gameOver){
	// 	con, y, x, lineType := b.whereConnected(PLAYER_ONE_COLOR)
	// 	if !con{
	// 		con, y, x, lineType = b.whereConnected(PLAYER_ONE_COLOR)
	// 	}
	// 	op := &ebiten.DrawImageOptions{}
	// 	op.GeoM.Translate(tileOffset+float64(x)*tileHeight + 30, tileOffset+float64(y)*tileHeight+30)
	// 	if lineType == 0{
	// 		screen.DrawImage(lineH, op)
	// 	}else if lineType == 1{
	// 		screen.DrawImage(lineV, op)
	// 	}else if lineType == 2{
	// 		screen.DrawImage(lineD, op)
	// 	}else if lineType == 3{
	// 		screen.DrawImage(lineU, op)
	// 	}
	// }
}

var animated [7][6]bool

func drawTile(x, y int, player string, screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	destY := tileOffset + float64(y)*tileHeight
	if animated[x][y] {
		op.GeoM.Translate(tileOffset+float64(x)*tileHeight, tileOffset+float64(y)*tileHeight)
	} else {
		fallY += fallSpeed
		fallSpeed += gravity
		if fallY > destY {
			fallY = destY
			fallSpeed = 0
			animated[x][y] = true
		}
		op.GeoM.Translate(tileOffset+float64(x)*tileHeight, fallY)
		if animated[x][y] {
			fallY = 0
		}
	}
	if player == PLAYER_TWO_COLOR {
		screen.DrawImage(red, op)
	} else {
		screen.DrawImage(yellow, op)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 626, 417
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
	boardCopy := NewBoard()
	gameStarted = true

	readyToStartGui <- 1
	for !b.gameOver() {
		if waiting {
			_, bestMove := alphabeta(boardCopy, true, 0, SMALL, BIG, difficulty)
			b.drop(bestMove, PLAYER_TWO_COLOR)
			boardCopy.drop(bestMove, PLAYER_TWO_COLOR)
			time.Sleep(1 * time.Second)
			waiting = false
		} else {
			column := <-mouseClickBuffer
			if b.drop(column, PLAYER_ONE_COLOR) {
				boardCopy.drop(column, PLAYER_ONE_COLOR)
				time.Sleep(1 * time.Second)
				waiting = true
				frameCount = 0
			}
		}
	}
	gameOver = true
	if b.areFourConnected(PLAYER_ONE_COLOR) {
		playerOneWin = true
	} else if b.areFourConnected(PLAYER_TWO_COLOR) {
		playerOneWin = false
	} else {
		// fmt.Println("Tie")
	}
}

var readyToStartGui chan int = make(chan int, 1)

func playMultiplayer() {
	var conn net.Conn
	var color string
	var opponentColor string

	waiting, conn = lobby()

	if waiting {
		color = PLAYER_TWO_COLOR
		opponentColor = PLAYER_ONE_COLOR
	} else {
		color = PLAYER_ONE_COLOR
		opponentColor = PLAYER_TWO_COLOR
	}
	gameStarted = true
	readyToStartGui <- 1
	for !b.gameOver() {

		if waiting {
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
			waiting = false
		} else {
			column := <-mouseClickBuffer
			if b.drop(column, color) {
				frameCount = 0
				waiting = true
				_, err := fmt.Fprintf(conn, "%d\n", column)
				if err != nil {
					panic(err)
				}
				time.Sleep(1 * time.Second)
			}
		}
	}
	gameOver = true
	fmt.Fprintf(conn, "end")
	if b.areFourConnected(color) {
		playerOneWin = true
	} else if b.areFourConnected(opponentColor) {
		playerOneWin = false
	} else {

	}
}

func main() {
	ebiten.SetWindowSize(626, 417)
	ebiten.SetWindowTitle("Render an image")

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
	if err := ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}

}
