package main

import (
	"fmt"
	"math/rand"
	"net"
	"time"
)

var tokenToConn map[string]net.Conn = make(map[string]net.Conn)

// Application constants, defining host, port, and protocol.
const (
	CONN_HOST = "localhost"
	CONN_PORT = "12345"
	CONN_TYPE = "tcp"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func generateToken() string {
	token := ""
	for{
		for i := 0; i < 5; i++ {
			token += string(rune(rand.Intn(26) + 'A'))
		}
		if _, isDup := tokenToConn[token]; !isDup {
			return token
		}
	}
	return token
}

func main() {
	connectors := make(chan net.Conn, 128)
	waiters := make(chan net.Conn, 128)

	// Start the server and listen for incoming connections.
	listener, err := net.Listen("tcp", ":12345")
	if err != nil {
		panic(err)
	}

	go func() {
		// run loop forever, until exit.
		for {
			// Listen for an incoming connection.
			conn, err := listener.Accept()
			if err != nil {
				panic(err)
			}
			fmt.Println("Client connected.")
			fmt.Println("Client " + conn.RemoteAddr().String() + " connected.")

			var waitOrConnect string
			fmt.Fscan(conn, &waitOrConnect)

			if waitOrConnect == "connect" {
				connectors <- conn
			} else if waitOrConnect == "wait" {
				waiters <- conn
			}
		}
	}()

	for {
		select {
		case conn := <-connectors:
			opponentToken := ""
			fmt.Fscan(conn, &opponentToken)
			// check if conn is in map
			if connectTo, ok := tokenToConn[opponentToken]; ok {
				fmt.Fprintf(connectTo, "go\n")
				fmt.Fprintf(conn, "go\n")
				delete(tokenToConn, opponentToken)
				go handleConnection(conn, connectTo)
			} else {
				//error hanle
			}
		case conn := <-waiters:
			token := generateToken()
			fmt.Fprintf(conn, "%s\n", token)
			tokenToConn[token] = conn
		}
	}

}

func readMsgAndSend(from, to net.Conn) bool {
	var msg string
	_, err := fmt.Fscan(from, &msg)
	if err != nil {
		fmt.Println("Client " + from.RemoteAddr().String() + " disconnected.")
		return false
	}
	fmt.Fprintf(to, "%s\n", msg)
	return true
}

func handleConnection(conn1, conn2 net.Conn) {
	defer conn1.Close()
	defer conn2.Close()

	for {
		if !readMsgAndSend(conn1, conn2) {
			return
		}
		if !readMsgAndSend(conn2, conn1) {
			return
		}
	}
}
