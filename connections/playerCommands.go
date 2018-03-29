package connections

import "doomrooms/types"

func onPlayerCommand(player *Player, conn *Connection, msg *types.Message) {
	handled := false

	roomOthers := func() []*Player {
		res := player.CurrentRoom().Players[:0]
		for _, p := range player.CurrentRoom().Players {
			if p != player {
				res = append(res, p)
			}
		}
		return res
	}

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

	handleGameCommand := func(method string, argCount int, fn func() (interface{}, string)) {
		handleCommand(method, argCount, func() (interface{}, string) {
			if player.Game() == nil {
				return nil, "no game set"
			}

			return fn()
		})
	}

	handleRoomCommand := func(method string, argCount int, fn func() (interface{}, string)) {
		handleGameCommand(method, argCount, func() (interface{}, string) {
			if player.CurrentRoom() == nil {
				return nil, "not in a room"
			}

			return fn()
		})
	}

	handleCommand("set-game", 1, func() (interface{}, string) {
		gameID := msg.Args[0].(string)

		g := GetGame(gameID)
		if g == nil {
			return nil, "game not found"
		}

		player.CurrentGameID = gameID
		return g, ""
	})

	handleCommand("send-private-chat", 2, func() (interface{}, string) {
		nick := msg.Args[0].(string)
		line := msg.Args[1].(string)

		target := GetPlayer(nick)
		if target == nil {
			return nil, "player-not-found"
		}

		target.Emit("private-chat", player.Nickname, line)
		return nil, ""
	})

	handleCommand("get-tags", 0, func() (interface{}, string) { // REVIEW
		return player.Tags, ""
	})

	handleCommand("set-tags", 1, func() (interface{}, string) {
		tags := msg.Args[0].(map[string]interface{})

		player.Tags = tags
		return tags, ""
	})

	handleGameCommand("open-pipe", 0, func() (interface{}, string) {
		gs := player.Game().GameServer()
		ps, err := MakePipeSession()
		if err != nil {
			return nil, err.Error()
		}
		gs.Emit("pipe-opened", player, ps.PrivateID)
		return ps.PrivateID, ""
	})

	handleGameCommand("get-current-room", 0, func() (interface{}, string) {
		return player.CurrentRoom(), ""
	})

	handleGameCommand("make-room", 3, func() (interface{}, string) {
		name := msg.Args[0].(string)
		hidden := msg.Args[1].(bool)
		options := msg.Args[2].(map[string]interface{})

		game := player.Game()

		room := game.MakeRoom(player, name, hidden, options)
		room.AddPlayer(player)

		game.gameServer.Emit("room-creation", room)

		return room, ""
	})

	handleGameCommand("join-room", -1, func() (interface{}, string) {
		nargs := len(msg.Args)
		if nargs == 0 {
			return nil, "not-enough-args"
		}

		id := msg.Args[0].(string)
		givenPass := ""
		if nargs > 1 {
			givenPass = msg.Args[1].(string)
		}

		room := player.Game().GetRoom(id)
		if room == nil {
			return nil, "room not found"
		}

		if room.Hidden && !room.PlayerInvited(player) {
			return nil, "not-invited"
		}

		if pass := room.Options["password"]; pass != nil && givenPass != pass {
			return nil, "incorrect-password"
		}

		err := room.AddPlayer(player)
		if err != nil {
			return nil, err.Error()
		}

		player.Game().gameServer.Emit("room-join", room, player)

		return room, ""
	})

	handleGameCommand("get-room", 1, func() (interface{}, string) {
		id := msg.Args[0].(string)
		room := player.Game().GetRoom(id)
		return room, ""
	})

	handleGameCommand("search-rooms", 1, func() (interface{}, string) {
		query := msg.Args[0].(string)
		rooms := player.Game().SearchRooms(query, false)
		return rooms, ""
	})

	handleRoomCommand("send-room-chat", 1, func() (interface{}, string) {
		line := msg.Args[0].(string)

		for _, p := range roomOthers() {
			p.Emit("room-chat", player.Nickname, line)
		}

		return nil, ""
	})
	handleRoomCommand("send-filtered-room-chat", 2, func() (interface{}, string) {
		line := msg.Args[0].(string)
		filter := msg.Args[1].(map[string]interface{})

		for _, p := range roomOthers() {
			if !p.TagsMatch(filter) {
				continue
			}

			p.Emit("filtered-room-chat", player.Nickname, line, filter)
		}

		return nil, ""
	})

	handleRoomCommand("invite-player", 1, func() (interface{}, string) {
		nick := msg.Args[0].(string)

		p := GetPlayer(nick)
		if p == nil {
			return nil, "player-not-found"
		}

		player.CurrentRoom().InvitePlayer(player, p)

		return nil, ""
	})
	handleRoomCommand("uninvite-player", 1, func() (interface{}, string) {
		nick := msg.Args[0].(string)

		p := GetPlayer(nick)
		if p == nil {
			return nil, "player-not-found"
		}

		player.CurrentRoom().UninvitePlayer(p)

		return nil, ""
	})

	handleRoomCommand("kick-player", 2, func() (interface{}, string) {
		nick := msg.Args[0].(string)
		reason := msg.Args[1].(string)
		room := player.CurrentRoom()

		if room.Admin != player {
			return nil, "not-admin"
		}

		p := GetPlayer(nick)
		if p == nil {
			return nil, "player-not-found"
		}

		p.Emit("kick", reason)
		room.RemovePlayer(p)

		return nil, ""
	})

	handleRoomCommand("leave-room", 0, func() (interface{}, string) {
		player.CurrentRoom().RemovePlayer(player)
		return nil, ""
	})

	handleRoomCommand("start", 0, func() (interface{}, string) {
		room := player.CurrentRoom()

		if room.Admin != player {
			return nil, "not-admin"
		}

		err := room.Start()
		if err != nil {
			return nil, err.Error()
		}

		return room, ""
	})

	// left for basic communication, use PipeSessions for bigger amounts of
	// communcation instead.
	handleRoomCommand("message-game-server", -1, func() (interface{}, string) {
		gs := player.Game().gameServer
		res, err := gs.Send("player-message", msg.Args...)
		if err != nil {
			return nil, err.Error()
		}

		return res, ""
	})

	if !handled {
		conn.Reply(msg.ID, "unknown command", nil)
	}
}
