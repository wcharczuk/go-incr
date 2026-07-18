package mapi

import (
	"context"
	"math"
	"math/rand"
	"testing"
	"time"

	"github.com/wcharczuk/go-incr"
	"github.com/wcharczuk/go-incr/incrutil/pmap"
)

func Test_FilterMapValues(t *testing.T) {
	ctx := context.Background()
	g := incr.New()

	base := pmap.FromGoMap(map[string]int{"a": 1, "b": 2, "c": 3, "d": 4})
	v := incr.Var(g, base)
	// keep even values only, doubled
	evens := FilterMapValues(g, v, intsEqual, func(_ string, value int) (int, bool) {
		return value * 2, value%2 == 0
	})
	o := incr.MustObserve(g, evens)

	if err := g.Stabilize(ctx); err != nil {
		t.Fatal(err)
	}
	got := pmap.ToGoMap(o.Value())
	if len(got) != 2 || got["b"] != 4 || got["d"] != 8 {
		t.Fatalf("initial pass gave %v", got)
	}

	// a value flipping from even to odd must leave the output
	v.Set(base.Set("b", 3))
	if err := g.Stabilize(ctx); err != nil {
		t.Fatal(err)
	}
	if got := pmap.ToGoMap(o.Value()); len(got) != 1 || got["d"] != 8 {
		t.Fatalf("after b became odd, got %v", got)
	}

	// and flipping back must bring it in again
	v.Set(base.Set("b", 10))
	if err := g.Stabilize(ctx); err != nil {
		t.Fatal(err)
	}
	if got := pmap.ToGoMap(o.Value()); len(got) != 2 || got["b"] != 20 {
		t.Fatalf("after b became even again, got %v", got)
	}
}

func Test_UnorderedFold(t *testing.T) {
	ctx := context.Background()
	g := incr.New()

	base := pmap.FromGoMap(map[string]int{"a": 1, "b": 2, "c": 3})
	v := incr.Var(g, base)
	sum := UnorderedFold(g, v, 0, intsEqual,
		func(acc int, _ string, value int) int { return acc + value },
		func(acc int, _ string, value int) int { return acc - value })
	o := incr.MustObserve(g, sum)

	if err := g.Stabilize(ctx); err != nil {
		t.Fatal(err)
	}
	if o.Value() != 6 {
		t.Fatalf("initial sum %d, want 6", o.Value())
	}

	v.Set(base.Set("b", 20))
	if err := g.Stabilize(ctx); err != nil {
		t.Fatal(err)
	}
	if o.Value() != 24 {
		t.Fatalf("after rebinding b, sum %d, want 24", o.Value())
	}

	v.Set(base.Delete("a").Set("d", 100))
	if err := g.Stabilize(ctx); err != nil {
		t.Fatal(err)
	}
	if o.Value() != 105 {
		t.Fatalf("after remove and add, sum %d, want 105", o.Value())
	}
}

// Test_UnorderedFold_matchesFullFold cross-checks the maintained accumulator
// against summing from scratch over many mutations. A wrong add or remove drifts
// silently, so this is the property that matters.
func Test_UnorderedFold_matchesFullFold(t *testing.T) {
	ctx := context.Background()
	g := incr.New()
	rng := rand.New(rand.NewSource(3))

	current := pmap.New[int, int]()
	v := incr.Var(g, current)
	sum := UnorderedFold(g, v, 0, intsEqual,
		func(acc int, _ int, value int) int { return acc + value },
		func(acc int, _ int, value int) int { return acc - value })
	o := incr.MustObserve(g, sum)
	if err := g.Stabilize(ctx); err != nil {
		t.Fatal(err)
	}

	for step := range 500 {
		key := rng.Intn(40)
		if rng.Intn(4) == 0 {
			current = current.Delete(key)
		} else {
			current = current.Set(key, rng.Intn(100))
		}
		v.Set(current)
		if err := g.Stabilize(ctx); err != nil {
			t.Fatal(err)
		}
		var want int
		for _, value := range current.All() {
			want += value
		}
		if o.Value() != want {
			t.Fatalf("step %d: maintained sum %d, full fold %d", step, o.Value(), want)
		}
	}
}

