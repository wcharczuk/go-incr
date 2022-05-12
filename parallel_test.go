package incr

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func Test_ParallelStabilize(t *testing.T) {
	ctx := testContext()

	v0 := Var("foo")
	v1 := Var("bar")
	m0 := Apply2(v0.Read(), v1.Read(), func(_ context.Context, a, b string) (string, error) {
		return a + " " + b, nil
	})

	err := ParallelStabilize(ctx, m0)
	ItsNil(t, err)

	ItsEqual(t, 0, v0.Node().setAt)
	ItsEqual(t, 1, v0.Node().changedAt)
	ItsEqual(t, 0, v1.Node().setAt)
	ItsEqual(t, 1, v1.Node().changedAt)
	ItsEqual(t, 1, m0.Node().changedAt)
	ItsEqual(t, 1, v0.Node().recomputedAt)
	ItsEqual(t, 1, v1.Node().recomputedAt)
	ItsEqual(t, 1, m0.Node().recomputedAt)

	ItsEqual(t, "foo bar", m0.Value())

	v0.Set("not foo")
	ItsEqual(t, 2, v0.Node().setAt)
	ItsEqual(t, 0, v1.Node().setAt)

	err = ParallelStabilize(ctx, m0)
	ItsNil(t, err)

	ItsEqual(t, 2, v0.Node().changedAt)
	ItsEqual(t, 1, v1.Node().changedAt)
	ItsEqual(t, 2, m0.Node().changedAt)

	ItsEqual(t, 2, v0.Node().recomputedAt)
	ItsEqual(t, 1, v1.Node().recomputedAt)
	ItsEqual(t, 2, m0.Node().recomputedAt)

	ItsEqual(t, "not foo bar", m0.Value())
}

func Test_parallelWorker(t *testing.T) {
	ctx := testContext()

	gotValues := make(chan string)
	var didCallFinalizer bool
	calledFinalizer := make(chan struct{})
	pw := parallelWorker[string]{
		work:    make(chan string),
		errors:  make(chan error, 1),
		ctx:     ctx,
		stop:    make(chan struct{}),
		stopped: make(chan struct{}),
		action: func(ictx context.Context, v string) error {
			itsBlueDye(ictx, t)
			gotValues <- v
			if !strings.HasPrefix(v, "test-value") {
				return errors.New("invalid value")
			}
			return nil
		},
		finalizer: func(ictx context.Context, ipw *parallelWorker[string]) {
			defer close(calledFinalizer)
			itsBlueDye(ictx, t)
			didCallFinalizer = true
		},
	}

	go pw.dispatch()

	pw.work <- "test-value-00"

	ItsEqual(t, "test-value-00", <-gotValues)
	<-calledFinalizer
	ItsEqual(t, true, didCallFinalizer)

	//
	// assert how the worker handles errors
	// specifically that it returns out of the dispatch loop
	//

	pw.work <- "not-test-value-00"
	ItsEqual(t, "not-test-value-00", <-gotValues)
	<-pw.stopped

	ItsEqual(t, "invalid value", (<-pw.errors).Error())
}
