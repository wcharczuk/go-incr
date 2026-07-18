package mapi

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/wcharczuk/go-incr"
	"github.com/wcharczuk/go-incr/incrutil/pmap"
)

func intsEqual(a, b int) bool { return a == b }

func Test_MapValues(t *testing.T) {
	ctx := context.Background()
	g := incr.New()

	base := pmap.FromGoMap(map[string]int{"a": 1, "b": 2, "c": 3})
	v := incr.Var(g, base)
	doubled := MapValues(g, v, intsEqual, func(_ string, value int) int { return value * 2 })
	o := incr.MustObserve(g, doubled)

	if err := g.Stabilize(ctx); err != nil {
		t.Fatal(err)
	}
	if got := pmap.ToGoMap(o.Value()); len(got) != 3 || got["a"] != 2 || got["c"] != 6 {
		t.Fatalf("initial pass gave %v", got)
	}

	// add, rebind and remove in one pass
	v.Set(base.Set("d", 10).Set("b", 20).Delete("a"))
	if err := g.Stabilize(ctx); err != nil {
		t.Fatal(err)
	}
	got := pmap.ToGoMap(o.Value())
	want := map[string]int{"b": 40, "c": 6, "d": 20}
	if len(got) != len(want) {
		t.Fatalf("after update got %v, want %v", got, want)
	}
	for key, wantValue := range want {
		if got[key] != wantValue {
			t.Fatalf("after update %q = %d, want %d", key, got[key], wantValue)
		}
	}
}

// Test_MapValues_onlyChangedKeys is the behavioral claim: an unchanged key must
// not be recomputed.
func Test_MapValues_onlyChangedKeys(t *testing.T) {
	ctx := context.Background()
	g := incr.New()

	base := pmap.FromGoMap(map[string]int{"a": 1, "b": 2, "c": 3})
	v := incr.Var(g, base)
	var calls []string
	mapped := MapValues(g, v, intsEqual, func(key string, value int) int {
		calls = append(calls, key)
		return value * 2
	})
	incr.MustObserve(g, mapped)

	if err := g.Stabilize(ctx); err != nil {
		t.Fatal(err)
	}
	if len(calls) != 3 {
		t.Fatalf("initial pass called fn %d times, want 3", len(calls))
	}

	calls = nil
	v.Set(base.Set("b", 99))
	if err := g.Stabilize(ctx); err != nil {
		t.Fatal(err)
	}
	if len(calls) != 1 || calls[0] != "b" {
		t.Fatalf("changing one key called fn for %v, want only [b]", calls)
	}

	// re-setting an identical map must recompute nothing at all
	calls = nil
	v.Set(base.Set("b", 99))
	if err := g.Stabilize(ctx); err != nil {
		t.Fatal(err)
	}
	if len(calls) != 0 {
		t.Fatalf("an unchanged input recomputed %v", calls)
	}
}

// Test_MapValues_scaling checks that per-change cost does not track map size.
//
// The bound is loose for the same reason as Test_operators_scaling: at the largest
// size the tree spans several megabytes, so walking it misses cache in a way that
// grows with size without any extra work being done, and the measurement includes
// updating the input map too. A tighter bound flaked on the middle size while the
// series was not even monotonic. Test_MapValues_onlyChangedKeys is the exact
// statement of the claim -- one callback per changed key -- and this is a backstop
// against a genuine return to linear behavior.
//
// The equivalent over a builtin map is linear; see Added and Removed in this
// package, which scan and clone on every pass.
func Test_MapValues_scaling(t *testing.T) {
	skipTimingUnlessRequested(t)
	ctx := context.Background()
	sizes := []int{1024, 8192, 65536}
	costs := make([]time.Duration, len(sizes))

	for index, size := range sizes {
		g := incr.New()
		var base pmap.Map[int, int]
		for k := range size {
			base = base.Set(k, k)
		}
		v := incr.Var(g, base)
		mapped := MapValues(g, v, intsEqual, func(_ int, value int) int { return value * 2 })
		o := incr.MustObserve(g, mapped)
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
		if o.Value().Len() != size {
			t.Fatalf("size %d: output has %d entries", size, o.Value().Len())
		}
		t.Logf("size %6d: changing one key costs %v", size, costs[index])
	}

	worst := math.Inf(-1)
	for i := 1; i < len(sizes); i++ {
		e := math.Log(float64(costs[i])/float64(costs[i-1])) / math.Log(float64(sizes[i])/float64(sizes[i-1]))
		t.Logf("  %6d -> %6d: exponent %.2f", sizes[i-1], sizes[i], e)
		worst = math.Max(worst, e)
	}
	if worst > 0.7 {
		t.Errorf("cost tracks map size: exponent %.2f; expected roughly logarithmic", worst)
	}
}
