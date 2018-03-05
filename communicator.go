package main

import "fmt"

type Communicator interface {
	ConnectionCh() <-chan *Connection
	Started() bool
	Start(host string, port string) error
	Stop() error
}

type NetConnection interface {
	Write(bytes []byte) error
	Close() error
}

type CommunicatorManager struct {
	connCh chan *Connection
}

func MakeCommunicatorManager() *CommunicatorManager {
	return &CommunicatorManager{
		connCh: make(chan *Connection),
	}
}

func (cm *CommunicatorManager) StartService(service string, host string, port string) error {
	// var comm Communicator = nil // TODO
	var err error = nil

	switch service {
	case "gameserver-tcp":
		go ListenGameservers(host, port)

	case "player-tcp":
		// HACK
		tcpComm := MakeTCPCommunicator()
		err = tcpComm.Start(host, port)
		go func() {
			for {
				cm.connCh <- <-tcpComm.ConnectionCh()
			}
		}()
	case "player-wc":
		// HACK
		wsComm := MakeWebsocketCommunicator()
		err = wsComm.Start()
		go func() {
			for {
				cm.connCh <- <-wsComm.ConnectionCh()
			}
		}()

	default:
		err = fmt.Errorf("no service with name '%s' found", service)
	}

	if err != nil {
		return err
	}

	return nil
}

func (cm *CommunicatorManager) ConnectionCh() <-chan *Connection {
	return cm.connCh
}
