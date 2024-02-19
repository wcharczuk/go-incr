package incr

import (
	"fmt"
	"sync"
)

// parallelBatch is a collection of goroutines working on subtasks that are part of
// the same overall task.
//
// A zero Group is valid and does not cancel on error.
type parallelBatch struct {
	wg      sync.WaitGroup
	sem     chan parallelBatchToken
	errOnce sync.Once
	err     error
}

// parallelBatchToken is a token within the parallel batch system.
type parallelBatchToken struct{}

func (pb *parallelBatch) done() {
	if pb.sem != nil {
		<-pb.sem
	}
	pb.wg.Done()
}

// Wait blocks until all function calls from the Go method have returned, then
// returns the first non-nil error (if any) from them.
func (pb *parallelBatch) Wait() error {
	pb.wg.Wait()
	return pb.err
}

// Go calls the given function in a new goroutine.
// It blocks until the new goroutine can be added without the number of
// active goroutines in the group exceeding the configured limit.
//
// The first call to return a non-nil error cancels the group; its error will be
// returned by Wait.
func (pb *parallelBatch) Go(f func() error) {
	if pb.sem != nil {
		pb.sem <- parallelBatchToken{}
	}

	pb.wg.Add(1)
	go func() {
		defer pb.done()
		if err := f(); err != nil {
			pb.errOnce.Do(func() {
				pb.err = err
			})
		}
	}()
}

// SetLimit limits the number of active goroutines in this group to at most n.
// A negative value indicates no limit.
//
// Any subsequent call to the Go method will block until it can add an active
// goroutine without exceeding the configured limit.
//
// The limit must not be modified while any goroutines in the group are active.
func (pb *parallelBatch) SetLimit(n int) {
	if n < 0 {
		pb.sem = nil
		return
	}
	if len(pb.sem) != 0 {
		panic(fmt.Errorf("errgroup: modify limit while %v goroutines in the group are still active", len(pb.sem)))
	}
	pb.sem = make(chan parallelBatchToken, n)
}
