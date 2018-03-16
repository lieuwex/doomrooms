package types

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
