package main

func onPlayerCommand(player *Player, msg Message) {
	handled := false

	reply := func(err string, res interface{}) {
		player.connection.Reply(msg.ID, err, res)
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

	handleRoomCommand := func(method string, argCount int, fn func()) {
		handleCommand(method, argCount, func() {
			if player.currentRoom == nil {
				reply("not in a room", nil)
				return
			}

			fn()
		})
	}

	handleCommand("make-room", 1, func() { // TODO
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

	handleRoomCommand("start", 0, func() {
		player.currentRoom.Broadcast("start", "jaja")

		player.currentRoom.Game().connection.Send("gamestartofzo", player.currentRoom.ID)

		reply("", "started!")
	})

	if !handled {
		reply("unknown command", nil)
	}
}
