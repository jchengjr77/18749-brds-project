// 18749 Building Reliable Distributed Systems - Project GFD

package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

const (
	SERVER_TYPE = "tcp"
	PORT        = "8000"
)

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

func sendRelaunchToLFD(serverID int, gfdConn net.Conn) {
	_, err := gfdConn.Write([]byte("RELAUNCH:" + strconv.Itoa(serverID)))
	if err != nil {
		fmt.Println("Error sending message to server: ", err.Error())
	}
	fmt.Println("Relaunch sent!")
}

func sendElectionToLFD(serverID *int, lfdConn net.Conn) {
	_, err := lfdConn.Write([]byte("ELECTED:" + strconv.Itoa(*serverID)))
	if err != nil {
		fmt.Println("Error sending message to server: ", err.Error())
	}
	fmt.Println("Reelection sent!")
}

/*
*

	NOTE: LFD messages are assumed to be of the form "[id],[action]"
	(no space in between, just a comma)
	[action] can be either 'add_primary', "add" or "remove"
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
func handleUpdate(update LFDUpdate, servers *map[int]bool, lfds map[int]net.Conn, mode string, primary *int) {
	if update.action == "add" {
		_, exists := (*servers)[update.serverID]
		if exists {
			fmt.Println("Cannot add, already registered: ", update.serverID)
			return
		}
		(*servers)[update.serverID] = true
	} else if update.action == "add_primary" {
		_, exists := (*servers)[update.serverID]
		if exists {
			fmt.Println("Cannot add, already registered: ", update.serverID)
			return
		}
		(*servers)[update.serverID] = true
		*primary = update.serverID
		fmt.Println("SET PRIMARY TO SERVER " + strconv.Itoa(*primary))
	} else if update.action == "remove" {
		_, exists := (*servers)[update.serverID]
		if !exists {
			fmt.Println("Cannot remove, isn't registered: ", update.serverID)
			return
		}
		delete(*servers, update.serverID)
		if len(*servers) < 3 {
			sendRelaunchToLFD(update.serverID, lfds[update.serverID])
		}
		// Tolerate the primary failing once
		// Can be extended to more than one but for the project purposes we only tolerate one
		if *primary == update.serverID {
			key := -1
			for k, _ := range *servers {
				key = k
				break
			}
			*primary = key
			sendElectionToLFD(primary, lfds[*primary])
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

	args := os.Args[1:]
	if len(args) < 1 {
		fmt.Printf("go run gfd/main.go <mode: 'active' or 'passive'>")
		return
	}
	mode := args[0]

	// lfd IDs, monotonically increasing
	lfdID := 1

	// No primary yet
	primary := -1

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
			handleUpdate(update, &servers, lfds, mode, &primary)
		case conn := <-newLFDChan:
			go handleLFD(conn, lfdID, lfdUpdatesChan)
			lfds[lfdID] = conn
			lfdID++
		}
		printGFDState(&servers)
	}
}
