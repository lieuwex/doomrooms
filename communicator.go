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
	connCh        chan *Connection
	communicators map[string]Communicator
}

func MakeCommunicatorManager() *CommunicatorManager {
	return &CommunicatorManager{
		connCh: make(chan *Connection),
		communicators: map[string]Communicator{
			"player-tcp": MakeTCPCommunicator(),
			"player-wc":  MakeWebsocketCommunicator(),
		},
	}
}

func (cm *CommunicatorManager) StartService(service string, host string, port string) error {
	// var comm Communicator = nil // TODO

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
			cm.connCh <- <-comm.ConnectionCh()
		}
	}()

	return nil
}

func (cm *CommunicatorManager) ConnectionCh() <-chan *Connection {
	return cm.connCh
}
