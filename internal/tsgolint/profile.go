package tsgolint

import (
	"fmt"
	"os"
	"runtime/pprof"
	"runtime/trace"
)

func RecordTrace(traceOut string) (func(), error) {
	if traceOut != "" {
		f, err := os.Create(traceOut)
		if err != nil {
			return nil, fmt.Errorf("error creating trace file: %w", err)
		}
		trace.Start(f)
		return func() {
			trace.Stop()
			f.Close()
		}, nil
	}
	return func() {}, nil
}
func RecordCpuprof(cpuprofOut string) (func(), error) {
	if cpuprofOut != "" {
		f, err := os.Create(cpuprofOut)
		if err != nil {
			return nil, fmt.Errorf("error creating cpuprof file: %w", err)
		}
		err = pprof.StartCPUProfile(f)
		if err != nil {
			return nil, fmt.Errorf("error starting cpu profiling: %w", err)
		}
		return func() {
			pprof.StopCPUProfile()
			f.Close()
		}, nil
	}
	return func() {}, nil
}

func WriteMemProfiles(heapOut string, allocsOut string) {
	if heapOut != "" {
		if f, err := os.Create(heapOut); err == nil {
			_ = pprof.WriteHeapProfile(f)
			_ = f.Close()
		}
	}

	if allocsOut != "" {
		if f, err := os.Create(allocsOut); err == nil {
			// debug=0 → compressed protobuf suitable for pprof
			_ = pprof.Lookup("allocs").WriteTo(f, 0)
			_ = f.Close()
		}
	}
}