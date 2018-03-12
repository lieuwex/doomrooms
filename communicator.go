package main

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

type Communicator interface {
	ConnectionCh() <-chan NetConnection
	Started() bool
	Start(host string, port string) error
	Stop() error
}

type NetConnection interface {
	Write(msg Message) error
	WriteRes(res Result) error
	Channel() chan Thing
	Close() error
	Closed() bool
}

type CommunicatorManager struct {
	connCh        chan *Connection
	communicators map[string]Communicator
	log           *logrus.Logger
}

func MakeCommunicatorManager() *CommunicatorManager {
	cm := &CommunicatorManager{
		connCh: make(chan *Connection),
		communicators: map[string]Communicator{
			"player-tcp-json": MakeTCPJSONCommunicator(),
			"player-ws-json":  MakeWebsocketJSONCommunicator(),
		},
		log: logrus.New(),
	}
	cm.log.Formatter = Formatter
	return cm
}

func (cm *CommunicatorManager) StartService(service string, host string, port string) error {
	comm := cm.communicators[service]
	if comm == nil {
		return fmt.Errorf("no service with name '%s' found", service)
	}

	err := comm.Start(host, port)
	if err != nil {
		return err
	}

	go func() {
		for {
			netConn := <-comm.ConnectionCh()
			cm.connCh <- MakeConnection(netConn)
		}
	}()

	return nil
}

func (cm *CommunicatorManager) ConnectionCh() <-chan *Connection {
	return cm.connCh
}
