package main

import (
	"errors"

	log "github.com/sirupsen/logrus"
)

type Connection struct {
	ch chan *Message

	currentID     uint64
	netConn       NetConnection
	closed        bool
	resultWaiters map[uint64][]chan *Result
}

func MakeConnection(netConn NetConnection) *Connection {
	conn := &Connection{
		ch: make(chan *Message),

		netConn:       netConn,
		currentID:     0,
		closed:        false,
		resultWaiters: make(map[uint64][]chan *Result),
	}

	go func() {
		for {
			msg, ok := <-netConn.Channel()
			if !ok {
				conn.closed = true
				close(conn.ch)
				return
			}

			if id := msg.GetID(); id > conn.currentID {
				conn.currentID = id
			}

			switch msg.GetType() {
			case TResult: // REVIEW
				channels := conn.resultWaiters[msg.GetID()]
				if channels != nil && len(channels) > 0 {
					for _, ch := range channels {
						ch <- msg.GetResult()
					}
					conn.resultWaiters[msg.GetID()] = nil
				}

			case TMessage:
				conn.ch <- msg.GetMessage()

			default: // REVIEW
				panic("unknown type")
			}
		}
	}()

	return conn
}

func (conn *Connection) Chan() <-chan *Message {
	return conn.ch
}

func (conn *Connection) Write(method string, args ...interface{}) error {
	if args == nil {
		args = make([]interface{}, 0)
	}
	conn.currentID++
	return conn.write(Message{
		ID:     conn.currentID,
		Method: method,
		Args:   args,
	})
}

func (conn *Connection) Send(method string, args ...interface{}) (interface{}, error) {
	id := conn.currentID + 1
	ch := make(chan *Result)
	if conn.resultWaiters[id] == nil {
		conn.resultWaiters[id] = []chan *Result{ch}
	} else {
		conn.resultWaiters[id] = append(conn.resultWaiters[id], ch)
	}

	err := conn.Write(method, args...)
	if err != nil {
		return nil, err
	}

	for {
		res := <-ch
		if res.Error != "" {
			err = errors.New(res.Error)
		}
		return res.Result, err
	}
}

func (conn *Connection) Reply(id uint64, err string, res interface{}) error {
	return conn.writeRes(Result{
		ID:     id,
		Error:  err,
		Result: res,
	})
}

func (conn *Connection) write(msg Message) error {
	log.WithFields(log.Fields{
		"data": msg,
	}).Info("sending")

	return conn.netConn.Write(msg)
}

func (conn *Connection) writeRes(res Result) error {
	log.WithFields(log.Fields{
		"data": res,
	}).Info("sending")

	return conn.netConn.WriteRes(res)
}
