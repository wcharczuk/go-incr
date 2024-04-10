package incr

import (
	"context"
	"sync"
)

// parallelBatch is an iterator processor that runs in parallel, calling a given delegate for each iterator item seen.
func parallelBatch[A any](ctx context.Context, fn func(context.Context, A) error, iter func() (A, bool), parallelism int) (err error) {
	var errOnce sync.Once
	sem := make(chan A, parallelism)
	wg := new(sync.WaitGroup)

	process := func() {
		defer wg.Done()
		workErr := fn(ctx, <-sem)
		if workErr != nil {
			errOnce.Do(func() {
				err = workErr
			})
		}
	}
	w, ok := iter()
	for ok {
		sem <- w
		wg.Add(1)
		go process()
		w, ok = iter()
	}
	wg.Wait()
	return
}
