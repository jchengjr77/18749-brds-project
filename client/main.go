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

const (
	RESET  = "\033[0m"
	RED    = "\033[31m"
	GREEN  = "\033[32m"
	YELLOW = "\033[33m"
	BLUE   = "\033[34m"
	PURPLE = "\033[35m"
	CYAN   = "\033[36m"
	WHITE  = "\033[37m"
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
	fmt.Printf(BLUE+"[%s] %s <%d, %d, %s, %s>\n"+RESET, time.Now().Format(time.RFC850), action, clientID, serverID, msg, msgType)
}

/*
 * sendMessageToServer sends a single message to the server
 */
func sendMessageToServer(conn net.Conn, msg string, clientID int, serverID int, connToNameMap *map[net.Conn]string, connMap *map[net.Conn]int, servMap *map[net.Conn]int, idToConnMap *map[int]net.Conn, serverName string) {
	_, err := conn.Write([]byte(msg))
	if err != nil {
		fmt.Println("Error sending message to server: ", err.Error())
		fmt.Println("Redialing", serverName)
		var newConn net.Conn
		for {
			newConn, err = net.Dial("tcp", serverName+":8080")
			if err == nil {
				break
			}
		}
		fmt.Println("Redialed ", serverName)
		delete((*connToNameMap), conn)
		delete((*connMap), conn)
		delete((*servMap), conn)
		(*connToNameMap)[newConn] = serverName
		(*connMap)[newConn] = clientID
		(*servMap)[newConn] = serverID
		(*idToConnMap)[serverID] = newConn
	}
	printMsg(clientID, serverID, msg, "request")
}

/*
 * manually send a message (your clientID) to the server
 */
func manuallySendIDRoutine(connMap *map[net.Conn]int, servMap *map[net.Conn]int, primaryConn *net.Conn, passive bool, connToNameMap *map[net.Conn]string, idToConnMap *map[int]net.Conn) {
	reqNum := 0
	go automaticallySendIDRoutine(connMap, servMap, primaryConn, passive, &reqNum, connToNameMap, idToConnMap)
	for {
		fmt.Println("Press 'Enter' to send message to server...")
		fmt.Scanln()
		for conn, serverName := range (*connToNameMap) {
			clientId := (*connMap)[conn]
			fmt.Println("CURR CONN: " + strconv.Itoa((*servMap)[conn]))
			fmt.Println("PRIM CONN: " + strconv.Itoa((*servMap)[*primaryConn]))
			if passive && (*servMap)[conn] != (*servMap)[*primaryConn] {
				continue
			}
			s := "requestnum:" + strconv.Itoa(reqNum) + ",clientid:" + strconv.Itoa(clientId)
			go sendMessageToServer(conn, s, clientId, (*servMap)[conn], connToNameMap, connMap, servMap, idToConnMap, serverName)
		}
		time.Sleep(2 * time.Second) //can send max once every 2 seconds
		reqNum++
	}
}

/*
 * automatically send a message (your clientID) to the server
 */
func automaticallySendIDRoutine(connMap *map[net.Conn]int, servMap *map[net.Conn]int, primaryConn *net.Conn, passive bool, reqNum *int, connToNameMap *map[net.Conn]string, idToConnMap *map[int]net.Conn) {
	for {
		fmt.Println("sending messages out")
		for conn, serverName := range (*connToNameMap) {
			clientId := (*connMap)[conn]
			if passive && conn != *primaryConn {
				continue
			}
			s := "requestnum:" + strconv.Itoa(*reqNum) + ",clientid:" + strconv.Itoa(clientId)
			go sendMessageToServer(conn, s, clientId, (*servMap)[conn], connToNameMap, connMap, servMap, idToConnMap, serverName)
		}
		time.Sleep(5 * time.Second) //can send max once every 5 seconds
		*reqNum++
	}
}

/*
 * listenToServerRoutine is a routine that listens to messages from server
 */
