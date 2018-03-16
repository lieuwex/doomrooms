package doomrooms

import (
	"doomrooms/communicators/json"
	"doomrooms/types"
	"fmt"

	"github.com/sirupsen/logrus"
)

type CommunicatorManager struct {
	connCh        chan *Connection
	communicators map[string]types.Communicator
	log           *logrus.Logger
}

func MakeCommunicatorManager() *CommunicatorManager {
	cm := &CommunicatorManager{
		connCh: make(chan *Connection),
		communicators: map[string]types.Communicator{
			"player-tcp-json": json.MakeTCPJSONCommunicator(),
			"player-ws-json":  json.MakeWebsocketJSONCommunicator(),
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

func (cm *CommunicatorManager) StopServices() error {
	for _, comm := range cm.communicators {
		if err := comm.Stop(); err != nil {
			// REVIEW
			return err
		}
	}
	return nil
}

func (cm *CommunicatorManager) ConnectionCh() <-chan *Connection {
	return cm.connCh
}
