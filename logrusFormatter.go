package main

import (
	log "github.com/sirupsen/logrus"
)

var Formatter *log.TextFormatter

func SetLogrusFormatter() {
	Formatter = new(log.TextFormatter)
	Formatter.TimestampFormat = "2006-01-02 15:04:05"
	log.SetFormatter(Formatter)
	Formatter.FullTimestamp = true
}
