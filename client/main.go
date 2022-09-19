// 18749 Building Reliable Distrbuted Systems - Project Client

package main

import (
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"time"
)

/*
 * sendMessageToServer sends a single message to the server
 */
func sendMessageToServer(conn net.Conn, msg string) {
	_, err := conn.Write([]byte(msg))
	if err != nil {
		// handle write error
		fmt.Println("Error sending ID: ", err.Error())
	}
	fmt.Printf("[%s] Sent '%s' to server\n", time.Now().Format(time.RFC850), msg)
}

/*
 * sendMessagesRoutine is a routine that sends a message to server every 10 seconds
 */
func sendMessagesRoutine(conn net.Conn, MSGS []string, myID int) {
	for {
		randomIndex := rand.Intn(len(MSGS))
		msg := MSGS[randomIndex]
		sendMessageToServer(conn, msg)
		time.Sleep(10 * time.Second)
	}
}

/*
 * sendIDRoutine only sends the client ID to the server
 */
func sendIDRoutine(conn net.Conn, myID int) {
	for {
		sendMessageToServer(conn, strconv.Itoa(myID))
		time.Sleep(10 * time.Second)
	}
}

/*
 * listenToServerRoutine is a routine that listens to messages from server
 */
func listenToServerRoutine(conn net.Conn) {
	for {
		buf := make([]byte, 1024)
		mlen, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Error reading: ", err.Error())
			return
		}
		fmt.Printf("[%s] Recieved %s from server\n", time.Now().Format(time.RFC850), string(buf[:mlen]))
	}
}

func main() {
	fmt.Println("---------- Client started ----------")

	// connect to server
	conn, err := net.Dial("tcp", ":8080")
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

	// go sendMessagesRoutine(conn, MSGS, myID)
	go sendIDRoutine(conn, myID)
	go listenToServerRoutine(conn)
	for {
	}
}
