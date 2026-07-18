package incr

import (
	"context"
	"testing"
	"time"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_StepFunction(t *testing.T) {
	ctx := context.Background()
	base := time.Date(2026, 4, 1, 9, 0, 0, 0, time.UTC)
	clock := NewClock(base)
	g := New()

	// deliberately out of order, to check they are sorted
	rate := StepFunction(g, clock, 100,
		Step[int]{At: base.Add(2 * time.Hour), Value: 300},
		Step[int]{At: base.Add(time.Hour), Value: 200},
	)
	o := MustObserve(g, rate)

	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 100, o.Value(), "before the first step, the initial value holds")

	clock.AdvanceBy(59 * time.Minute)
	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 100, o.Value())

	clock.AdvanceBy(time.Minute)
	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 200, o.Value(), "a step takes effect at its time")

	clock.AdvanceBy(time.Hour)
	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 300, o.Value())

	// past the last step the value stops changing
	clock.AdvanceBy(24 * time.Hour)
	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 300, o.Value())
}

// Test_StepFunction_jumpsSkipSteps covers a clock that moves past several steps at once,
// which is what happens after a process was busy or asleep. The answer has to be the same
// as if it had advanced through them.
func Test_StepFunction_jumpsSkipSteps(t *testing.T) {
	ctx := context.Background()
	base := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	clock := NewClock(base)
	g := New()

	steps := make([]Step[int], 0, 10)
	for i := 1; i <= 10; i++ {
		steps = append(steps, Step[int]{At: base.Add(time.Duration(i) * time.Minute), Value: i})
	}
	o := MustObserve(g, StepFunction(g, clock, 0, steps...))

	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 0, o.Value())

	// straight past the first seven
	clock.AdvanceBy(7 * time.Minute)
	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 7, o.Value(), "a jump should land on the last step it passed")

	clock.AdvanceBy(time.Hour)
	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 10, o.Value())
}

// Test_StepFunction_wakesOnlyOnBoundaries checks that a step function costs nothing
// between its steps, rather than recomputing on every pass.
func Test_StepFunction_wakesOnlyOnBoundaries(t *testing.T) {
	ctx := context.Background()
	base := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	clock := NewClock(base)
	g := New()

	var recomputes int
	steps := StepFunction(g, clock, 0, Step[int]{At: base.Add(time.Hour), Value: 1})
	MustObserve(g, Map(g, steps, func(v int) int { recomputes++; return v }))

	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 1, recomputes)

	// advancing without crossing the boundary must not recompute
	for range 5 {
		clock.AdvanceBy(time.Minute)
		testutil.Nil(t, g.Stabilize(ctx))
	}
	testutil.Equal(t, 1, recomputes, "no step boundary was crossed")

	clock.AdvanceBy(time.Hour)
	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 2, recomputes)

	// and nothing wakes it again once the steps are exhausted
	for range 5 {
		clock.AdvanceBy(time.Hour)
		testutil.Nil(t, g.Stabilize(ctx))
	}
	testutil.Equal(t, 2, recomputes, "the steps are exhausted, so there is nothing to wake for")
}

// Test_StepFunction_sameTime covers two steps sharing a time, where the later one wins.
func Test_StepFunction_sameTime(t *testing.T) {
	ctx := context.Background()
	base := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	clock := NewClock(base)
	g := New()

	at := base.Add(time.Hour)
	o := MustObserve(g, StepFunction(g, clock, 0,
		Step[int]{At: at, Value: 1},
		Step[int]{At: at, Value: 2},
	))

	clock.AdvanceBy(time.Hour)
	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 2, o.Value())
}
