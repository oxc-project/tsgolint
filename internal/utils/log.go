package utils

import (
	"os"
	"sync"
)

type LogLevel uint8

const (
	LogLevelNormal LogLevel = iota
	LogLevelDebug
)

var (
	logLevel     LogLevel
	logLevelOnce sync.Once
)

func GetLogLevel() LogLevel {
	logLevelOnce.Do(func() {
		switch os.Getenv("OXC_LOG") {
		case "debug":
			logLevel = LogLevelDebug
		default:
			logLevel = LogLevelNormal
		}
	})

	return logLevel
}
