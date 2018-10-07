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

	var err error
	if b {
		args = append([]interface{}{event}, args...)
		err = gs.Connection.Write("emit", args...)
	}

	if err != nil {
		log.WithFields(log.Fields{
			"event": event,
			"args":  args,
			"error": err.Error(),
		}).Info("error while emitting event to gameserver")
	}
	return err
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
			"room-remove":   "on",
			"room-join":     "off",
			"room-leave":    "off",

			"game-stop": "on",

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
		gameID, ok := msg.Args[0].(string)
		if !ok {
			return nil, "invalid-type"
		}
		force, ok := msg.Args[1].(bool)
		if !ok {
			return nil, "invalid-type"
		}

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
		gameID, ok := msg.Args[0].(string)
		if !ok {
			return nil, "invalid-type"
		}
		gameName, ok := msg.Args[1].(string)
		if !ok {
			return nil, "invalid-type"
		}

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
		key, ok := msg.Args[0].(string)
		if !ok {
			return nil, "invalid-type"
		}
		val, ok := msg.Args[1].(string)
		if !ok {
			return nil, "invalid-type"
		}

		gs.NotifyOptions[key] = val
		return gs.NotifyOptions, ""
	})

	handleCommand("list-rooms", 0, func() (interface{}, string) {
		return gs.Game().Rooms(true), ""
	})

	handleCommand("search-rooms", 1, func() (interface{}, string) {
		query, ok := msg.Args[0].(string)
		if !ok {
			return nil, "invalid-type"
		}
		rooms := gs.Game().SearchRooms(query, true)
		return rooms, ""
	})

	// left for basic communication, use PipeSessions for bigger amounts of
	// communcation instead.
	handleCommand("message-player", -1, func() (interface{}, string) {
		nick, ok := msg.Args[0].(string)
		if !ok {
			return nil, "invalid-type"
		}
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
		nick, ok := msg.Args[0].(string)
		if !ok {
			return nil, "invalid-type"
		}

		p := GetPlayer(nick)
		if p == nil {
			return nil, "player-not-found"
		}

		tags := p.privateTags[gs.Game().ID]
		return tags, ""
	})

	handleCommand("set-private-player-tags", 2, func() (interface{}, string) {
		nick, ok := msg.Args[0].(string)
		if !ok {
			return nil, "invalid-type"
		}
		tags, ok := msg.Args[1].(map[string]interface{})
		if !ok {
			return nil, "invalid-type"
		}

		p := GetPlayer(nick)
		if p == nil {
			return nil, "player-not-found"
		}

		p.privateTags[gs.Game().ID] = tags
		return tags, ""
	})

	handleCommand("start-game", 1, func() (interface{}, string) {
		roomID, ok := msg.Args[0].(string)
		if !ok {
			return nil, "invalid-type"
		}

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
