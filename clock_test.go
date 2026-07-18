package incr

import (
	"context"
	"testing"
	"time"

	"github.com/wcharczuk/go-incr/testutil"
)

// epoch is a fixed instant, so that every expectation here is exact rather than
// relative to when the test ran.
var epoch = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

func Test_Clock_At(t *testing.T) {
	ctx := context.Background()
	g := New()
	clock := NewClock(epoch)

	fired := At(g, clock, epoch.Add(time.Minute))
	o := MustObserve(g, fired)

	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, false, o.Value(), "before the time, At is false")

	// short of the trigger
	clock.AdvanceBy(30 * time.Second)
	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, false, o.Value())

	// exactly at the trigger counts as reached
	clock.AdvanceBy(30 * time.Second)
	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, true, o.Value(), "at the time, At is true")

	// and stays true
	clock.AdvanceBy(time.Hour)
	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, true, o.Value())
}

// Test_Clock_At_jumpPast covers advancing straight past a trigger, which is what
// happens when a process was busy or asleep.
func Test_Clock_At_jumpPast(t *testing.T) {
	ctx := context.Background()
	g := New()
	clock := NewClock(epoch)

	o := MustObserve(g, At(g, clock, epoch.Add(time.Minute)))
	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, false, o.Value())

	clock.AdvanceBy(24 * time.Hour)
	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, true, o.Value())
}

func Test_Clock_AtIntervals(t *testing.T) {
	ctx := context.Background()
	g := New()
	clock := NewClock(epoch)

	ticks := AtIntervals(g, clock, time.Minute)
	o := MustObserve(g, ticks)

	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 0, o.Value())

	// a partial interval does not count
	clock.AdvanceBy(59 * time.Second)
	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 0, o.Value())

	clock.AdvanceBy(time.Second)
	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 1, o.Value())

	// several intervals passing between stabilizations advances by several, so the
	// count tracks elapsed time rather than how often the graph ran
	clock.AdvanceBy(5 * time.Minute)
	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 6, o.Value())
}

// Test_Clock_AtIntervals_drivesDependents checks that dependents recompute as
// intervals elapse, which is the point of the node.
func Test_Clock_AtIntervals_drivesDependents(t *testing.T) {
	ctx := context.Background()
	g := New()
	clock := NewClock(epoch)

	var recomputes int
	ticks := AtIntervals(g, clock, time.Minute)
	watcher := Map(g, ticks, func(count int) int {
		recomputes++
		return count * 10
	})
	o := MustObserve(g, watcher)

	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 1, recomputes)
	testutil.Equal(t, 0, o.Value())

	// a stabilization with no time elapsed must not recompute the dependent
	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 1, recomputes, "stabilizing without advancing should do nothing")

	clock.AdvanceBy(time.Minute)
	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 2, recomputes)
	testutil.Equal(t, 10, o.Value())
}

func Test_Clock_Snapshot(t *testing.T) {
	ctx := context.Background()
	g := New()
	clock := NewClock(epoch)

	v := Var(g, 1)
	captured := Snapshot(g, clock, v, epoch.Add(time.Minute), -1)
	o := MustObserve(g, captured)

	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, -1, o.Value(), "before the time, the snapshot holds its default")

	// the input moving before the capture time does not matter
	v.Set(2)
	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, -1, o.Value())

	// at the capture time it takes whatever the input holds then
	v.Set(3)
	clock.AdvanceBy(time.Minute)
	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 3, o.Value())

	// and does not follow the input afterwards, which is what makes it a snapshot
	v.Set(4)
	clock.AdvanceBy(time.Hour)
	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 3, o.Value(), "a captured snapshot must not follow its input")
}

// Test_Clock_noAdvanceNoWork checks that the clock only wakes nodes whose trigger has
// passed, rather than everything registered.
func Test_Clock_noAdvanceNoWork(t *testing.T) {
	ctx := context.Background()
	g := New()
	clock := NewClock(epoch)

	var early, late int
	earlyNode := Map(g, At(g, clock, epoch.Add(time.Minute)), func(v bool) bool { early++; return v })
	lateNode := Map(g, At(g, clock, epoch.Add(time.Hour)), func(v bool) bool { late++; return v })
	MustObserve(g, earlyNode)
	MustObserve(g, lateNode)

	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 1, early)
	testutil.Equal(t, 1, late)

	// reaching only the first trigger must not recompute the second
	clock.AdvanceBy(time.Minute)
	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 2, early)
	testutil.Equal(t, 1, late, "a node whose time has not come should not recompute")
}

// Test_Clock_rejectsRewind covers advancing backwards, which would otherwise let a
// node that has already fired report that it had not.
func Test_Clock_rejectsRewind(t *testing.T) {
	ctx := context.Background()
	g := New()
	clock := NewClock(epoch)

	o := MustObserve(g, At(g, clock, epoch.Add(time.Minute)))
	clock.AdvanceBy(time.Minute)
	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, true, o.Value())

	clock.Advance(epoch)
	testutil.Equal(t, epoch.Add(time.Minute), clock.Now(), "the clock should not move backwards")
	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, true, o.Value())
}

// Test_Clock_deterministic is the motivating property: the same sequence of advances
// produces the same results every run, with no dependence on real elapsed time.
func Test_Clock_deterministic(t *testing.T) {
	ctx := context.Background()
	run := func() []int {
		g := New()
		clock := NewClock(epoch)
		o := MustObserve(g, AtIntervals(g, clock, 10*time.Second))
		var out []int
		for range 5 {
			clock.AdvanceBy(15 * time.Second)
			if err := g.Stabilize(ctx); err != nil {
				t.Fatal(err)
			}
			out = append(out, o.Value())
		}
		return out
	}
	first := run()
	for range 3 {
		next := run()
		for i := range first {
			if first[i] != next[i] {
				t.Fatalf("runs differ at step %d: %v then %v", i, first, next)
			}
		}
	}
	testutil.Equal(t, []int{1, 3, 4, 6, 7}, first)
}
