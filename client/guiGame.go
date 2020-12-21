package main

import (
	_ "image/png"
	"log"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"fmt"
	"strconv"
	"net"
    "image/color"   
)

var boardImage *ebiten.Image
var red *ebiten.Image
var yellow *ebiten.Image

func init() {
	var err error
	boardImage, _, err = ebitenutil.NewImageFromFile("images/con4.png")
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
	h, _ := boardImage.Size()
	tileHeight = (float64(h))/7.0

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
var tileHeight float64
var mplusNormalFont font.Face

const (
	CONN_HOST            = "localhost"
	CONN_PORT            = "12345"
	CONN_TYPE            = "tcp"
	PLAYER_ONE_COLOR     = "◯"
	PLAYER_TWO_COLOR     = "⬤"
	MIN_DIFFICULTY       = 1
	MAX_DIFFICULTY       = 12
	SECONDS_TO_MAKE_TURN = 60
)

var lastFrameClick bool = false
var gameOver bool = false
var waiting bool = false
var playerOneWin bool = false
var gameStarted bool

var c chan int = make(chan int, 1)

func (g *Game) Update() error {
	press := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
	
	if press && !lastFrameClick && !gameOver && !waiting{
       mouseX, _ := ebiten.CursorPosition()
       c <- col(mouseX)	
    }
    if press{
		lastFrameClick = true
	}else{
		lastFrameClick = false
	}
    return nil
}

var cnt int

func col(x int) int {
	return int(float64(x-10)/tileHeight) 
}

var b *Board = NewBoard()

func (g *Game) Draw(screen *ebiten.Image) {
	var msg string
	if !gameStarted{
		msg = "waiting"
	}else if gameOver {
		if playerOneWin{
			msg = "you win :D"
		}else{
			msg = "you lose :("
		}
	} else if !waiting{
		msg = "your turn"
	} else{
		msg = "opponent turn"
	}
	text.Draw(screen, msg, mplusNormalFont, 200, 450, color.White)
	screen.DrawImage(boardImage, nil)
	for i := 0; i < len(b.board); i++ {
		for j := 0; j < len(b.board[0]); j++ {
			if b.board[i][j] == PLAYER_TWO_COLOR{
				drawTile(j, i, PLAYER_TWO_COLOR, screen)
			} else if b.board[i][j] == PLAYER_ONE_COLOR{
				drawTile(j, i, PLAYER_ONE_COLOR, screen)
			}
		}
	}
}

func drawTile(x, y int, player string, screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}

	// var tileWidth float64 = (float64(h))/7.0
	op.GeoM.Translate(10 + float64(x) * tileHeight, 10+ float64(y)*tileHeight)
	// op.GeoM.Translate(screenWidth/2, screenHeight/2)
	if(player == PLAYER_TWO_COLOR){
		screen.DrawImage(red, op)
	} else {
		screen.DrawImage(yellow, op)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 490, 480
}

func playAgainstAi(){
	boardCopy := NewBoard()
	for !b.gameOver(){
		if waiting {
			_, bestMove := alphabeta(boardCopy, true, 0, SMALL, BIG, 12)
			b.drop(bestMove, PLAYER_TWO_COLOR)
			boardCopy.drop(bestMove, PLAYER_TWO_COLOR)
			waiting = false
		} else {
			var column int
			column = <- c
			if !b.drop(column, PLAYER_ONE_COLOR) {
			} else {
				boardCopy.drop(column, PLAYER_ONE_COLOR)
				waiting = true
			}
		}
		gameOver = true
		if b.areFourConnected(PLAYER_ONE_COLOR) {
		playerOneWin = true
		} else if b.areFourConnected(PLAYER_TWO_COLOR) {
			playerOneWin =false
		} else {
			// fmt.Println("Tie")
		}
	}
}

var c1 chan int = make(chan int, 1)

func playMultiplayer(){
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
	c1 <- 1
	for !b.gameOver() {

		if waiting {
			// fmt.Println("waiting for oponent move...\n")

			var msg string
			_, err := fmt.Fscan(conn, &msg)
			if err != nil{
				panic(err)
			}
			if msg == "timeout" || msg == "error"{
				fmt.Println("opponent disconnected!")
				panic(nil)
				return
			}
			column, _ := strconv.Atoi(msg)
			b.drop(column, opponentColor)
			waiting = false
		} else {
				column := <- c
				if b.drop(column, color) {
					waiting = true
				_, err := fmt.Fprintf(conn, "%d\n", column)
				if err != nil{
					panic(err)
				}
			}
		}
	}
	gameOver = true
	fmt.Fprintf(conn, "end")
	if b.areFourConnected(color){
		playerOneWin = true
	}else if b.areFourConnected(opponentColor){
		playerOneWin = false
	}else{

	}
}

func main() {
	ebiten.SetWindowSize(490, 480)
	ebiten.SetWindowTitle("Render an image")

	// boardCopy := NewBoard()
				
	go playMultiplayer()
	<-c1
	if err := ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}
	
}

