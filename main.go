package main

import "fmt"

func main() {
	cm := MakeCommunicatorManager()

	err := ReadConfig(cm, "./config")
	if err != nil {
		panic(err)
	}

	for {
		connection := <-cm.ConnectionCh()
		fmt.Printf("got new connection in main.go: %#v, passing to HandleConnection...\n", connection)
		go HandlePlayerConnection(connection)
	}
}
