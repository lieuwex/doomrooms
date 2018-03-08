package main

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

type GameServer struct {
	Connection *Connection
}

func (gs *GameServer) Game() *Game {
	for _, g := range Games {
		if g.GameServer() == gs {
			return g
		}
	}
	return nil
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

	g := gs.Game()
	if g != nil {
		g.gameServer = nil
	}

	return nil
}

func ListenGameservers(host string, port string) error {
	comm := MakeTCPCommunicator()
	err := comm.Start(host, port)
	if err != nil {
		return err
	}

	go func() {
		for {
			gs := &GameServer{
				Connection: <-comm.ConnectionCh(),
			}
			go HandleGameServer(gs)
		}
	}()

	return nil
}

func HandleGameServer(gs *GameServer) {
	conn := gs.Connection
	defer conn.netConn.Close()

	addGameServer(gs)
	defer removeGameServer(gs)

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

	handleCommand("attach-game", 2, func() {
		gameID := msg.Args[0].(string)
		force := msg.Args[1].(bool)

		g := GetGame(gameID)
		if g == nil {
			reply("game not found", nil)
			return
		}

		if g.gameServer != nil {
			if !force {
				reply("a gameserver has already been attached and force arg is false", nil)
				return
			}

			g.gameServer.Connection.Send("emit", "conn-overrule")
		}

		g.gameServer = gs
		reply("", g)
	})

	handleCommand("make-game", 2, func() {
		gameID := msg.Args[0].(string)
		gameName := msg.Args[1].(string)

		g, err := MakeGame(gameID, gameName)
		if err != nil {
			reply(err.Error(), nil)
			return
		}

		g.gameServer = gs
		reply("", g)
	})

	if !handled {
		reply("unknown command", nil)
	}
}
