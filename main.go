package main

import (
	log "github.com/sirupsen/logrus"
)

func main() {
	SetLogrusFormatter()

	cm := MakeCommunicatorManager()

	err := ReadConfig(cm, "./config")
	if err != nil {
		panic(err)
	}

	for {
		connection := <-cm.ConnectionCh()
		log.WithFields(log.Fields{
			"conn": connection,
		}).Info("got new connection in main.go")
		go HandlePlayerConnection(connection)
	}
}
