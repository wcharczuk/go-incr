package incr

import (
	"context"
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_Join(t *testing.T) {
	ctx := context.Background()
	g := New(OptGraphMaxHeight(64))

	left := Var(g, 10)
	right := Var(g, 20)
	which := Var(g, true)

	// an incremental whose value is one of two other incrementals
	selector := Map(g, which, func(useLeft bool) Incr[int] {
		if useLeft {
			return left
		}
		return right
	})
	joined := Join(g, selector)
	o := MustObserve(g, joined)

	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 10, o.Value())

	// a change to the selected inner incremental propagates
	left.Set(11)
	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 11, o.Value())

	// switching which inner incremental is selected
	which.Set(false)
	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 20, o.Value())

	// the deselected one no longer drives the output
	left.Set(999)
	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 20, o.Value())

	right.Set(21)
	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 21, o.Value())
}

// Test_Join_releasesDeselected checks that an inner incremental stops being needed
// once the outer value moves away from it, which is what makes Join safe to use over
// a large set of alternatives.
func Test_Join_releasesDeselected(t *testing.T) {
	ctx := context.Background()
	g := New(OptGraphMaxHeight(64))

	var leftRecomputes int
	base := Var(g, 1)
	left := Map(g, base, func(x int) int { leftRecomputes++; return x })
	right := Var(g, 100)
	which := Var(g, true)

	selector := Map(g, which, func(useLeft bool) Incr[int] {
		if useLeft {
			return left
		}
		return right
	})
	o := MustObserve(g, Join(g, selector))

	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 1, o.Value())
	testutil.Equal(t, 1, leftRecomputes)

	which.Set(false)
	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 100, o.Value())

	// with left deselected, changing its input must not recompute it
	before := leftRecomputes
	base.Set(2)
	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, before, leftRecomputes, "a deselected inner incremental should not recompute")
}
