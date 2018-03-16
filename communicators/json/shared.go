package json

import (
	"doomrooms/types"
	"encoding/json"

	log "github.com/sirupsen/logrus"
)

func parseBytes(bytes []byte) types.Thing {
	var m map[string]interface{}
	if json.Unmarshal(bytes, &m) != nil {
		return nil
	}

	if m["method"] != nil {
		var msg types.Message
		if json.Unmarshal(bytes, &msg) == nil {
			return &msg
		}
	} else {
		var res types.Result
		if json.Unmarshal(bytes, &res) == nil {
			return &res
		}
	}

	log.WithFields(log.Fields{
		"msg": string(bytes),
	}).Error("invalid message received")

	return nil
}
