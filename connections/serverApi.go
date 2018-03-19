package connections

import (
	"fmt"

	"doomrooms/types"

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

func (gs *GameServer) Send(method string, args ...interface{}) (interface{}, error) {
	return gs.Connection.Send(method, args...)
}

func (gs *GameServer) Emit(event string, args ...interface{}) error {
	val := gs.NotifyOptions[event]
	b := val == "on" || val == ""

	if b {
		args = append([]interface{}{event}, args...)
		return gs.Connection.Write("emit", args...)
	}
	return nil
}

var GameServers = make([]*GameServer, 0)

func gameServerIndex(gs *GameServer) int {
	for i, x := range GameServers {
		if x == gs {
			return i
		}
	}
	return -1
}

func addGameServer(gs *GameServer) error {
	i := gameServerIndex(gs)
	if i != -1 {
		return fmt.Errorf("server already added")
	}

	GameServers = append(GameServers, gs)

	return nil
}
func removeGameServer(gs *GameServer) error {
	i := gameServerIndex(gs)
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

func HandleGameServerConnection(conn *Connection) {
	defer conn.Close()

	gs := &GameServer{
		Connection: conn,
		NotifyOptions: map[string]string{
			"room-creation": "on",
			"room-join":     "off",
			"room-leave":    "off",

			// why would anyone want to set these to "off"?
			"game-start":  "on",
			"pipe-opened": "on",
		},
	}

	addGameServer(gs)
	defer removeGameServer(gs)

	for {
		msg := <-conn.Chan()
		if conn.Closed() {
			log.Info("connection closed")
			break
		}

		onGameServerCommand(gs, msg)
	}
}

func onGameServerCommand(gs *GameServer, msg *types.Message) {
	handled := false
	conn := gs.Connection

	handleCommand := func(method string, argCount int, fn func() (interface{}, string)) {
		if msg.Method != method {
			return
		}
		handled = true

		if len(msg.Args) != argCount && argCount != -1 {
			conn.Reply(msg.ID, "not enough arguments", nil)
			return
		}

		res, err := fn()
		conn.Reply(msg.ID, err, res)
	}

	handleCommand("attach-game", 2, func() (interface{}, string) {
		gameID := msg.Args[0].(string)
		force := msg.Args[1].(bool)

		if gs.Game() != nil {
			// TODO: handle old game
		}

		g := GetGame(gameID)
		if g == nil {
			return nil, "game not found"
		}

		if g.gameServer != nil {
			if !force {
				return nil, "a gameserver has already been attached and force arg is false"
			}

			g.gameServer.Emit("conn-overrule")
		}

		g.gameServer = gs
		return g, ""
	})

	handleCommand("make-game", 2, func() (interface{}, string) {
		gameID := msg.Args[0].(string)
		gameName := msg.Args[1].(string)

		if gs.Game() != nil {
			// TODO: handle old game
		}

		g, err := MakeGame(gameID, gameName)
		if err != nil {
			return nil, err.Error()
		}

		g.gameServer = gs
		return g, ""
	})

	handleCommand("set-notif-option", 2, func() (interface{}, string) {
		key := msg.Args[0].(string)
		val := msg.Args[1].(string)

		gs.NotifyOptions[key] = val
		return gs.NotifyOptions, ""
	})

	handleCommand("list-rooms", 0, func() (interface{}, string) {
		return gs.Game().rooms, ""
	})

	handleCommand("search-rooms", 1, func() (interface{}, string) {
		query := msg.Args[0].(string)
		rooms := gs.Game().SearchRooms(query, true)
		return rooms, ""
	})

	// left for basic communication, use PipeSessions for bigger amounts of
	// communcation instead.
	handleCommand("message-player", -1, func() (interface{}, string) {
		nick := msg.Args[0].(string)
		args := msg.Args[1:]

		p := GetPlayer(nick)
		if p == nil {
			return nil, "player-not-found"
		}

		res, err := p.Send("game-server-message", args...)
		if err != nil {
			return nil, err.Error()
		}

		return res, ""
	})

	handleCommand("get-private-player-tags", 1, func() (interface{}, string) {
		nick := msg.Args[0].(string)

		p := GetPlayer(nick)
		if p == nil {
			return nil, "player-not-found"
		}

		tags := p.privateTags[gs.Game().ID]
		return tags, ""
	})

	handleCommand("set-private-player-tags", 2, func() (interface{}, string) {
		nick := msg.Args[0].(string)
		tags := msg.Args[1].(map[string]interface{})

		p := GetPlayer(nick)
		if p == nil {
			return nil, "player-not-found"
		}

		p.privateTags[gs.Game().ID] = tags
		return tags, ""
	})

	handleCommand("start-game", 1, func() (interface{}, string) {
		roomID := msg.Args[0].(string)

		room := gs.Game().GetRoom(roomID)
		if room == nil {
			return nil, "room-not-found"
		}

		err := room.Start()
		if err != nil {
			return nil, err.Error()
		}

		return room, ""
	})

	if !handled {
		conn.Reply(msg.ID, "unknown command", nil)
	}
}
