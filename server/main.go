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

func handleClient(conn net.Conn, clientID int, stateChan chan int) {
	fmt.Println("New Client Connected! ID: ", clientID)
	_, err := conn.Write([]byte(strconv.Itoa(clientID)))
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
		fmt.Printf("[%s] Recieved '%s' from Client%d\n", time.Now().Format(time.RFC850), string(buf[:mlen]), clientID)
		stateChan <- 1
		_, err = conn.Write([]byte("Message ack"))
		fmt.Printf("[%s] Replied to Client%d with ack\n", time.Now().Format(time.RFC850), clientID)

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

func connectToLFD(listener net.Listener) (conn net.Conn) {
	// First connect to LFD
	conn, err := listener.Accept()
	if err != nil {
		// handle connection error
		fmt.Println("Error accepting: ", err.Error())
		return
	}

	fmt.Println("LFD1 Connected!")
	_, err = conn.Write([]byte(strconv.Itoa(1)))
	if err != nil {
		// handle write error
		fmt.Println("Error sending ID: ", err.Error())
	}
	return conn
}

func listenLFD(conn net.Conn) {
	for {
		// Recieve heartbeat from LFD
		buf := make([]byte, 1024)
		mlen, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Error reading: ", err.Error())
			return
		}
		fmt.Printf("[%s] Recieved %s\n", time.Now().Format(time.RFC850), string(buf[:mlen]))
		// Send reply to LFD
		_, err = conn.Write([]byte("Heartbeat Ack"))
		fmt.Printf("[%s] Replied to LFD with ack\n", time.Now().Format(time.RFC850))
	}
}

func incState(my_state *int) {
	*my_state++
	fmt.Println("New state: ", *my_state)
}
func main() {
	my_state := 0
	fmt.Println("---------- Server started ----------")

	// client IDs, monotonically increasing
	clientID := 1

	// map of clients [client id] -> [net connection]
	clients := map[int]net.Conn{}

	// make channel for new connections
	newClientChan := make(chan net.Conn)

	// make channel for new incrementing state
	stateChan := make(chan int)

	// start the server
	listener, err := net.Listen(SERVER_TYPE, HOST+":"+PORT)
	if err != nil {
		// handle server initialization error
		fmt.Println("Error initializing: ", err.Error())
		return
	}

	lfdConn := connectToLFD(listener)
	go listenLFD(lfdConn)

	go listenerChannelWrapper(listener, newClientChan)

	defer listener.Close()

	/**

	1. spawn a listener thread
	2. create a channel for new connections [between main thread and listener thread]
	3. new select/case block in the main thread waits for new connections OR client death notifications

	**/

	for {
		select {
		case <-stateChan:
			incState(&my_state)
		case conn := <-newClientChan:
			go handleClient(conn, clientID, stateChan)
			clients[clientID] = conn
			clientID++
		default:
			//time.Sleep(time.Second)
		}
	}
}
