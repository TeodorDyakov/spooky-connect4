package main

import (
	"sync"
	"fmt"
	"math/rand"
	"net"
	"time"
)

var tokenToConn map[string]net.Conn = make(map[string]net.Conn)
var tokenGenMutex sync.Mutex

const(
	CONN_TYPE = "tcp"
	CONN_PORT = "12345"
)
func init() {
	rand.Seed(time.Now().UnixNano())
}

func generateToken(conn net.Conn) string {
	tokenGenMutex.Lock()
	var tok string
	for {
		token := ""
		for i := 0; i < 5; i++ {
			token += string(rune(rand.Intn(26) + 'A'))
		}
		if _, isDup := tokenToConn[token]; !isDup {
			tok = token
			break
		}
	}
	tokenToConn[tok] = conn
	tokenGenMutex.Unlock()
	return tok
}

var toClose chan net.Conn = make(chan net.Conn, 128)

func main() {
	connectors := make(chan net.Conn, 128)
	waiters := make(chan net.Conn, 128)
	quick := make(chan net.Conn, 128)

	// Start the server and listen for incoming connections.
	listener, err := net.Listen("tcp", ":"+ CONN_PORT)
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
			fmt.Println("Client " + conn.RemoteAddr().String() + " connected.")

			var playerType string
			fmt.Fscan(conn, &playerType)

			if playerType == "connect" {
				connectors <- conn
			} else if playerType == "wait" {
				waiters <- conn
			} else if playerType == "quick" {
				quick <- conn
			}
		}
	}()

	var mutex sync.Mutex

	for {
		select {
		case conn := <-connectors:
			go func(){
				opponentToken := ""
				_, err := fmt.Fscan(conn, &opponentToken)
				if err != nil{
					toClose <- conn
				}
				var connectTo net.Conn
				ok := false
				// check if conn is in map, synchronized so we dont get two player to connect to one
				mutex.Lock()
				if connectTo, ok = tokenToConn[opponentToken]; ok {
					delete(tokenToConn, opponentToken)
				}
				mutex.Unlock()

				if ok {
					startGame(conn, connectTo)
				} else {
					fmt.Fprintf(conn, "error\n")
					toClose <- conn
				}
			}()
		case conn := <-waiters:
			go func(){
				token := generateToken
				_, err := fmt.Fprintf(conn, "%s\n", token)
				if err != nil{
					toClose <- conn
				}
			}()
		case conn := <-quick:
			go startGame(conn, <-quick)
		case conn := <-toClose:
			go func(){
				conn.Close()
				fmt.Println("Client " + conn.RemoteAddr().String() + " disconnected.")
			}()
		}
	}

}

func readMsgAndSend(from, to net.Conn) bool {
	var msg string
	_, err1 := fmt.Fscan(from, &msg)
	_, err2 := fmt.Fprintf(to, "%s\n", msg)
	if err1 != nil || err2 != nil || msg == "end" {
		toClose <- to
		toClose <- from
		return false
	}
	return true
}

func startGame(conn1, conn2 net.Conn) {
	fmt.Fprintf(conn2, "second\n")
	fmt.Fprintf(conn1, "first\n")
	for {
		if !readMsgAndSend(conn1, conn2) {
			return 
		}
		if !readMsgAndSend(conn2, conn1) {
			return
		}
	}
}
