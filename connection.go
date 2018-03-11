package main

import (
	"errors"

	log "github.com/sirupsen/logrus"
)

type Connection struct {
	ch chan Thing

	currentID     uint64
	netConn       NetConnection
	closed        bool
	resultWaiters map[uint64][]chan *Result
}

func (conn *Connection) Chan() <-chan Thing {
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
