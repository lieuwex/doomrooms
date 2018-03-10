package main

type ThingType int

const (
	TMessage ThingType = iota
	TResult
)

type Message struct {
	ID     uint64        `json:"id"`
	Method string        `json:"method"`
	Args   []interface{} `json:"args"`
}

func CreateMessage(id uint64, method string, args ...interface{}) Message {
	return Message{
		ID:     id,
		Method: method,
		Args:   args,
	}
}
func (msg *Message) GetID() uint64        { return msg.ID }
func (msg *Message) GetType() ThingType   { return TMessage }
func (msg *Message) GetMessage() *Message { return msg }
func (msg *Message) GetResult() *Result   { return nil }

type Result struct {
	ID     uint64      `json:"id"`
	Error  string      `json:"err,omitempty"`
	Result interface{} `json:"res,omitempty"`
}

func CreateResult(id uint64, err string, res interface{}) Result {
	return Result{
		ID:     id,
		Error:  err,
		Result: res,
	}
}
func (res *Result) GetID() uint64        { return res.ID }
func (res *Result) GetType() ThingType   { return TResult }
func (res *Result) GetMessage() *Message { return nil }
func (res *Result) GetResult() *Result   { return res }

type Thing interface {
	GetID() uint64
	GetType() ThingType
	GetMessage() *Message
	GetResult() *Result
}
