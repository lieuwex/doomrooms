package types

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
