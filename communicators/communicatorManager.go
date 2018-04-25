package communicators

import (
	"doomrooms/communicators/json"
	"doomrooms/connections"
	"doomrooms/utils"
	"fmt"

	log "github.com/sirupsen/logrus"
)

type CommunicatorManager struct {
	Communicators []Communicator

	playerConnCh      chan *connections.Connection
	gameServerConnCh  chan *connections.Connection
	pipeSessionConnCh chan *connections.Connection

	log *log.Logger
}

func MakeCommunicatorManager() *CommunicatorManager {
	cm := &CommunicatorManager{
		playerConnCh:      make(chan *connections.Connection),
		gameServerConnCh:  make(chan *connections.Connection),
		pipeSessionConnCh: make(chan *connections.Connection),
		log:               log.New(),
	}
	cm.log.Formatter = utils.Formatter
	return cm
}

func (cm *CommunicatorManager) StartService(service string, host string, port string, typ string) error {
	var comm Communicator
	switch service {
	case "tcp-json":
		comm = json.MakeTCPJSONCommunicator()
	case "ws-json":
		comm = json.MakeWebsocketJSONCommunicator()
	default:
		return fmt.Errorf("no service with name '%s' found", service)
	}

	cm.Communicators = append(cm.Communicators, comm)

	err := comm.Start(host, port)
	if err != nil {
		return err
	}

	var connCh chan *connections.Connection
	switch typ {
	case "player":
		connCh = cm.playerConnCh
	case "gameserver":
		connCh = cm.gameServerConnCh
	case "pipesession":
		connCh = cm.pipeSessionConnCh

	default:
		return fmt.Errorf("unknown type '%s'", typ)
	}

	go func() {
		for {
			netConn := <-comm.ConnectionCh()
			connCh <- connections.MakeConnection(netConn)
		}
	}()

	cm.log.WithFields(log.Fields{
		"service": service,
		"host":    host,
		"port":    port,
		"type":    typ,
	}).Info("started service")

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

func (cm *CommunicatorManager) PlayerConnectionCh() <-chan *connections.Connection {
	return cm.playerConnCh
}
func (cm *CommunicatorManager) GameServerConnectionCh() <-chan *connections.Connection {
	return cm.gameServerConnCh
}
func (cm *CommunicatorManager) PipeSessionConnectionCh() <-chan *connections.Connection {
	return cm.pipeSessionConnCh
}
