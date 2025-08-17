package core

import (
	"fmt"
	"runtime"
	"sync"
)

// RunExtractCommentsConcurrently runs ExtractComments in parallel for multiple files.
func RunExtractCommentsConcurrently(
	files []string,
	maxWorkers int,
	tags string,
	ignoreErrors bool,
) map[string][]Comment {
	if maxWorkers <= 0 {
		maxWorkers = runtime.NumCPU()
	}
	if maxWorkers > len(files) {
		maxWorkers = len(files)
	}

	results := make(map[string][]Comment)
	var resultsMutex sync.Mutex
	var waitGroup sync.WaitGroup
	fileChannel := make(chan string)

	// Worker goroutine for extracting comments
	worker := func() {
		defer waitGroup.Done()
		for file := range fileChannel {
			extractedComments, err := ExtractComments(file, tags)
			if err != nil && !ignoreErrors {
				fmt.Printf("Error processing %s: %v\n", file, err)
			}
			if extractedComments != nil {
				resultsMutex.Lock()
				results[file] = extractedComments
				resultsMutex.Unlock()
			}
		}
	}

	waitGroup.Add(maxWorkers)
	for i := 0; i < maxWorkers; i++ {
		go worker()
	}

	// Feed files into the worker pool
	go func() {
		for _, filePath := range files {
			fileChannel <- filePath
		}
		close(fileChannel)
	}()

	waitGroup.Wait()
	return results
}

