package connections

import (
	"doomrooms/utils"
	"errors"
	"fmt"
	"log"
	"strings"
)

var Games = make(map[string]*Game)

type Game struct {
	ID   string `json:"id"`
	Name string `json:"name"`

	rooms map[string]*Room

	idGen      utils.IDGenerator
	gameServer *GameServer

	// TODO: player connections
}

func GetGame(id string) *Game {
	return Games[id]
}

func MakeGame(id string, name string) (*Game, error) {
	if GetGame(id) != nil {
		return nil, fmt.Errorf("game with ID '%s' already exists", id)
	}

	g := &Game{
		ID:   id,
		Name: name,

		rooms: make(map[string]*Room),

		idGen: utils.MakeIDGenerator(),
	}
	Games[id] = g // REVIEW: safe?

	log.Printf("made game %#v", g)

	return g, nil
}

func (g *Game) MakeRoom(name string, hidden bool, options map[string]interface{}) *Room {
	id := g.idGen.UniqIDf() // this should always be unique
	room := &Room{
		ID:      id,
		Name:    name,
		Hidden:  hidden,
		Options: options,
		Started: false,
		GameID:  g.ID,
	}
	g.rooms[id] = room
	return room
}

func (g *Game) GetRoom(id string) *Room {
	return g.rooms[id]
}

func (g *Game) RemoveRoom(id string) error {
	room := g.GetRoom(id)
	if room == nil {
		return errors.New("room not found")
	}

	// REVIEW
	for _, p := range room.Players {
		if err := room.RemovePlayer(p); err != nil {
			return err
		}
	}

	room.Broadcast("room-remove")
	g.rooms[id] = nil
	g.GameServer().Emit("room-remove", room)

	return nil
}

func (g *Game) Rooms(includeHidden bool) []*Room {
	var res []*Room

	for _, r := range g.rooms {
		if !r.Hidden || includeHidden {
			res = append(res, r)
		}
	}

	return res
}

// TODO: also support some shit like r.Options
func (g *Game) SearchRooms(query string, includeHidden bool) []*Room {
	res := make([]*Room, 0)
	query = strings.ToLower(query)

	for _, r := range g.rooms {
		if r.Hidden && !includeHidden {
			continue
		}

		name := strings.ToLower(r.Name)
		if strings.Contains(name, query) || strings.Contains(query, name) {
			res = append(res, r)
		}
	}

	return res
}

func (g *Game) GameServer() *GameServer {
	return g.gameServer
}
