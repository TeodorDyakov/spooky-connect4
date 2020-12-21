package main

import(
	"fmt"
	"net"
)
/*
Start the command line in which you connect to another player
return a boolean value indicating whenther you are first or 
you should wait your tun when game starts and a connection to the server
*/
func lobby() (bool, net.Conn) {
	var waiting bool

	fmt.Println("Connecting to", CONN_TYPE, "server", CONN_HOST+":"+CONN_PORT)
	conn, err := net.Dial(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
	if err != nil {
		panic(err)
	}

	fmt.Println("enter 1 to make a room or 2 to connect to room or 3 for quikckplay")
	var input int
	fmt.Scan(&input)

	if input == 1 {
		_, err := fmt.Fprintf(conn, "wait\n")
		if err != nil{
			panic(err)
		}
		var token string
		_, err = fmt.Fscan(conn, &token)
		if err != nil{
			panic(err)
		}
		fmt.Printf("You token is:%s\n", token)
		fmt.Println("waiting for a friend to connect...")
	} else if input == 2 {
		_, err := fmt.Fprintf(conn, "connect\n")
		if err != nil{
			panic(err)
		}
		var token string
		fmt.Printf("Enter friend token\n")
		fmt.Scan(&token)
		_, err = fmt.Fprintf(conn, "%s\n", token)
		if err != nil{
			panic(err)
		}
	} else {
		fmt.Println("Searhing for opponent...")
		_, err := fmt.Fprintf(conn, "quick\n")
		if err != nil{
			panic(err)
		}
	}

	var msg string
	_, err = fmt.Fscan(conn, &msg)
	if err != nil{
			panic(err)
	}
	if msg == "second" {
		waiting = true
	} else if msg == "first" {
		waiting = false
	} else {
		return waiting, nil
	}
	return waiting, conn
}