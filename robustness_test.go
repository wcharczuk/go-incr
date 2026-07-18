package incr

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/wcharczuk/go-incr/testutil"
)

// Test_Stabilize_errorIsRetried is the case that motivated recomputeFailed: a node whose
// stabilization fails must be tried again, not silently abandoned holding a stale value.
//
// The stamp that marks a node as recomputed is applied before its function runs, so
// without putting it back a failed node is indistinguishable from one that succeeded --
// it is no longer stale with respect to its inputs and nothing returns it to the heap. A
// transient failure would strand the value permanently while every later pass reported
// success, which is a worse outcome than the original error.
func Test_Stabilize_errorIsRetried(t *testing.T) {
	ctx := context.Background()
	g := New()
	v := Var(g, 1)

	var attempts int
	fail := true
	o := MustObserve(g, MapContext(g, v, func(_ context.Context, x int) (int, error) {
		attempts++
		if fail {
			return 0, errors.New("transient")
		}
		return x * 10, nil
	}))

	testutil.NotNil(t, g.Stabilize(ctx), "the first pass should surface the error")
	testutil.Equal(t, 1, attempts)

	// nothing about the graph changed, so the retry has to come from the node having been
	// returned to the recompute heap rather than from a new write
	fail = false
	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 2, attempts, "a failed node should be tried again on the next pass")
	testutil.Equal(t, 10, o.Value())

	// and once it succeeds it settles, rather than recomputing forever
	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 2, attempts, "a node that succeeded should not be retried")
}

// Test_Stabilize_errorLeavesGraphUsable checks that an error does not corrupt the graph:
// the stabilizing status is released, the structure stays consistent, and later passes
// work normally.
func Test_Stabilize_errorLeavesGraphUsable(t *testing.T) {
	ctx := context.Background()
	g := New()
	v := Var(g, 1)
	fail := true
	o := MustObserve(g, MapContext(g, v, func(_ context.Context, x int) (int, error) {
		if fail {
			return 0, errors.New("nope")
		}
		return x, nil
	}))

	testutil.NotNil(t, g.Stabilize(ctx))
	testutil.Nil(t, ExpertGraph(g).CheckInvariants(), "an error should not corrupt the graph")

	fail = false
	v.Set(2)
	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 2, o.Value())
	testutil.Nil(t, ExpertGraph(g).CheckInvariants())
}

// Test_Stabilize_panicBecomesError covers a panic escaping user code. It is reported as a
// PanicError rather than unwinding through the caller, and the node is retried like any
// other failure.
func Test_Stabilize_panicBecomesError(t *testing.T) {
	ctx := context.Background()
	g := New()
	v := Var(g, 1)

	var attempts int
	boom := true
	var handled error
	m := Map(g, v, func(x int) int {
		attempts++
		if boom {
			panic("user code panicked")
		}
		return x
	})
	m.Node().OnError(func(_ context.Context, err error) { handled = err })
	o := MustObserve(g, m)

	err := g.Stabilize(ctx)
	testutil.NotNil(t, err, "a panic should be reported as an error")
	testutil.Equal(t, 1, attempts)

	var panicErr *PanicError
	testutil.Equal(t, true, errors.As(err, &panicErr), "the error should be a PanicError")
	testutil.Equal(t, "user code panicked", panicErr.Value)
	testutil.Equal(t, true, len(panicErr.Stack) > 0, "the stack should be kept")
	testutil.Equal(t, true, strings.Contains(panicErr.Error(), "user code panicked"))
	testutil.NotNil(t, handled, "the node's error handler should see it")

	// the graph is not left marked as stabilizing, and the node is retried
	boom = false
	testutil.Nil(t, g.Stabilize(ctx), "a panic should not leave the graph stabilizing")
	testutil.Equal(t, 2, attempts, "a node that panicked should be tried again")
	testutil.Equal(t, 1, o.Value())
	testutil.Nil(t, ExpertGraph(g).CheckInvariants())
}

// Test_Stabilize_panicUnwrapsError covers panicking with an error value, which errors.Is
// should see through.
func Test_Stabilize_panicUnwrapsError(t *testing.T) {
	ctx := context.Background()
	g := New()
	v := Var(g, 1)
	sentinel := errors.New("the underlying cause")
	MustObserve(g, Map(g, v, func(x int) int { panic(sentinel) }))

	err := g.Stabilize(ctx)
	testutil.Equal(t, true, errors.Is(err, sentinel), "a panicked error should be unwrappable")
}

