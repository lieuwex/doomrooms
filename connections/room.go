package connections

import (
	"errors"
	"fmt"
)

type Room struct {
	ID      string                 `json:"id"`
	Name    string                 `json:"name"`
	Public  bool                   `json:"public"`
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

	// remove the player from their current room, if any
	if room := player.CurrentRoom(); room != nil {
		if err := room.RemovePlayer(player); err != nil {
			return err
		}
	}

	r.UninvitePlayer(player)

	// add player
	player.CurrentRoomID = r.ID
	r.Players = append(r.Players, player)

	// broadcast their join
	if err := r.Broadcast("player-join", player.Nickname); err != nil {
		return err
	}

	// set current player as admin, if no other player has been set as one
	if r.Admin == nil {
		r.Admin = player
		if err := r.Broadcast("admin-change", r.Admin.Nickname); err != nil {
			return err
		}
	}

	return nil
}

func (r *Room) RemovePlayer(player *Player) error {
	playerIndex := playerIndex(r.Players, player)
	if playerIndex == -1 {
		return errors.New("player not in room")
	} else if len(r.Players) == 1 {
		return errors.New("current player only player in room")
	}

	// remove player
	r.Players = append(r.Players[:playerIndex], r.Players[playerIndex+1:]...)
	player.CurrentRoomID = ""

	// if the current player is the admin, promote another user to admin
	if r.Admin == player {
		r.Admin = r.Players[0]
		if err := r.Broadcast("admin-change", r.Admin.Nickname); err != nil {
			return err
		}
	}

	// broadcast leave
	if err := r.Broadcast("player-leave", player.Nickname); err != nil {
		return err
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
