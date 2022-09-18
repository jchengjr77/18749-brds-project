// 18749 Building Reliable Distrbuted Systems - Project Client

package main

import (
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"time"
)

func sendMessageToServer(conn net.Conn, msg string) {
	_, err := conn.Write([]byte(msg))
	if err != nil {
		// handle write error
		fmt.Println("Error sending ID: ", err.Error())
	}
	fmt.Printf("[%s] Sent '%s' to server\n", time.Now().Format(time.RFC850), msg)
}

func sendMessages(conn net.Conn, MSGS []string, myID int) {
	for {
		randomIndex := rand.Intn(len(MSGS))
		msg := MSGS[randomIndex]
		sendMessageToServer(conn, msg)
		time.Sleep(10 * time.Second)
	}
}

func listenToServer(conn net.Conn) {
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

	go sendMessages(conn, MSGS, myID)
	go listenToServer(conn)
	for {
	}
}