// Test_ParallelStabilize_panicBecomesError is the case that cannot be handled by the
// caller at all: a panic in a worker goroutine ends the process, since a recover in the
// caller never sees it. It has to be caught where the work runs.
func Test_ParallelStabilize_panicBecomesError(t *testing.T) {
	ctx := context.Background()
	g := New(OptGraphMaxHeight(64))

	inputs := make([]Incr[int], 0, 16)
	for i := range 16 {
		v := Var(g, i)
		inputs = append(inputs, Map(g, v, func(x int) int {
			if x == 7 {
				panic("a worker panicked")
			}
			return x
		}))
	}
	MustObserve(g, ReduceBalanced(g, func(a, b int) int { return a + b }, inputs...))

	err := g.ParallelStabilize(ctx)
	testutil.NotNil(t, err, "a panic in a worker should be reported, not fatal")
	var panicErr *PanicError
	testutil.Equal(t, true, errors.As(err, &panicErr))
	testutil.Nil(t, ExpertGraph(g).CheckInvariants())
}

// Test_Stabilize_contextCancellation checks that a pass can be interrupted, and that what
// it leaves behind is resumable rather than half applied and forgotten.
func Test_Stabilize_contextCancellation(t *testing.T) {
	ctx := context.Background()

	newDeepGraph := func() (*Graph, ObserveIncr[int]) {
		g := New(OptGraphMaxHeight(4096))
		v := Var(g, 0)
		var cur Incr[int] = v
		for range 2000 {
			cur = Map(g, cur, func(x int) int { return x + 1 })
		}
		return g, MustObserve(g, cur)
	}

	// an already canceled context should do no work at all
	g, o := newDeepGraph()
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	testutil.Equal(t, context.Canceled, g.Stabilize(canceled))
	testutil.Equal(t, 0, o.Value(), "a canceled pass should not have computed anything")

	// the work is still pending, so stabilizing with a live context finishes it
	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 2000, o.Value(), "a canceled pass should be resumable")
	testutil.Nil(t, ExpertGraph(g).CheckInvariants())

	// a context that cannot be canceled is unaffected
	g2, o2 := newDeepGraph()
	testutil.Nil(t, g2.Stabilize(context.Background()))
	testutil.Equal(t, 2000, o2.Value())

	// and the cause is reported, rather than a bare context.Canceled
	g3, _ := newDeepGraph()
	cause := errors.New("shutting down")
	withCause, cancelCause := context.WithCancelCause(ctx)
	cancelCause(cause)
	testutil.Equal(t, cause, g3.Stabilize(withCause), "the cancellation cause should be reported")
}

// Test_ParallelStabilize_contextCancellation is the same for the parallel path, where
// cancellation lands on height block boundaries.
func Test_ParallelStabilize_contextCancellation(t *testing.T) {
	g := New(OptGraphMaxHeight(4096))
	v := Var(g, 0)
	var cur Incr[int] = v
	for range 500 {
		cur = Map(g, cur, func(x int) int { return x + 1 })
	}
	o := MustObserve(g, cur)

	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	testutil.Equal(t, context.Canceled, g.ParallelStabilize(canceled))
	testutil.Equal(t, 0, o.Value())

	testutil.Nil(t, g.ParallelStabilize(context.Background()))
	testutil.Equal(t, 500, o.Value(), "a canceled parallel pass should be resumable")
	testutil.Nil(t, ExpertGraph(g).CheckInvariants())
}

// Test_Graph_drainsWhenUnobserved states the property the fuzzer checks on every random
// program: releasing everything leaves nothing behind. A leak here is invisible through
// values, which is how an unbounded one survived in bind relinking for as long as it did.
func Test_Graph_drainsWhenUnobserved(t *testing.T) {
	ctx := context.Background()
	g := New(OptGraphMaxHeight(256))
	sw := Var(g, 1)
	o := MustObserve(g, Bind(g, sw, func(bs Scope, which int) Incr[int] {
		return Map(bs, Return(bs, which), func(x int) int { return x * 2 })
	}))

	// rebuild the right hand side many times over, which is where edges accumulated
	for i := range 100 {
		sw.Set(i)
		testutil.Nil(t, g.Stabilize(ctx))
	}
	testutil.Nil(t, ExpertGraph(g).CheckInvariants())

	o.Unobserve(ctx)
	testutil.Nil(t, g.Stabilize(ctx))

	testutil.Equal(t, 0, g.numNodes, "unobserving everything should drain the graph")
	testutil.Equal(t, 0, g.recomputeHeap.len())
	testutil.Equal(t, 0, len(g.observers))
	testutil.Equal(t, 0, len(sw.Node().parents), "the switch should keep no dependents")
	testutil.Nil(t, ExpertGraph(g).CheckInvariants())
}

