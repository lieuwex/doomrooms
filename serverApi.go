package main

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

type GameServer struct {
	Connection *Connection
	Game       *Game
}

var GameServers = make([]*GameServer, 0)

func addGameServer(gs *GameServer) error {
	i := GameServerIndex(GameServers, gs)
	if i != -1 {
		return fmt.Errorf("server already added")
	}

	GameServers = append(GameServers, gs)

	return nil
}
func removeGameServer(gs *GameServer) error {
	i := GameServerIndex(GameServers, gs)
	if i == -1 {
		return fmt.Errorf("no matching server found")

	}

	GameServers[i] = GameServers[len(GameServers)-1]
	GameServers = GameServers[:len(GameServers)-1]

	return nil
}

func ListenGameservers(host string, port string) error {
	comm := MakeTCPCommunicator()
	err := comm.Start(host, port)
	if err != nil {
		panic(err) // TODO
		return err
	}
	defer comm.Stop() // REVIEW

	for {
		gs := &GameServer{
			Connection: <-comm.ConnectionCh(),
		}
		go HandleGameServer(gs)
	}

	return nil
}

func HandleGameServer(gs *GameServer) {
	conn := gs.Connection
	defer conn.netConn.Close()

	addGameServer(gs)
	defer removeGameServer(gs)

	msg := <-conn.Chan()
	if conn.closed {
		return
	}

	if msg.Method != "hello" {
		// REVIEW
		conn.Reply(msg.ID, "expected 'hello' msg", nil)
		return
	} else if len(msg.Args) != 2 {
		// REVIEW
		conn.Reply(msg.ID, "expected 2 args", nil)
		return
	}

	g, err := MakeGame(conn, msg.Args[0].(string), msg.Args[1].(string))
	if err != nil {
		conn.Reply(msg.ID, err.Error(), nil)
		return
	}

	log.WithFields(log.Fields{
		"game": g,
	}).Info("made game")

	gs.Game = g
	conn.Reply(msg.ID, "", g)

	for {
		msg := <-conn.Chan()
		if conn.closed {
			log.Info("connection closed")
			break
		}

		// REVIEW
		onGameServerCommand(gs, msg)
	}
}

func onGameServerCommand(gs *GameServer, msg Message) {
	handled := false
	conn := gs.Connection

	reply := func(err string, res interface{}) {
		conn.Reply(msg.ID, err, res)
	}

	handleCommand := func(method string, argCount int, fn func()) {
		if msg.Method != method {
			return
		}
		handled = true

		if len(msg.Args) != argCount {
			reply("not enough arguments", nil)
			return
		}

		fn()
	}

	handleCommand("ping", 0, func() {
		reply("", "pong")
	})

	if !handled {
		reply("unknown command", nil)
	}
}
