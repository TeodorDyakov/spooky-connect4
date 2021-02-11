package main

import (
	"net"
	"fmt"
	"testing"
	"time"
)

var _ = func() bool {
	testing.Init()
	return true
}()

func TestServer(t *testing.T) {
   	go start()

   	time.Sleep(1*time.Second)
   	
   	conn, err := net.Dial("tcp", ":12345")
    if err != nil {
        t.Fatal(err)
    }

    defer conn.Close()

    //first the one who creates room connects 
    fmt.Fprintf(conn, "wait\n")

    var token string
    fmt.Fscan(conn, &token)

    //we expect the server to return us a 5 characters random token
    if len(token) != 5 {
		t.Errorf("server did not return 5 character token but returned %s\n", token)
	}

	connectToRoom, err := net.Dial("tcp", ":12345")
    if err != nil {
        t.Fatal(err)
    }	

	fmt.Fprintf(connectToRoom, "connect\n")
    fmt.Fprintf(connectToRoom, "%s\n", token)

    var response string

    fmt.Fscan(connectToRoom, &response)

    //we expect the server to return "first" for the one who connects
    if response != "first" {
		t.Errorf("expected response to be \"first\" but it was %s\n", response)
	}

	//expect the other player to be second
	fmt.Fscan(conn, &response)

	if response != "second" {
		t.Errorf("expected response to be \"second\" but it was %s\n", response)
	}

	//the first player places
	fmt.Fprintf(connectToRoom, "2\n")

	//expect the server to return where the other player has placed

	fmt.Fscan(conn, &response)

	if response != "2" {
		t.Errorf("expected response to be 2 but it was %s\n", response)
	}

}