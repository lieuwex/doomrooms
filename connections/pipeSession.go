package connections

import (
	"doomrooms/utils"
	"fmt"
	"math/rand"
)

const bufferSize = 100

var PipeSessions = make([]*PipeSession, 0)

type PipeSession struct {
	A *Connection
	B *Connection

	PrivateID string

	AToB chan []byte
	BToA chan []byte
}

func MakePipeSession() *PipeSession {
	session := &PipeSession{
		// TODO: make this cryptographically secure
		PrivateID: utils.FormatID(rand.Uint64()),

		AToB: make(chan []byte, bufferSize),
		BToA: make(chan []byte, bufferSize),
	}

	PipeSessions = append(PipeSessions, session)
	return session
}

func (ps *PipeSession) AddConnection(conn *Connection) error {
	var recvCh chan []byte
	var sendCh chan []byte

	if ps.A == nil {
		ps.A = conn

		sendCh = ps.AToB
		recvCh = ps.BToA
	} else if ps.B == nil {
		ps.B = conn

		sendCh = ps.BToA
		recvCh = ps.AToB
	} else {
		return fmt.Errorf("PipeSession is fully loaded")
	}
	waitch := make(chan bool)

	go func() {
		for {
			bytes := <-recvCh
			err := conn.netConn.WriteRaw(bytes)
			if err != nil {
				// TODO
			}
		}
	}()

	go func() {
		for {
			bytes := <-conn.netConn.RawChannel()
			sendCh <- bytes
		}
	}()

	<-waitch
	return nil
}
