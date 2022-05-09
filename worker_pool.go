package incr

import (
	"context"
	"errors"
	"runtime"
	"sync"
	"sync/atomic"
)

func newWorkerPool[T any](action workerAction[T], finalizer workerFinalizer[T]) *workerPool[T] {
	return &workerPool[T]{
		latch:     newLatch(),
		action:    action,
		finalizer: finalizer,
		work:      make(chan T),
	}
}

type workerPool[T any] struct {
	Errors      chan error
	Parallelism int
	SkipRecover bool

	latch            *latch
	action           workerAction[T]
	finalizer        workerFinalizer[T]
	work             chan T
	workers          []*worker[T]
	availableWorkers chan *worker[T]
}

func (wp *workerPool[T]) Start(ctx context.Context) error {
	if !wp.latch.CanStart() {
		return errCannotStart
	}
	if wp.action == nil {
		return errCannotStartActionRequired
	}
	wp.initializeWorkers(ctx)
	wp.latch.SendStarting()
	started := wp.latch.ReceiveStarted()
	go wp.dispatch(ctx)
	<-started
	return nil
}

func (wp *workerPool[T]) Shutdown(ctx context.Context) error {
	if !wp.latch.CanStop() {
		return errCannotStop
	}
	for x := 0; x < len(wp.workers); x++ {
		wp.workers[x].Shutdown(ctx)
	}
	return nil
}

func (wp *workerPool[T]) Push(ctx context.Context, t T) {
	select {
	case wp.work <- t:
		return
	case <-ctx.Done():
		return
	}
}

func (wp *workerPool[T]) dispatch(ctx context.Context) {
	wp.latch.SendStarted()
	defer wp.latch.SendStopped()

	var worker *worker[T]
	var workItem T
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		select {
		case <-ctx.Done():
			return
		case workItem = <-wp.work:
			select {
			case <-ctx.Done():
				return
			case worker = <-wp.availableWorkers:
				select {
				case <-ctx.Done():
					return
				case worker.work <- workItem:
				}
			}
		}
	}
}

func (wp *workerPool[T]) returnWorker(ctx context.Context, w *worker[T]) error {
	if wp.finalizer != nil {
		if err := wp.finalizer(ctx, w); err != nil {
			return err
		}
	}
	select {
	case wp.availableWorkers <- w:
		return nil
	case <-ctx.Done():
		return nil
	}
}

func (wp *workerPool[T]) initializeWorkers(ctx context.Context) {
	effectiveParallelism := wp.Parallelism
	if effectiveParallelism == 0 {
		effectiveParallelism = runtime.NumCPU()
	}

	wp.workers = make([]*worker[T], effectiveParallelism)
	wp.availableWorkers = make(chan *worker[T], effectiveParallelism)

	// create and start workers.
	for x := 0; x < effectiveParallelism; x++ {
		worker := newWorker(wp.action)
		worker.ctx = ctx
		worker.errors = wp.Errors
		worker.skipRecover = wp.SkipRecover
		worker.finalizer = wp.returnWorker

		workerStarted := worker.ReceiveStarted()
		go func() { _ = worker.Start() }()
		<-workerStarted

		wp.workers[x] = worker
		wp.availableWorkers <- worker
	}
}

func newLatch() *latch {
	l := new(latch)
	l.Reset()
	return l
}

const (
	latchStopped  int32 = 0
	latchStarting int32 = 1
	latchResuming int32 = 2
	latchStarted  int32 = 3
	latchActive   int32 = 4
	latchPausing  int32 = 5
	latchPaused   int32 = 6
	latchStopping int32 = 7
)

var (
	errCannotStart               = errors.New("cannot start; already started")
	errCannotStop                = errors.New("cannot stop; already stopped")
	errCannotCancel              = errors.New("cannot cancel; already canceled")
	errCannotStartActionRequired = errors.New("cannot start; action is required")
)

type latch struct {
	sync.Mutex
	state    int32
	starting chan struct{}
	started  chan struct{}
	stopping chan struct{}
	stopped  chan struct{}
}

func (l *latch) Reset() {
	l.Lock()
	atomic.StoreInt32(&l.state, latchStopped)
	l.starting = make(chan struct{}, 1)
	l.started = make(chan struct{}, 1)
	l.stopping = make(chan struct{}, 1)
	l.stopped = make(chan struct{}, 1)
	l.Unlock()
}

func (l *latch) CanStart() bool {
	return atomic.LoadInt32(&l.state) == latchStopped
}

func (l *latch) CanStop() bool {
	return atomic.LoadInt32(&l.state) == latchStarted
}

func (l *latch) IsStarting() bool {
	return atomic.LoadInt32(&l.state) == latchStarting
}

func (l *latch) IsStarted() bool {
	return atomic.LoadInt32(&l.state) == latchStarted
}

func (l *latch) IsStopping() bool {
	return atomic.LoadInt32(&l.state) == latchStopping
}

func (l *latch) IsStopped() bool {
	return atomic.LoadInt32(&l.state) == latchStopped
}

func (l *latch) ReceiveStarting() (notifyStarting <-chan struct{}) {
	l.Lock()
	notifyStarting = l.starting
	l.Unlock()
	return
}

func (l *latch) ReceiveStarted() (notifyStarted <-chan struct{}) {
	l.Lock()
	notifyStarted = l.started
	l.Unlock()
	return
}

func (l *latch) ReceiveStopping() (notifyStopping <-chan struct{}) {
	l.Lock()
	notifyStopping = l.stopping
	l.Unlock()
	return
}

func (l *latch) ReceiveStopped() (notifyStopped <-chan struct{}) {
	l.Lock()
	notifyStopped = l.stopped
	l.Unlock()
	return
}

func (l *latch) SendStarting() {
	if l.IsStarting() {
		return
	}
	atomic.StoreInt32(&l.state, latchStarting)
	l.starting <- struct{}{}
}

func (l *latch) SendStarted() {
	if l.IsStarted() {
		return
	}
	atomic.StoreInt32(&l.state, latchStarted)
	l.started <- struct{}{}
}

func (l *latch) SendStopping() {
	if l.IsStopping() {
		return
	}
	atomic.StoreInt32(&l.state, latchStopping)
	l.stopping <- struct{}{}
}

func (l *latch) SendStopped() {
	if l.IsStopped() {
		return
	}
	atomic.StoreInt32(&l.state, latchStopped)
	l.stopped <- struct{}{}
}

func (l *latch) WaitStarted() {
	if !l.CanStart() {
		return
	}
	started := l.ReceiveStarted()
	l.SendStarting()
	<-started
}

func (l *latch) WaitStopped() {
	if !l.CanStop() {
		return
	}
	stopped := l.ReceiveStopped()
	l.SendStopping()
	<-stopped
}
