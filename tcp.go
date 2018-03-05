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

type TCPCommunicator struct {
	started      bool
	listener     *net.TCPListener
	connectionCh chan *Connection
}

func MakeTCPCommunicator() *TCPCommunicator {
	return &TCPCommunicator{
		started:      false,
		listener:     nil,
		connectionCh: make(chan *Connection),
	}
}

func (comm *TCPCommunicator) ConnectionCh() <-chan *Connection {
	return comm.connectionCh
}

func (comm *TCPCommunicator) Started() bool {
	return comm.started
}

func (comm *TCPCommunicator) Start(host string, port string) error {
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

func (comm *TCPCommunicator) Stop() error {
	if !comm.started {
		return fmt.Errorf("not started")
	}

	comm.started = false
	comm.listener.Close()

	return nil
}

type TCPConnection struct {
	socket net.Conn
}

func makeConnection(socket *net.TCPConn) *Connection {
	conn := &Connection{
		ch: make(chan Message),

		netConn:   &TCPConnection{socket},
		currentID: 0,
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
				conn.closed = true // REVIEW
				close(conn.ch)
				return
			}

			var msg Message
			if json.Unmarshal(raw, &msg) != nil {
				log.WithFields(log.Fields{
					"msg": raw,
				}).Error("invalid message received")
				continue
			}
			if msg.ID > conn.currentID {
				conn.currentID = msg.ID
			}
			conn.ch <- msg
		}
	}()

	return conn
}

func (conn *TCPConnection) Write(bytes []byte) error {
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
	return conn.socket.Close()
}
