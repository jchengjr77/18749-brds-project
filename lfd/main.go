// 18749 Building Reliable Distrbuted Systems - Project Client

package main

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const (
	SERVER_TYPE    = "tcp"
	PORT           = "8081"
	SERVER_RUN_CMD = "server/main.go"
	HOST1 = "nathans-mbp-7.wifi.local.cmu.edu"
	HOST2 = "vigneshsmbp230.wifi.local.cmu.edu"
	HOST3 = "jeremys-mbp-200.wifi.local.cmu.edu"
)

/*
 * sendHeartbeatToServer sends a single heartbeat to the server
 */
func sendHeartbeatToServer(conn net.Conn, myName string) error {
	_, err := conn.Write([]byte(myName))
	if err != nil {
		// If server crashes, we should get a timeout / broken pipe error
		fmt.Println("Error sending heartbeat: ", err.Error())
		return err
	}
	fmt.Printf("[%s] Sent heartbeat to server\n", time.Now().Format(time.RFC850))
	return nil
}

/*
 * sendReelectionToServer sends a reelect message to the server
 */
func sendReelectionToServer(conn net.Conn) error {
	_, err := conn.Write([]byte("ELECTED"))
	if err != nil {
		// If server crashes, we should get a timeout / broken pipe error
		fmt.Println("Error sending heartbeat: ", err.Error())
		return err
	}
	fmt.Printf("[%s] Sent reelection to server\n", time.Now().Format(time.RFC850))
	return nil
}

/*
 * sendHeartbeatsRoutine is a routine that sends a heartbeat to server every heartbeatFreq seconds
 */
func sendHeartbeatsRoutine(conn net.Conn, heartbeatFreq int, myId int, gfdConn net.Conn, isPrimary bool) {
	first := true
	for {
		err := sendHeartbeatToServer(conn, "LFD"+strconv.Itoa(myId)+" heartbeat")
		// If we have an error, likely the server has crashed and we will stop running
		if err != nil {
			fmt.Printf("[%s] Server has crashed!\n", time.Now().Format(time.RFC850))
			_, err := gfdConn.Write([]byte(strconv.Itoa(myId) + ",remove"))
			if err != nil {
				fmt.Printf("GFD has crashed!\n", time.Now().Format(time.RFC850))
				return
			}
			return
		}
		if first {
			if isPrimary {
				_, err = gfdConn.Write([]byte(strconv.Itoa(myId) + ",add_primary"))
			} else {
				_, err = gfdConn.Write([]byte(strconv.Itoa(myId) + ",add"))
			}
			if err != nil {
				fmt.Printf("GFD has crashed!\n", time.Now().Format(time.RFC850))
				return
			}
			first = false
		}
		time.Sleep(time.Duration(heartbeatFreq) * time.Second)
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
		fmt.Printf("[%s] Received %s from server\n", time.Now().Format(time.RFC850), string(buf[:mlen]))
	}
}

/*
 * server message protocol is <serverID>:<isPrimary>
 */
func listenForServers(listener net.Listener, serverChan chan net.Conn) {
	for {
		serverConn, err := listener.Accept()
		if err != nil {
			// handle connection error
			fmt.Println("Error accepting: ", err.Error())
		} else {
			serverChan <- serverConn
		}
	}
}

/*
 * Protocol for relaunch messages are assumed to be RELAUNCH:<serverID>
* Protocol for reelection messages are assumed to be ELECTED:<serverID>
*/
func listenForGFDCommands(gfdConn net.Conn, serverConn net.Conn) {
	for {
		buf := make([]byte, 1024)
		mlen, err := gfdConn.Read(buf)
		if err != nil {
			fmt.Println("Error reading: ", err.Error())
			return
		}
		fmt.Printf("[%s] Received %s from GFD\n", time.Now().Format(time.RFC850), string(buf[:mlen]))
		gfdCommand := strings.Split(string(buf[:mlen]), ":")[0]
		serverIdStr := strings.Split(string(buf[:mlen]), ":")[1]

		if gfdCommand == "RELAUNCH" {
			cmd := exec.Command("go", "run", "server/main.go", serverIdStr, "10", "0", HOST1, HOST2, HOST3)
			stdout, err := cmd.Output()
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			fmt.Print(string(stdout))
		} else if gfdCommand == "ELECTED" {
			sendReelectionToServer(serverConn)
		}

		fmt.Printf("[%s] Received %s from GFD\n", time.Now().Format(time.RFC850), string(buf[:mlen]))

	}
}

func main() {
	fmt.Println("---------- Local Fault Detector Started ----------")
	var err error
	heartbeatFreq := 1

	/* args[0] is the heartbeatFreq
	 * args[1] is the Host Name
	 */
	args := os.Args[1:]
	if len(args) > 0 {
		heartbeatFreq, err = strconv.Atoi(args[0])
		if err != nil {
			fmt.Println("Error parsing args: ", err.Error())
			return
		}
		fmt.Printf("Set heartbeat frequency to %d seconds\n", heartbeatFreq)
	}

	//connect to gfd
	gfdConn, err := net.Dial("tcp", args[1]+":8000")
	if err != nil {
		// handle connection error
		fmt.Println("Error dialing: ", err.Error())
		return
	}
	defer gfdConn.Close()
	buf := make([]byte, 1024)
	mlen, err := gfdConn.Read(buf)
	if err != nil {
		fmt.Println("Error reading: ", err.Error())
		return
	}
	lfdID, err := strconv.Atoi(string(buf[:mlen]))
	if err != nil {
		fmt.Println("Error converting ID data:", err.Error())
		return
	}
	fmt.Println("Received LFD ID: ", lfdID)

	// listen for server connection
	listener, err := net.Listen(SERVER_TYPE, ":"+PORT)
	if err != nil {
		// handle server initialization error
		fmt.Println("Error initializing: ", err.Error())
		return
	}
	defer listener.Close()

	newServerChan := make(chan net.Conn)
	go listenForServers(listener, newServerChan)
	for {
		select {
		case conn := <-newServerChan:
			buf := make([]byte, 1024)
			mlen, err := conn.Read(buf)
			if err != nil {
				fmt.Println("Error reading: ", err.Error())
				return
			}
			serverMsg := string(buf[:mlen])
			fmt.Println("SERVERMSG: " + serverMsg)
			serverID, err := strconv.Atoi(strings.Split(serverMsg, ":")[0])
			isPrimary, err := strconv.ParseBool(strings.Split(serverMsg, ":")[1])
			if err != nil {
				fmt.Println("Error converting ID data:", err.Error())
				return
			}
			fmt.Println("Received ID from server: ", serverID)

			go sendHeartbeatsRoutine(conn, heartbeatFreq, serverID, gfdConn, isPrimary)
			go listenToServerRoutine(conn)
			go listenForGFDCommands(gfdConn, conn)
		}
	}
}
