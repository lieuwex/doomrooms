package json

import (
	"bufio"
	"doomrooms/types"
	"encoding/json"
	"fmt"
	"io"
	"net"

	"log"
)

const delim = '\n'

type TCPJSONCommunicator struct {
	started      bool
	listener     *net.TCPListener
	connectionCh chan types.NetConnection
}

func MakeTCPJSONCommunicator() *TCPJSONCommunicator {
	return &TCPJSONCommunicator{
		started:      false,
		listener:     nil,
		connectionCh: make(chan types.NetConnection),
	}
}

func (comm *TCPJSONCommunicator) ConnectionCh() <-chan types.NetConnection {
	return comm.connectionCh
}

func (comm *TCPJSONCommunicator) Started() bool {
	return comm.started
}

func (comm *TCPJSONCommunicator) Start(host string, port string) error {
	if comm.started {
		return fmt.Errorf("already started")
	}

	addr, err := net.ResolveTCPAddr("tcp", ":"+port)
	if err != nil {
		return err
	}

	comm.listener, err = net.ListenTCP("tcp", addr)
	if err != nil {
		return err
	}

	comm.started = true
	go func() {
		for {
			socket, err := comm.listener.AcceptTCP()
			if !comm.started {
				return
			} else if err != nil {
				fmt.Printf("err %s\n", err)
			}

			comm.connectionCh <- makeTCPConnection(socket)
		}
	}()

	return nil
}

func (comm *TCPJSONCommunicator) Stop() error {
	if !comm.started {
		return fmt.Errorf("not started")
	}

	comm.started = false
	comm.listener.Close()

	return nil
}

type TCPConnection struct {
	socket net.Conn
	ch     chan types.Thing
	rawCh  chan []byte
	closed bool
}

func makeTCPConnection(socket *net.TCPConn) types.NetConnection {
	netConn := &TCPConnection{
		socket: socket,
		ch:     make(chan types.Thing),
		rawCh:  make(chan []byte),
		closed: false,
	}
	reader := bufio.NewReader(socket)

	go func() {
		for {
			raw, err := reader.ReadBytes('\n')
			if err != nil {
				if err != io.EOF {
					log.Printf("error reading from connection: %s", err)
				}
				break
			}

			select {
			case netConn.rawCh <- raw:
				// nop
			default:
				if msg := parseBytes(raw); msg != nil {
					netConn.ch <- msg
				}
			}
		}

		netConn.closed = true
		close(netConn.ch)
		close(netConn.rawCh)
	}()

	return netConn
}

func (conn *TCPConnection) Write(msg types.Message) error {
	bytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return conn.write(bytes)
}
func (conn *TCPConnection) WriteRes(res types.Result) error {
	bytes, err := json.Marshal(res)
	if err != nil {
		return err
	}

	return conn.write(bytes)
}

func (conn *TCPConnection) write(bytes []byte) error {
	bytes = append(bytes, delim)
	return conn.WriteRaw(bytes)
}

func (conn *TCPConnection) WriteRaw(bytes []byte) error {
	n, err := conn.socket.Write(bytes)
	if err != nil {
		return err
	} else if n != len(bytes) {
		return fmt.Errorf("only sent %d bytes out of %d", n, len(bytes))
	}

	return nil
}

func (conn *TCPConnection) Close() error {
	// REVIEW: how does this compare to makeTCPConnection?

	if conn.closed {
		return nil
	}

	conn.closed = true
	return conn.socket.Close()
}

func (conn *TCPConnection) Closed() bool {
	return conn.closed
}

func (conn *TCPConnection) Channel() <-chan types.Thing {
	return conn.ch
}

func (conn *TCPConnection) RawChannel() <-chan []byte {
	return conn.rawCh
}
