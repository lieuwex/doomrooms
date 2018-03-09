package main

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
)

var Players = make(map[string]*Player)

func GetPlayer(nick string) *Player {
	return Players[nick]
}

func HandlePlayerConnection(conn *Connection) {
	defer conn.netConn.Close()

	msg := <-conn.Chan()
	if conn.closed {
		log.Info("connection closed")
		return
	}

	expectArgs := func(expected int) bool {
		if len(msg.Args) != expected {
			errMsg := fmt.Sprintf("expected %d arg(s)", expected)
			conn.Reply(msg.ID, errMsg, nil)
			return false
		}
		return true
	}

	var p *Player
	var err error

	switch msg.Method {
	case "login":
		if !expectArgs(2) {
			return
		}

		p = Players[msg.Args[0].(string)]
		if p == nil {
			err = fmt.Errorf("User not found")
		}
	case "make-player":
		if !expectArgs(2) {
			return
		}

		username := msg.Args[0].(string)
		password := msg.Args[1].(string)

		p, err = MakePlayer(username, password)
	case "pipe-session":
		if !expectArgs(2) {
			return
		}

		err = fmt.Errorf("TODO")

	default:
		conn.Reply(msg.ID, "expected greeting message", nil)
		return
	}

	// ????????

	if err != nil {
		conn.Reply(msg.ID, err.Error(), nil)
		return
	}

	p.addConnection(conn)
	defer p.removeConnection(conn)

	log.WithFields(log.Fields{
		"player": p,
	}).Info("got player")
	conn.Reply(msg.ID, "", p)

	for {
		msg := <-conn.Chan()
		if conn.closed {
			log.Info("connection closed")
			break
		}

		// REVIEW
		onPlayerCommand(p, conn, msg)
	}
}

type Player struct {
	Nickname      string                 `json:"nick"`
	Tags          map[string]interface{} `json:"tags"`
	CurrentGameID string                 `json:"currentGameId"`

	currentRoom *Room
	password    string
	connections []*Connection
	privateTags map[string]map[string]interface{}
}

func checkNickname(nick string) bool {
	lowerNick := strings.ToLower(nick)
	for n, _ := range Players {
		if strings.ToLower(n) == lowerNick {
			return false
		}
	}
	return true
}

func MakePlayer(nick string, password string) (*Player, error) {
	for !checkNickname(nick) {
		return nil, fmt.Errorf("nickname already in use")
	}

	p := &Player{
		Nickname: nick,
		password: password,

		connections: make([]*Connection, 0),
		privateTags: make(map[string]map[string]interface{}),
	}
	Players[nick] = p
	return p, nil
}

func (p *Player) Game() *Game {
	return GetGame(p.CurrentGameID)
}

func (p *Player) CurrentRoom() *Room {
	return p.currentRoom
}

func (p *Player) Send(method string, args ...interface{}) error {
	for _, conn := range p.connections {
		err := conn.Send(method, args...)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Player) addConnection(conn *Connection) error {
	i := ConnectionIndex(p.connections, conn)
	if i != -1 {
		return fmt.Errorf("connection already added")
	}

	p.connections = append(p.connections, conn) // REVIEW
	return nil
}

func (p *Player) removeConnection(conn *Connection) error {
	i := ConnectionIndex(p.connections, conn)
	if i == -1 {
		return fmt.Errorf("no matching connection found")
	}

	p.connections[i] = p.connections[len(p.connections)-1]
	p.connections = p.connections[:len(p.connections)-1]

	return nil
}
