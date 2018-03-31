package connections

import (
	"doomrooms/utils"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
)

const bufferSize = 100
const IDLength = 25

var PipeSessions = make([]*PipeSession, 0)

type PipeSession struct {
	PrivateID string

	a *Connection
	b *Connection

	aToB chan []byte
	bToA chan []byte

	waitch chan bool
}

func MakePipeSession() (*PipeSession, error) {
	id, err := utils.GenerateUID(25)
	if err != nil {
		return nil, err
	}

	session := &PipeSession{
		PrivateID: id,

		aToB: make(chan []byte, bufferSize),
		bToA: make(chan []byte, bufferSize),

		waitch: make(chan bool),
	}

	PipeSessions = append(PipeSessions, session)
	return session, nil
}

func removePipeSession(ps *PipeSession) {
	for i, x := range PipeSessions {
		if x == ps {
			PipeSessions = append(PipeSessions[:i], PipeSessions[i+1:]...)
			return
		}
	}
}

func (ps *PipeSession) BindConnection(conn *Connection) error {
	sendCh, recvCh, err := ps.addConnection(conn)
	if err != nil {
		return err
	}
	defer ps.removeConnection(conn)

	go func() {
		for {
			select {
			case <-ps.waitch:
				return

			case bytes := <-recvCh:
				err := conn.netConn.WriteRaw(bytes)
				if err != nil {
					fmt.Printf("ps err %#v\n", err)
					close(ps.waitch)
					return
				}
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ps.waitch:
				return

			case bytes := <-conn.netConn.RawChannel():
				if conn.netConn.Closed() {
					close(ps.waitch)
					return
				}
				sendCh <- bytes
			}
		}
	}()

	<-ps.waitch

	conn.Close()
	return nil
}

func (ps *PipeSession) addConnection(conn *Connection) (sendCh chan []byte, recvCh chan []byte, err error) {
	if ps.a == nil {
		ps.a = conn
		return ps.aToB, ps.bToA, nil
	} else if ps.b == nil {
		ps.b = conn
		return ps.bToA, ps.aToB, nil
	}

	return nil, nil, fmt.Errorf("PipeSession is fully loaded")
}

func (ps *PipeSession) removeConnection(conn *Connection) error {
	if ps.a == conn {
		ps.a = nil
	} else if ps.b == conn {
		ps.b = nil
	} else {
		return fmt.Errorf("Connection not bound to PipeSession")
	}

	if ps.a == nil && ps.b == nil {
		removePipeSession(ps)
	}

	return nil
}

func HandlePipeSesionConnection(conn *Connection) {
	defer conn.Close()

	bytes := <-conn.netConn.RawChannel()
	if conn.closed {
		log.Info("connection closed")
		return
	}

	privateID := strings.TrimSpace(string(bytes))

	var ps *PipeSession
	for _, x := range PipeSessions {
		if x.PrivateID == privateID {
			ps = x
		}
	}
	if ps == nil {
		return
	}

	err := ps.BindConnection(conn)
	if err != nil {
		// TODO
		fmt.Printf("PIPE ERROR: %#v\n", err)
	}
}
