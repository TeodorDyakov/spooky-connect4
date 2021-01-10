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

//read the host and port form cmd args
func init() {
	hostPtr := flag.String("host", "localhost", "ip of host")
	portPtr := flag.String("port", "12345", "port on which to run server")
	flag.Parse()
	host = *hostPtr
	port = *portPtr
}

//serverMessage is used to pass info for the server connection, whether the client is first or second turn in the game
// and the token for the room(nil for quickplay)
type serverMessage struct {
	conn     net.Conn
	isSecond bool
	token    string
}

//createRoom connects to the server, sends the type ("wait") of client then sends the token to the channel so
// it can be visualized in the client
//after that it waits for the server to return a message
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
	var isSecond bool
	if msg == "second" {
		isSecond = true
	} else if msg == "first" {
		isSecond = false
	}
	info <- serverMessage{conn, isSecond, ""}
}

//connectToRoom connects to the server sends a message to the server to acknowledge
// the game ("connect" meaning you connect to room with token)
// sends the token to the server than recives a message from the server
//serverMessage channel channel is used to pass the server message to other goroutines
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
	var isSecond bool
	if msg == "second" {
		isSecond = true
	} else if msg == "first" {
		isSecond = false
	}
	info <- serverMessage{conn, isSecond, ""}
}

//quickplayLobby connects to the server sends a message to the server to acknowledge
// the game type (quick) and waits for a message from the server
//serverMessage channel channel is used to pass the server message to other goroutines
func quickplayLobby(info chan<- serverMessage) {
	conn, err := net.Dial(connType, host+":"+port)
	if err != nil {
		info <- serverMessage{nil, false, ""}
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
	info <- serverMessage{conn, isSecond, ""}
}
