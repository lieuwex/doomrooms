package connections

import "doomrooms/types"

func onPlayerCommand(player *Player, conn *Connection, msg *types.Message) {
	handled := false

	reply := func(err string, res interface{}) {
		conn.Reply(msg.ID, err, res)
	}

	roomOthers := func() []*Player {
		res := player.CurrentRoom().Players[:0]
		for _, p := range player.CurrentRoom().Players {
			if p != player {
				res = append(res, p)
			}
		}
		return res
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

	handleGameCommand := func(method string, argCount int, fn func()) {
		handleCommand(method, argCount, func() {
			if player.Game() == nil {
				reply("no game set", nil)
				return
			}

			fn()
		})
	}

	handleRoomCommand := func(method string, argCount int, fn func()) {
		handleGameCommand(method, argCount, func() {
			if player.currentRoom == nil {
				reply("not in a room", nil)
				return
			}

			fn()
		})
	}

	handleCommand("ping", 0, func() {
		reply("", "pong")
	})

	handleCommand("set-game", 1, func() {
		gameID := msg.Args[0].(string)

		g := GetGame(gameID)
		if g == nil {
			reply("game not found", nil)
		}

		player.CurrentGameID = gameID
		reply("", g)
	})

	handleCommand("send-private-chat", 2, func() {
		nick := msg.Args[0].(string)
		line := msg.Args[1].(string)

		target := GetPlayer(nick)
		if target == nil {
			reply("player-not-found", nil)
			return
		}

		target.Send("emit", "private-chat", player.Nickname, line)
		reply("", nil)
	})

	handleGameCommand("make-room", 3, func() {
		name := msg.Args[0].(string)
		hidden := msg.Args[1].(bool)
		options := msg.Args[2].(map[string]interface{})

		game := player.Game()

		room := game.MakeRoom(player, name, hidden, options)
		room.AddPlayer(player)

		game.gameServer.Emit("room-creation", room)
		reply("", room)

		player.currentRoom = room
	})

	handleGameCommand("join-room", -1, func() {
		nargs := len(msg.Args)
		if nargs == 0 {
			reply("not-enough-args", nil)
			return
		}

		id := msg.Args[0].(string)
		givenPass := ""
		if nargs > 1 {
			givenPass = msg.Args[1].(string)
		}

		room := player.Game().GetRoom(id)
		if room == nil {
			reply("room not found", nil)
			return
		}

		if room.Hidden && !room.PlayerInvited(player) {
			reply("not-invited", nil)
			return
		}

		if pass := room.Options["password"]; pass != nil && givenPass != pass {
			reply("incorrect-password", nil)
			return
		}

		err := room.AddPlayer(player)
		if err != nil {
			reply(err.Error(), nil)
			return
		}

		player.Game().gameServer.Emit("room-join", room, player)
		reply("", room)

		player.currentRoom = room
	})

	handleGameCommand("search-rooms", 1, func() {
		query := msg.Args[0].(string)
		rooms := player.Game().SearchRooms(query, false)
		reply("", rooms)
	})

	handleRoomCommand("send-room-chat", 1, func() {
		line := msg.Args[0].(string)
		players := roomOthers()

		for _, p := range players {
			p.Send("emit", "room-chat", player.Nickname, line)
		}
	})

	// REVIEW: Team chat?

	handleRoomCommand("invite-player", 1, func() {
		nick := msg.Args[0].(string)
		p := GetPlayer(nick)
		if p == nil {
			reply("player-not-found", nil)
			return
		}

		player.CurrentRoom().InvitePlayer(p)
		reply("", nil)
	})

	handleRoomCommand("kick-player", 2, func() {
		nick := msg.Args[0].(string)
		reason := msg.Args[1].(string)
		room := player.CurrentRoom()

		if room.Admin != player {
			reply("not-admin", nil)
			return
		}

		p := GetPlayer(nick)
		if p == nil {
			reply("player-not-found", nil)
			return
		}

		p.Send("emit", "kick", reason)
		room.RemovePlayer(p)

		reply("", nil)
	})

	handleRoomCommand("leave-room", 0, func() {
		err := player.CurrentRoom().RemovePlayer(player)
		if err != nil {
			reply(err.Error(), nil)
			return
		}

		reply("", nil)
	})

	handleRoomCommand("start", 0, func() {
		room := player.CurrentRoom()

		if room.Admin != player {
			reply("not-admin", nil)
			return
		}

		err := room.Start()
		if err != nil {
			reply(err.Error(), nil)
			return
		}

		reply("", room)
	})

	handleRoomCommand("message-game-server", -1, func() {
		gs := player.Game().gameServer
		res, err := gs.Send("player-message", msg.Args...)
		if err != nil {
			reply(err.Error(), nil)
			return
		}

		reply("", res)
	})

	if !handled {
		reply("unknown command", nil)
	}
}
