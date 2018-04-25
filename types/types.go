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

type NetConnection interface {
	Write(msg Message) error
	WriteRes(res Result) error
	WriteRaw(bytes []byte) error
	Channel() <-chan Thing
	RawChannel() <-chan []byte
	Close() error
	Closed() bool
}
