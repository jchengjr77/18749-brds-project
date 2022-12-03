# 18749 Building Reliable Distributed Systems

Fall 2022

Jonathan Cheng (jcheng3)

Nathan Ang (nathanan)

Vignesh Rajmohan (vrajmoha)

William Foy (wfoy)

Jeremy Ryu (jihoonr) 

## How To Run

First, in the root directory, 

Run GFD with:

`go run gfd/main.go <1 if active 0 if passive>`

Run LFD with:

`go run lfd/main.go 10 <checkpoint freq> <server hostname>`

Run server with:

`go run server/main.go <server id> <checkpoint freq> <primary?> <backup replica hosts>`

You may add a number as a CLI argument that will set a custom heartbeat_freq

Finally, run clients with: 

`go run client/main.go <host name 1> <host name 2> <host name 3>`

