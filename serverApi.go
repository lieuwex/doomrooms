package main

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

type GameServer struct {
	Connection    *Connection
	NotifyOptions map[string]string
}

func (gs *GameServer) Game() *Game {
	for _, g := range Games {
		if g.GameServer() == gs {
			return g
		}
	}
	return nil
}

func (gs *GameServer) Send(method string, args ...interface{}) error {
	return gs.Connection.Send(method, args...)
}

func (gs *GameServer) Emit(event string, args ...interface{}) error {
	val := gs.NotifyOptions[event]
	b := val == "on" || val == ""

	if b {
		args = append([]interface{}{event}, args...)
		return gs.Send("emit", args...)
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
	comm := MakeTCPJSONCommunicator()
	err := comm.Start(host, port)
	if err != nil {
		return err
	}

	go func() {
		for {
			gs := &GameServer{
				Connection: <-comm.ConnectionCh(),
				NotifyOptions: map[string]string{
					"room-creation": "on",
					"room-join":     "off",
					"room-leave":    "off",

					"game-start": "on", // why would anyone want to set this to "off"?
				},
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

	handleCommand("set-notif-option", 2, func() {
		key := msg.Args[0].(string)
		val := msg.Args[1].(string)

		gs.NotifyOptions[key] = val
		reply("", gs.NotifyOptions)
	})

	handleCommand("message-player", 2, func() {
		nick := msg.Args[0].(string)
		thing := msg.Args[1]

		p := GetPlayer(nick)
		if p == nil {
			reply("player-not-found", nil)
		}

		p.Send("emit", "gameserver-message", thing)
	})

	handleCommand("get-private-player-tags", 1, func() {
		nick := msg.Args[0].(string)

		p := GetPlayer(nick)
		if p == nil {
			reply("player-not-found", nil)
		}

		tags := p.privateTags[gs.Game().ID]
		reply("", tags)
	})

	handleCommand("set-private-player-tags", 2, func() {
		nick := msg.Args[0].(string)
		tags := msg.Args[1].(map[string]interface{})

		p := GetPlayer(nick)
		if p == nil {
			reply("player-not-found", nil)
		}

		p.privateTags[gs.Game().ID] = tags
		reply("", tags)
	})

	if !handled {
		reply("unknown command", nil)
	}
}