func Test_Merge(t *testing.T) {
	ctx := context.Background()
	g := incr.New()

	left := pmap.FromGoMap(map[string]int{"a": 1, "both": 10})
	right := pmap.FromGoMap(map[string]int{"b": 2, "both": 20})
	lv := incr.Var(g, left)
	rv := incr.Var(g, right)

	merged := Merge(g, lv, rv, intsEqual, intsEqual,
		func(_ string, element MergeElement[int, int]) (int, bool) {
			switch {
			case element.HasLeft && element.HasRight:
				return element.Left + element.Right, true
			case element.HasLeft:
				return element.Left, true
			default:
				return element.Right, true
			}
		})
	o := incr.MustObserve(g, merged)

	if err := g.Stabilize(ctx); err != nil {
		t.Fatal(err)
	}
	got := pmap.ToGoMap(o.Value())
	want := map[string]int{"a": 1, "b": 2, "both": 30}
	if len(got) != len(want) {
		t.Fatalf("initial merge gave %v, want %v", got, want)
	}
	for key, wantValue := range want {
		if got[key] != wantValue {
			t.Fatalf("initial merge %q = %d, want %d", key, got[key], wantValue)
		}
	}

	// a change on one side only
	lv.Set(left.Set("both", 100))
	if err := g.Stabilize(ctx); err != nil {
		t.Fatal(err)
	}
	if got := pmap.ToGoMap(o.Value()); got["both"] != 120 {
		t.Fatalf("after left change, both = %d, want 120", got["both"])
	}

	// a key leaving one side falls back to the other
	lv.Set(left.Delete("both"))
	if err := g.Stabilize(ctx); err != nil {
		t.Fatal(err)
	}
	if got := pmap.ToGoMap(o.Value()); got["both"] != 20 {
		t.Fatalf("after left removal, both = %d, want 20", got["both"])
	}

	// a key leaving both sides leaves the output
	rv.Set(right.Delete("both"))
	if err := g.Stabilize(ctx); err != nil {
		t.Fatal(err)
	}
	if _, ok := o.Value().Get("both"); ok {
		t.Fatal("key present after leaving both inputs")
	}
}

// Test_Merge_matchesFullMerge cross-checks against merging from scratch while both
// inputs are mutated independently.
func Test_Merge_matchesFullMerge(t *testing.T) {
	ctx := context.Background()
	g := incr.New()
	rng := rand.New(rand.NewSource(11))

	left, right := pmap.New[int, int](), pmap.New[int, int]()
	lv, rv := incr.Var(g, left), incr.Var(g, right)
	merged := Merge(g, lv, rv, intsEqual, intsEqual,
		func(_ int, element MergeElement[int, int]) (int, bool) {
			// exclude keys present on the right only, to exercise the filtering path
			if !element.HasLeft {
				return 0, false
			}
			return element.Left*1000 + element.Right, true
		})
	o := incr.MustObserve(g, merged)
	if err := g.Stabilize(ctx); err != nil {
		t.Fatal(err)
	}

	for step := range 400 {
		key := rng.Intn(30)
		switch rng.Intn(4) {
		case 0:
			left = left.Delete(key)
		case 1:
			right = right.Delete(key)
		case 2:
			left = left.Set(key, rng.Intn(9)+1)
		default:
			right = right.Set(key, rng.Intn(9)+1)
		}
		lv.Set(left)
		rv.Set(right)
		if err := g.Stabilize(ctx); err != nil {
			t.Fatal(err)
		}

		want := map[int]int{}
		for key, leftValue := range left.All() {
			rightValue, _ := right.Get(key)
			want[key] = leftValue*1000 + rightValue
		}
		got := pmap.ToGoMap(o.Value())
		if len(got) != len(want) {
			t.Fatalf("step %d: got %d entries, want %d", step, len(got), len(want))
		}
		for key, wantValue := range want {
			if got[key] != wantValue {
				t.Fatalf("step %d: key %d = %d, want %d", step, key, got[key], wantValue)
			}
		}
	}
}

