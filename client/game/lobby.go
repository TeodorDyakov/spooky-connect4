package connect4FMI

import (
	"fmt"
	"net"
)

const (
	CONN_TYPE = "tcp"
	CONN_PORT = "12345"
	CONN_HOST = "localhost"
)

type gameInfo struct {
	conn    net.Conn
	waiting bool
	token   string
}

func createRoom(info chan gameInfo, tokenChan chan string) {
	var waiting bool
	fmt.Println("Connecting to", CONN_TYPE, "server", CONN_HOST+":"+CONN_PORT)
	conn, err := net.Dial(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
	if err != nil {
		panic(err)
	}
	_, err = fmt.Fprintf(conn, "wait\n")
	if err != nil {
		panic(err)
	}
	var token string
	_, err = fmt.Fscan(conn, &token)
	if err != nil {
		panic(err)
	}
	tokenChan <- token

	var msg string
	_, err = fmt.Fscan(conn, &msg)
	if err != nil {
		panic(err)
	}
	if msg == "second" {
		waiting = true
	} else if msg == "first" {
		waiting = false
	}
	info <- gameInfo{conn, waiting, ""}
}

func connectToRoom(token string, info chan gameInfo) {
	fmt.Println("Connecting to", CONN_TYPE, "server", CONN_HOST+":"+CONN_PORT)
	conn, err := net.Dial(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
	var waiting bool
	_, err = fmt.Fprintf(conn, "connect\n")
	if err != nil {
		panic(err)
	}
	_, err = fmt.Fprintf(conn, "%s\n", token)
	if err != nil {
		panic(err)
	}

	var msg string
	_, err = fmt.Fscan(conn, &msg)
	if err != nil {
		panic(err)
	}
	if msg == "second" {
		waiting = true
	} else if msg == "first" {
		waiting = false
	}
	info <- gameInfo{conn, waiting, ""}
}

func quickplayLobby(info chan gameInfo) {
	var waiting bool
	fmt.Println("Connecting to", CONN_TYPE, "server", CONN_HOST+":"+CONN_PORT)
	conn, err := net.Dial(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
	if err != nil {
		panic(err)
	}
	_, err = fmt.Fprintf(conn, "quick\n")
	if err != nil {
		panic(err)
	}
	var msg string
	_, err = fmt.Fscan(conn, &msg)
	if err != nil {
		panic(err)
	}
	if msg == "second" {
		waiting = true
	} else if msg == "first" {
		waiting = false
	}
	info <- gameInfo{conn, waiting, ""}
}
