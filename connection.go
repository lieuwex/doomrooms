package main

import (
	"encoding/json"
	"log"
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
	return conn.writeres(Result{
		ID:     id,
		Error:  err,
		Result: res,
	})
}

func (conn *Connection) write(msg Message) error {
	bytes, err := json.Marshal(msg)
	log.Printf("sending '%s'\n", string(bytes))
	if err != nil {
		return err
	}
	return conn.netConn.Write(bytes)
}

func (conn *Connection) writeres(res Result) error { // HACK: REVIEW
	bytes, err := json.Marshal(res)
	log.Printf("sending '%s'\n", string(bytes))
	if err != nil {
		return err
	}
	return conn.netConn.Write(bytes)
}
