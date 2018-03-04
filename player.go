package main

import (
	"fmt"
)

func HandlePlayerConnection(conn *Connection) {
	defer conn.netConn.Close()

	msg := <-conn.Chan()
	if conn.closed {
		fmt.Printf("connection closed lol\n")
		return
	}

	if msg.Method != "hello" {
		// REVIEW
		conn.Reply(msg.ID, "expected 'hello' msg", nil)
		return
	} else if len(msg.Args) != 1 {
		// REVIEW
		conn.Reply(msg.ID, "expected 1 arg", nil)
		return
	}

	p, err := MakePlayer(conn, msg.Args[0].(string))
	if err != nil {
		// REVIEW
		conn.Reply(msg.ID, err.Error(), nil)
		return
	}

	fmt.Printf("got player: %#v\n", p)
	conn.Reply(msg.ID, "", p)

	for {
		msg := <-conn.Chan()
		if conn.closed {
			fmt.Printf("connection closed lol\n")
			break
		}

		// REVIEW
		onPlayerCommand(p, msg)
	}
}

type Player struct {
	// connection the the connection which is used to send commands to Doomrooms.
	connection *Connection
	// pipedConnection is the connection which gets piped to the game server.
	pipedConnection *Connection

	GameID   string                 `json:"gameId"`
	Nickname string                 `json:"nick"`
	Info     map[string]interface{} `json:"info"`

	currentRoom *Room
}

func MakePlayer(conn *Connection, gameID string) (*Player, error) {
	g := GetGame(gameID)
	if g == nil {
		return nil, fmt.Errorf("no game with id '%s' found", gameID)
	}

	nick := UniqIDf()
	for !g.CheckNickname(nick) {
		nick = UniqIDf()
	}

	p := &Player{
		connection: conn,

		GameID:   gameID,
		Nickname: nick,
	}
	return p, nil
}

func (p *Player) Game() *Game {
	return GetGame(p.GameID)
}

func (p *Player) CurrentRoom() *Room {
	g := p.Game()
	return g.GetRoom(p.GameID) // REVIEW
}
