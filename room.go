package main

import "fmt"

type Room struct {
	ID      string                 `json:"id"`
	Name    string                 `json:"name"`
	Hidden  bool                   `json:"hidden"`
	Options map[string]interface{} `json:"options"`

	Players []*Player `json:"players"`
	Admin   *Player   `json:"admin"`

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

func (r *Room) PlayerInvited(player *Player) bool {
	return PlayerIndex(r.invited, player) > -1
}

func (r *Room) InvitePlayer(player *Player) error {
	if r.PlayerInvited(player) {
		return fmt.Errorf("player already invited")
	}

	r.invited = append(r.invited, player) // REVIEW: safe?
	return player.Send("emit", "room-invite", r.ID)
}

func (r *Room) Broadcast(method string, args ...interface{}) error {
	for _, player := range r.Players {
		err := player.Send(method, args...)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Room) Game() *Game {
	return r.game
}

func (r *Room) Start() error {
	r.Broadcast("game-start", r)
	r.Game().GameServer().Emit("game-start", r)

	return nil
}
