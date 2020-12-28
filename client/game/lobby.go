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

func createRoom(info chan gameInfo) {
	conn, err := net.Dial(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
	if err != nil {
		info <- gameInfo{nil, false, ""}
		return
	}
	fmt.Fprintf(conn, "wait\n")
	var token string
	fmt.Fscan(conn, &token)
	info <- gameInfo{conn, false, token}

	var msg string
	fmt.Fscan(conn, &msg)
	var waiting bool
	if msg == "second" {
		waiting = true
	} else if msg == "first" {
		waiting = false
	}
	info <- gameInfo{conn, waiting, ""}
}

func connectToRoom(token string, info chan gameInfo) {
	conn, err := net.Dial(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
	if err != nil {
		info <- gameInfo{nil, false, ""}
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
	info <- gameInfo{conn, waiting, ""}
}

func quickplayLobby(info chan gameInfo) {
	conn, err := net.Dial(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
	if err != nil {
		info <- gameInfo{nil, false, ""}
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
	info <- gameInfo{conn, waiting, ""}
}
