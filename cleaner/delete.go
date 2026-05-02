package cleaner

import (
	"os"
	"sync"
)

type DeleteResult struct {
	Path  string
	Error error
}

func DeleteModules(paths []string, concurrency int, progressFn func(current int, total int, path string)) []DeleteResult {
	if concurrency <= 0 {
		concurrency = 4
	}

	results := make([]DeleteResult, len(paths))
	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, concurrency)

	total := len(paths)
	current := 0

	for i, path := range paths {
		wg.Add(1)
		go func(idx int, p string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			err := os.RemoveAll(p)

			mu.Lock()
			current++
			if progressFn != nil {
				progressFn(current, total, p)
			}
			results[idx] = DeleteResult{Path: p, Error: err}
			mu.Unlock()
		}(i, path)
	}

	wg.Wait()
	return results
}
