package core

import (
	"fmt"
	"runtime"
	"sync"
)

// RunExtractCommentsConcurrently processes multiple files in parallel.
// Uses worker goroutines to avoid bottlenecks on large repos.
func RunExtractCommentsConcurrently(
	files []string, maxWorkers int, tags string, ignoreErrors bool,
) map[string][]Comment {
	results := make(map[string][]Comment)
	if len(files) == 0 {
		return results
	}

	// Default worker count to CPU cores
	if maxWorkers <= 0 {
		maxWorkers = runtime.NumCPU()
	}
	if maxWorkers > len(files) {
		maxWorkers = len(files) // don’t spawn more workers than files
	}

	var mu sync.Mutex
	var wg sync.WaitGroup
	ch := make(chan string, len(files)) // buffered channel holds all files

	// Worker: consumes file paths, extracts comments, stores them in results
	worker := func() {
		defer wg.Done()
		for file := range ch {
			cmts, err := ExtractComments(file, tags)
			if err != nil && !ignoreErrors {
				fmt.Printf("Error processing %s: %v\n", file, err)
			}
			if len(cmts) > 0 {
				mu.Lock()
				results[file] = cmts
				mu.Unlock()
			}
		}
	}

	// Start workers
	wg.Add(maxWorkers)
	for i := 0; i < maxWorkers; i++ {
		go worker()
	}

	// Send all files into channel
	for _, f := range files {
		ch <- f
	}
	close(ch)
	wg.Wait()

	return results
}

