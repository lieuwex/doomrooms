package connections

import "doomrooms/types"

func onPlayerCommand(player *Player, conn *Connection, msg *types.Message) {
	handled := false

	roomOthers := func() []*Player {
		var res []*Player
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
		gameID, ok := msg.Args[0].(string)
		if !ok {
			return nil, "invalid-type"
		}

		g := GetGame(gameID)
		if g == nil {
			return nil, "game not found"
		}

		player.SetGame(g)
		return g, ""
	})

	handleCommand("send-private-chat", 2, func() (interface{}, string) {
		nick, ok := msg.Args[0].(string)
		if !ok {
			return nil, "invalid-type"
		}
		line, ok := msg.Args[1].(string)
		if !ok {
			return nil, "invalid-type"
		}

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
		tags, ok := msg.Args[0].(map[string]interface{})
		if !ok {
			return nil, "invalid-type"
		}

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
		name, ok := msg.Args[0].(string)
		if !ok {
			return nil, "invalid-type"
		}
		hidden, ok := msg.Args[1].(bool)
		if !ok {
			return nil, "invalid-type"
		}
		options, ok := msg.Args[2].(map[string]interface{})
		if !ok {
			return nil, "invalid-type"
		}

		game := player.Game()

		room := game.MakeRoom(name, hidden, options)
		if err := player.JoinRoom(room); err != nil {
			return nil, err.Error()
		}

		game.gameServer.Emit("room-creation", room)

		return room, ""
	})

	handleGameCommand("join-room", -1, func() (interface{}, string) {
		nargs := len(msg.Args)
		if nargs == 0 {
			return nil, "not-enough-args"
		}

		id, ok := msg.Args[0].(string)
		if !ok {
			return nil, "invalid-type"
		}
		givenPass := ""
		if nargs > 1 {
			givenPass, ok = msg.Args[1].(string)
			if !ok {
				return nil, "invalid-type"
			}
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

		if err := player.JoinRoom(room); err != nil {
			return nil, err.Error()
		}

		player.Game().gameServer.Emit("room-join", room, player)

		return room, ""
	})

	handleGameCommand("get-room", 1, func() (interface{}, string) {
		id, ok := msg.Args[0].(string)
		if !ok {
			return nil, "invalid-type"
		}
		room := player.Game().GetRoom(id)
		return room, ""
	})

	handleGameCommand("search-rooms", 1, func() (interface{}, string) {
		query, ok := msg.Args[0].(string)
		if !ok {
			return nil, "invalid-type"
		}
		rooms := player.Game().SearchRooms(query, false)
		return rooms, ""
	})

	handleRoomCommand("send-room-chat", 1, func() (interface{}, string) {
		line, ok := msg.Args[0].(string)
		if !ok {
			return nil, "invalid-type"
		}

		for _, p := range roomOthers() {
			p.Emit("room-chat", player.Nickname, line)
		}

		return nil, ""
	})
	handleRoomCommand("send-filtered-room-chat", 2, func() (interface{}, string) {
		line, ok := msg.Args[0].(string)
		if !ok {
			return nil, "invalid-type"
		}
		filter, ok := msg.Args[1].(map[string]interface{})
		if !ok {
			return nil, "invalid-type"
		}

		for _, p := range roomOthers() {
			if !p.TagsMatch(filter) {
				continue
			}

			p.Emit("filtered-room-chat", player.Nickname, line, filter)
		}

		return nil, ""
	})

	handleRoomCommand("invite-player", 1, func() (interface{}, string) {
		nick, ok := msg.Args[0].(string)
		if !ok {
			return nil, "invalid-type"
		}

		p := GetPlayer(nick)
		if p == nil {
			return nil, "player-not-found"
		}

		player.CurrentRoom().InvitePlayer(player, p)

		return nil, ""
	})
	handleRoomCommand("uninvite-player", 1, func() (interface{}, string) {
		nick, ok := msg.Args[0].(string)
		if !ok {
			return nil, "invalid-type"
		}

		p := GetPlayer(nick)
		if p == nil {
			return nil, "player-not-found"
		}

		player.CurrentRoom().UninvitePlayer(p)

		return nil, ""
	})

	handleRoomCommand("kick-player", 2, func() (interface{}, string) {
		nick, ok := msg.Args[0].(string)
		if !ok {
			return nil, "invalid-type"
		}
		reason, ok := msg.Args[1].(string)
		if !ok {
			return nil, "invalid-type"
		}
		room := player.CurrentRoom()

		if room.Admin != player {
			return nil, "not-admin"
		}

		p := GetPlayer(nick)
		if p == nil {
			return nil, "player-not-found"
		}

		p.Emit("kick", reason)
		if err := room.RemovePlayer(p); err != nil {
			return nil, err.Error()
		}

		return nil, ""
	})

	handleRoomCommand("leave-room", 0, func() (interface{}, string) {
		if err := player.CurrentRoom().RemovePlayer(player); err != nil {
			return nil, err.Error()
		}
		return nil, ""
	})

	handleRoomCommand("remove-room", 0, func() (interface{}, string) {
		room := player.CurrentRoom()

		if room.Admin != player {
			return nil, "not-admin"
		}

		if err := player.Game().RemoveRoom(room.ID); err != nil {
			return nil, err.Error()
		}
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
