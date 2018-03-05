package main

import (
	"io/ioutil"
	"strings"

	log "github.com/sirupsen/logrus"
)

func HandleService(cm *CommunicatorManager, service string, port string, host string) error {
	if service == "gameserver-tcp" {
		go ListenGameservers(host, port)
	} else {
		err := cm.StartService(service, host, port)
		if err != nil {
			return err
		}
	}

	cm.log.WithFields(log.Fields{
		"service": service,
		"port":    port,
	}).Info("started service")

	return nil
}

// REVIEW: warn the user or something when shit doesn't have a port?
func ReadConfig(cm *CommunicatorManager, filename string) error {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	lines := strings.Split(string(bytes), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || // empty
			line[0] == '#' { // comment
			continue
		}

		words := strings.Split(line, " ")
		service := words[0]
		port := words[1]
		host := ""
		if len(words) == 3 {
			host = words[3]
		}

		err = HandleService(cm, service, port, host)
		if err != nil {
			// REVIEW
			return err
		}
	}

	for name, comm := range cm.communicators {
		if !comm.Started() {
			cm.log.WithFields(log.Fields{
				"service": name,
			}).Warn("service not enabled")
		}
	}

	return nil
}
