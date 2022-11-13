// 18749 Building Reliable Distrbuted Systems - Project Client

package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

/*
 * formats and prints important logging messages
 */
func printMsg(clientID int, serverID int, msg string, msgType string) {
	var action string
	if msgType == "request" {
		action = "Sent"
	} else {
		action = "Received"
	}
	fmt.Printf("[%s] %s <%d, %d, %s, %s>\n", time.Now().Format(time.RFC850), action, clientID, serverID, msg, msgType)
}

/*
 * sendMessageToServer sends a single message to the server
 */
func sendMessageToServer(conn net.Conn, msg string, clientID int, serverID int) {
	_, err := conn.Write([]byte(msg))
	if err != nil {
		fmt.Println("Error sending message to server: ", err.Error())
	}
	printMsg(clientID, serverID, msg, "request")
}

/*
 * manually send a message (your clientID) to the server
 */
func manuallySendIDRoutine(connMap, servMap map[net.Conn]int) {
	reqNum := 0
	for {
		fmt.Println("Press 'Enter' to send message to server...")
		fmt.Scanln()
		for conn, id := range connMap {
			s := "requestnum:" + strconv.Itoa(reqNum) + ",clientid:" + strconv.Itoa(id)
			go sendMessageToServer(conn, s, id, servMap[conn])
		}
		time.Sleep(2 * time.Second) //can send max once every 2 seconds
		reqNum++
	}
}

/*
 * listenToServerRoutine is a routine that listens to messages from server
 */
func listenToServerRoutine(conn net.Conn, myID int, serverID int, msgChan chan string, repChan chan bool) {
	for {
		buf := make([]byte, 1024)
		mlen, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Error reading: ", err.Error())
			return
		}
		msgChan <- string(buf[:mlen])
		dup := <-repChan
		if !dup {
			printMsg(myID, serverID, string(buf[:mlen]), "reply")
		}
	}
}

func processMsgs(msgChan chan string, repChan chan bool) {
	resSet := make(map[int]struct{})
	for {
		msg := <-msgChan //block until message is received
		endInd := strings.Index(msg, ",")
		req := msg[len("requestnum:"):endInd]
		reqNum, _ := strconv.Atoi(req)
		idInd := strings.Index(msg, "serverid:")
		id := msg[idInd+len("serverid:"):]
		_, exists := resSet[reqNum]
		dup := false
		if exists { //if duplicate
			fmt.Println("request_num " + req + ": Discarded duplicate reply from S" + id)
			dup = true
		} else {
			resSet[reqNum] = struct{}{}
		}
		repChan <- dup
	}
}

func main() {
	fmt.Println("---------- Client started ----------")

	/* args[0] is the Host Name 1
	 * args[1] is the Host Name 2
	 * args[2] is the Host Name 3
	 */

	args := os.Args[1:]

	connMap := map[net.Conn]int{}
	servMap := map[net.Conn]int{}
	msgChan := make(chan string)
	repChan := make(chan bool)
	primaryReplica := args[0]
	primaryReplicaId := 1
	// connect to primary replica server
	conn, err := net.Dial("tcp", primaryReplica+":8080")
	if err != nil {
		// handle connection error
		fmt.Println("Error dialing: ", err.Error())
		return
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
	connMap[conn] = myID
	servMap[conn] = primaryReplicaId
	go listenToServerRoutine(conn, myID, primaryReplicaId, msgChan, repChan)
	go processMsgs(msgChan, repChan)
	manuallySendIDRoutine(connMap, servMap)
}
