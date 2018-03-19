package connections

import (
	"doomrooms/utils"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
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

	log.WithFields(log.Fields{
		"game": g,
	}).Info("made game")

	return g, nil
}

func (g *Game) MakeRoom(creator *Player, name string, hidden bool, options map[string]interface{}) *Room {
	id := g.idGen.UniqIDf() // this should always be unique
	room := &Room{
		ID:      id,
		Name:    name,
		Hidden:  hidden,
		Options: options,
		Started: false,
		GameID:  g.ID,

		Admin: creator,
	}
	g.rooms[id] = room
	return room
}

func (g *Game) GetRoom(id string) *Room {
	return g.rooms[id]
}

// TODO: also support some shit like r.Options
func (g *Game) SearchRooms(name string, includeHidden bool) []*Room {
	res := make([]*Room, 0)
	name = strings.ToLower(name)

	for _, r := range g.rooms {
		if r.Hidden && !includeHidden {
			continue
		}

		roomName := strings.ToLower(r.Name)
		if strings.Contains(roomName, name) ||
			strings.Contains(name, roomName) {
			res = append(res, r)
		}
	}

	return res
}

func (g *Game) GameServer() *GameServer {
	return g.gameServer
}
