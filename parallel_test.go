package incr

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func Test_ParallelStabilize(t *testing.T) {
	ctx := testContext()

	v0 := Var("foo")
	v1 := Var("bar")
	m0 := Map2(v0.Read(), v1.Read(), func(a, b string) string {
		return a + " " + b
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

func Test_ParallelStabilize_jsDocs(t *testing.T) {
	ctx := testContext()

	type Entry struct {
		Entry string
		Time  time.Time
	}

	now := time.Date(2022, 05, 04, 12, 11, 10, 9, time.UTC)

	data := []Entry{
		{"0", now},
		{"1", now.Add(time.Second)},
		{"2", now.Add(2 * time.Second)},
		{"3", now.Add(3 * time.Second)},
		{"4", now.Add(4 * time.Second)},
	}

	i := Var(data)
	output := Map(
		i.Read(),
		func(entries []Entry) (output []string) {
			for _, e := range entries {
				if e.Time.Sub(now) > 2*time.Second {
					output = append(output, e.Entry)
				}
			}
			return
		},
	)

	err := ParallelStabilize(
		ctx,
		output,
	)
	ItsNil(t, err)
	ItsEqual(t, 2, len(output.Value()))

	data = append(data, Entry{
		"5", now.Add(5 * time.Second),
	})
	err = ParallelStabilize(
		ctx,
		output,
	)
	ItsNil(t, err)
	ItsEqual(t, 2, len(output.Value()))

	i.Set(data)
	err = ParallelStabilize(
		context.Background(),
		output,
	)
	ItsNil(t, err)
	ItsEqual(t, 3, len(output.Value()))
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
