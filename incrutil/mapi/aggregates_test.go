package mapi

import (
	"context"
	"math/rand"
	"testing"

	"github.com/wcharczuk/go-incr"
	"github.com/wcharczuk/go-incr/incrutil/pmap"
)

func Test_aggregates(t *testing.T) {
	ctx := context.Background()
	g := incr.New()

	base := pmap.FromGoMap(map[string]int{"a": 1, "b": 2, "c": 3, "d": 4})
	v := incr.Var(g, base)

	count := incr.MustObserve(g, Cardinality(g, v))
	sum := incr.MustObserve(g, Sum(g, v, intsEqual))
	evens := incr.MustObserve(g, Counti(g, v, intsEqual, func(_ string, value int) bool { return value%2 == 0 }))
	keys := incr.MustObserve(g, Keys(g, v))

	if err := g.Stabilize(ctx); err != nil {
		t.Fatal(err)
	}
	if count.Value() != 4 || sum.Value() != 10 || evens.Value() != 2 {
		t.Fatalf("initial: count %d sum %d evens %d", count.Value(), sum.Value(), evens.Value())
	}
	if got := keys.Value(); len(got) != 4 || got[0] != "a" || got[3] != "d" {
		t.Fatalf("keys %v", got)
	}

	// rebinding a value must not move the count, but must move the sum
	v.Set(base.Set("a", 11))
	if err := g.Stabilize(ctx); err != nil {
		t.Fatal(err)
	}
	if count.Value() != 4 {
		t.Fatalf("count changed to %d on a rebind", count.Value())
	}
	if sum.Value() != 20 {
		t.Fatalf("sum %d, want 20", sum.Value())
	}
	// 11 is odd, so b and d remain
	if evens.Value() != 2 {
		t.Fatalf("evens %d, want 2", evens.Value())
	}

	v.Set(base.Delete("d").Set("e", 6))
	if err := g.Stabilize(ctx); err != nil {
		t.Fatal(err)
	}
	if count.Value() != 4 || sum.Value() != 12 || evens.Value() != 2 {
		t.Fatalf("after edit: count %d sum %d evens %d", count.Value(), sum.Value(), evens.Value())
	}
}

// Test_aggregates_matchReference cross-checks each maintained aggregate against
// recomputing it from scratch over many mutations.
func Test_aggregates_matchReference(t *testing.T) {
	ctx := context.Background()
	g := incr.New()
	rng := rand.New(rand.NewSource(5))

	current := pmap.New[int, int]()
	v := incr.Var(g, current)
	count := incr.MustObserve(g, Cardinality(g, v))
	sum := incr.MustObserve(g, Sum(g, v, intsEqual))
	evens := incr.MustObserve(g, Counti(g, v, intsEqual, func(_ int, value int) bool { return value%2 == 0 }))
	if err := g.Stabilize(ctx); err != nil {
		t.Fatal(err)
	}

	for step := range 400 {
		key := rng.Intn(30)
		if rng.Intn(4) == 0 {
			current = current.Delete(key)
		} else {
			current = current.Set(key, rng.Intn(50))
		}
		v.Set(current)
		if err := g.Stabilize(ctx); err != nil {
			t.Fatal(err)
		}
		var wantSum, wantEvens int
		for _, value := range current.All() {
			wantSum += value
			if value%2 == 0 {
				wantEvens++
			}
		}
		if count.Value() != current.Len() {
			t.Fatalf("step %d: count %d, want %d", step, count.Value(), current.Len())
		}
		if sum.Value() != wantSum {
			t.Fatalf("step %d: sum %d, want %d", step, sum.Value(), wantSum)
		}
		if evens.Value() != wantEvens {
			t.Fatalf("step %d: evens %d, want %d", step, evens.Value(), wantEvens)
		}
	}
}

func Test_Subrange(t *testing.T) {
	ctx := context.Background()
	g := incr.New()

	var base pmap.Map[int, int]
	for i := range 100 {
		base = base.Set(i, i*10)
	}
	v := incr.Var(g, base)
	bounds := incr.Var(g, Bounds[int]{Low: 10, High: 19})
	window := Subrange(g, v, bounds, intsEqual)
	o := incr.MustObserve(g, window)

	if err := g.Stabilize(ctx); err != nil {
		t.Fatal(err)
	}
	if o.Value().Len() != 10 {
		t.Fatalf("window has %d entries, want 10", o.Value().Len())
	}
	if got, _ := o.Value().Get(15); got != 150 {
		t.Fatalf("window value for 15 is %d", got)
	}
	if _, ok := o.Value().Get(20); ok {
		t.Fatal("window contains a key above the bound")
	}

	// a change inside the window shows up
	v.Set(base.Set(15, -1))
	if err := g.Stabilize(ctx); err != nil {
		t.Fatal(err)
	}
	if got, _ := o.Value().Get(15); got != -1 {
		t.Fatalf("in-window change gave %d", got)
	}

	// a change outside the window does not
	v.Set(base.Set(15, -1).Set(80, -2))
	if err := g.Stabilize(ctx); err != nil {
		t.Fatal(err)
	}
	if _, ok := o.Value().Get(80); ok {
		t.Fatal("out-of-window change leaked into the window")
	}
	if o.Value().Len() != 10 {
		t.Fatalf("window grew to %d entries", o.Value().Len())
	}

	// moving the window rebuilds it
	bounds.Set(Bounds[int]{Low: 50, High: 54})
	if err := g.Stabilize(ctx); err != nil {
		t.Fatal(err)
	}
	if o.Value().Len() != 5 {
		t.Fatalf("moved window has %d entries, want 5", o.Value().Len())
	}
	if got, _ := o.Value().Get(52); got != 520 {
		t.Fatalf("moved window value for 52 is %d", got)
	}
	if _, ok := o.Value().Get(15); ok {
		t.Fatal("moved window still holds an old key")
	}
}

// Test_Subrange_matchesReference cross-checks the window against filtering the map
// directly, while both the map and the bounds move.
func Test_Subrange_matchesReference(t *testing.T) {
	ctx := context.Background()
	g := incr.New()
	rng := rand.New(rand.NewSource(13))

	current := pmap.New[int, int]()
	for i := range 60 {
		current = current.Set(i, i)
	}
	v := incr.Var(g, current)
	bounds := incr.Var(g, Bounds[int]{Low: 0, High: 10})
	o := incr.MustObserve(g, Subrange(g, v, bounds, intsEqual))
	if err := g.Stabilize(ctx); err != nil {
		t.Fatal(err)
	}

	window := Bounds[int]{Low: 0, High: 10}
	for step := range 400 {
		switch rng.Intn(3) {
		case 0:
			low := rng.Intn(50)
			window = Bounds[int]{Low: low, High: low + rng.Intn(15)}
			bounds.Set(window)
		case 1:
			current = current.Delete(rng.Intn(60))
			v.Set(current)
		default:
			key := rng.Intn(60)
			current = current.Set(key, rng.Intn(1000))
			v.Set(current)
		}
		if err := g.Stabilize(ctx); err != nil {
			t.Fatal(err)
		}

		want := map[int]int{}
		for key, value := range current.All() {
			if key >= window.Low && key <= window.High {
				want[key] = value
			}
		}
		got := pmap.ToGoMap(o.Value())
		if len(got) != len(want) {
			t.Fatalf("step %d: window has %d entries, want %d", step, len(got), len(want))
		}
		for key, wantValue := range want {
			if got[key] != wantValue {
				t.Fatalf("step %d: window key %d = %d, want %d", step, key, got[key], wantValue)
			}
		}
	}
}
