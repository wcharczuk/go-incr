package incr

import (
	"context"
	"runtime"
	"sync"
)

func parallelBatch[A any](ctx context.Context, fn func(context.Context, A) error, work ...A) (err error) {
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
	for _, w := range work {
		sem <- w
		wg.Add(1)
		go process()
	}
	wg.Wait()
	return
}
