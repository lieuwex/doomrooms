package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"

	log "github.com/sirupsen/logrus"
)

const delim = '\n'

type TCPJSONCommunicator struct {
	started      bool
	listener     *net.TCPListener
	connectionCh chan NetConnection
}

func MakeTCPJSONCommunicator() *TCPJSONCommunicator {
	return &TCPJSONCommunicator{
		started:      false,
		listener:     nil,
		connectionCh: make(chan NetConnection),
	}
}

func (comm *TCPJSONCommunicator) ConnectionCh() <-chan NetConnection {
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
			if err != nil {
				fmt.Printf("err %s\n", err)
			}

			comm.connectionCh <- makeConnection(socket)
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
	ch     chan Thing
	closed bool
}

func parseBytes(bytes []byte) Thing {
	var m map[string]interface{}
	if json.Unmarshal(bytes, &m) != nil {
		return nil
	}

	if m["method"] != nil {
		var msg Message
		if json.Unmarshal(bytes, &msg) == nil {
			return &msg
		}
	} else {
		var res Result
		if json.Unmarshal(bytes, &res) == nil {
			return &res
		}
	}

	log.WithFields(log.Fields{
		"msg": string(bytes),
	}).Error("invalid message received")

	return nil
}

func makeConnection(socket *net.TCPConn) NetConnection {
	netConn := &TCPConnection{
		socket: socket,
		ch:     make(chan Thing),
		closed: false,
	}
	reader := bufio.NewReader(socket)

	go func() {
		for {
			raw, err := reader.ReadBytes('\n')
			if err != nil {
				if err != io.EOF {
					log.WithFields(log.Fields{
						"error": err,
					}).Error("error while reading from connection")
				}
				netConn.closed = true
				close(netConn.ch)
				return
			}

			msg := parseBytes(raw)
			if msg == nil {
				continue
			}

			netConn.ch <- msg
		}
	}()

	return netConn
}

func (conn *TCPConnection) Write(msg Message) error {
	bytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return conn.write(bytes)
}
func (conn *TCPConnection) WriteRes(res Result) error {
	bytes, err := json.Marshal(res)
	if err != nil {
		return err
	}

	return conn.write(bytes)
}

func (conn *TCPConnection) write(bytes []byte) error {
	bytes = append(bytes, delim)

	n, err := conn.socket.Write(bytes)
	if err != nil {
		return err
	} else if n != len(bytes) {
		return fmt.Errorf("only sent %d bytes out of %d", n, len(bytes))
	}

	return nil
}

func (conn *TCPConnection) Close() error {
	if conn.closed {
		return nil
	}

	conn.closed = true
	return conn.socket.Close()
}

func (conn *TCPConnection) Closed() bool {
	return conn.closed
}

func (conn *TCPConnection) Channel() chan Thing {
	return conn.ch
}
