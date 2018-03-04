package main

type Room struct {
	ID      string                 `json:"id"`
	Name    string                 `json:"name"`
	Hidden  bool                   `json:"hidden"`
	Options map[string]interface{} `json:"options"`
	Players []*Player              `json:"players"`

	invited []*Player // REVIEW: hidden?
	game    *Game
}

func (r *Room) AddPlayer(player *Player) error {
	r.Broadcast("emit", "player-join", player.Nickname) // REVIEW

	if i := PlayerIndex(r.invited, player); i > -1 {
		// remove player from invited
		r.invited = append(r.invited[:i], r.invited[i+1:]...)
	}

	r.Players = append(r.Players, player)
	return nil
}

// REVIEW: do we want nickname instead?
func (r *Room) InvitePlayer(player *Player) error {
	return player.connection.Send("emit", "room-invite", r.ID)
}

func (r *Room) Broadcast(method string, args ...interface{}) error {
	for _, player := range r.Players {
		err := player.connection.Send(method, args...)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Room) Game() *Game {
	return r.game
}
