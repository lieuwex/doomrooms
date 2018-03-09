package main

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
