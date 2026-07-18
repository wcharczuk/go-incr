package mapi

import (
	"context"
	"math/rand"
	"testing"

	"github.com/wcharczuk/go-incr"
	"github.com/wcharczuk/go-incr/incrutil/pmap"
)

func Test_Partition(t *testing.T) {
	ctx := context.Background()
	g := incr.New()

	base := pmap.FromGoMap(map[string]int{"a": 1, "b": 2, "c": 3, "d": 4})
	v := incr.Var(g, base)
	o := incr.MustObserve(g, Partition(g, v, intsEqual,
		func(_ string, value int) bool { return value%2 == 0 }))

	if err := g.Stabilize(ctx); err != nil {
		t.Fatal(err)
	}
	if got := o.Value(); got.Matching.Len() != 2 || got.NotMatching.Len() != 2 {
		t.Fatalf("initial split: %d matching, %d not", got.Matching.Len(), got.NotMatching.Len())
	}

	// a value flipping the predicate has to move sides, not appear on both
	v.Set(base.Set("a", 10))
	if err := g.Stabilize(ctx); err != nil {
		t.Fatal(err)
	}
	got := o.Value()
	if _, ok := got.Matching.Get("a"); !ok {
		t.Fatal("a should have moved to the matching side")
	}
	if _, ok := got.NotMatching.Get("a"); ok {
		t.Fatal("a is on both sides")
	}
	if got.Matching.Len() != 3 || got.NotMatching.Len() != 1 {
		t.Fatalf("after flip: %d matching, %d not", got.Matching.Len(), got.NotMatching.Len())
	}

	// and a removal leaves both sides
	v.Set(base.Set("a", 10).Delete("b"))
	if err := g.Stabilize(ctx); err != nil {
		t.Fatal(err)
	}
	got = o.Value()
	if _, ok := got.Matching.Get("b"); ok {
		t.Fatal("a removed key is still on the matching side")
	}
	if got.Matching.Len()+got.NotMatching.Len() != 3 {
		t.Fatalf("sides hold %d entries total, want 3",
			got.Matching.Len()+got.NotMatching.Len())
	}
}

// Test_Partition_matchesReference cross-checks the maintained halves against splitting
// from scratch, since a key left on the wrong side is a plausible wrong answer rather than
// a failure.
func Test_Partition_matchesReference(t *testing.T) {
	ctx := context.Background()
	g := incr.New()
	rng := rand.New(rand.NewSource(23))

	current := pmap.New[int, int]()
	v := incr.Var(g, current)
	predicate := func(_ int, value int) bool { return value%3 == 0 }
	o := incr.MustObserve(g, Partition(g, v, intsEqual, predicate))
	if err := g.Stabilize(ctx); err != nil {
		t.Fatal(err)
	}

	for step := range 400 {
		key := rng.Intn(40)
		if rng.Intn(4) == 0 {
			current = current.Delete(key)
		} else {
			current = current.Set(key, rng.Intn(30))
		}
		v.Set(current)
		if err := g.Stabilize(ctx); err != nil {
			t.Fatal(err)
		}

		wantMatching, wantNot := 0, 0
		for key, value := range current.All() {
			if predicate(key, value) {
				wantMatching++
			} else {
				wantNot++
			}
		}
		got := o.Value()
		if got.Matching.Len() != wantMatching || got.NotMatching.Len() != wantNot {
			t.Fatalf("step %d: got %d/%d, want %d/%d", step,
				got.Matching.Len(), got.NotMatching.Len(), wantMatching, wantNot)
		}
		// and no key on both sides
		for key := range got.Matching.Keys() {
			if _, ok := got.NotMatching.Get(key); ok {
				t.Fatalf("step %d: key %d is on both sides", step, key)
			}
		}
	}
}
