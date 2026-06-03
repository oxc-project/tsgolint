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

var GetLogLevel = sync.OnceValue(func() LogLevel {
	switch os.Getenv("OXC_LOG") {
	case "debug":
		return LogLevelDebug
	default:
		return LogLevelNormal
	}
})
