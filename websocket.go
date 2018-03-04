package main

import (
	"fmt"

	"golang.org/x/net/websocket"
)

type WebsocketCommunicator struct {
	started      bool
	connectionCh chan *Connection
}

func (comm *WebsocketCommunicator) ConnectionCh() chan *Connection {
	return comm.connectionCh
}

func (comm *WebsocketCommunicator) Started() bool {
	return comm.started
}

func (comm *WebsocketCommunicator) Start() error {
	if comm.started {
		return fmt.Errorf("already started")
	}

	comm.started = true
	// TODO
	return nil
}

func (comm *WebsocketCommunicator) Stop() error {
	if !comm.started {
		return fmt.Errorf("not started")
	}

	comm.started = false
	// TODO
	return nil
}

type WebsocketConnection struct {
	socket *websocket.Conn
}

func (conn *WebsocketConnection) Write(bytes []byte) error {
	n, err := conn.socket.Write(bytes)
	if err != nil {
		return err
	} else if n != len(bytes) {
		return fmt.Errorf("only sent %d bytes out of %d", n, len(bytes))
	}

	return nil
}

func (conn *WebsocketConnection) Close() error {
	return conn.socket.Close()
}
