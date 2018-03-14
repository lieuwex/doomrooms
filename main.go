package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"

	log "github.com/sirupsen/logrus"
)

func onInterrupt(fn func()) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		fn()
		os.Exit(1)
	}()
}

var configPath string

func parseFlags() {
	flag.StringVar(&configPath, "config", "./config.toml", "location of the config file")

	flag.Parse()
}

func main() {
	SetLogrusFormatter()
	parseFlags()

	exists, err := FileExists(configPath)
	if err != nil {
		panic(err)
	} else if !exists {
		fmt.Printf("no config file found at '%s'\n\n", configPath)
		flag.Usage()
		os.Exit(1)
	}

	cm := MakeCommunicatorManager()

	onInterrupt(func() {
		log.Info("stopping all running services")
		cm.StopServices()
	})

	err = ReadConfig(cm, configPath)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Fatal("error while reading config")
	}

	for {
		connection := <-cm.ConnectionCh()
		log.WithFields(log.Fields{
			"conn": connection,
		}).Info("got new connection in main.go")
		go HandlePlayerConnection(connection)
	}
}
