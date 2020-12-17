package main

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
	"time"
)

const (
	CONN_HOST        = "localhost"
	CONN_PORT        = "12345"
	CONN_TYPE        = "tcp"
	PLAYER_ONE_COLOR = "○"
	PLAYER_TWO_COLOR = "◙"
	MIN_DIFFICULTY   = 1
	MAX_DIFFICULTY   = 12
)

var b *Board = NewBoard()

func init() {
	rand.Seed(time.Now().UnixNano())
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

	waiting := false

	for !b.areFourConnected(PLAYER_ONE_COLOR) && !b.areFourConnected(PLAYER_TWO_COLOR) {

		clearConsole()
		b.printBoard()

		if waiting {
			fmt.Println("waiting for oponent move...\n")

			_, bestMove := alphabeta(b, true, 0, SMALL, BIG, difficulty)
			b.drop(bestMove, PLAYER_TWO_COLOR)
			waiting = false
		} else {
			for {
				fmt.Printf("Enter column to drop: ")

				var column int
				_, err = fmt.Scan(&column)

				if err != nil || !b.drop(column, PLAYER_ONE_COLOR) {
					fmt.Println("You cant place here! Try another column")
				} else {
					waiting = true
					break
				}
			}
		}
	}

	clearConsole()
	b.printBoard()
	if b.areFourConnected(PLAYER_ONE_COLOR) {
		fmt.Println("You won!")
	} else {
		fmt.Println("You lost.")
	}
}

func playMultiplayer() {
	var conn net.Conn
	var color string
	var opponentColor string

	var waiting bool

	fmt.Println("Connecting to", CONN_TYPE, "server", CONN_HOST+":"+CONN_PORT)
	conn, err := net.Dial(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
	if err != nil {
		fmt.Println("Error connecting:", err.Error())
		os.Exit(1)
	}

	fmt.Println("enter 1 to make a room or 2 to connect to room")
	var input int
	fmt.Scan(&input)

	if input == 1 {
		fmt.Fprintf(conn, "wait\n")

		var token string
		fmt.Fscan(conn, &token)
		fmt.Printf("You token is:%s\n", token)
		fmt.Println("waiting for a friend to connect...")

		var msg string
		fmt.Fscan(conn, &msg)

		if msg == "go"{
			color = PLAYER_TWO_COLOR
			opponentColor = PLAYER_ONE_COLOR
			waiting = true
		}else{
			fmt.Println("cant connect to friend!")
			return
		}
	} else {
		fmt.Fprintf(conn, "connect\n")

		var token string
		fmt.Printf("Enter friend token\n")
		fmt.Scan(&token)

		fmt.Fprintf(conn, "%s\n", token)
		var msg string
		fmt.Fscan(conn, &msg)

		if msg == "go"{
			color = PLAYER_ONE_COLOR
			opponentColor = PLAYER_TWO_COLOR
			waiting = false
		} else{
			fmt.Println("cant connect to friend!")
			return
		}
	}

	for !b.areFourConnected(color) && !b.areFourConnected(opponentColor) {

		clearConsole()
		b.printBoard()

		if waiting {
			fmt.Println("waiting for oponent move...\n")

			c1 := make(chan int, 1)

			go func() {
				var column int
				fmt.Fscan(conn, &column)
				c1 <- column
			}()

			select {
			case otherPlayerColumn := <-c1:
				b.drop(otherPlayerColumn, opponentColor)
				waiting = false
			case <-time.After(60 * time.Second):
				fmt.Println("timeout Opponent failed to make a move in 60 seconds")
				return
			}

		} else {
			for {
				fmt.Printf("Enter column to drop: ")

				var column int
				_, err = fmt.Scan(&column)

				if err != nil || !b.drop(column, color) {
					fmt.Println("You cant place here! Try another column")
				} else {
					fmt.Fprintf(conn, "%d\n", column)
					waiting = true
					break
				}
			}
		}
	}

	fmt.Fprintf(conn, "end")

	clearConsole()
	b.printBoard()
	if b.areFourConnected(color) {
		fmt.Println("You won!")
	} else {
		fmt.Println("You lost.")
	}
}

func main() {

	fmt.Println("Hello! Welcome to connect four CMD!\n" +
		"To enter multiplayer lobby press [1]\n" + "To play against AI press [2]\n")

	var option string
	fmt.Scan(&option)

	for !(option == "1" || option == "2") {
		fmt.Println("Unknown command! Try again:")
		fmt.Scan(&option)
	}

	if option == "2" {
		playAgainstAi()
		return
	} else {
		playMultiplayer()
	}

}
	