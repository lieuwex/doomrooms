package config

import (
	"doomrooms/communicators"
	"fmt"
	"io/ioutil"
	"strconv"

	"log"

	"github.com/BurntSushi/toml"
)

type listenerSetting struct {
	Type    string
	Host    string
	Port    uint64
	Timeout uint
}

func HandleService(cm *communicators.CommunicatorManager, service string, settings listenerSetting) error {
	if settings.Port == 0 {
		return fmt.Errorf("port required")
	}
	portStr := strconv.FormatUint(settings.Port, 10)

	err := cm.StartService(service, settings.Host, portStr, settings.Type)
	if err != nil {
		return err
	}

	return nil
}

func ReadConfig(cm *communicators.CommunicatorManager, filename string) error {
	log.Printf("reading config file %s", filename)

	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	var config map[string][]listenerSetting
	if _, err := toml.Decode(string(bytes), &config); err != nil {
		return err
	}

	for service, settings := range config {
		for _, setting := range settings {
			if err := HandleService(cm, service, setting); err != nil {
				// REVIEW
				return err
			}
		}
	}

	return nil
}
