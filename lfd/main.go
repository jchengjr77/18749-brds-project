// 18749 Building Reliable Distrbuted Systems - Project Client

package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
)

func sendBeatToServer(conn net.Conn, myName string) {
	_, err := conn.Write([]byte(myName))
	if err != nil {
		// handle write error
		fmt.Println("Error sending ID: ", err.Error())
	}
	fmt.Printf("[%s] Sent heartbeat to server\n", time.Now().Format(time.RFC850))
}

func sendHeartbeats(conn net.Conn, heartbeat_freq int, myID int) {
	defer conn.Close()
	for {
		sendBeatToServer(conn, "LFD"+strconv.Itoa(myID)+" heartbeat")
		time.Sleep(time.Duration(heartbeat_freq) * time.Second)
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
	fmt.Println("---------- Local Fault Detector started ----------")
	var err error
	heartbeat_freq := 1

	args := os.Args[1:]
	if len(args) > 0 {
		heartbeat_freq, err = strconv.Atoi(args[0])
		if err != nil {
			fmt.Println("Error parsing args: ", err.Error())
			return
		}
		fmt.Printf("Set heartbeat frequency to %d seconds\n", heartbeat_freq)
	}

	// connect to server
	conn, err := net.Dial("tcp", ":8080")
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
	fmt.Println("Received LFD ID: ", myID)

	go sendHeartbeats(conn, heartbeat_freq, myID)
	go listenToServer(conn)
	for {
	}
}
