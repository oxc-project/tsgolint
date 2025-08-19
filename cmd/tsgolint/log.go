package main

import "os"

type LogLevel uint8

const (
	LogLevelNormal LogLevel = iota
	LogLevelDebug
)

func getLogLevel() LogLevel {
	logLevel := os.Getenv("OXC_LOG")

	if logLevel == "debug" {
		return LogLevelDebug
	}

	return LogLevelNormal

}
