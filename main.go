package main

import (
	"flag"

	log "github.com/sirupsen/logrus"
)

var configPath string

func parseFlags() {
	flag.StringVar(&configPath, "config", "./config", "location of the config file")

	flag.Parse()
}

func main() {
	SetLogrusFormatter()
	parseFlags()

	cm := MakeCommunicatorManager()

	err := ReadConfig(cm, configPath)
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
