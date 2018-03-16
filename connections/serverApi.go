package connections

import (
	"fmt"

	"doomrooms/communicators/json"
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

func ListenGameservers(host string, port string) error {
	comm := json.MakeTCPJSONCommunicator()
	err := comm.Start(host, port)
	if err != nil {
		return err
	}

	go func() {
		for {
			netConn := <-comm.ConnectionCh()
			gs := &GameServer{
				Connection: MakeConnection(netConn),
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
	defer conn.Close()

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

	reply := func(err string, res interface{}) {
		conn.Reply(msg.ID, err, res)
	}

	handleCommand := func(method string, argCount int, fn func()) {
		if msg.Method != method {
			return
		}
		handled = true

		if len(msg.Args) != argCount && argCount != -1 {
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

		if gs.Game() != nil {
			// TODO: handle old game
		}

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

		if gs.Game() != nil {
			// TODO: handle old game
		}

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

	handleCommand("list-rooms", 0, func() {
		reply("", gs.Game().rooms)
	})

	handleCommand("search-rooms", 1, func() {
		query := msg.Args[0].(string)
		rooms := gs.Game().SearchRooms(query, true)
		reply("", rooms)
	})

	handleCommand("message-player", -1, func() {
		nick := msg.Args[0].(string)
		args := msg.Args[1:]

		p := GetPlayer(nick)
		if p == nil {
			reply("player-not-found", nil)
		}

		res, err := p.Send("game-server-message", args...)
		if err != nil {
			reply(err.Error(), nil)
			return
		}

		reply("", res)
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

	handleCommand("start-game", 1, func() {
		roomID := msg.Args[0].(string)

		room := gs.Game().GetRoom(roomID)
		if room == nil {
			reply("room-not-found", nil)
		}

		err := room.Start()
		if err != nil {
			reply(err.Error(), nil)
			return
		}

		reply("", room)
	})

	if !handled {
		reply("unknown command", nil)
	}
}
