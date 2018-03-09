package main

import (
	log "github.com/sirupsen/logrus"
)

type Connection struct {
	ch chan Message

	currentID uint64
	netConn   NetConnection
	closed    bool
}

// var connections

func (conn *Connection) Chan() <-chan Message {
	return conn.ch
}

func (conn *Connection) Send(method string, args ...interface{}) error {
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
