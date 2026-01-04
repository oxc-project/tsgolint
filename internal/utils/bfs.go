// NOTE: This is a direct copy of `typescript-go/internal/core/bfs.go`
// The shim can't expose generic types yet, so we have to copy the code here.
// https://github.com/golang/go/issues/60425
package utils

import (
	"github.com/typescript-eslint/tsgolint/internal/collections"
)

type BreadthFirstSearchResult[N comparable] struct {
	Stopped bool
	Path    []N
}

type breadthFirstSearchJob[N comparable] struct {
	node   N
	parent *breadthFirstSearchJob[N]
}

type BreadthFirstSearchLevel[N comparable] struct {
	jobs *collections.OrderedMap[N, *breadthFirstSearchJob[N]]
}

func (l *BreadthFirstSearchLevel[N]) Has(node N) bool {
	return l.jobs.Has(node)
}

func (l *BreadthFirstSearchLevel[N]) Delete(node N) {
	l.jobs.Delete(node)
}

func (l *BreadthFirstSearchLevel[N]) Range(f func(node N) bool) {
	for node := range l.jobs.Keys() {
		if !f(node) {
			return
		}
	}
}

type BreadthFirstSearchOptions[N comparable] struct {
	// Visited is a set of nodes that have already been visited.
	// If nil, a new set will be created.
	Visited *collections.SyncSet[N]
	// PreprocessLevel is a function that, if provided, will be called
	// before each level, giving the caller an opportunity to remove nodes.
	PreprocessLevel func(*BreadthFirstSearchLevel[N])
}

func BreadthFirstSearch[N comparable](
	start N,
	neighbors func(N) []N,
	visit func(node N) (isResult bool, stop bool),
	options BreadthFirstSearchOptions[N],
) BreadthFirstSearchResult[N] {
	visited := options.Visited
	if visited == nil {
		visited = &collections.SyncSet[N]{}
	}
	createPath := func(job *breadthFirstSearchJob[N]) []N {
		var path []N
		for job != nil {
			path = append(path, job.node)
			job = job.parent
		}
		return path
	}

	// processLevel processes each node at the current level.
	// It produces either a list of jobs to be processed in the next level,
	// or a result if the visit function returns true for any node.
	var fallback *breadthFirstSearchJob[N]
	level := collections.NewOrderedMapFromList([]collections.MapEntry[N, *breadthFirstSearchJob[N]]{
		{Key: start, Value: &breadthFirstSearchJob[N]{node: start}},
	})

	for levelIndex := 0; level.Size() > 0; levelIndex++ {
		if options.PreprocessLevel != nil {
			options.PreprocessLevel(&BreadthFirstSearchLevel[N]{jobs: level})
		}

		nextLevel := collections.NewOrderedMapWithSizeHint[N, *breadthFirstSearchJob[N]](level.Size())

		for i := range level.Size() {
			_, job, _ := level.EntryAt(i)

			if !visited.AddIfAbsent(job.node) {
				continue
			}

			isResult, stopVisit := visit(job.node)
			if isResult {
				if stopVisit {
					return BreadthFirstSearchResult[N]{Stopped: true, Path: createPath(job)}
				}
				if fallback == nil {
					fallback = job
				}
			}

			for _, child := range neighbors(job.node) {
				if !nextLevel.Has(child) {
					nextLevel.Set(child, &breadthFirstSearchJob[N]{node: child, parent: job})
				}
			}
		}

		level = nextLevel
	}

	return BreadthFirstSearchResult[N]{Stopped: false, Path: createPath(fallback)}
}
