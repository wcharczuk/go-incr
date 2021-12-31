package incr

import (
	"context"
	"runtime"
)

// Batch processes a given work channel with a given action.
func Batch[T any](ctx context.Context, work chan T, action func(context.Context, T) error) (err error) {
	// sanity check on empty work.
	if len(work) == 0 {
		return nil
	}

	// handle an edge case where we are on a 128 core
	// cpu but there are ~5 work items.
	effectiveParallelism := runtime.NumCPU()
	if len(work) < effectiveParallelism {
		effectiveParallelism = len(work)
	}

	batchErrors := make(chan error, effectiveParallelism)
	allWorkers := make([]*batchWorker[T], effectiveParallelism)
	availableWorkers := make(chan *batchWorker[T], effectiveParallelism)

	finalizer := func(worker *batchWorker[T]) {
		availableWorkers <- worker
	}

	for x := 0; x < effectiveParallelism; x++ {
		worker := &batchWorker[T]{
			work:      make(chan T),
			action:    action,
			finalizer: finalizer,
			errors:    batchErrors,
			stop:      make(chan struct{}),
			stopped:   make(chan struct{}),
		}
		go worker.Process(ctx)
		allWorkers[x] = worker
		availableWorkers <- worker
	}

	defer func() {
		for _, w := range allWorkers {
			close(w.stop)
			<-w.stopped
		}
		if err == nil && len(batchErrors) > 0 {
			err = <-batchErrors
		}
	}()

	var worker *batchWorker[T]
	var workItem T
	for len(work) > 0 {
		select {
		case <-ctx.Done():
			return nil
		case err = <-batchErrors:
			return err
		default:
		}

		select {
		case <-ctx.Done():
			return
		case err = <-batchErrors:
			return err
		case workItem = <-work:
			select {
			case <-ctx.Done():
				return
			case err = <-batchErrors:
				return err
			case worker = <-availableWorkers:
				select {
				case <-ctx.Done():
					return
				case err = <-batchErrors:
					return err
				case worker.work <- workItem:
					continue
				}
			}
		}
	}
	return
}

type batchWorker[T any] struct {
	work      chan T
	action    func(context.Context, T) error
	finalizer func(*batchWorker[T])
	stop      chan struct{}
	stopped   chan struct{}
	errors    chan error
}

func (bw *batchWorker[T]) Process(ctx context.Context) {
	defer func() { close(bw.stopped) }()

	var w T
	for {
		// do a work pass first to
		// prioritize doing work instead
		// of stopping
		select {
		case <-ctx.Done():
			return
		case w = <-bw.work:
			bw.process(ctx, w)
			continue
		default:
		}

		select {
		case <-ctx.Done():
			return
		case <-bw.stop:
			return
		case w = <-bw.work:
			bw.process(ctx, w)
			continue
		}
	}
}

func (bw *batchWorker[T]) process(ctx context.Context, w T) {
	if err := bw.action(ctx, w); err != nil {
		println("batch worker pushing error")
		bw.errors <- err
	}
	bw.finalizer(bw)
}