// Test_Clock_concurrentReads covers reading a clock while it is advanced, which is the
// shape of a clock driven by a ticker goroutine while a graph consults it.
func Test_Clock_concurrentReads(t *testing.T) {
	clock := NewClock(time.Now())

	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 500 {
				_ = clock.Now()
			}
		}()
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for range 500 {
			clock.AdvanceBy(time.Millisecond)
		}
	}()
	wg.Wait()
}

// Test_ParallelStabilize_withBinds exercises the parallel path over a graph whose shape
// changes underneath it, which is where the locking around structural work matters. Run
// under -race this is the check that binds and parallel stabilization compose.
func Test_ParallelStabilize_withBinds(t *testing.T) {
	ctx := context.Background()
	g := New(OptGraphMaxHeight(256))

	switches := make([]VarIncr[int], 0, 16)
	for i := range 16 {
		sw := Var(g, i%4)
		switches = append(switches, sw)
		MustObserve(g, Bind(g, sw, func(bs Scope, which int) Incr[int] {
			leaves := make([]Incr[int], 0, 8)
			for j := range 8 {
				leaves = append(leaves, Map(bs, Return(bs, j*which), func(x int) int { return x + 1 }))
			}
			return ReduceBalanced(bs, func(a, b int) int { return a + b }, leaves...)
		}))
	}

	for round := range 20 {
		for i, sw := range switches {
			sw.Set((round + i) % 5)
		}
		testutil.Nil(t, g.ParallelStabilize(ctx))
		testutil.Nil(t, ExpertGraph(g).CheckInvariants())
	}
}

// Test_Var_concurrentSet covers several goroutines writing different vars between passes,
// which is how a graph fed by concurrent producers is driven.
func Test_Var_concurrentSet(t *testing.T) {
	ctx := context.Background()
	g := New()

	vars := make([]VarIncr[int], 0, 32)
	inputs := make([]Incr[int], 0, 32)
	for i := range 32 {
		v := Var(g, i)
		vars = append(vars, v)
		inputs = append(inputs, v)
	}
	o := MustObserve(g, ReduceBalanced(g, func(a, b int) int { return a + b }, inputs...))
	testutil.Nil(t, g.Stabilize(ctx))

	for range 20 {
		var wg sync.WaitGroup
		for i, v := range vars {
			wg.Add(1)
			go func(i int, v VarIncr[int]) {
				defer wg.Done()
				v.Set(i * 2)
			}(i, v)
		}
		wg.Wait()
		testutil.Nil(t, g.Stabilize(ctx))
	}

	// every var was last written i*2, so the sum is fixed regardless of interleaving
	want := 0
	for i := range 32 {
		want += i * 2
	}
	testutil.Equal(t, want, o.Value())
}

// Test_Graph_maxHeightIsReported checks that exceeding the height limit is an error a
// caller can act on rather than a panic or a corrupt graph.
func Test_Graph_maxHeightIsReported(t *testing.T) {
	g := New(OptGraphMaxHeight(8))
	v := Var(g, 0)
	var cur Incr[int] = v
	for range 20 {
		cur = Map(g, cur, func(x int) int { return x + 1 })
	}

	_, err := Observe(g, cur)
	testutil.NotNil(t, err, "exceeding the height limit should be an error")
}

// Test_PanicError_withoutNode covers a panic that did not come from a node's computation.
// There is nothing to retry or to attribute it to, but it must still be reported rather
// than swallowed, and formatting it must not assume a node is present.
func Test_PanicError_withoutNode(t *testing.T) {
	err := newPanicError(nil, "something in the graph itself")
	testutil.Equal(t, "", err.Node)
	testutil.Equal(t, true, len(err.Stack) > 0)
	testutil.Equal(t, true, strings.Contains(err.Error(), "something in the graph itself"))
	testutil.Equal(t, false, strings.Contains(err.Error(), "node  panicked"))
}

