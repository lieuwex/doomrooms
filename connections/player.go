package connections

import (
	"errors"
	"fmt"
	"strings"

	"log"

	"github.com/kmanley/golang-tuple"
)

var Players = make(map[string]*Player)

func GetPlayer(nick string) *Player {
	return Players[nick]
}

func HandlePlayerConnection(conn *Connection) {
	defer conn.Close()

	msg := <-conn.Chan()
	if conn.closed {
		log.Println("connection closed")
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

		username, ok := msg.Args[0].(string)
		if !ok {
			err = errors.New("invalid-type")
			break
		}
		password, ok := msg.Args[1].(string)
		if !ok {
			err = errors.New("invalid-type")
			break
		}

		p = GetPlayer(username)
		if p == nil {
			err = errors.New("User not found")
		} else if p.password != password {
			p = nil
			err = errors.New("Invalid password")
		}
	case "make-player":
		if !expectArgs(2) {
			return
		}

		username, ok := msg.Args[0].(string)
		if !ok {
			err = errors.New("invalid-type")
			break
		}
		password, ok := msg.Args[1].(string)
		if !ok {
			err = errors.New("invalid-type")
			break
		}

		p, err = MakePlayer(username, password)

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

	log.Printf("got player %#v", p)
	conn.Reply(msg.ID, "", p)

	for {
		msg := <-conn.Chan()
		if conn.closed {
			log.Println("connection closed")
			break
		}

		// REVIEW: do we want this to be a goroutine?
		go onPlayerCommand(p, conn, msg)
	}
}

type Player struct {
	Nickname string                 `json:"nick"`
	Tags     map[string]interface{} `json:"tags"`

	CurrentGameID string `json:"currentGameId"`
	CurrentRoomID string `json:"currentRoomID"`

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
		return nil, errors.New("nickname already in use")
	}

	p := &Player{
		Nickname: nick,
		password: password,

		Tags: make(map[string]interface{}),

		connections: make([]*Connection, 0),
		privateTags: make(map[string]map[string]interface{}),
	}
	Players[nick] = p
	return p, nil
}

func (p *Player) Game() *Game {
	return GetGame(p.CurrentGameID)
}
func (p *Player) SetGame(game *Game) {
	p.CurrentGameID = game.ID
}

func (p *Player) CurrentRoom() *Room {
	return p.Game().GetRoom(p.CurrentRoomID)
}

func (p *Player) Send(method string, args ...interface{}) (interface{}, error) {
	nconn := len(p.connections)
	ch := make(chan *tuple.Tuple, nconn)

	for _, connection := range p.connections {
		go func(conn *Connection) {
			res, err := conn.Send(method, args...)
			ch <- tuple.NewTupleFromItems(res, err)
		}(connection)
	}

	// REVIEW
	first := <-ch
	close(ch)

	res := first.Get(0)
	var err error
	if rawErr := first.Get(1); rawErr != nil {
		err = rawErr.(error)
	}

	return res, err
}

func (p *Player) Emit(event string, args ...interface{}) error {
	args = append([]interface{}{event}, args...)

	for _, conn := range p.connections {
		err := conn.Write("emit", args...)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Player) addConnection(conn *Connection) error {
	for _, x := range p.connections {
		if x == conn {
			return errors.New("connection already added")
		}
	}

	p.connections = append(p.connections, conn) // REVIEW
	return nil
}

func (p *Player) removeConnection(conn *Connection) error {
	index := -1
	for i, x := range p.connections {
		if x == conn {
			index = i
		}
	}
	if index == -1 {
		return errors.New("no matching connection found")
	}

	p.connections[index] = p.connections[len(p.connections)-1]
	p.connections = p.connections[:len(p.connections)-1]

	return nil
}

func (p *Player) TagsMatch(tags map[string]interface{}) bool {
	for key, val := range tags {
		if p.Tags[key] != val {
			return false
		}
	}
	return true
}
