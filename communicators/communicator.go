package communicators

import (
	"doomrooms/communicators/json"
	"doomrooms/connections"
	"doomrooms/types"
	"doomrooms/utils"
	"fmt"

	"github.com/sirupsen/logrus"
)

type CommunicatorManager struct {
	Log           *logrus.Logger
	Communicators map[string]types.Communicator

	connCh chan *connections.Connection
}

func MakeCommunicatorManager() *CommunicatorManager {
	cm := &CommunicatorManager{
		connCh: make(chan *connections.Connection),
		Communicators: map[string]types.Communicator{
			"player-tcp-json": json.MakeTCPJSONCommunicator(),
			"player-ws-json":  json.MakeWebsocketJSONCommunicator(),
		},
		Log: logrus.New(),
	}
	cm.Log.Formatter = utils.Formatter
	return cm
}

func (cm *CommunicatorManager) StartService(service string, host string, port string) error {
	comm := cm.Communicators[service]
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
			cm.connCh <- connections.MakeConnection(netConn)
		}
	}()

	return nil
}

func (cm *CommunicatorManager) StopServices() error {
	for _, comm := range cm.Communicators {
		if err := comm.Stop(); err != nil {
			// REVIEW
			return err
		}
	}
	return nil
}

func (cm *CommunicatorManager) ConnectionCh() <-chan *connections.Connection {
	return cm.connCh
}
