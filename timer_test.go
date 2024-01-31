package incr

import (
	"context"
	"testing"
	"time"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_Timer(t *testing.T) {
	clock := time.Now()

	timer := Timer(Return(0), 500*time.Millisecond)
	timer.(*timerIncr[int]).clockSource = func(_ context.Context) time.Time {
		return clock
	}

	var counterTimed int
	timed := Map(timer, func(base int) int {
		counterTimed++
		return base + counterTimed + 1
	})
	timed.Node().SetLabel("timed")

	var counterUntimed int
	untimed := Map(Return(0), func(base int) int {
		counterUntimed++
		return base + counterUntimed + 1
	})
	untimed.Node().SetLabel("untimed")

	final := Map2(timed, untimed, func(a, b int) int {
		return a + b
	})
	final.Node().SetLabel("final")

	g := New()
	o := Observe(g, final)

	ctx := testContext()

	err := g.Stabilize(ctx)
	testutil.ItsNil(t, err)
	testutil.ItsEqual(t, 0, timer.Value())
	testutil.ItsEqual(t, 2, timed.Value())
	testutil.ItsEqual(t, 2, untimed.Value())
	testutil.ItsEqual(t, 4, o.Value())

	err = g.Stabilize(ctx)
	testutil.ItsNil(t, err)
	testutil.ItsEqual(t, 4, o.Value())

	clock = clock.Add(time.Second)

	err = g.Stabilize(ctx)
	testutil.ItsNil(t, err)
	testutil.ItsEqual(t, 5, o.Value())
}
