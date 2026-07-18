package incr

import (
	"context"
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_ArrayFold(t *testing.T) {
	ctx := context.Background()
	g := New()

	vars := make([]VarIncr[int], 4)
	inputs := make([]Incr[int], 4)
	for i := range vars {
		vars[i] = Var(g, i+1)
		inputs[i] = vars[i]
	}
	// a non-commutative fold, so the ordering is observable
	folded := ArrayFold(g, "", func(acc string, value int) string {
		return acc + string(rune('0'+value))
	}, inputs...)
	o := MustObserve(g, folded)

	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, "1234", o.Value())

	vars[1].Set(9)
	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, "1934", o.Value())
}

func Test_ForAll_Exists(t *testing.T) {
	ctx := context.Background()
	g := New()

	flags := make([]VarIncr[bool], 4)
	inputs := make([]Incr[bool], 4)
	for i := range flags {
		flags[i] = Var(g, true)
		inputs[i] = flags[i]
	}
	all := MustObserve(g, ForAll(g, inputs...))
	any := MustObserve(g, Exists(g, inputs...))

	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, true, all.Value())
	testutil.Equal(t, true, any.Value())

	flags[2].Set(false)
	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, false, all.Value())
	testutil.Equal(t, true, any.Value())

	for _, f := range flags {
		f.Set(false)
	}
	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, false, all.Value())
	testutil.Equal(t, false, any.Value())

	// the empty cases follow the identities of the operations
	testutil.Nil(t, g.Stabilize(ctx))
	emptyAll := MustObserve(g, ForAll(g))
	emptyAny := MustObserve(g, Exists(g))
	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, true, emptyAll.Value())
	testutil.Equal(t, false, emptyAny.Value())
}

func Test_DependOn(t *testing.T) {
	ctx := context.Background()
	g := New()

	value := Var(g, 1)
	var sideEffects int
	dependency := Map(g, Var(g, 0), func(x int) int { sideEffects++; return x })

	// the result takes value's value, but dependency is necessary and drives recomputes
	combined := DependOn(g, value, dependency)
	o := MustObserve(g, combined)

	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 1, o.Value())
	testutil.Equal(t, 1, sideEffects, "the dependency should be necessary and have run")

	value.Set(2)
	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 2, o.Value(), "the value still comes from the input")
}

func Test_namedCutoffs(t *testing.T) {
	ctx := context.Background()
	g := New()

	v := Var(g, 1)
	var alwaysSeen, equalSeen, neverSeen int
	MustObserve(g, Map(g, CutoffAlways(g, v), func(x int) int { alwaysSeen++; return x }))
	MustObserve(g, Map(g, CutoffEqual(g, v), func(x int) int { equalSeen++; return x }))
	MustObserve(g, Map(g, CutoffNever(g, v), func(x int) int { neverSeen++; return x }))

	testutil.Nil(t, g.Stabilize(ctx))
	// the first pass has no previous value to compare, so everything propagates
	testutil.Equal(t, 1, alwaysSeen)
	testutil.Equal(t, 1, equalSeen)
	testutil.Equal(t, 1, neverSeen)

	// a genuine change: always still blocks, equal and never let it through
	v.Set(2)
	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 1, alwaysSeen, "CutoffAlways should block a real change")
	testutil.Equal(t, 2, equalSeen)
	testutil.Equal(t, 2, neverSeen)

	// setting the same value: equal now blocks too, never does not
	v.Set(2)
	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 1, alwaysSeen)
	testutil.Equal(t, 2, equalSeen, "CutoffEqual should block an unchanged value")
	testutil.Equal(t, 3, neverSeen, "CutoffNever should propagate regardless")
}

func Test_Var_Update(t *testing.T) {
	ctx := context.Background()
	g := New()

	v := Var(g, 10)
	o := MustObserve(g, Map(g, v, func(x int) int { return x * 2 }))
	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 20, o.Value())

	v.Update(func(current int) int { return current + 5 })
	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 30, o.Value())

	// successive updates compose rather than overwrite
	v.Update(func(current int) int { return current * 2 })
	v.Update(func(current int) int { return current + 1 })
	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 62, o.Value())
}
