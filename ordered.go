package main

import "runtime"

// forEachOrdered calls work on each item using up to limit goroutines, and
// calls yield with each result in input order, from the calling goroutine. The
// next item starts only after a result has been yielded, so at most limit
// results are held at a time, even when an early item is slow.
func forEachOrdered[T, R any](items []T, limit int, work func(T) R, yield func(T, R)) {
	limit = max(limit, 1)
	results := make([]chan R, len(items))
	start := func(i int) {
		results[i] = make(chan R, 1)
		go func() {
			results[i] <- work(items[i])
		}()
	}
	for i := range min(limit, len(items)) {
		start(i)
	}
	for i, item := range items {
		yield(item, <-results[i])
		if next := i + limit; next < len(items) {
			start(next)
		}
	}
}

// workers returns how many files to format at once.
func workers() int {
	return runtime.GOMAXPROCS(0)
}
