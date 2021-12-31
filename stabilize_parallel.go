package incr

import (
	"context"
	"runtime"
	"sync"
	"time"
)

// Stabilize stabilizes a computation.
func Stabilize(ctx context.Context, outputs ...Stabilizer) error {
	stabilizationQueue, err := stabilizeDiscoverStale(ctx, outputs)
	if err != nil {
		return err
	}
	if err = stabilize(ctx, stabilizationQueue); err != nil {
		return err
	}
	return nil
}

func stabilizeDiscoverStale(ctx context.Context, outputs []Stabilizer) (chan Stabilizer, error) {
	mu := sync.Mutex{}
	staleNodes := map[nodeID]bool{}
	sq := new(Queue[Stabilizer])

	discoverWork := make(chan Stabilizer, len(outputs))
	for _, n := range outputs {
		discoverWork <- n
	}
	queueStabilization := func(s Stabilizer) error {
		mu.Lock()
		sq.Push(s)
		mu.Unlock()
		return nil
	}
	queueWork := func(ctx context.Context, s Stabilizer) {
		select {
		case <-ctx.Done():
			return
		case discoverWork <- s:
			return
		}
	}
	action := func(ctx context.Context, w Stabilizer) error {
		if stabilizeIsStale(ctx, staleNodes, w) {
			if err := queueStabilization(w); err != nil {
				return err
			}
		}
		for _, p := range w.getNode().parents {
			queueWork(ctx, p)
		}
		return nil
	}
	if err := stabilizeBatch(ctx, discoverWork, action); err != nil {
		return nil, err
	}
	output := make(chan Stabilizer, sq.Len())
	sq.ReverseEach(func(v Stabilizer) {
		output <- v
	})
	return output, nil
}

func stabilizeIsStale(ctx context.Context, staleNodes map[nodeID]bool, w Stabilizer) bool {
	if isStale, ok := staleNodes[w.getNode().id]; ok {
		return isStale
	}
	if w.getNode().isStale() {
		staleNodes[w.getNode().id] = true
		return true
	}
	for _, p := range w.getNode().parents {
		if stabilizeIsStale(ctx, staleNodes, p) {
			return true
		}
	}
	return false
}

func stabilize(ctx context.Context, stabilizationQueue chan Stabilizer) error {
	action := func(ctx context.Context, w Stabilizer) error {
		tracePrintf(ctx, "stabilizing %T %s", w, w.getNode().id.String())
		if err := w.Stabilize(ctx); err != nil {
			return err
		}
		w.getNode().recomputedAt = time.Now()
		return nil
	}
	return stabilizeBatch(ctx, stabilizationQueue, action)
}

func stabilizeBatch(ctx context.Context, work chan Stabilizer, action func(context.Context, Stabilizer) error) error {
	effectiveParallelism := runtime.NumCPU()
	batchErrors := make(chan error, effectiveParallelism)
	allWorkers := make([]*stabilizeWorker, effectiveParallelism)
	availableWorkers := make(chan *stabilizeWorker, effectiveParallelism)

	finalizer := func(worker *stabilizeWorker) {
		select {
		case <-ctx.Done():
			return
		case availableWorkers <- worker:
			return
		}
	}

	for x := 0; x < effectiveParallelism; x++ {
		worker := &stabilizeWorker{
			work:      make(chan Stabilizer),
			action:    action,
			finalizer: finalizer,
			stop:      make(chan struct{}),
			stopped:   make(chan struct{}),
			errors:    batchErrors,
		}
		go worker.discover(ctx)
		allWorkers[x] = worker
		availableWorkers <- worker
	}

	defer func() {
		for _, w := range allWorkers {
			close(w.stop)
			<-w.stopped
		}
	}()

	var worker *stabilizeWorker
	var workItem Stabilizer
	for len(work) > 0 {
		select {
		case <-ctx.Done():
			return nil
		case err := <-batchErrors:
			return err
		case workItem = <-work:
			select {
			case <-ctx.Done():
				return nil
			case worker = <-availableWorkers:
				select {
				case <-ctx.Done():
					return nil
				case worker.work <- workItem:
				}
			}
		}
	}
	return nil
}

type stabilizeWorker struct {
	work      chan Stabilizer
	action    func(context.Context, Stabilizer) error
	finalizer func(*stabilizeWorker)
	stop      chan struct{}
	stopped   chan struct{}
	errors    chan error
}

func (sw *stabilizeWorker) discover(ctx context.Context) {
	defer func() { close(sw.stopped) }()

	var w Stabilizer
	var err error
	for {
		select {
		case <-ctx.Done():
			return
		case <-sw.stop:
			return
		case w = <-sw.work:
			if err = sw.action(ctx, w); err != nil {
				sw.errors <- err
				return
			}
			sw.finalizer(sw)
		}
	}
}
