package incr

import (
	"context"
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_ReduceBalanced(t *testing.T) {
	ctx := context.Background()
	g := New()

	vars := make([]VarIncr[int], 7)
	inputs := make([]Incr[int], 7)
	for i := range vars {
		vars[i] = Var(g, i+1)
		inputs[i] = vars[i]
	}
	// an odd count exercises the trailing input being carried up a level
	root := ReduceBalanced(g, func(a, b int) int { return a + b }, inputs...)
	o := MustObserve(g, root)

	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 28, o.Value())

	vars[3].Set(40)
	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 64, o.Value())
}

// Test_ReduceBalanced_nonCommutative checks that inputs are combined in the order
// given, so that an associative but non-commutative operation is safe.
func Test_ReduceBalanced_nonCommutative(t *testing.T) {
	ctx := context.Background()
	g := New()

	parts := []string{"a", "b", "c", "d", "e"}
	inputs := make([]Incr[string], len(parts))
	for i, p := range parts {
		inputs[i] = Var(g, p)
	}
	root := ReduceBalanced(g, func(a, b string) string { return a + b }, inputs...)
	o := MustObserve(g, root)

	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, "abcde", o.Value())
}

func Test_ReduceBalanced_edgeCases(t *testing.T) {
	ctx := context.Background()
	g := New()

	testutil.Nil(t, ReduceBalanced(g, func(a, b int) int { return a + b }))

	only := Var(g, 9)
	single := ReduceBalanced(g, func(a, b int) int { return a + b }, only)
	o := MustObserve(g, single)
	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 9, o.Value())
}

func Test_UnorderedArrayFold(t *testing.T) {
	ctx := context.Background()
	g := New()

	vars := make([]VarIncr[int], 5)
	inputs := make([]Incr[int], 5)
	for i := range vars {
		vars[i] = Var(g, i+1)
		inputs[i] = vars[i]
	}
	sum := UnorderedArrayFold(g, 0,
		func(acc, v int) int { return acc + v },
		func(acc, old, new int) int { return acc - old + new },
		inputs...)
	o := MustObserve(g, sum)

	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 15, o.Value())

	// a single change is applied as a delta
	vars[2].Set(30)
	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 42, o.Value())

	// several changes in one pass
	vars[0].Set(100)
	vars[4].Set(200)
	testutil.Nil(t, g.Stabilize(ctx))
	// [100, 2, 30, 4, 200]
	testutil.Equal(t, 336, o.Value())

	// setting a var to the value it already holds must not double-count
	vars[0].Set(100)
	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 336, o.Value())
}

// Test_UnorderedArrayFold_matchesFullFold cross-checks the incremental
// accumulator against folding from scratch over many random updates, which is the
// property that actually matters: a wrong update function silently drifts.
func Test_UnorderedArrayFold_matchesFullFold(t *testing.T) {
	ctx := context.Background()
	g := New()

	const n = 24
	vars := make([]VarIncr[int], n)
	inputs := make([]Incr[int], n)
	expected := make([]int, n)
	for i := range vars {
		vars[i] = Var(g, i)
		inputs[i] = vars[i]
		expected[i] = i
	}
	sum := UnorderedArrayFold(g, 0,
		func(acc, v int) int { return acc + v },
		func(acc, old, new int) int { return acc - old + new },
		inputs...)
	o := MustObserve(g, sum)
	testutil.Nil(t, g.Stabilize(ctx))

	// a fixed sequence rather than a random one, so a failure reproduces
	for step := 1; step <= 200; step++ {
		slot := (step * 7) % n
		value := step * 3
		vars[slot].Set(value)
		expected[slot] = value
		testutil.Nil(t, g.Stabilize(ctx))

		var want int
		for _, v := range expected {
			want += v
		}
		testutil.Equal(t, want, o.Value())
	}
}

// Test_UnorderedArrayFold_duplicateInput covers the same input appearing in
// several slots, where each occurrence contributes to the accumulator separately.
func Test_UnorderedArrayFold_duplicateInput(t *testing.T) {
	ctx := context.Background()
	g := New()

	v := Var(g, 5)
	sum := UnorderedArrayFold(g, 0,
		func(acc, x int) int { return acc + x },
		func(acc, old, new int) int { return acc - old + new },
		v, v, v)
	o := MustObserve(g, sum)

	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 15, o.Value())

	v.Set(10)
	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 30, o.Value())
}

// Test_UnorderedArrayFold_overMappedInputs covers inputs that are computed rather
// than vars, so the fold hears about changes propagating from further up.
func Test_UnorderedArrayFold_overMappedInputs(t *testing.T) {
	ctx := context.Background()
	g := New()

	base := Var(g, 1)
	inputs := make([]Incr[int], 4)
	for i := range inputs {
		scale := i + 1
		inputs[i] = Map(g, base, func(x int) int { return x * scale })
	}
	sum := UnorderedArrayFold(g, 0,
		func(acc, x int) int { return acc + x },
		func(acc, old, new int) int { return acc - old + new },
		inputs...)
	o := MustObserve(g, sum)

	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 10, o.Value())

	base.Set(3)
	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 30, o.Value())
}

func Test_All(t *testing.T) {
	ctx := context.Background()
	g := New()

	vars := make([]VarIncr[int], 4)
	inputs := make([]Incr[int], 4)
	for i := range vars {
		vars[i] = Var(g, i)
		inputs[i] = vars[i]
	}
	collected := All(g, inputs...)
	o := MustObserve(g, collected)

	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, []int{0, 1, 2, 3}, o.Value())

	// the returned slice must not alias anything the next pass reuses
	first := o.Value()
	vars[1].Set(99)
	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, []int{0, 99, 2, 3}, o.Value())
	testutil.Equal(t, []int{0, 1, 2, 3}, first)

	testutil.Nil(t, All[int](g))
}
