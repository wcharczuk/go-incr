package incr

import (
	"context"
	"errors"
	"runtime"
	"sync"
)

// ParallelStabilize stabilizes graphs in parallel as entered
// through representative nodes.
//
// For each input node, if the node is not attached to a graph, the full
// graph is discovered from the relationships on that given input node.
//
// If multiple nodes are supplied that are actually connected, initialization
// will skip the already connected (and as a result, initialized) graph.
//
// ParallelStabilize differs from Stabilize in that it reads the current
// recompute heap in pseudo-height chunks, processing each pseudo-height in
// parallel before moving on to the next, smallest height chunk.
//
// Each parallel recompute cycle may produce new nodes to process, and as a result
// parallel stabilization can move up and down in height before fully recomputing
// the graph.
func ParallelStabilize(ctx context.Context, n INode) error {
	if shouldInitialize(n.Node()) {
		tracePrintf(ctx, "parallel stabilize; initializing graph rooted at: %v", n)
		Initialize(ctx, n)
	}
	if err := parallelStabilize(ctx, n.Node().g); err != nil {
		return err
	}
	return nil
}

func parallelStabilize(ctx context.Context, g *graph) (err error) {
	if err = ensureNotStabilizing(ctx, g); err != nil {
		return
	}
	stabilizeStart(ctx, g)
	defer stabilizeEnd(ctx, g)
	return parallelRecomputeAll(ctx, g)
}

func parallelRecomputeAll(ctx context.Context, g *graph) error {
	wg := sync.WaitGroup{}
	workerPool := &parallelWorkerPool[*Node]{
		work: make(chan *Node),
		action: func(ictx context.Context, n *Node) error {
			defer wg.Done()
			return n.recompute(ctx)
		},
		started: make(chan struct{}),
	}
	if g.recomputeHeap.Len() == 0 {
		return nil
	}

	go func() {
		_ = workerPool.Start(ctx)
	}()
	<-workerPool.started
	defer workerPool.Stop()

	var minHeightBlock []INode
	var err error
	for g.recomputeHeap.Len() > 0 {
		minHeightBlock = g.recomputeHeap.RemoveMinHeight()
		for _, n := range minHeightBlock {
			wg.Add(1)
			if err = workerPool.Submit(ctx, n.Node()); err != nil {
				return err
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
func (p *parallelWorkerPool[T]) Submit(ctx context.Context, t T) error {
	select {
	case <-ctx.Done():
		return context.Canceled
	case <-p.stop:
		return errParallelWorkerPoolStopping
	case p.work <- t:
		return nil
	}
}

func (p *parallelWorkerPool[T]) Start(ctx context.Context) error {
	effectiveNumWorkers := p.numWorkers
	if effectiveNumWorkers == 0 {
		effectiveNumWorkers = runtime.NumCPU()
	}
	p.readyWorkers = make(chan *parallelWorker[T], effectiveNumWorkers)
	workerErrors := make(chan error, effectiveNumWorkers)
	for x := 0; x < effectiveNumWorkers; x++ {
		w := &parallelWorker[T]{
			work:      make(chan T),
			ctx:       ctx,
			action:    p.action,
			finalizer: parallelWorkerPoolFinalizer(p.readyWorkers),
			errors:    workerErrors,
			stop:      make(chan struct{}),
			stopped:   make(chan struct{}),
		}
		go w.dispatch()
		p.workers = append(p.workers, w)
		p.readyWorkers <- w
	}

	p.stop = make(chan struct{})
	p.stopped = make(chan struct{})

	close(p.started)

	defer close(p.stopped)
	var workItem T
	var readyWorker *parallelWorker[T]
	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-workerErrors:
			return err
		case <-p.stop:
			return nil
		default:
		}

		select {
		case <-ctx.Done():
			return nil
		case err := <-workerErrors:
			return err
		case <-p.stop:
			return nil
		case workItem = <-p.work:
			select {
			case <-ctx.Done():
				return nil
			case err := <-workerErrors:
				return err
			case <-p.stop:
				return nil
			case readyWorker = <-p.readyWorkers:
				select {
				case <-ctx.Done():
					return nil
				case err := <-workerErrors:
					return err
				case <-p.stop:
					return nil
				case readyWorker.work <- workItem:
					continue
				}
			}
		}
	}
}

func (p *parallelWorkerPool[T]) Stop() {
	close(p.stop)
	<-p.stopped
	for _, w := range p.workers {
		close(w.stop)
		<-w.stopped
	}
}

func parallelWorkerPoolFinalizer[T any](readyWorkers chan *parallelWorker[T]) func(context.Context, *parallelWorker[T]) {
	return func(ctx context.Context, w *parallelWorker[T]) {
		readyWorkers <- w
	}
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
