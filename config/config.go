package config

import (
	"doomrooms/communicators"
	"doomrooms/connections"
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/BurntSushi/toml"
	log "github.com/sirupsen/logrus"
)

type listenerSetting struct {
	Host    string
	Port    uint64
	Timeout uint
}

func HandleService(cm *communicators.CommunicatorManager, service string, settings listenerSetting) error {
	if settings.Port == 0 {
		return fmt.Errorf("port required")
	}
	portStr := strconv.FormatUint(settings.Port, 10)

	var err error
	if service == "gameserver-tcp-json" {
		err = connections.ListenGameservers(settings.Host, portStr)
	} else {
		err = cm.StartService(service, settings.Host, portStr)
	}

	if err != nil {
		return err
	}

	cm.Log.WithFields(log.Fields{
		"service": service,
		"port":    settings.Port,
	}).Info("started service")

	return nil
}

func ReadConfig(cm *communicators.CommunicatorManager, filename string) error {
	log.WithFields(log.Fields{
		"path": filename,
	}).Info("reading config file")
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	var config map[string]listenerSetting
	if _, err := toml.Decode(string(bytes), &config); err != nil {
		return err
	}

	for service, settings := range config {
		err = HandleService(cm, service, settings)
		if err != nil {
			// REVIEW
			return err
		}
	}

	for name, comm := range cm.Communicators {
		if !comm.Started() {
			cm.Log.WithFields(log.Fields{
				"service": name,
			}).Warn("service not enabled")
		}
	}

	return nil
}