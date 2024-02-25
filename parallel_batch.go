package incr

import (
	"context"
)

// parallelBatch is a collection of goroutines working on subtasks that are part of
// the same overall task.
//
// A zero Group is valid and does not cancel on error.
type parallelBatch[A any] struct {
	action  func(context.Context, A) error
	work    chan A
	workers chan *parallelBatchWorker[A]
}

type parallelBatchWorker[A any] struct {
	action func(context.Context, A) error
	work   chan A
	errors chan error
}

func (pbw *parallelBatchWorker[A]) run(ctx context.Context) {
	var wi A
	var err error
	for {
		select {
		case wi = <-pbw.work:
			if err = pbw.action(ctx, wi); err != nil {
				pbw.errors <- err
			}
		}
	}
}
