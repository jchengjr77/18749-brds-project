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
	SVR_PORT = "8082"
)

type Pair struct {
	First int
	Second int
}

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

func handlePrimary(conn net.Conn, cpChan chan Pair) {
	fmt.Println("New Primary Replica Detected.")
	for {
		buf := make([]byte, 1024)
		mlen, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Error reading: ", err.Error())
			return
		}	
		msg := string(buf[:mlen])
		fmt.Println("[%s] received checkpoint %s", time.Now().Format(time.RFC850), msg)

		// checkpoint format <checkpoint num>,<state>
		parts := strings.Split(msg, ",")
		cpNum, err := strconv.Atoi(parts[0])
		if err != nil {
			fmt.Println("Error converting using Atoi: ", err.Error())
			continue
		}
		cpState, err := strconv.Atoi(parts[1])
		if err != nil {
			fmt.Println("Error converting using Atoi: ", err.Error())
			continue
		}
		cpChan <- Pair{cpNum, cpState}
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

func primaryReplicaListenerWrapper(listener net.Listener, newPrimaryChan chan net.Conn) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			// handle connection error
			fmt.Println("Error accepting: ", err.Error())
			return
		}
		newPrimaryChan <- conn
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
		fmt.Printf("Usage: go run server/main.go <server id> <checkpoint freq> <primary?> <backup replica hosts>")
		return
	}
	serverId, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Println("Error parsing args: ", err.Error())
		return
	}
	// checkpointFreq, err := strconv.Atoi(args[1])
	// if err != nil {
	// 	fmt.Println("Error parsing args: ", err.Error())
	// 	return
	// }
	// isPrimary := (len(args) >= 3)
	// backupHostnames := args[3:]

	// TODO: only send checkpoints messages to other replicas if primary
	/***
		- uncomment the codelines above. (I needed to silence the "declared but not used" warning)
		- if server is primary replica, use checkpointFreq to send checkpoint messages.
		- the primary replica will have to dial() into the two backup replicas. See the client codebase for an example of this. (I defined the backup replica ports to be 8082)
		- iterate through the backupHostnames and dial into each one at port 8082
		- you can check arg[3], or isPrimary, to see if this server is the primary replica
		- checkpoint_freq should be in seconds
		- every time a new checkpoint message is sent, increment checkpoint_count
		- checkpoint messages are of the format "<checkpointCount>,<state>"
	***/

	my_state := 0
	my_checkpoint_count := 1
	fmt.Printf("---------- Server %d started ----------\n", serverId)

	// client IDs, monotonically increasing
	clientID := 1

	// map of clients [client id] -> [net connection]
	clients := map[int]net.Conn{}

	// make channel for new connections
	newClientChan := make(chan net.Conn)

	// make channel for new incrementing state
	stateChan := make(chan int)

	// make channels for new primary replica messages
	newPrimaryChan := make(chan net.Conn)
	checkpointChan := make(chan Pair) // (checkpoint num, state)

	// start the server to receive clients
	listener, err := net.Listen(SERVER_TYPE, ":"+PORT)
	if err != nil {
		// handle server initialization error
		fmt.Println("Error initializing: ", err.Error())
		return
	}

	// listen for new primary replica if needed
	primaryReplicaListener, err := net.Listen(SERVER_TYPE, ":"+SVR_PORT)
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
	go primaryReplicaListenerWrapper(primaryReplicaListener, newPrimaryChan)

	defer listener.Close()
	defer primaryReplicaListener.Close()

	for {
		select {
		case recentID := <-stateChan:
			setState(&my_state, recentID)
		case conn := <-newClientChan:
			go handleClient(conn, clientID, serverId, stateChan)
			clients[clientID] = conn
			clientID++
		case pair := <-checkpointChan:
			cpNum := pair.First
			cpState := pair.Second
			fmt.Printf("[%s] received checkpoint %d -> %d, state %d -> %d \n", time.Now().Format(time.RFC850), my_checkpoint_count, cpNum, my_state, cpState)	
			my_checkpoint_count = cpNum
			my_state = cpState
		case conn := <-newPrimaryChan:
			go handlePrimary(conn, checkpointChan)
		} 
	}
}
