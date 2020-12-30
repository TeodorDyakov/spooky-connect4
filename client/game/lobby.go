package connect4FMI

import (
	"fmt"
	"net"
	"os"
	"bufio"
	"strings"
)

var (
	host string
	port string
	connType = "tcp"
)

func init(){
	file, err := os.Open("connectionConfig")
    if err != nil {
        panic(err)
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    scanner.Scan()
    host = strings.Split(scanner.Text(), ":")[1]
    scanner.Scan()
    port = strings.Split(scanner.Text(), ":")[1]
}

type serverMessage struct {
	conn    net.Conn
	waiting bool
	token   string
}

func createRoom(info chan<- serverMessage) {
	conn, err := net.Dial(connType, host+":"+port)
	if err != nil {
		info <- serverMessage{nil, false, ""}
		return
	}
	fmt.Fprintf(conn, "wait\n")
	var token string
	fmt.Fscan(conn, &token)
	info <- serverMessage{conn, false, token}

	var msg string
	fmt.Fscan(conn, &msg)
	var waiting bool
	if msg == "second" {
		waiting = true
	} else if msg == "first" {
		waiting = false
	}
	info <- serverMessage{conn, waiting, ""}
}

func connectToRoom(token string, info chan<- serverMessage) {
	conn, err := net.Dial(connType, host+":"+port)
	if err != nil {
		info <- serverMessage{nil, false, ""}
		return
	}
	fmt.Fprintf(conn, "connect\n")
	fmt.Fprintf(conn, "%s\n", token)
	var msg string
	fmt.Fscan(conn, &msg)
	var waiting bool
	if msg == "second" {
		waiting = true
	} else if msg == "first" {
		waiting = false
	}
	info <- serverMessage{conn, waiting, ""}
}

func quickplayLobby(info chan<- serverMessage) {
	conn, err := net.Dial(connType, host+":"+port)
	if err != nil {
		info <- serverMessage{nil, false, ""}
		return
	}
	fmt.Fprintf(conn, "quick\n")
	var msg string
	fmt.Fscan(conn, &msg)
	var waiting bool
	if msg == "second" {
		waiting = true
	} else if msg == "first" {
		waiting = false
	}
	info <- serverMessage{conn, waiting, ""}
}
