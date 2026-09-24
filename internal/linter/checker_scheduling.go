package linter

import (
	"context"
	"os"
	"sync"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/core"
)

type checkerScheduling int

const (
	checkerSchedulingBaseline checkerScheduling = iota
	checkerSchedulingAffinity
	checkerSchedulingStealing
)

func checkerSchedulingFromEnvironment() checkerScheduling {
	// See https://github.com/oxc-project/tsgolint/pull/1240 for mode values and limitations.
	switch os.Getenv("OXLINT_TSGOLINT_CHECKER_SCHEDULING") {
	case "affinity":
		return checkerSchedulingAffinity
	case "stealing", "sorted":
		return checkerSchedulingStealing
	default:
		return checkerSchedulingBaseline
	}
}

// checkerScheduler owns queue selection and initial checker reservations.
// Construct it before launching workers; each worker requests its initial workload once.
type checkerScheduler struct {
	workers          int
	queue            <-chan checkerWorkload
	initialWorkloads []checkerWorkload
}

func newCheckerScheduler(scheduling checkerScheduling, program *compiler.Program, files []*ast.SourceFile, workers int) checkerScheduler {
	var workloadQueue chan checkerWorkload
	if scheduling == checkerSchedulingBaseline {
		workloadQueue = makeCheckerWorkloadQueue(program, files)
	} else {
		workloadQueue = makeAffinityCheckerWorkloadQueue(program, files, scheduling == checkerSchedulingStealing)
	}

	var initialWorkloads []checkerWorkload
	if scheduling == checkerSchedulingStealing {
		// Reserve one checker per worker before starting any of them. Otherwise
		// a fast worker can claim empty workloads and leave peers unable to steal.
		// More workers than selected files cannot add parallelism; a single file
		// should keep its warmed owner instead of being stolen by an idle peer.
		workers = min(workers, len(workloadQueue), len(files))
		initialWorkloads = make([]checkerWorkload, workers)
		for i := range initialWorkloads {
			initialWorkloads[i] = <-workloadQueue
		}
	}

	return checkerScheduler{
		workers: workers, queue: workloadQueue, initialWorkloads: initialWorkloads,
	}
}

func (s checkerScheduler) initialWorkload(workerIndex int) (checkerWorkload, bool) {
	if s.initialWorkloads != nil {
		return s.initialWorkloads[workerIndex], true
	}
	workload, ok := <-s.queue
	return workload, ok
}

type checkerWorkload struct {
	checker *checker.Checker
	program *compiler.Program
	queue   chan *ast.SourceFile
	peers   []chan *ast.SourceFile
	pending <-chan checkerWorkload
}

func (w checkerWorkload) nextFile() *ast.SourceFile {
	if file, ok := <-w.queue; ok {
		return file
	}
	// Claim unstarted checker workloads before stealing. This preserves cache
	// affinity when there are fewer workers than checkers, including one worker.
	if len(w.pending) > 0 {
		return nil
	}
	for {
		var largest chan *ast.SourceFile
		for _, queue := range w.peers {
			if len(queue) > len(largest) {
				largest = queue
			}
		}
		if largest == nil {
			return nil
		}
		// All queues are prefilled and closed. Another worker may win the last
		// file, in which case rescan. Only the file moves: the thief continues
		// using its own checker, never the victim's mutable checker state.
		if file, ok := <-largest; ok {
			return file
		}
	}
}

func (w checkerWorkload) nextWorkload() (checkerWorkload, bool) {
	if next, ok := <-w.pending; ok {
		return next, true
	}
	// Another worker may claim the last pending workload after nextFile saw
	// it. Keep this checker available to steal instead of abandoning its lane.
	// Queues only shrink, so an empty scan means no stealable work remains.
	for _, queue := range w.peers {
		if len(queue) > 0 {
			return w, true
		}
	}
	return checkerWorkload{}, false
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
		flatQueue = append(flatQueue, checkerWorkload{checker: ch, program: program, queue: queue})
		flatQueueMu.Unlock()
	})

	workloadQueue := make(chan checkerWorkload, len(flatQueue))
	for _, w := range flatQueue {
		w.pending = workloadQueue
		workloadQueue <- w
	}
	close(workloadQueue)
	return workloadQueue
}

func makeAffinityCheckerWorkloadQueue(program *compiler.Program, files []*ast.SourceFile, steal bool) chan checkerWorkload {
	// Only the compiler-owned pool exposes stable checkers through this API.
	// A custom pool can return request-scoped checkers that cannot be retained
	// after their release callback has been called.
	checkers := make(map[int]*checker.Checker)
	var checkersMu sync.Mutex
	program.ForEachCheckerParallel(func(idx int, ch *checker.Checker) {
		checkersMu.Lock()
		checkers[idx] = ch
		checkersMu.Unlock()
	})

	workloadQueue := make(chan checkerWorkload, len(checkers))
	if len(checkers) == 0 || len(files) == 0 {
		close(workloadQueue)
		return workloadQueue
	}

	// Reuse the checker that performed semantic diagnostics for each file.
	// Keep one workload per checker so that lint workers never use it concurrently.
	ctx := core.WithRequestID(context.Background(), "__single_run__")
	filesByChecker := make(map[*checker.Checker][]*ast.SourceFile, len(checkers))
	for _, file := range files {
		ch, done := program.GetTypeCheckerForFile(ctx, file)
		filesByChecker[ch] = append(filesByChecker[ch], file)
		done()
	}
	queues := make([]chan *ast.SourceFile, len(checkers))
	for idx := range queues {
		queues[idx] = makeSourceFileQueue(filesByChecker[checkers[idx]])
	}
	for idx := range queues {
		if !steal && len(queues[idx]) == 0 {
			continue
		}
		w := checkerWorkload{
			checker: checkers[idx], program: program, queue: queues[idx],
			pending: workloadQueue,
		}
		if steal {
			w.peers = queues
		}
		workloadQueue <- w
	}
	close(workloadQueue)
	return workloadQueue
}
