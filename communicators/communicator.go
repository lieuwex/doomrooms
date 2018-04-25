package communicators

import "doomrooms/types"

type Communicator interface {
	ConnectionCh() <-chan types.NetConnection
	Started() bool
	Start(host string, port string) error
	Stop() error
}
