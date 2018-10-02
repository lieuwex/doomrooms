package connections

import (
	"errors"
	"fmt"
)

type Room struct {
	ID      string                 `json:"id"`
	Name    string                 `json:"name"`
	Hidden  bool                   `json:"hidden"`
	Options map[string]interface{} `json:"options"`
	Started bool                   `json:"started"`
	GameID  string                 `json:"gameID"`

	Players []*Player `json:"players"`
	Admin   *Player   `json:"admin"`

	invited []*Player // REVIEW: hidden?
}

func playerIndex(players []*Player, p *Player) int {
	for i, x := range players {
		if x == p {
			return i
		}
	}
	return -1
}

func (r *Room) AddPlayer(player *Player) error {
	if r.PlayerInRoom(player) {
		return fmt.Errorf("player already in room")
	}

	if room := player.CurrentRoom(); room != nil {
		if err := room.RemovePlayer(player); err != nil {
			return err
		}
	}

	player.CurrentRoomID = r.ID

	r.UninvitePlayer(player)

	r.Players = append(r.Players, player)
	if err := r.Broadcast("player-join", player.Nickname); err != nil {
		return err
	}

	if r.Admin == nil {
		r.Admin = player
		if err := r.Broadcast("admin-change", r.Admin.Nickname); err != nil {
			return err
		}
	}

	return nil
}

func (r *Room) RemovePlayer(player *Player) error {
	i := playerIndex(r.Players, player)
	if i == -1 {
		return errors.New("player not in room")
	}

	r.Players = append(r.Players[:i], r.Players[i+1:]...)
	player.CurrentRoomID = ""

	if r.Admin == player {
		r.Admin = r.Players[0]
		if err := r.Broadcast("admin-change", r.Admin.Nickname); err != nil {
			return err
		}
	}

	if err := r.Broadcast("player-leave", player.Nickname); err != nil {
		return err
	}

	if len(r.Players) == 0 {
		return r.Game().RemoveRoom(r.ID)
	}
	return nil
}

func (r *Room) PlayerInRoom(player *Player) bool {
	return playerIndex(r.Players, player) > -1
}

func (r *Room) PlayerInvited(player *Player) bool {
	return playerIndex(r.invited, player) > -1
}

func (r *Room) InvitePlayer(inviter, player *Player) error {
	if r.PlayerInvited(player) {
		return fmt.Errorf("player already invited")
	} else if r.PlayerInRoom(player) {
		return fmt.Errorf("player already in room")
	}

	r.invited = append(r.invited, player) // REVIEW: safe?

	if err := player.Emit("room-invite", inviter, r); err != nil {
		return err
	}

	return r.Broadcast("player-invited", inviter.Nickname, player)
}

func (r *Room) UninvitePlayer(player *Player) {
	if i := playerIndex(r.invited, player); i > -1 {
		r.invited = append(r.invited[:i], r.invited[i+1:]...)
	}
}

func (r *Room) Broadcast(event string, args ...interface{}) error {
	args = append([]interface{}{r.ID}, args...)

	for _, player := range r.Players {
		if err := player.Emit(event, args...); err != nil {
			return err
		}
	}

	return nil
}

func (r *Room) Game() *Game {
	return GetGame(r.GameID)
}

func (r *Room) Start() error {
	if r.Started {
		return fmt.Errorf("already started")
	}

	r.Started = true
	if err := r.Broadcast("game-start", r); err != nil {
		return err
	}
	return r.Game().GameServer().Emit("game-start", r)
}

func (r *Room) Stop() error { // REVIEW
	if !r.Started {
		return fmt.Errorf("already stopped")
	}

	r.Started = false
	if err := r.Broadcast("game-stop", r); err != nil {
		return err
	}
	return r.Game().GameServer().Emit("game-stop", r)
}
