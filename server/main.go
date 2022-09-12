// 18749 Building Reliable Distributed Systems - Project Server

package main

import (
	"fmt"
	"net"
	"strconv"
)

const (
	SERVER_TYPE = "tcp"
	HOST = "localhost"
	PORT = "8080"
)

func handleClient(conn net.Conn, nextClientID int) {
	fmt.Println("New Client Connected! ID: ", nextClientID)
	_, err := conn.Write([]byte(strconv.Itoa(nextClientID)))
	if err != nil {
		// handle write error
		fmt.Println("Error sending ID: ", err.Error())
	}
}

func main() {
	fmt.Println("---------- Server started ----------")

	// client IDs, monotonically increasing
	nextClientID := 1

	// map of clients [client id] -> [net connection]
	clients := map[int]net.Conn {}

	// start the server
	listener, err := net.Listen(SERVER_TYPE, HOST+":"+PORT)
	if err != nil {
		// handle server initialization error
		fmt.Println("Error initializing: ", err.Error())
		return
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			// handle connection error
			fmt.Println("Error accepting: ", err.Error())
			return
		}
		go handleClient(conn, nextClientID)
		clients[nextClientID] = conn
		defer conn.Close()
		nextClientID ++
	}
}