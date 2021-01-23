package gameLogic

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

//read the host and port form cmd args
func init() {
	hostPtr := flag.String("host", "localhost", "ip of host")
	portPtr := flag.String("port", "12345", "port on which to run server")
	flag.Parse()
	host = *hostPtr
	port = *portPtr
}

//ServerMessage is used to pass info for the server connection, whether the client is first or second turn in the game
// and the token for the room(nil for quickplay)
type ServerMessage struct {
	Conn     net.Conn
	IsSecond bool
	Token    string
	Status   string
}

//createRoom connects to the server, sends the type ("wait") of client then sends the token to the channel so
// it can be visualized in the client
//after that it waits for the server to return a message
func CreateRoom(info chan<- ServerMessage) {
	status := ""

	conn, err := net.Dial(connType, host+":"+port)
	if err != nil {
		info <- ServerMessage{nil, false, "", "error when connecting"}
		return
	}
	fmt.Fprintf(conn, "wait\n")
	var token string
	fmt.Fscan(conn, &token)
	info <- ServerMessage{conn, false, token, ""}

	var msg string
	fmt.Fscan(conn, &msg)
	var isSecond bool
	if msg == "second" {
		isSecond = true
	} else if msg == "first" {
		isSecond = false
	}

	info <- ServerMessage{conn, isSecond, "", status}
}

//connectToRoom connects to the server sends a message to the server to acknowledge
// the game ("connect" meaning you connect to room with token)
// sends the token to the server than recives a message from the server
//ServerMessage channel channel is used to pass the server message to other goroutines
func ConnectToRoom(token string, info chan<- ServerMessage) {
	status := ""

	conn, err := net.Dial(connType, host+":"+port)
	if err != nil {
		info <- ServerMessage{nil, false, "", "error when connecting"}
		return
	}
	fmt.Fprintf(conn, "connect\n")
	fmt.Fprintf(conn, "%s\n", token)
	var msg string
	fmt.Fscan(conn, &msg)
	var isSecond bool
	if msg == "second" {
		isSecond = true
	} else if msg == "first" {
		isSecond = false
	} else if msg == "wrong_token" {
		status = msg
	}
	info <- ServerMessage{conn, isSecond, "", status}
}

//quickplayLobby connects to the server sends a message to the server to acknowledge
// the game type (quick) and waits for a message from the server
//ServerMessage channel channel is used to pass the server message to other goroutines
func QuickplayLobby(info chan<- ServerMessage) {
	status := ""

	conn, err := net.Dial(connType, host+":"+port)
	if err != nil {
		info <- ServerMessage{nil, false, "", "error when connecting"}
		return
	}
	fmt.Fprintf(conn, "quick\n")
	var msg string
	fmt.Fscan(conn, &msg)
	var isSecond bool
	if msg == "second" {
		isSecond = true
	} else if msg == "first" {
		isSecond = false
	}
	info <- ServerMessage{conn, isSecond, "", status}
}
