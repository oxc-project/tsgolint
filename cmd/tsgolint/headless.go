package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/typescript-eslint/tsgolint/internal/headless"
	"github.com/typescript-eslint/tsgolint/internal/rules"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

// framing of diagnostics & errors now handled by internal/headless

func runHeadless(args []string) int {
	logLevel := utils.GetLogLevel()

	var (
		traceOut   string
		cpuprofOut string
		heapOut    string
		allocsOut  string
	)
	flag.StringVar(&traceOut, "trace", "", "file to put trace to")
	flag.StringVar(&cpuprofOut, "cpuprof", "", "file to put cpu profiling to")
	flag.StringVar(&heapOut, "heap", "", "file to put heap profiling to")
	flag.StringVar(&allocsOut, "allocs", "", "file to put allocs profiling to")
	flag.CommandLine.Parse(args)

	log.SetOutput(os.Stderr)
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	if logLevel == utils.LogLevelDebug {
		log.Printf("Starting tsgolint")
	}

	if done, err := recordTrace(traceOut); err != nil {
		os.Stderr.WriteString(err.Error())
		return 1
	} else {
		defer done()
	}
	if done, err := recordCpuprof(cpuprofOut); err != nil {
		os.Stderr.WriteString(err.Error())
		return 1
	} else {
		defer done()
	}

	cwd, err := os.Getwd(); if err != nil { fmt.Fprintf(os.Stderr, "error getting cwd: %v", err); return 1 }
	configRaw, err := io.ReadAll(os.Stdin); if err != nil { fmt.Fprintf(os.Stderr, "error reading stdin: %v", err); return 1 }
	exitCode := headless.Run(configRaw, rules.AllRulesByName, cwd, logLevel, os.Stdout)
	writeMemProfiles(heapOut, allocsOut)
	return exitCode
}
