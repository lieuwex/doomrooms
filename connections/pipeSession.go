package connections

import (
	"context"
	"doomrooms/utils"
	"fmt"
	"strings"
	"sync"

	"log"
)

const bufferSize = 100
const IDLength = 25

var PipeSessions = make([]*PipeSession, 0)

type PipeSession struct {
	PrivateID string

	ctx    context.Context
	cancel context.CancelFunc

	mu sync.Mutex
	a  *Connection
	b  *Connection

	aToB chan []byte
	bToA chan []byte
}

func MakePipeSession() (*PipeSession, error) {
	id, err := utils.GenerateUID(25)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	session := &PipeSession{
		PrivateID: id,

		ctx:    ctx,
		cancel: cancel,

		aToB: make(chan []byte, bufferSize),
		bToA: make(chan []byte, bufferSize),
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

	go func() {
		defer ps.cancel()

		for {
			select {
			case <-ps.ctx.Done():
				return

			case bytes := <-recvCh:
				if err := conn.netConn.WriteRaw(bytes); err != nil {
					fmt.Printf("ps write err %#v\n", err)
					return
				}
			}
		}
	}()

	go func() {
		defer ps.cancel()

		for {
			select {
			case <-ps.ctx.Done():
				return

			case bytes, ok := <-conn.netConn.RawChannel():
				if !ok {
					return
				}

				sendCh <- bytes
			}
		}
	}()

	<-ps.ctx.Done()

	conn.Close()
	ps.removeConnection(conn)

	return nil
}

func (ps *PipeSession) addConnection(conn *Connection) (sendCh chan []byte, recvCh chan []byte, err error) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

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
	ps.mu.Lock()
	defer ps.mu.Unlock()

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
	bytes := <-conn.netConn.RawChannel()
	if conn.closed {
		log.Println("connection closed")
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
