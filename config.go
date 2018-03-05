package main

import (
	"io/ioutil"
	"log"
	"strings"
)

func HandleService(cm *CommunicatorManager, service string, port string, host string) error {
	err := cm.StartService(service, host, port)
	if err != nil {
		return err
	}

	log.Printf("enabled service %s on port %s", service, port)

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
			return err
		}
	}

	return nil
}
