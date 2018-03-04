package main

import (
	"fmt"
	"net"
)

type GameServer struct {
	Socket net.Conn
	Game   *Game
}

var GameServers = make([]GameServer, 0)

func ListenGameservers(host string, port string) error {
	comm := MakeTCPCommunicator()
	err := comm.Start(host, port)
	if err != nil {
		panic(err) // TODO
		return err
	}
	defer comm.Stop() // REVIEW

	for {
		connection := <-comm.ConnectionCh()
		go HandleGameServer(connection)
	}

	fmt.Println("doei")
	return nil
}

func HandleGameServer(conn *Connection) {
	defer conn.netConn.Close()

	msg := <-conn.Chan()
	if conn.closed {
		return
	}

	if msg.Method != "hello" {
		// REVIEW
		conn.Reply(msg.ID, "expected 'hello' msg", nil)
		return
	} else if len(msg.Args) != 2 {
		// REVIEW
		conn.Reply(msg.ID, "expected 2 args", nil)
		return
	}

	g := MakeGame(conn, msg.Args[0].(string), msg.Args[1].(string))

	fmt.Printf("made game: %#v\n", g)
	conn.Reply(msg.ID, "", g)

	for {
		msg := <-conn.Chan()
		if conn.closed {
			fmt.Printf("connection closed lol\n")
			break
		}

		// REVIEW
		onGameServerCommand(g, msg)
	}
}

func onGameServerCommand(game *Game, msg Message) {

}
