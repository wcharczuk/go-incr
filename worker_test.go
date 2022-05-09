package incr

import (
	"context"
	"sync"
	"testing"
)

func Test_worker(t *testing.T) {
	var didWork bool
	wg := sync.WaitGroup{}
	wg.Add(1)
	var gotObj interface{}
	w := newWorker(func(_ context.Context, obj interface{}) error {
		defer wg.Done()
		didWork = true
		gotObj = obj
		return nil
	})
	go func() { _ = w.Start() }()
	<-w.ReceiveStarted()

	ItsEqual(t, true, w.latch.IsStarted())
	w.work <- "hello"
	wg.Wait()
	ItsEqual(t, "hello", gotObj)
	ItsNil(t, w.Stop())

	ItsEqual(t, false, w.latch.IsStarted())
	ItsEqual(t, true, didWork)
}

func Test_worker_finalizer(t *testing.T) {
	var didWork, didFinalize bool
	wg := sync.WaitGroup{}
	wg.Add(1)
	var gotObj string
	w := newWorker(func(_ context.Context, obj string) error {
		didWork = true
		gotObj = obj
		return nil
	})
	w.finalizer = func(ctx context.Context, iw *worker[string]) error {
		defer wg.Done()
		didFinalize = true
		return nil
	}
	go func() { _ = w.Start() }()
	<-w.ReceiveStarted()
	defer func() { _ = w.Stop() }()

	ItsEqual(t, true, w.latch.IsStarted())
	w.work <- "hello"
	wg.Wait()
	ItsEqual(t, "hello", gotObj)
	ItsEqual(t, true, didWork)
	ItsEqual(t, true, didFinalize)
}

func Test_worker_recoverPanic(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(1)
	w := newWorker(func(_ context.Context, obj interface{}) error {
		defer wg.Done()
		panic("only a test")
	})
	w.errors = make(chan error)
	w.skipRecover = false
	go func() { _ = w.Start() }()
	<-w.ReceiveStarted()
	defer func() { _ = w.Stop() }()

	ItsEqual(t, true, w.latch.IsStarted())
	w.work <- "hello"
	wg.Wait()
	err := <-w.errors
	ItsNotNil(t, err)
}
