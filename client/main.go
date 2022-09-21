// 18749 Building Reliable Distrbuted Systems - Project Client

package main

import (
	"fmt"
	"net"
	"strconv"
	"time"
)

/*
 * formats and prints important logging messages
 */
func printMsg(clientID int, serverID int, msg string, msgType string) {
	var action string
	if (msgType == "request") {
		action = "Sent"
	} else {
		action = "Received"
	}
	fmt.Printf("[%s] %s <%d, %d, %s, %s>\n", time.Now().Format(time.RFC850), action, clientID, serverID, msg, msgType)
}

/*
 * sendMessageToServer sends a single message to the server
 */
func sendMessageToServer(conn net.Conn, msg string, clientID int) {
	_, err := conn.Write([]byte(msg))
	if err != nil {
		// handle write error
		fmt.Println("Error sending ID: ", err.Error())
	}

	printMsg(clientID, 1, msg, "request")
}

/*
 * sendIDRoutine only sends the client ID to the server
 */
func sendIDRoutine(conn net.Conn, myID int) {
	for {
		time.Sleep(10 * time.Second)
		sendMessageToServer(conn, strconv.Itoa(myID), myID)
	}
}

/*
 * manually send a message (your clientID) to the server
 */
func manuallySendIDRoutine(conn net.Conn, myID int) {
	for {
		fmt.Println("Press 'Enter' to send message to server...")
		fmt.Scanln()
		sendMessageToServer(conn, strconv.Itoa(myID), myID)
	}
}

/*
 * listenToServerRoutine is a routine that listens to messages from server
 */
func listenToServerRoutine(conn net.Conn, myID int) {
	for {
		buf := make([]byte, 1024)
		mlen, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Error reading: ", err.Error())
			return
		}

		printMsg(myID, 1, string(buf[:mlen]), "reply")
	}
}

func main() {
	fmt.Println("---------- Client started ----------")

	// connect to server
	// conn, err := net.Dial("tcp", "Nathans-Macbook-Pro-7.local:8080")
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		// handle connection error
		fmt.Println("Error dialing: ", err.Error())
		return
	}

	var MSGS = []string{
		" is still alive and kicking!",
		" is having a blast!",
		" wants to head out soon",
		" says the server is kind of mean to it",
		" thinks you're cute",
		" wants to change its ID",
		" didn't respond to my texts",
		" secretly watches youtube during lecture",
		" does a backflip",
		" thinks you programmers should write more client messages",
		" smells a segfault coming",
		" is ordering a caramel macchiato",
		" faints for a sec, and comes back up",
		" has to call its parents",
	}
	// silence declared-but-unused error
	_ = MSGS

	buf := make([]byte, 1024)
	mlen, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading: ", err.Error())
		return
	}
	myID, err := strconv.Atoi(string(buf[:mlen]))
	if err != nil {
		fmt.Println("Error converting ID data:", err.Error())
		return
	}
	fmt.Println("Received Client ID: ", myID)

	go sendIDRoutine(conn, myID)
	go manuallySendIDRoutine(conn, myID)
	go listenToServerRoutine(conn, myID)
	for {
	}
}
