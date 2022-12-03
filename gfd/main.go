// 18749 Building Reliable Distributed Systems - Project GFD

package main

import (
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"strings"
)

const (
	SERVER_TYPE = "tcp"
	PORT        = "8000"
)

// To test (Need to enable ssh and all too)
var serverDestMap = map[int]string{
	1: "M",
	2: "CM",
	3: "D",
}

type LFDUpdate struct {
	serverID int
	action   string
}

func listenerChannelWrapper(listener net.Listener, newLFDChan chan net.Conn) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			// handle connection error
			fmt.Println("Error accepting: ", err.Error())
			return
		}
		newLFDChan <- conn
	}
}

func relaunchDeadReplica(serverID int) {
	dest := serverDestMap[serverID]
	cmd := exec.Command("ssh " + dest + " go run server/main.go 1 10 0")
	stdout, err := cmd.Output()

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Print(string(stdout))
}

/*
*

	NOTE: LFD messages are assumed to be of the form "[id],[action]"
	(no space in between, just a comma)
	[action] can be either "add" or "remove"
*/
func handleLFD(conn net.Conn, lfdID int, lfdUpdatesChan chan LFDUpdate) {
	fmt.Println("New LFD Connected! ID: ", lfdID)
	_, err := conn.Write([]byte(strconv.Itoa(lfdID)))
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
		parts := strings.Split(msg, ",")
		id, err := strconv.Atoi(parts[0])
		if err != nil {
			fmt.Println("Error converting to id: ", err.Error())
		}
		newUpdate := LFDUpdate{id, parts[1]}
		lfdUpdatesChan <- newUpdate
	}
}

// parses LFD update and adds/removes server from list
func handleUpdate(update LFDUpdate, servers *map[int]bool) {
	if update.action == "add" {
		_, exists := (*servers)[update.serverID]
		if exists {
			fmt.Println("Cannot add, already registered: ", update.serverID)
			return
		}
		(*servers)[update.serverID] = true
	} else if update.action == "remove" {
		_, exists := (*servers)[update.serverID]
		if !exists {
			fmt.Println("Cannot remove, isn't registered: ", update.serverID)
			return
		}
		delete(*servers, update.serverID)
		for len(*servers) < 1 {
			relaunchDeadReplica(update.serverID)
			// Delay ?
		}
	} else {
		fmt.Println("----- GFD / RM has no idea what update that was -----")
		fmt.Println(update)
	}
}

func printGFDState(servers *map[int]bool) {
	if len(*servers) == 0 {
		fmt.Println("GFD / RM: 0 members")
	} else {
		keys := make([]int, len(*servers))
		i := 0
		for k := range *servers {
			keys[i] = k
			i++
		}
		fmt.Println("GFD / RM: ", len(*servers), "member(s): ", keys)
	}
}

func main() {
	fmt.Println("---------- Global Fault Detector started ----------")

	// lfd IDs, monotonically increasing
	lfdID := 1

	// map of LFDs [lfd id] -> [net connection]
	lfds := map[int]net.Conn{}

	// set of live servers
	servers := map[int]bool{}

	// make channel for new LFD registration
	newLFDChan := make(chan net.Conn)

	// make channel for LFD update messages
	lfdUpdatesChan := make(chan LFDUpdate)

	// start the GFD / RM
	listener, err := net.Listen(SERVER_TYPE, ":"+PORT)
	if err != nil {
		// handle GFD initialization error
		fmt.Println("Error initializing: ", err.Error())
		return
	}

	go listenerChannelWrapper(listener, newLFDChan)

	defer listener.Close()

	printGFDState(&servers)
	for {
		select {
		case update := <-lfdUpdatesChan:
			handleUpdate(update, &servers)
		case conn := <-newLFDChan:
			go handleLFD(conn, lfdID, lfdUpdatesChan)
			lfds[lfdID] = conn
			lfdID++
		}
		printGFDState(&servers)
	}
}