func listenToServerRoutine(conn net.Conn, myID int, serverID int, serverName string, msgChan chan string, repChan chan bool, connMap *map[net.Conn]int, servMap *map[net.Conn]int, connToNameMap *map[net.Conn]string, idToConnMap *map[int]net.Conn) {
	for {
		buf := make([]byte, 1024)
		mlen, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Error reading: ", err.Error())
			fmt.Println("Redialing", serverName)
			var newConn net.Conn
			for {
				newConn, err = net.Dial("tcp", serverName+":8080")
				if err == nil {
					break
				}
			}
			fmt.Println("Redialed ", serverName)
			delete((*connToNameMap), conn)
			delete((*connMap), conn)
			delete((*servMap), conn)
			(*connToNameMap)[newConn] = serverName
			(*connMap)[newConn] = myID
			(*servMap)[newConn] = serverID
			(*idToConnMap)[serverID] = newConn
		}
		msgChan <- string(buf[:mlen])
		dup := <-repChan
		if !dup {
			printMsg(myID, serverID, string(buf[:mlen]), "reply")
		}
	}
}

func processMsgs(msgChan chan string, repChan chan bool, reassignPrimChan chan int) {
	resSet := make(map[int]struct{})
	for {
		msg := <-msgChan //block until message is received
		electedInd := strings.Index(msg, "ELECTED:")

		reqInd := strings.Index(msg, "requestnum:")
		endInd := strings.Index(msg, ",")
		// a new elected server id exists
		fmt.Println(msg)
		if electedInd > -1 {
			newPrimaryId, err := strconv.Atoi((string)(strings.Split(msg, ":")[1][0]))
			if err != nil {
				fmt.Println(err)
			}
			reassignPrimChan <- newPrimaryId
		} 
		if reqInd > -1 {
			req := msg[reqInd:endInd]
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
}

func reassignPrimary(reassignPrimChan chan int, primaryConn *net.Conn, idToConnMap *map[int]net.Conn) {
	for {
		select {
		case newPrimary := <-reassignPrimChan:
			*primaryConn = (*idToConnMap)[newPrimary]
			fmt.Printf(YELLOW+"[%s] re-election to %d\n"+RESET, time.Now().Format(time.RFC850), newPrimary)

		}
	}
}

func main() {
	fmt.Println("---------- Client started ----------")

	/* args[0] is the Host Name 1
	 * args[1] is the Host Name 2
	 * args[2] is the Host Name 3
	 */

	args := os.Args[1:]

	host1 := args[1:][0]
	host2 := args[1:][1]
	host3 := args[1:][2]

	connMap := map[net.Conn]int{}
	
	connToNameMap := map[net.Conn]string{}
	idToNameMap := map[int]string{}

	idToNameMap[1] = host1
	idToNameMap[2] = host2
	idToNameMap[3] = host3

	servMap := map[net.Conn]int{}
	idToConnMap := map[int]net.Conn{}
	msgChan := make(chan string)
	repChan := make(chan bool)
	reassignPrimChan := make(chan int)
	var primaryConn net.Conn
	mode := string(args[0])
	for i, server := range args[1:] {
		// connect to primary replica server
		conn, err := net.Dial("tcp", server+":8080")
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

		connToNameMap[conn] = idToNameMap[i+1]

		fmt.Println(string(buf[:mlen]))
		myID, err := strconv.Atoi(string(buf[:mlen]))
		sendMessageToServer(conn, "ACK", myID, i+1, &connToNameMap, &connMap, &servMap, &idToConnMap, server)
		if err != nil {
			fmt.Println("Error converting ID data:", err.Error())
			return
		}
		fmt.Println("Received Client ID: ", myID)
		connMap[conn] = myID
		servMap[conn] = i + 1
		idToConnMap[i+1] = conn

		go listenToServerRoutine(conn, myID, i+1, server, msgChan, repChan, &connMap, &servMap, &connToNameMap, &idToConnMap)
		// if i == 0 {
		// 	primaryConn = conn
		// }
		go reassignPrimary(reassignPrimChan, &primaryConn, &idToConnMap)
	}

	go processMsgs(msgChan, repChan, reassignPrimChan)
	isPassive := mode == "passive"
	manuallySendIDRoutine(&connMap, &servMap, &primaryConn, isPassive, &connToNameMap, &idToConnMap)

}
