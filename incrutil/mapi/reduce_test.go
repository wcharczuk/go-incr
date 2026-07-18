package mapi

import (
	"context"
	"math"
	"math/rand"
	"testing"

	"github.com/wcharczuk/go-incr"
	"github.com/wcharczuk/go-incr/incrutil/pmap"
)

func Test_Reduce_maxAndMin(t *testing.T) {
	ctx := context.Background()
	g := incr.New()

	base := pmap.FromGoMap(map[string]int{"a": 3, "b": 7, "c": 5})
	v := incr.Var(g, base)
	maxValue := incr.MustObserve(g, MaxValue(g, v))
	minValue := incr.MustObserve(g, MinValue(g, v))

	if err := g.Stabilize(ctx); err != nil {
		t.Fatal(err)
	}
	if got := maxValue.Value(); !got.Present || got.Value != 7 {
		t.Fatalf("max %+v, want 7", got)
	}
	if got := minValue.Value(); !got.Present || got.Value != 3 {
		t.Fatalf("min %+v, want 3", got)
	}

	// removing the maximum has to find the next one, which is the case an inverse
	// cannot handle
	v.Set(base.Delete("b"))
	if err := g.Stabilize(ctx); err != nil {
		t.Fatal(err)
	}
	if got := maxValue.Value(); got.Value != 5 {
		t.Fatalf("after removing the maximum, max is %+v, want 5", got)
	}

	// lowering the maximum likewise
	v.Set(base.Set("b", 1))
	if err := g.Stabilize(ctx); err != nil {
		t.Fatal(err)
	}
	if got := maxValue.Value(); got.Value != 5 {
		t.Fatalf("after lowering the maximum, max is %+v, want 5", got)
	}
	if got := minValue.Value(); got.Value != 1 {
		t.Fatalf("min %+v, want 1", got)
	}

	// an empty map reports absence
	v.Set(pmap.New[string, int]())
	if err := g.Stabilize(ctx); err != nil {
		t.Fatal(err)
	}
	if got := maxValue.Value(); got.Present {
		t.Fatalf("empty map reported a maximum: %+v", got)
	}
}

// Test_Reduce_nonCommutative checks that subtrees are combined in key order, so an
// associative but non-commutative operation is well defined.
func Test_Reduce_nonCommutative(t *testing.T) {
	ctx := context.Background()
	g := incr.New()

	base := pmap.New[int, string]()
	for i, part := range []string{"a", "b", "c", "d", "e", "f", "g"} {
		base = base.Set(i, part)
	}
	v := incr.Var(g, base)
	joined := incr.MustObserve(g, Reduce(g, v, "",
		func(_ int, value string) string { return value },
		func(a, b string) string { return a + b }))

	if err := g.Stabilize(ctx); err != nil {
		t.Fatal(err)
	}
	if got := joined.Value(); got != "abcdefg" {
		t.Fatalf("concatenation gave %q, want %q", got, "abcdefg")
	}

	v.Set(base.Set(3, "D"))
	if err := g.Stabilize(ctx); err != nil {
		t.Fatal(err)
	}
	if got := joined.Value(); got != "abcDefg" {
		t.Fatalf("after edit gave %q, want %q", got, "abcDefg")
	}
}

// Test_Reduce_matchesReference cross-checks the memoized fold against folding from
// scratch over many mutations. A stale memo entry would produce a plausible but wrong
// answer rather than an error.
func Test_Reduce_matchesReference(t *testing.T) {
	ctx := context.Background()
	g := incr.New()
	rng := rand.New(rand.NewSource(17))

	current := pmap.New[int, int]()
	v := incr.Var(g, current)
	maxValue := incr.MustObserve(g, MaxValue(g, v))
	if err := g.Stabilize(ctx); err != nil {
		t.Fatal(err)
	}

	for step := range 600 {
		key := rng.Intn(50)
		if rng.Intn(3) == 0 {
			current = current.Delete(key)
		} else {
			current = current.Set(key, rng.Intn(1000))
		}
		v.Set(current)
		if err := g.Stabilize(ctx); err != nil {
			t.Fatal(err)
		}

		want, wantPresent := 0, false
		for _, value := range current.All() {
			if !wantPresent || value > want {
				want, wantPresent = value, true
			}
		}
		got := maxValue.Value()
		if got.Present != wantPresent || (wantPresent && got.Value != want) {
			t.Fatalf("step %d: max %+v, want %d present=%v", step, got, want, wantPresent)
		}
	}
}

// Test_Reduce_work is the claim that matters: a changed key must only re-combine the
// subtrees on its path to the root, so the number of combines per change grows with
// the log of the map rather than with the map.
func Test_Reduce_work(t *testing.T) {
	ctx := context.Background()
	type result struct {
		size     int
		combines int
	}
	results := make([]result, 0, 3)

	for _, size := range []int{1024, 8192, 65536} {
		g := incr.New()
		var base pmap.Map[int, int]
		for k := range size {
			base = base.Set(k, k)
		}
		v := incr.Var(g, base)
		var combines int
		out := incr.MustObserve(g, Reduce(g, v, 0,
			func(_ int, value int) int { return value },
			func(a, b int) int {
				combines++
				return max(a, b)
			}))
		if err := g.Stabilize(ctx); err != nil {
			t.Fatal(err)
		}
		// the first pass necessarily folds the whole map
		if combines < size/2 {
			t.Fatalf("size %d: initial fold made only %d combines", size, combines)
		}

		combines = 0
		current := base.Set(size/2, size*10)
		v.Set(current)
		if err := g.Stabilize(ctx); err != nil {
			t.Fatal(err)
		}
		if out.Value() != size*10 {
			t.Fatalf("size %d: result %d, want %d", size, out.Value(), size*10)
		}
		results = append(results, result{size: size, combines: combines})
		t.Logf("size %6d: one changed key cost %d combines", size, combines)
	}

	// combines per change should track log2(size), which grows by 3 for each 8x in
	// size; the check is that it is not tracking size itself.
	first, last := results[0], results[len(results)-1]
	sizeRatio := float64(last.size) / float64(first.size)
	combineRatio := float64(last.combines) / float64(first.combines)
	exponent := math.Log(combineRatio) / math.Log(sizeRatio)
	t.Logf("combines grew %.2fx for a %.0fx larger map: exponent %.2f", combineRatio, sizeRatio, exponent)
	if exponent > 0.35 {
		t.Errorf("combines per change track map size: exponent %.2f, want roughly logarithmic", exponent)
	}
	// and in absolute terms a change should be tens of combines, not thousands
	if last.combines > 200 {
		t.Errorf("one changed key in a %d entry map cost %d combines", last.size, last.combines)
	}
}

// Test_Reduce_memoIsBounded checks that the memo does not grow without limit as the
// map is updated over and over, since it holds an entry per subtree and nothing
// reports when an old version is discarded.
func Test_Reduce_memoIsBounded(t *testing.T) {
	const size = 512
	var base pmap.Map[int, int]
	for k := range size {
		base = base.Set(k, k)
	}
	reducer := pmap.NewReducer(
		func(_ int, value int) int { return value },
		func(a, b int) int { return max(a, b) })

	// drive the reducer directly so the memo is observable
	current := base
	if _, ok := reducer.Reduce(current); !ok {
		t.Fatal("initial reduce reported empty")
	}
	for i := range 4000 {
		current = current.Set(i%size, i)
		if _, ok := reducer.Reduce(current); !ok {
			t.Fatal("reduce reported empty")
		}
	}
	if got := reducer.MemoLen(); got > 8*size {
		t.Errorf("memo grew to %d entries for a %d entry map", got, size)
	}
}
