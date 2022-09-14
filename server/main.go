// 18749 Building Reliable Distributed Systems - Project Server

package main

import (
	"fmt"
	"net"
	"strconv"
	"time"
)

const (
	SERVER_TYPE = "tcp"
	HOST        = "localhost"
	PORT        = "8080"
)

func handleClient(conn net.Conn, nextClientID int) {
	fmt.Println("New Client Connected! ID: ", nextClientID)
	_, err := conn.Write([]byte(strconv.Itoa(nextClientID)))
	if err != nil {
		// handle write error
		fmt.Println("Error sending ID: ", err.Error())
	}
	for {
		buf := make([]byte, 1024)
		mlen, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Error reading: ", err.Error())
			return
		}
		fmt.Println(string(buf[:mlen]))

	}
}

func listenerChannelWrapper(listener net.Listener, newClientChan chan net.Conn) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			// handle connection error
			fmt.Println("Error accepting: ", err.Error())
			return
		}
		newClientChan <- conn
	}
}

func main() {
	fmt.Println("---------- Server started ----------")

	// client IDs, monotonically increasing
	nextClientID := 1

	// map of clients [client id] -> [net connection]
	clients := map[int]net.Conn{}

	// make channel for new connections
	newClientChan := make(chan net.Conn)

	// start the server
	listener, err := net.Listen(SERVER_TYPE, HOST+":"+PORT)
	if err != nil {
		// handle server initialization error
		fmt.Println("Error initializing: ", err.Error())
		return
	}

	go listenerChannelWrapper(listener, newClientChan)

	defer listener.Close()

	/**

	1. spawn a listener thread
	2. create a channel for new connections [between main thread and listener thread]
	3. new select/case block in the main thread waits for new connections OR client death notifications

	**/

	for {
		select {
		case conn := <-newClientChan:
			go handleClient(conn, nextClientID)
			clients[nextClientID] = conn
			nextClientID++
		default:
			time.Sleep(time.Second)
		}

	}
}
