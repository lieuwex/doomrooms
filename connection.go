package doomrooms

import (
	"doomrooms/types"
	"errors"

	log "github.com/sirupsen/logrus"
)

type Connection struct {
	ch chan *types.Message

	currentID     uint64
	netConn       types.NetConnection
	closed        bool
	resultWaiters map[uint64][]chan *types.Result
}

func MakeConnection(netConn types.NetConnection) *Connection {
	conn := &Connection{
		ch: make(chan *types.Message),

		netConn:       netConn,
		currentID:     0,
		closed:        false,
		resultWaiters: make(map[uint64][]chan *types.Result),
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
			case types.TResult: // REVIEW
				channels := conn.resultWaiters[msg.GetID()]
				if channels != nil && len(channels) > 0 {
					for _, ch := range channels {
						ch <- msg.GetResult()
					}
					conn.resultWaiters[msg.GetID()] = nil
				}

			case types.TMessage:
				conn.ch <- msg.GetMessage()

			default: // REVIEW
				panic("unknown type")
			}
		}
	}()

	return conn
}

func (conn *Connection) Chan() <-chan *types.Message {
	return conn.ch
}

func (conn *Connection) Write(method string, args ...interface{}) error {
	if args == nil {
		args = make([]interface{}, 0)
	}
	conn.currentID++
	return conn.write(types.Message{
		ID:     conn.currentID,
		Method: method,
		Args:   args,
	})
}

func (conn *Connection) Send(method string, args ...interface{}) (interface{}, error) {
	id := conn.currentID + 1
	ch := make(chan *types.Result)
	if conn.resultWaiters[id] == nil {
		conn.resultWaiters[id] = []chan *types.Result{ch}
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
	return conn.writeRes(types.Result{
		ID:     id,
		Error:  err,
		Result: res,
	})
}

func (conn *Connection) write(msg types.Message) error {
	log.WithFields(log.Fields{
		"data": msg,
	}).Info("sending")

	return conn.netConn.Write(msg)
}

func (conn *Connection) writeRes(res types.Result) error {
	log.WithFields(log.Fields{
		"data": res,
	}).Info("sending")

	return conn.netConn.WriteRes(res)
}