// Test_OnInvalidated_firesOnce covers a node reached by invalidation more than once, which
// happens when a bind's right-hand side is shared or when teardown and the bind's own
// invalidation walk both reach the same node. The handler reports a transition, so firing
// it twice for one transition would double-release whatever it was holding.
func Test_OnInvalidated_firesOnce(t *testing.T) {
	ctx := context.Background()
	g := New(OptGraphMaxHeight(64))

	var fired int
	v := Var(g, 1)
	m := Map(g, v, func(x int) int { return x })
	m.Node().OnInvalidated(func() { fired++ })
	MustObserve(g, m)
	testutil.Nil(t, g.Stabilize(ctx))

	// invalidate the same node twice over; the transition happens once
	ExpertGraph(g).InvalidateNode(m)
	testutil.Equal(t, 1, fired, "the first invalidation should report the transition")
	ExpertGraph(g).InvalidateNode(m)
	testutil.Equal(t, 1, fired, "an already invalid node should not report it again")
}

// Test_necessaryDependentEdges covers the invariant that a dependent edge exists only
// while the dependent is necessary.
//
// It is the property that makes releasing part of a graph work. Necessity is derived from
// having dependents, so an edge pointing at an unnecessary node both keeps that node
// reachable and props up the necessity of everything beneath it, and nothing will ever
// remove it. Jane Street's incremental states the same invariant directly.
//
// The shapes below are the ones where it could plausibly break: a shared input whose
// dependents are released one at a time, a node held by two observers, and a bind whose
// discarded right-hand side reads an input that outlives it.
func Test_necessaryDependentEdges(t *testing.T) {
	ctx := context.Background()

	t.Run("sharedInputPartiallyReleased", func(t *testing.T) {
		g := New(OptGraphMaxHeight(64))
		shared := Var(g, 1)
		left := Map(g, shared, func(x int) int { return x + 1 })
		right := Map(g, shared, func(x int) int { return x * 2 })

		ol := MustObserve(g, left)
		or := MustObserve(g, right)
		testutil.Nil(t, g.Stabilize(ctx))
		testutil.Equal(t, 2, len(shared.Node().children), "both dependents are necessary")

		or.Unobserve(ctx)
		testutil.Nil(t, g.Stabilize(ctx))
		testutil.Equal(t, 1, len(shared.Node().children),
			"the released dependent should no longer be recorded on its input")
		testutil.Nil(t, ExpertGraph(g).CheckInvariants())

		ol.Unobserve(ctx)
		testutil.Nil(t, g.Stabilize(ctx))
		testutil.Equal(t, 0, len(shared.Node().children))
		testutil.Nil(t, ExpertGraph(g).CheckInvariants())
	})

	t.Run("twoObserversOneRemoved", func(t *testing.T) {
		g := New(OptGraphMaxHeight(64))
		v := Var(g, 1)
		m := Map(g, v, func(x int) int { return x + 1 })
		first := MustObserve(g, m)
		second := MustObserve(g, m)
		testutil.Nil(t, g.Stabilize(ctx))

		// still necessary through the other observer, so the edge must stay
		first.Unobserve(ctx)
		testutil.Nil(t, g.Stabilize(ctx))
		testutil.Equal(t, 1, len(v.Node().children), "still necessary through the second observer")
		testutil.Nil(t, ExpertGraph(g).CheckInvariants())

		second.Unobserve(ctx)
		testutil.Nil(t, g.Stabilize(ctx))
		testutil.Equal(t, 0, len(v.Node().children))
		testutil.Nil(t, ExpertGraph(g).CheckInvariants())
	})

	t.Run("bindRightHandSideReadingAnOuterInput", func(t *testing.T) {
		g := New(OptGraphMaxHeight(64))
		outer := Var(g, 100)
		sw := Var(g, 0)

		// the right-hand side depends on a node that outlives every rebuild, so a
		// discarded rhs is exactly the case where a stale dependent edge would be kept
		o := MustObserve(g, Bind(g, sw, func(bs Scope, which int) Incr[int] {
			return Map2(bs, outer, Return(bs, which), func(a, b int) int { return a + b })
		}))

		for i := range 20 {
			sw.Set(i)
			testutil.Nil(t, g.Stabilize(ctx))
			testutil.Nil(t, ExpertGraph(g).CheckInvariants())
			testutil.Equal(t, 1, len(outer.Node().children),
				"each rebuild should leave exactly one dependent on the outer input")
		}

		o.Unobserve(ctx)
		testutil.Nil(t, g.Stabilize(ctx))
		testutil.Equal(t, 0, len(outer.Node().children),
			"releasing the bind should leave no dependents on the outer input")
		testutil.Nil(t, ExpertGraph(g).CheckInvariants())
	})
}
