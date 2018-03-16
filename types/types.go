package types

type ThingType int

const (
	TMessage ThingType = iota
	TResult
)

type Thing interface {
	GetID() uint64
	GetType() ThingType
	GetMessage() *Message
	GetResult() *Result
}

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
