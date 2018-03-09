package main

func onPlayerCommand(player *Player, conn *Connection, msg Message) {
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

		if len(msg.Args) != argCount {
			reply("not enough arguments", nil)
			return
		}

		fn()
	}

	handleGameCommand := func(method string, argCount int, fn func()) {
		handleCommand(method, argCount, func() {
			if player.Game() == nil {
				reply("no game selected", nil)
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

	handleCommand("make-room", 1, func() {
		name := msg.Args[0].(string)

		room := player.Game().MakeRoom(name)
		room.AddPlayer(player)

		reply("", room)

		player.currentRoom = room
	})

	handleCommand("join-room", 1, func() {
		id := msg.Args[0].(string)
		room := player.Game().GetRoom(id)
		if room == nil {
			reply("room not found", nil)
			return
		}

		// TODO: do some checks whatever to check if the player can join the
		// game.

		err := room.AddPlayer(player)
		if err != nil {
			reply(err.Error(), nil)
			return
		}

		reply("", room)

		player.currentRoom = room
	})

	handleCommand("search", 1, func() {
		query := msg.Args[0].(string)
		rooms := player.Game().SearchRooms(query)
		reply("", rooms)
	})

	handleCommand("send-private-chat", 2, func() {
		nick := msg.Args[0].(string)
		line := msg.Args[1].(string)

		target := Players[nick]
		if target == nil {
			reply("player-not-found", nil)
			return
		}

		target.Send("emit", "private-chat", player.Nickname, line)
		reply("", nil)
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
		p := Players[nick]
		if p == nil {
			reply("player-not-found", nil)
			return
		}

		player.CurrentRoom().InvitePlayer(p)
		reply("", nil)
	})

	handleRoomCommand("start", 0, func() {
		player.currentRoom.Broadcast("start", "jaja")

		// HACK: for demo
		player.currentRoom.Game().GameServer().Connection.Send("gamestartofzo", player.currentRoom.ID)

		reply("", "started!")
	})

	if !handled {
		reply("unknown command", nil)
	}
}
