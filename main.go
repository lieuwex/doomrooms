package main

import "fmt"

var Communicators = make([]Communicator, 0)

func main() {
	tcpComm := MakeTCPCommunicator()
	err := tcpComm.Start("", "1337")
	if err != nil {
		panic(err)
	}

	go ListenGameservers("localhost", "6060")

	for {
		connection := <-tcpComm.ConnectionCh()
		fmt.Printf("got new connection in main.go: %#v, passing to HandleConnection...\n", connection)
		go HandlePlayerConnection(connection)
	}
}
