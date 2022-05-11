package incr

import (
	"context"
	"errors"
	"runtime"
	"sync"
)

// ParallelStabilize stabilizes a graph in parallel.
func ParallelStabilize(ctx context.Context, nodes ...INode) error {
	seenGraphs := make(set[Identifier])
	for _, gn := range nodes {
		if shouldInitialize(gn.Node()) {
			tracePrintf(ctx, "parallel stabilize; initializing graph rooted at: %v", gn)
			Initialize(ctx, gn)
		}
		if seenGraphs.has(gn.Node().gs.id) {
			continue
		}
		seenGraphs.add(gn.Node().gs.id)
		if err := parallelStabilizeNode(ctx, gn); err != nil {
			return err
		}
	}
	return nil
}

type set[T comparable] map[T]struct{}

func (s set[T]) has(t T) (ok bool) {
	_, ok = s[t]
	return
}

func (s set[T]) add(t T) {
	s[t] = struct{}{}
}

func parallelStabilizeNode(ctx context.Context, gn INode) error {
	gnn := gn.Node()

	gnn.gs.mu.Lock()
	defer gnn.gs.mu.Unlock()
	if gnn.gs.s != StatusNotStabilizing {
		tracePrintf(ctx, "parallel stabilize; already stabilizing, cannot continue")
		return errors.New("parallel stabilize; already stabilizing, cannot continue")
	}
	defer func() {
		tracePrintf(ctx, "parallel stabilize; stabilization %s.%d complete", gnn.gs.id.Short(), gnn.gs.sn)
		gnn.gs.sn++
		gnn.gs.s = StatusNotStabilizing
	}()
	gnn.gs.s = StatusStabilizing
	tracePrintf(ctx, "parallel stabilize; stabilization %s.%d starting", gnn.gs.id.Short(), gnn.gs.sn)
	return parallelRecomputeAll(ctx, gnn.gs)
}

func parallelRecomputeAll(ctx context.Context, gs *graphState) error {
	wg := sync.WaitGroup{}
	workerPool := &parallelWorkerPool[*Node]{
		work: make(chan *Node),
		action: func(ictx context.Context, n *Node) error {
			defer wg.Done()
			tracePrintf(ctx, "parallel stabilize[%d]; recomputing %s", gs.sn, n.id.Short())
			return n.recompute(ctx)
		},
		started: make(chan struct{}),
	}
	if gs.rh.Len() == 0 {
		return nil
	}

	go func() {
		tracePrintf(ctx, "parallel stabilize[%d]; worker pool starting", gs.sn)
		_ = workerPool.Start(ctx)
	}()
	<-workerPool.started
	defer workerPool.Stop()

	var minHeight []INode
	var nn *Node
	for gs.rh.Len() > 0 {
		minHeight = gs.rh.RemoveMinHeight()
		tracePrintf(ctx, "parallel stabilize[%d]; stabilizing %d node block", gs.sn, len(minHeight))
		for _, n := range minHeight {
			nn = n.Node()
			if nn.shouldRecompute() {
				if nn.maybeCutoff(ctx) {
					continue
				}
				nn.changedAt = gs.sn
				wg.Add(1)
				tracePrintf(ctx, "parallel stabilize[%d]; submitting %v", gs.sn, n)
				workerPool.Submit(ctx, nn)
			}
		}
		wg.Wait()
	}
	return nil
}

var (
	errParallelWorkerPoolStopping = errors.New("parallel worker pool; stopping")
)

type parallelWorkerPool[T any] struct {
	work       chan T
	action     func(context.Context, T) error
	numWorkers int

	started      chan struct{}
	stop         chan struct{}
	stopped      chan struct{}
	workers      []*parallelWorker[T]
	readyWorkers chan *parallelWorker[T]
}

// Submit process a work item by dispatching it to a ready worker.
func (pwp *parallelWorkerPool[T]) Submit(ctx context.Context, t T) error {
	select {
	case <-ctx.Done():
		return context.Canceled
	case <-pwp.stop:
		return errParallelWorkerPoolStopping
	case pwp.work <- t:
		return nil
	}
}

func (pwp *parallelWorkerPool[T]) Start(ctx context.Context) error {
	effectiveNumWorkers := pwp.numWorkers
	if effectiveNumWorkers == 0 {
		effectiveNumWorkers = runtime.NumCPU()
	}
	pwp.readyWorkers = make(chan *parallelWorker[T], effectiveNumWorkers)
	workerErrors := make(chan error, effectiveNumWorkers)
	for x := 0; x < effectiveNumWorkers; x++ {
		w := &parallelWorker[T]{
			work:      make(chan T),
			ctx:       ctx,
			action:    pwp.action,
			finalizer: pwp.finalizer,
			errors:    workerErrors,
			stop:      make(chan struct{}),
			stopped:   make(chan struct{}),
		}
		go w.dispatch()
		pwp.workers = append(pwp.workers, w)
		pwp.readyWorkers <- w
	}

	pwp.stop = make(chan struct{})
	pwp.stopped = make(chan struct{})

	close(pwp.started)

	defer close(pwp.stopped)
	var workItem T
	var readyWorker *parallelWorker[T]
	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-workerErrors:
			return err
		case <-pwp.stop:
			return nil
		default:
		}

		select {
		case <-ctx.Done():
			return nil
		case err := <-workerErrors:
			return err
		case <-pwp.stop:
			return nil
		case workItem = <-pwp.work:
			select {
			case <-ctx.Done():
				return nil
			case err := <-workerErrors:
				return err
			case <-pwp.stop:
				return nil
			case readyWorker = <-pwp.readyWorkers:
				select {
				case <-ctx.Done():
					return nil
				case err := <-workerErrors:
					return err
				case <-pwp.stop:
					return nil
				case readyWorker.work <- workItem:
					continue
				}
			}
		}
	}
}

func (pwp *parallelWorkerPool[T]) Stop() {
	close(pwp.stop)
	<-pwp.stopped
	for _, w := range pwp.workers {
		close(w.stop)
		<-w.stopped
	}
}

func (pwp *parallelWorkerPool[T]) finalizer(ctx context.Context, w *parallelWorker[T]) {
	pwp.readyWorkers <- w
}

type parallelWorker[T any] struct {
	work      chan T
	errors    chan error
	action    func(context.Context, T) error
	finalizer func(context.Context, *parallelWorker[T])
	ctx       context.Context
	stop      chan struct{}
	stopped   chan struct{}
}

func (p *parallelWorker[T]) dispatch() {
	defer close(p.stopped)
	var workItem T
	var err error
	for {
		select {
		case <-p.stop:
			return
		case <-p.ctx.Done():
			return
		default:
		}

		select {
		case <-p.stop:
			return
		case <-p.ctx.Done():
			return
		case workItem = <-p.work:
			if err = p.action(p.ctx, workItem); err != nil {
				select {
				case <-p.stop:
					return
				case <-p.ctx.Done():
					return
				case p.errors <- err:
					return
				}
			}
			if p.finalizer != nil {
				p.finalizer(p.ctx, p)
			}
		}
	}
}
