package incr

import (
	"context"
	"runtime"
	"sync"
)

func parallelBatch[A any](ctx context.Context, fn func(context.Context, A) error, work ...A) (err error) {
	var errOnce sync.Once
	sem := make(chan struct{}, runtime.NumCPU())
	wg := new(sync.WaitGroup)

	process := func(i int) {
		defer func() {
			<-sem
			wg.Done()
		}()
		workErr := fn(ctx, work[i])
		if workErr != nil {
			errOnce.Do(func() {
				err = workErr
			})
		}
	}
	for index := range work {
		sem <- struct{}{}
		wg.Add(1)
		go process(index)
	}
	wg.Wait()
	return
}
