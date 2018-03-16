package main

import (
	"fmt"
	"net/http"

	"golang.org/x/net/websocket"
)

type WebsocketJSONCommunicator struct {
	started      bool
	connectionCh chan NetConnection
	server       *http.Server
	openCh       chan bool
}

func MakeWebsocketJSONCommunicator() *WebsocketJSONCommunicator {
	return &WebsocketJSONCommunicator{
		started:      false,
		connectionCh: make(chan NetConnection),
		openCh:       make(chan bool),
	}
}

func (comm *WebsocketJSONCommunicator) ConnectionCh() <-chan NetConnection {
	return comm.connectionCh
}

func (comm *WebsocketJSONCommunicator) Started() bool {
	return comm.started
}

func (comm *WebsocketJSONCommunicator) Start(host string, port string) error {
	if comm.started {
		return fmt.Errorf("already started")
	}

	comm.server = &http.Server{
		Addr: host + ":" + port,
		Handler: websocket.Handler(func(ws *websocket.Conn) {
			comm.connectionCh <- makeWsConnection(ws)
			<-comm.openCh
		}),
	}

	go comm.server.ListenAndServe()
	comm.started = true
	return nil
}

func (comm *WebsocketJSONCommunicator) Stop() error {
	if !comm.started {
		return fmt.Errorf("not started")
	}

	close(comm.openCh)
	comm.server.Close()
	comm.started = false
	return nil
}

type WebsocketConnection struct {
	socket *websocket.Conn
	ch     chan Thing
	closed bool
}

func makeWsConnection(ws *websocket.Conn) NetConnection {
	netConn := &WebsocketConnection{
		socket: ws,
		ch:     make(chan Thing),
		closed: false,
	}

	go func() {
		for {
			var bytes []byte
			err := websocket.Message.Receive(ws, &bytes)
			if err != nil {
				// TODO

				netConn.closed = true
				close(netConn.ch)
				return
			}

			msg := parseBytes(bytes) // HACK: we should define this function ourselves.
			if msg == nil {
				continue
			}

			netConn.ch <- msg
		}
	}()

	return netConn
}

func (conn *WebsocketConnection) Write(msg Message) error {
	return websocket.JSON.Send(conn.socket, msg)
}
func (conn *WebsocketConnection) WriteRes(res Result) error {
	return websocket.JSON.Send(conn.socket, res)
}

func (conn *WebsocketConnection) Close() error {
	return conn.socket.Close()
}

func (conn *WebsocketConnection) Closed() bool {
	return conn.closed
}

func (conn *WebsocketConnection) Channel() chan Thing {
	return conn.ch
}
