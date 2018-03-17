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

func (ps *PipeSession) BindConnection(conn *Connection) error {
	isA, sendCh, recvCh, err := ps.addConnection(conn)
	if err != nil {
		return err
	}
	defer ps.removeConnection(conn)

	waitch := make(chan bool)

	go func() {
		for {
			select {
			case <-waitch:
				return

			case bytes := <-recvCh:
				err := conn.netConn.WriteRaw(bytes)
				if err != nil {
					fmt.Printf("ps err %#v\n", err)
					close(waitch)
					return
				}
			}
		}
	}()

	go func() {
		for {
			select {
			case <-waitch:
				return

			case bytes := <-conn.netConn.RawChannel():
				if conn.netConn.Closed() {
					close(waitch)
					return
				}
				sendCh <- bytes
			}
		}
	}()

	<-waitch

	var other *Connection
	if isA {
		other = ps.B
	} else {
		other = ps.A
	}

	if other != nil && !other.Closed() {
		return other.Close()
	}
	return nil
}

func (ps *PipeSession) addConnection(conn *Connection) (isA bool, sendCh chan []byte, recvCh chan []byte, err error) {
	if ps.A == nil {
		ps.A = conn
		return true, ps.AToB, ps.BToA, nil
	} else if ps.B == nil {
		ps.B = conn
		return false, ps.BToA, ps.AToB, nil
	}

	return false, nil, nil, fmt.Errorf("PipeSession is fully loaded")
}

func (ps *PipeSession) removeConnection(conn *Connection) error {
	if ps.A == conn {
		ps.A = nil
	} else if ps.B == conn {
		ps.B = nil
	} else {
		return fmt.Errorf("Connection not bound to PipeSession")
	}

	return nil
}
