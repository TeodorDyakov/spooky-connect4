package connect4FMI

import (
	"flag"
	"fmt"
	"net"
)

var (
	host     string
	port     string
	connType = "tcp"
)

func init() {
	hostPtr := flag.String("host", "localhost", "ip of host")
	portPtr := flag.String("port", "12345", "port on which to run server")
	flag.Parse()
	host = *hostPtr
	port = *portPtr
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
