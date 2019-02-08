package main

import (
	"doomrooms/communicators"
	"doomrooms/config"
	"doomrooms/connections"
	"doomrooms/utils"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"time"
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
	rand.Seed(time.Now().UnixNano())

	parseFlags()

	exists, err := utils.FileExists(configPath)
	if err != nil {
		panic(err)
	} else if !exists {
		fmt.Printf("no config file found at '%s'\n\n", configPath)
		flag.Usage()
		os.Exit(1)
	}

	cm := communicators.MakeCommunicatorManager()

	onInterrupt(func() {
		log.Println("stopping all running services")
		cm.StopServices()
	})

	err = config.ReadConfig(cm, configPath)
	if err != nil {
		log.Fatalf("error while reading config: %s", err)
	}

	go func() {
		for {
			connection := <-cm.PlayerConnectionCh()
			log.Printf("got new player connection in main.go: %#v", connection)
			go connections.HandlePlayerConnection(connection)
		}
	}()

	go func() {
		for {
			connection := <-cm.GameServerConnectionCh()
			log.Printf("got new gameserver connection in main.go: %#v", connection)
			go connections.HandleGameServerConnection(connection)
		}
	}()

	go func() {
		for {
			connection := <-cm.PipeSessionConnectionCh()
			log.Printf("got new pipesession connection in main.go: %#v", connection)
			go connections.HandlePipeSesionConnection(connection) // REVIEW
		}
	}()

	select {}
}
