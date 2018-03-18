package communicators

import (
	"doomrooms/communicators/json"
	"doomrooms/connections"
	"doomrooms/types"
	"doomrooms/utils"
	"fmt"

	log "github.com/sirupsen/logrus"
)

type CommunicatorManager struct {
	Communicators []types.Communicator

	playerConnCh     chan *connections.Connection
	gameServerConnCh chan *connections.Connection

	log *log.Logger
}

func MakeCommunicatorManager() *CommunicatorManager {
	cm := &CommunicatorManager{
		playerConnCh:     make(chan *connections.Connection),
		gameServerConnCh: make(chan *connections.Connection),
		log:              log.New(),
	}
	cm.log.Formatter = utils.Formatter
	return cm
}

func (cm *CommunicatorManager) StartService(service string, host string, port string, isPlayer bool) error {
	var comm types.Communicator
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

	connCh := cm.gameServerConnCh
	if isPlayer {
		connCh = cm.playerConnCh
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