// Test_operators_work asserts the exact thing these operators promise: the number
// of per-key callbacks for one changed key is constant, whatever the map's size.
//
// This is the algorithmic check, and it is not subject to the memory effects that
// muddy the timing check below.
func Test_operators_work(t *testing.T) {
	ctx := context.Background()
	for _, size := range []int{1024, 65536} {
		g := incr.New()
		var base pmap.Map[int, int]
		for k := range size {
			base = base.Set(k, k)
		}
		v := incr.Var(g, base)

		var mapCalls, addCalls, removeCalls, mergeCalls int
		incr.MustObserve(g, FilterMapValues(g, v, intsEqual, func(_ int, value int) (int, bool) {
			mapCalls++
			return value, true
		}))
		incr.MustObserve(g, UnorderedFold(g, v, 0, intsEqual,
			func(acc int, _ int, value int) int { addCalls++; return acc + value },
			func(acc int, _ int, value int) int { removeCalls++; return acc - value }))
		incr.MustObserve(g, Merge(g, v, v, intsEqual, intsEqual,
			func(_ int, element MergeElement[int, int]) (int, bool) {
				mergeCalls++
				return element.Left, true
			}))
		if err := g.Stabilize(ctx); err != nil {
			t.Fatal(err)
		}

		// the initial pass necessarily visits every key; after that a single changed
		// key must cost a fixed number of callbacks.
		mapCalls, addCalls, removeCalls, mergeCalls = 0, 0, 0, 0
		v.Set(base.Set(size/2, -1))
		if err := g.Stabilize(ctx); err != nil {
			t.Fatal(err)
		}
		if mapCalls != 1 {
			t.Errorf("size %d: FilterMapValues called fn %d times for one change, want 1", size, mapCalls)
		}
		// an updated key withdraws its old contribution and applies the new one
		if addCalls != 1 || removeCalls != 1 {
			t.Errorf("size %d: UnorderedFold made %d adds and %d removes for one change, want 1 and 1",
				size, addCalls, removeCalls)
		}
		if mergeCalls != 1 {
			t.Errorf("size %d: Merge called fn %d times for one change, want 1", size, mergeCalls)
		}
	}
}

// Test_operators_scaling checks that per-change wall clock does not track map size.
//
// The bound is loose because it cannot be tight: at the largest size the tree spans
// several megabytes, so walking it misses cache in a way that grows with size
// without any extra work being done, and the measurement includes updating the input
// map as well as the operator. Test_operators_work above is the exact statement;
// this is a backstop against a genuine return to linear behavior.
func Test_operators_scaling(t *testing.T) {
	ctx := context.Background()
	sizes := []int{1024, 8192, 65536}

	measure := func(t *testing.T, build func(g *incr.Graph, v incr.VarIncr[pmap.Map[int, int]]) incr.INode) []time.Duration {
		costs := make([]time.Duration, len(sizes))
		for index, size := range sizes {
			g := incr.New()
			var base pmap.Map[int, int]
			for k := range size {
				base = base.Set(k, k)
			}
			v := incr.Var(g, base)
			build(g, v)
			if err := g.Stabilize(ctx); err != nil {
				t.Fatal(err)
			}
			current := base
			iters := 300
			start := time.Now()
			for i := range iters {
				current = current.Set(i%size, -i-1)
				v.Set(current)
				if err := g.Stabilize(ctx); err != nil {
					t.Fatal(err)
				}
			}
			costs[index] = time.Since(start) / time.Duration(iters)
		}
		return costs
	}

	check := func(t *testing.T, name string, costs []time.Duration) {
		worst := math.Inf(-1)
		for i := 1; i < len(sizes); i++ {
			e := math.Log(float64(costs[i])/float64(costs[i-1])) / math.Log(float64(sizes[i])/float64(sizes[i-1]))
			t.Logf("%s: %6d -> %6d: %v -> %v exponent %.2f", name, sizes[i-1], sizes[i], costs[i-1], costs[i], e)
			worst = math.Max(worst, e)
		}
		if worst > 0.7 {
			t.Errorf("%s cost tracks map size: exponent %.2f", name, worst)
		}
	}

	check(t, "FilterMapValues", measure(t, func(g *incr.Graph, v incr.VarIncr[pmap.Map[int, int]]) incr.INode {
		out := FilterMapValues(g, v, intsEqual, func(_ int, value int) (int, bool) { return value * 2, value%2 == 0 })
		return incr.MustObserve(g, out)
	}))

	check(t, "UnorderedFold", measure(t, func(g *incr.Graph, v incr.VarIncr[pmap.Map[int, int]]) incr.INode {
		out := UnorderedFold(g, v, 0, intsEqual,
			func(acc int, _ int, value int) int { return acc + value },
			func(acc int, _ int, value int) int { return acc - value })
		return incr.MustObserve(g, out)
	}))

	check(t, "Merge", measure(t, func(g *incr.Graph, v incr.VarIncr[pmap.Map[int, int]]) incr.INode {
		// merge the input against itself, so both sides change each pass
		out := Merge(g, v, v, intsEqual, intsEqual,
			func(_ int, element MergeElement[int, int]) (int, bool) {
				return element.Left + element.Right, true
			})
		return incr.MustObserve(g, out)
	}))
}
