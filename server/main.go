// 18749 Building Reliable Distributed Systems - Project Server

package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
	"strings"
)

const (
	SERVER_TYPE = "tcp"
	PORT = "8080"
)

func printMsg(clientID int, serverID int, msg string, msgType string) {
	var action string
	if (msgType == "request") {
		action = "Received"
	} else {
		action = "Sent"
	}
	fmt.Printf("[%s] %s <%d, %d, %s, %s>\n", time.Now().Format(time.RFC850), action, clientID, serverID, msg, msgType)
}

func handleClient(conn net.Conn, clientID int, serverID int, stateChan chan int) {
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
		msg := string(buf[:mlen])
		printMsg(clientID, serverID, msg, "request")
		clientIDInd := strings.Index(msg, "clientid:")
		clientID, err := strconv.Atoi(msg[clientIDInd+len("clientid:"):])
		if err != nil {
			fmt.Println("Error converting using Atoi: ", err.Error())
			continue
		}
		stateChan <- clientID
		sepInd := strings.Index(msg, ",")
		ackMsg := msg[:sepInd] + ",serverid:" + strconv.Itoa(serverID)
		_, err = conn.Write([]byte(ackMsg))
		if err != nil {
			fmt.Println("Error sending ack: ", err.Error())
		}
		printMsg(clientID, serverID, ackMsg, "reply")
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

func connectToLFD(serverId int) (conn net.Conn) {
	// First connect to LFD
	conn, err := net.Dial("tcp", ":8081")
	if err != nil {
		// handle connection error
		fmt.Println("Error accepting: ", err.Error())
		return nil
	}

	fmt.Println("LFD1 Connected!")
	_, err = conn.Write([]byte(strconv.Itoa(serverId)))
	if err != nil {
		// handle write error
		fmt.Println("Error sending ID: ", err.Error())
	}
	return conn
}

func listenLFD(conn net.Conn) {
	for {
		// Receive heartbeat from LFD
		buf := make([]byte, 1024)
		mlen, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Error reading: ", err.Error())
			return
		}
		fmt.Printf("[%s] Received %s\n", time.Now().Format(time.RFC850), string(buf[:mlen]))
		// Send reply to LFD
		_, err = conn.Write([]byte("Heartbeat Ack"))
		fmt.Printf("[%s] Replied to LFD with ack\n", time.Now().Format(time.RFC850))
	}
}

func setState(my_state *int, val int) {
	fmt.Printf("[%s] my_state = %d -> %d \n", time.Now().Format(time.RFC850), *my_state, val)
	*my_state = val
}


func main() {
	args := os.Args[1:]
	if len(args) < 1 {
		fmt.Printf("Usage: go run server/main.go <server id>")
		return
	}
	serverId, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Println("Error parsing args: ", err.Error())
		return
	}

	my_state := 0
	fmt.Printf("---------- Server %d started ----------\n", serverId)

	// client IDs, monotonically increasing
	clientID := 1

	// map of clients [client id] -> [net connection]
	clients := map[int]net.Conn{}

	// make channel for new connections
	newClientChan := make(chan net.Conn)

	// make channel for new incrementing state
	stateChan := make(chan int)

	// start the server
	listener, err := net.Listen(SERVER_TYPE, ":"+PORT)
	if err != nil {
		// handle server initialization error
		fmt.Println("Error initializing: ", err.Error())
		return
	}

	lfdConn := connectToLFD(serverId)
	if lfdConn == nil {
		fmt.Println("Could not connect to LFD")
		return
	}
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
		case recentID := <-stateChan:
			setState(&my_state, recentID)
		case conn := <-newClientChan:
			go handleClient(conn, clientID, serverId, stateChan)
			clients[clientID] = conn
			clientID++
		}
	}
}
