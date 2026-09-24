package linter

import (
	"sync"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/compiler"
)

type checkerWorkload struct {
	checker *checker.Checker
	program *compiler.Program
	queue   chan *ast.SourceFile
}

func makeSourceFileQueue(files []*ast.SourceFile) chan *ast.SourceFile {
	queue := make(chan *ast.SourceFile, len(files))
	for _, file := range files {
		queue <- file
	}
	close(queue)
	return queue
}

func makeCheckerWorkloadQueue(program *compiler.Program, files []*ast.SourceFile) chan checkerWorkload {
	queue := makeSourceFileQueue(files)
	flatQueue := []checkerWorkload{}
	var flatQueueMu sync.Mutex
	program.ForEachCheckerParallel(func(idx int, ch *checker.Checker) {
		flatQueueMu.Lock()
		flatQueue = append(flatQueue, checkerWorkload{ch, program, queue})
		flatQueueMu.Unlock()
	})

	workloadQueue := make(chan checkerWorkload, len(flatQueue))
	for _, w := range flatQueue {
		workloadQueue <- w
	}
	close(workloadQueue)
	return workloadQueue
}
