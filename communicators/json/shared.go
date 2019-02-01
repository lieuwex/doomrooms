package json

import (
	"doomrooms/types"
	"encoding/json"

	"log"

	"github.com/mitchellh/mapstructure"
)

func parseBytes(bytes []byte) types.Thing {
	var m map[string]interface{}
	if json.Unmarshal(bytes, &m) != nil {
		return nil
	}

	if m["method"] != nil {
		var msg types.Message
		if mapstructure.Decode(m, &msg) == nil {
			return &msg
		}
	} else {
		var res types.Result
		if mapstructure.Decode(m, &res) == nil {
			return &res
		}
	}

	log.Printf("invalid message received: %#v", bytes)

	return nil
}
