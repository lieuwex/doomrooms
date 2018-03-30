package json

import (
	"doomrooms/types"
	"encoding/json"

	"github.com/mitchellh/mapstructure"
	log "github.com/sirupsen/logrus"
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

	log.WithFields(log.Fields{
		"msg": string(bytes),
	}).Error("invalid message received")

	return nil
}
