package main

import (
	"sync"
	"fmt"
	"math/rand"
	"net"
	"time"
)

var tokenToConn map[string]net.Conn = make(map[string]net.Conn)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var tokenGenMutex sync.Mutex

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

	var mutex sync.Mutex

	for {
		select {
		case conn := <-connectors:
			go func(){
				opponentToken := ""
				fmt.Fscan(conn, &opponentToken)
				
				var connectTo net.Conn
				ok := false

				// check if conn is in map, synchronized so we dont get two player to connect to one
				mutex.Lock()
				if connectTo, ok = tokenToConn[opponentToken]; ok {
					delete(tokenToConn, opponentToken)
				}
				mutex.Unlock()

				if ok {
					fmt.Fprintf(connectTo, "go\n")
					fmt.Fprintf(conn, "go\n")
					handleConnection(conn, connectTo)
				} else {
					fmt.Fprintf(conn, "error\n")
				}
			}()
		case conn := <-waiters:
			go func(){
				token := generateToken(conn)
				fmt.Fprintf(conn, "%s\n", token)
			}()
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
