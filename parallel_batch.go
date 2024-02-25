package incr

import (
	"context"
	"runtime"
)

func parallelBatch[A any](ctx context.Context, fn func(context.Context, A) error, work ...A) error {
	workers := make([]*parallelBatchWorker[A], runtime.NumCPU())
	readyWorkers := make(chan *parallelBatchWorker[A], runtime.NumCPU())
	errors := make(chan error, len(work))

	defer func() {
		close(readyWorkers)
		close(errors)
	}()
	for x := 0; x < runtime.NumCPU(); x++ {
		w := &parallelBatchWorker[A]{
			action:  fn,
			work:    make(chan A),
			errors:  errors,
			stop:    make(chan struct{}),
			stopped: make(chan struct{}),
			done: func(w *parallelBatchWorker[A]) {
				readyWorkers <- w
			},
		}
		go w.run(ctx)
		workers[x] = w
		readyWorkers <- w
	}
	var worker *parallelBatchWorker[A]
	for _, w := range work {
		worker = <-readyWorkers
		worker.work <- w
	}

	for _, worker := range workers {
		close(worker.stop)
		<-worker.stopped
	}
	if len(errors) > 0 {
		return <-errors
	}
	return nil
}

type parallelBatchWorker[A any] struct {
	action  func(context.Context, A) error
	work    chan A
	errors  chan error
	stop    chan struct{}
	stopped chan struct{}
	done    func(*parallelBatchWorker[A])
}

func (pbw *parallelBatchWorker[A]) run(ctx context.Context) {
	var wi A
	var err error
	for {
		select {
		case <-pbw.stop:
			close(pbw.stopped)
			return
		default:
		}
		select {
		case <-pbw.stop:
			close(pbw.stopped)
			return
		case wi = <-pbw.work:
			if err = pbw.action(ctx, wi); err != nil {
				pbw.errors <- err
			}
			pbw.done(pbw)
		}
	}
}
