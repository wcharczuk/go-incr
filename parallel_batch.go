package incr

import (
	"context"
	"runtime"
	"sync"
)

func parallelBatch[A any](ctx context.Context, fn func(context.Context, A) error, iter func() (A, bool)) (err error) {
	var errOnce sync.Once
	sem := make(chan A, runtime.NumCPU())
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
