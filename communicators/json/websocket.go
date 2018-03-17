package json

import (
	"doomrooms/types"
	"fmt"
	"net/http"

	"golang.org/x/net/websocket"
)

type WebsocketJSONCommunicator struct {
	started      bool
	connectionCh chan types.NetConnection
	server       *http.Server
	openCh       chan bool
}

func MakeWebsocketJSONCommunicator() *WebsocketJSONCommunicator {
	return &WebsocketJSONCommunicator{
		started:      false,
		connectionCh: make(chan types.NetConnection),
		openCh:       make(chan bool),
	}
}

func (comm *WebsocketJSONCommunicator) ConnectionCh() <-chan types.NetConnection {
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
	ch     chan types.Thing
	rawCh  chan []byte
	closed bool
}

func makeWsConnection(ws *websocket.Conn) types.NetConnection {
	netConn := &WebsocketConnection{
		socket: ws,
		ch:     make(chan types.Thing),
		rawCh:  make(chan []byte),
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

			msg := parseBytes(bytes)
			if msg == nil {
				continue
			}

			netConn.ch <- msg
		}
	}()

	return netConn
}

func (conn *WebsocketConnection) Write(msg types.Message) error {
	return websocket.JSON.Send(conn.socket, msg)
}
func (conn *WebsocketConnection) WriteRes(res types.Result) error {
	return websocket.JSON.Send(conn.socket, res)
}
func (conn *WebsocketConnection) WriteRaw(bytes []byte) error {
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

func (conn *WebsocketConnection) Closed() bool {
	return conn.closed
}

func (conn *WebsocketConnection) Channel() <-chan types.Thing {
	return conn.ch
}

func (conn *WebsocketConnection) RawChannel() <-chan []byte {
	return conn.rawCh
}
