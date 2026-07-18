package mapi

import (
	"context"
	"testing"

	"github.com/wcharczuk/go-incr"
	"github.com/wcharczuk/go-incr/incrutil/pmap"
)

func Test_Selector(t *testing.T) {
	ctx := context.Background()
	g := incr.New(incr.OptGraphMaxHeight(64))

	book := pmap.FromGoMap(map[string]int{"a": 1, "b": 2, "c": 3})
	v := incr.Var(g, book)
	selector := NewSelector(g, v, intsEqual)

	a := incr.MustObserve(g, selector.Select("a"))
	b := incr.MustObserve(g, selector.Select("b"))
	missing := incr.MustObserve(g, selector.Select("zz"))

	if err := g.Stabilize(ctx); err != nil {
		t.Fatal(err)
	}
	if a.Value() != 1 || b.Value() != 2 {
		t.Fatalf("initial values a=%d b=%d", a.Value(), b.Value())
	}
	if missing.Value() != 0 {
		t.Fatalf("an absent key should be the zero value, got %d", missing.Value())
	}

	// selecting the same key twice gives the same node, so consumers share one
	if selector.Select("a") != selector.Select("a") {
		t.Fatal("Select should return the same node for the same key")
	}

	v.Set(book.Set("a", 10))
	if err := g.Stabilize(ctx); err != nil {
		t.Fatal(err)
	}
	if a.Value() != 10 {
		t.Fatalf("changed key is %d, want 10", a.Value())
	}
	if b.Value() != 2 {
		t.Fatalf("unchanged key moved to %d", b.Value())
	}

	// a key arriving after it was selected
	v.Set(book.Set("a", 10).Set("zz", 99))
	if err := g.Stabilize(ctx); err != nil {
		t.Fatal(err)
	}
	if missing.Value() != 99 {
		t.Fatalf("a key that arrived later is %d, want 99", missing.Value())
	}

	// and leaving again
	v.Set(book.Set("a", 10))
	if err := g.Stabilize(ctx); err != nil {
		t.Fatal(err)
	}
	if missing.Value() != 0 {
		t.Fatalf("a key that left is %d, want the zero value", missing.Value())
	}
}

// Test_Selector_fanoutIsNarrow is the claim the type exists for: one key changing must
// only recompute the consumers of that key, however many other keys are being watched.
func Test_Selector_fanoutIsNarrow(t *testing.T) {
	ctx := context.Background()

	for _, watchers := range []int{16, 1024} {
		g := incr.New(incr.OptGraphMaxHeight(64))
		book := pmap.New[int, int]()
		for i := range watchers {
			book = book.Set(i, i)
		}
		v := incr.Var(g, book)
		selector := NewSelector(g, v, intsEqual)

		// a consumer per key, each counting its own recomputes
		var recomputes int
		for i := range watchers {
			incr.MustObserve(g, incr.Map(g, selector.Select(i), func(value int) int {
				recomputes++
				return value
			}))
		}
		if err := g.Stabilize(ctx); err != nil {
			t.Fatal(err)
		}
		if recomputes != watchers {
			t.Fatalf("watchers=%d: initial pass recomputed %d consumers, want %d",
				watchers, recomputes, watchers)
		}

		// one key changes
		recomputes = 0
		v.Set(book.Set(watchers/2, -1))
		if err := g.Stabilize(ctx); err != nil {
			t.Fatal(err)
		}
		if recomputes != 1 {
			t.Errorf("watchers=%d: one key changing recomputed %d consumers, want 1",
				watchers, recomputes)
		}
	}
}

// Test_Selector_multipleChanges checks that several keys changing in one pass wakes
// exactly those consumers.
func Test_Selector_multipleChanges(t *testing.T) {
	ctx := context.Background()
	g := incr.New(incr.OptGraphMaxHeight(64))

	book := pmap.New[int, int]()
	for i := range 100 {
		book = book.Set(i, i)
	}
	v := incr.Var(g, book)
	selector := NewSelector(g, v, intsEqual)

	seen := map[int]int{}
	for i := range 100 {
		key := i
		incr.MustObserve(g, incr.Map(g, selector.Select(key), func(value int) int {
			seen[key]++
			return value
		}))
	}
	if err := g.Stabilize(ctx); err != nil {
		t.Fatal(err)
	}

	clear(seen)
	next := book
	for _, key := range []int{3, 17, 42} {
		next = next.Set(key, -key)
	}
	v.Set(next)
	if err := g.Stabilize(ctx); err != nil {
		t.Fatal(err)
	}

	if len(seen) != 3 {
		t.Fatalf("expected 3 consumers to recompute, got %d: %v", len(seen), seen)
	}
	for _, key := range []int{3, 17, 42} {
		if seen[key] != 1 {
			t.Fatalf("consumer for key %d recomputed %d times", key, seen[key])
		}
	}
}

// Test_Selector_comparedWithNaiveFanout contrasts the Selector against every consumer
// taking the whole map, which is the shape it replaces.
func Test_Selector_comparedWithNaiveFanout(t *testing.T) {
	ctx := context.Background()
	const watchers = 256

	book := pmap.New[int, int]()
	for i := range watchers {
		book = book.Set(i, i)
	}

	// naive: each consumer reads the map and picks its key out, so each is a dependent of
	// the map itself and recomputes whenever any key changes
	naiveGraph := incr.New(incr.OptGraphMaxHeight(64))
	naiveVar := incr.Var(naiveGraph, book)
	var naiveRecomputes int
	for i := range watchers {
		key := i
		incr.MustObserve(naiveGraph, incr.Map(naiveGraph, naiveVar, func(m pmap.Map[int, int]) int {
			naiveRecomputes++
			value, _ := m.Get(key)
			return value
		}))
	}
	if err := naiveGraph.Stabilize(ctx); err != nil {
		t.Fatal(err)
	}

	selectorGraph := incr.New(incr.OptGraphMaxHeight(64))
	selectorVar := incr.Var(selectorGraph, book)
	selector := NewSelector(selectorGraph, selectorVar, intsEqual)
	var selectorRecomputes int
	for i := range watchers {
		incr.MustObserve(selectorGraph, incr.Map(selectorGraph, selector.Select(i), func(value int) int {
			selectorRecomputes++
			return value
		}))
	}
	if err := selectorGraph.Stabilize(ctx); err != nil {
		t.Fatal(err)
	}

	naiveRecomputes, selectorRecomputes = 0, 0
	naiveVar.Set(book.Set(0, -1))
	selectorVar.Set(book.Set(0, -1))
	if err := naiveGraph.Stabilize(ctx); err != nil {
		t.Fatal(err)
	}
	if err := selectorGraph.Stabilize(ctx); err != nil {
		t.Fatal(err)
	}

	if naiveRecomputes != watchers {
		t.Fatalf("the naive fan-out should recompute every consumer, got %d of %d",
			naiveRecomputes, watchers)
	}
	if selectorRecomputes != 1 {
		t.Fatalf("the selector should recompute one consumer, got %d", selectorRecomputes)
	}
	t.Logf("one key changing over %d watchers: naive recomputed %d, selector recomputed %d",
		watchers, naiveRecomputes, selectorRecomputes)
}
