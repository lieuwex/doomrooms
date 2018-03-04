package main

type Communicator interface {
	ConnectionCh() chan *Connection
	Started() bool
	Start(port string) error
	Stop() error
}

type NetConnection interface {
	Write(bytes []byte) error
	Close() error
}
