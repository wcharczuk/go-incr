package incr

import (
	"context"
	"testing"
	"time"

	"github.com/wcharczuk/go-incr/testutil"
)

// Test_VarEqual_noOpSetDoesNothing is the point of the constructor: writing the value a
// var already holds must not reach the graph at all.
func Test_VarEqual_noOpSetDoesNothing(t *testing.T) {
	ctx := context.Background()
	g := New()

	var recomputes int
	v := VarEqual(g, 1)
	o := MustObserve(g, Map(g, v, func(x int) int {
		recomputes++
		return x * 2
	}))

	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 2, o.Value())
	testutil.Equal(t, 1, recomputes)

	// the same value, several times over
	for range 5 {
		v.Set(1)
		testutil.Nil(t, g.Stabilize(ctx))
	}
	testutil.Equal(t, 1, recomputes, "setting the held value should not recompute anything")

	// a real change still propagates
	v.Set(2)
	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 4, o.Value())
	testutil.Equal(t, 2, recomputes)

	// and back to a no-op
	v.Set(2)
	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 2, recomputes)
}

// Test_VarEqual_comparedWithPlainVar contrasts the two, so the difference is stated
// rather than implied.
func Test_VarEqual_comparedWithPlainVar(t *testing.T) {
	ctx := context.Background()

	count := func(v VarIncr[int], g *Graph, recomputes *int) int {
		testutil.Nil(t, g.Stabilize(ctx))
		*recomputes = 0
		for range 10 {
			v.Set(7)
			testutil.Nil(t, g.Stabilize(ctx))
		}
		return *recomputes
	}

	plainGraph := New()
	var plainRecomputes int
	plain := Var(plainGraph, 7)
	MustObserve(plainGraph, Map(plainGraph, plain, func(x int) int { plainRecomputes++; return x }))

	equalGraph := New()
	var equalRecomputes int
	deduped := VarEqual(equalGraph, 7)
	MustObserve(equalGraph, Map(equalGraph, deduped, func(x int) int { equalRecomputes++; return x }))

	testutil.Equal(t, 10, count(plain, plainGraph, &plainRecomputes),
		"a plain var propagates every write")
	testutil.Equal(t, 0, count(deduped, equalGraph, &equalRecomputes),
		"VarEqual propagates none of them")
}

// Test_VarEqualFunc covers a type with no ==.
func Test_VarEqualFunc(t *testing.T) {
	ctx := context.Background()
	g := New()

	var recomputes int
	v := VarEqualFunc(g, []int{1, 2, 3}, func(a, b []int) bool {
		if len(a) != len(b) {
			return false
		}
		for i := range a {
			if a[i] != b[i] {
				return false
			}
		}
		return true
	})
	o := MustObserve(g, Map(g, v, func(xs []int) int {
		recomputes++
		total := 0
		for _, x := range xs {
			total += x
		}
		return total
	}))

	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 6, o.Value())
	testutil.Equal(t, 1, recomputes)

	// an equal but distinct slice is not a change
	v.Set([]int{1, 2, 3})
	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 1, recomputes)

	v.Set([]int{1, 2, 4})
	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 7, o.Value())
	testutil.Equal(t, 2, recomputes)
}

// Test_VarEqual_setDuringStabilization checks the case where a var is written from inside
// a stabilization, which is deferred and so cannot be compared against the value the var
// will have by then.
func Test_VarEqual_setDuringStabilization(t *testing.T) {
	ctx := context.Background()
	g := New()

	v := VarEqual(g, 1)
	other := Var(g, 0)
	// a node that writes the var while the graph is stabilizing
	writer := Map(g, other, func(x int) int {
		v.Set(x)
		return x
	})
	MustObserve(g, writer)
	o := MustObserve(g, v)

	testutil.Nil(t, g.Stabilize(ctx))

	other.Set(9)
	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 9, o.Value(), "a write from inside stabilization should still land")
}

// Test_VarEqual_updateComposes checks Update on a deduplicating var: an update that
// computes the same value is not a change either.
func Test_VarEqual_updateComposes(t *testing.T) {
	ctx := context.Background()
	g := New()

	var recomputes int
	v := VarEqual(g, 10)
	o := MustObserve(g, Map(g, v, func(x int) int { recomputes++; return x }))
	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 1, recomputes)

	v.Update(func(current int) int { return current })
	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 1, recomputes, "an update to the same value is not a change")

	v.Update(func(current int) int { return current + 1 })
	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 11, o.Value())
	testutil.Equal(t, 2, recomputes)
}

// Test_VarEqual_cost measures what a no-op write costs in a graph large enough for the
// difference to matter, since that is the reason to reach for this.
func Test_VarEqual_cost(t *testing.T) {
	ctx := context.Background()
	const width = 4096

	build := func(makeVar func(*Graph) VarIncr[int]) (*Graph, VarIncr[int]) {
		g := New(OptGraphMaxHeight(1024))
		v := makeVar(g)
		leaves := make([]Incr[int], width)
		for i := range leaves {
			leaves[i] = Map(g, v, func(x int) int { return x + 1 })
		}
		MustObserve(g, ReduceBalanced(g, func(a, b int) int { return a + b }, leaves...))
		if err := g.Stabilize(ctx); err != nil {
			t.Fatal(err)
		}
		return g, v
	}

	measure := func(g *Graph, v VarIncr[int]) time.Duration {
		const writes = 50
		start := time.Now()
		for range writes {
			v.Set(1)
			if err := g.Stabilize(ctx); err != nil {
				t.Fatal(err)
			}
		}
		return time.Since(start) / writes
	}

	plainGraph, plain := build(func(g *Graph) VarIncr[int] { return Var(g, 1) })
	equalGraph, deduped := build(func(g *Graph) VarIncr[int] { return VarEqual(g, 1) })

	plainCost := measure(plainGraph, plain)
	equalCost := measure(equalGraph, deduped)
	t.Logf("no-op write over %d dependents: Var %v, VarEqual %v", width, plainCost, equalCost)

	if equalCost >= plainCost {
		t.Errorf("VarEqual should make a no-op write cheaper: %v vs %v", equalCost, plainCost)
	}
}
