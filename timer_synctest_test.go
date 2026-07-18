package incr

import (
	"context"
	"testing"
	"testing/synctest"
	"time"

	"github.com/wcharczuk/go-incr/testutil"
)

// Test_Timer_underSynctest covers the wall-clock path, which [Clock] deliberately does
// not go through.
//
// [Timer] reads time.Now directly, so it cannot be driven by a [Clock]; before
// testing/synctest the only way to test it was to wait for real time to pass. Inside a
// bubble the clock is virtual and advances instantly once every goroutine is blocked,
// so a sleep costs nothing and the result is deterministic.
//
// This is the complement to clock_test.go rather than a replacement for it: synctest
// makes the real-time path testable, while a Clock additionally gives a stabilization
// a single consistent instant and lets production step time deliberately.
func Test_Timer_underSynctest(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx := context.Background()
		g := New()

		v := Var(g, 1)
		timer := Timer(g, v, 100*time.Millisecond)
		o := MustObserve(g, timer)

		testutil.Nil(t, g.Stabilize(ctx))
		testutil.Equal(t, 1, o.Value())

		// the timer has not elapsed, so a new input value must not come through
		v.Set(2)
		testutil.Nil(t, g.Stabilize(ctx))
		testutil.Equal(t, 1, o.Value(), "before the interval elapses the timer holds its value")

		// virtual time, so this returns immediately
		time.Sleep(150 * time.Millisecond)
		testutil.Nil(t, g.Stabilize(ctx))
		testutil.Equal(t, 2, o.Value(), "after the interval the timer takes the input value")
	})
}

// Test_ParallelStabilize_underSynctest covers the parallel path inside a bubble, where
// synctest.Wait can tell that every worker goroutine has finished rather than the test
// guessing at a sleep.
func Test_ParallelStabilize_underSynctest(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx := context.Background()
		g := New(OptGraphMaxHeight(64))

		v := Var(g, 1)
		fanout := make([]Incr[int], 8)
		for i := range fanout {
			scale := i + 1
			fanout[i] = Map(g, v, func(x int) int { return x * scale })
		}
		root := ReduceBalanced(g, func(a, b int) int { return a + b }, fanout...)
		o := MustObserve(g, root)

		testutil.Nil(t, g.ParallelStabilize(ctx))
		synctest.Wait()
		// 1*(1+2+...+8)
		testutil.Equal(t, 36, o.Value())

		v.Set(2)
		testutil.Nil(t, g.ParallelStabilize(ctx))
		synctest.Wait()
		testutil.Equal(t, 72, o.Value())
	})
}
