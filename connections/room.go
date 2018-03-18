package connections

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

func playerIndex(players []*Player, p *Player) int {
	for i, x := range players {
		if x == p {
			return i
		}
	}
	return -1
}

func (r *Room) AddPlayer(player *Player) error {
	r.Broadcast("emit", "player-join", player.Nickname) // REVIEW

	if i := playerIndex(r.invited, player); i > -1 {
		// remove player from invited
		r.invited = append(r.invited[:i], r.invited[i+1:]...)
	}

	r.Players = append(r.Players, player)
	return nil
}

func (r *Room) RemovePlayer(player *Player) error {
	i := playerIndex(r.invited, player)
	if i == -1 {
		return fmt.Errorf("player not found")
	}

	r.Players = append(r.Players[:i], r.Players[i+1:]...)

	r.Broadcast("emit", "player-leave", player.Nickname)

	if r.Admin == player {
		r.Admin = r.Players[0]
		r.Broadcast("emit", "admin-change", r.Admin.Nickname)
	}

	return nil
}

func (r *Room) PlayerInvited(player *Player) bool {
	return playerIndex(r.invited, player) > -1
}

func (r *Room) InvitePlayer(inviter, player *Player) error {
	if r.PlayerInvited(player) {
		return fmt.Errorf("player already invited")
	}

	r.invited = append(r.invited, player) // REVIEW: safe?

	return player.Emit("room-invite", inviter.Nickname, r.ID)
}

func (r *Room) Broadcast(event string, args ...interface{}) error {
	for _, player := range r.Players {
		err := player.Emit(event, args...)
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
