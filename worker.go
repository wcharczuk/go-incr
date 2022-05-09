package incr

import (
	"context"
	"fmt"
)

func newWorker[T any](action workerAction[T]) *worker[T] {
	return &worker[T]{
		latch:  newLatch(),
		ctx:    context.Background(),
		action: action,
		work:   make(chan T),
	}
}

type workerAction[T any] func(context.Context, T) error

type workerFinalizer[T any] func(context.Context, *worker[T]) error

type worker[T any] struct {
	latch *latch

	ctx       context.Context
	action    workerAction[T]
	finalizer workerFinalizer[T]

	skipRecover bool
	errors      chan error
	work        chan T
}

func (w *worker[T]) ReceiveStarted() <-chan struct{} {
	return w.latch.ReceiveStarted()
}

func (w *worker[T]) ReceiveStopped() <-chan struct{} {
	return w.latch.ReceiveStarted()
}

func (w *worker[T]) Start() error {
	if !w.latch.CanStart() {
		return errCannotStart
	}
	w.latch.SendStarting()
	w.dispatch()
	return nil
}

func (w *worker[T]) Stop() error {
	if !w.latch.CanStop() {
		return errCannotStop
	}
	w.latch.WaitStopped()
	w.latch.Reset()
	return nil
}

func (w *worker[T]) Shutdown(ctx context.Context) {
	stopped := make(chan struct{})
	go func() {
		defer func() {
			w.latch.Reset()
			close(stopped)
		}()
		w.latch.WaitStopped()
		if workLeft := len(w.work); workLeft > 0 {
			for x := 0; x < workLeft; x++ {
				w.execute(ctx, <-w.work)
			}
		}
	}()
	select {
	case <-stopped:
		return
	case <-ctx.Done():
		return
	}
}

func (w *worker[T]) background() context.Context {
	if w.ctx != nil {
		return w.ctx
	}
	return context.Background()
}

func (w *worker[T]) dispatch() {
	w.latch.SendStarted()
	defer w.latch.SendStopped()
	var workItem T
	var stopping <-chan struct{}
	for {
		stopping = w.latch.ReceiveStopping()
		select {
		case <-stopping:
			return
		case <-w.background().Done():
			return
		default:
		}
		select {
		case workItem = <-w.work:
			w.execute(w.background(), workItem)
		case <-stopping:
			return
		case <-w.background().Done():
			return
		}
	}
}

func (w *worker[T]) execute(ctx context.Context, workItem T) {
	defer func() {
		if !w.skipRecover {
			if r := recover(); r != nil {
				w.handlePanic(r)
			}
		}
		if w.finalizer != nil {
			w.handleError(w.finalizer(ctx, w))
		}
	}()
	if w.action != nil {
		w.handleError(w.action(ctx, workItem))
	}
}

// handlePanic sends a non-nil panic recovery result
// to the error channel if one is provided.
func (w *worker[T]) handlePanic(r interface{}) {
	if r == nil {
		return
	}
	if w.errors == nil {
		return
	}
	w.errors <- fmt.Errorf("%v", r)
}

// handleError sends a non-nil err to the error
// collector if one is provided.
func (w *worker[T]) handleError(err error) {
	if err == nil {
		return
	}
	if w.errors == nil {
		return
	}
	w.errors <- err
}
