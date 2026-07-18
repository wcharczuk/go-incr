package mapi

import (
	"context"
	"testing"

	"github.com/wcharczuk/go-incr"
	"github.com/wcharczuk/go-incr/incrutil/pmap"
)

func Test_Join(t *testing.T) {
	ctx := context.Background()
	g := incr.New(incr.OptGraphMaxHeight(64))

	a := incr.Var(g, 1)
	b := incr.Var(g, 2)
	outer := pmap.New[string, incr.Incr[int]]().Set("a", a).Set("b", b)
	v := incr.Var(g, outer)
	o := incr.MustObserve(g, Join(g, v))

	if err := g.Stabilize(ctx); err != nil {
		t.Fatal(err)
	}
	got := pmap.ToGoMap(o.Value())
	if len(got) != 2 || got["a"] != 1 || got["b"] != 2 {
		t.Fatalf("initial join gave %v", got)
	}

	// an inner incremental changing must reach the output
	a.Set(10)
	if err := g.Stabilize(ctx); err != nil {
		t.Fatal(err)
	}
	if got := pmap.ToGoMap(o.Value()); got["a"] != 10 || got["b"] != 2 {
		t.Fatalf("after inner change got %v", got)
	}

	// adding a key links its inner incremental
	c := incr.Var(g, 3)
	v.Set(outer.Set("c", c))
	if err := g.Stabilize(ctx); err != nil {
		t.Fatal(err)
	}
	if got := pmap.ToGoMap(o.Value()); len(got) != 3 || got["c"] != 3 {
		t.Fatalf("after adding c got %v", got)
	}
	c.Set(30)
	if err := g.Stabilize(ctx); err != nil {
		t.Fatal(err)
	}
	if got := pmap.ToGoMap(o.Value()); got["c"] != 30 {
		t.Fatalf("newly linked inner change gave %v", got)
	}

	// repointing a key at a different incremental
	v.Set(outer.Set("c", c).Set("a", b))
	if err := g.Stabilize(ctx); err != nil {
		t.Fatal(err)
	}
	if got := pmap.ToGoMap(o.Value()); got["a"] != 2 {
		t.Fatalf("after repointing a at b, got %v", got)
	}

	// removing a key drops it from the output
	v.Set(outer.Set("c", c).Delete("b"))
	if err := g.Stabilize(ctx); err != nil {
		t.Fatal(err)
	}
	if got := pmap.ToGoMap(o.Value()); len(got) != 2 {
		t.Fatalf("after removing b got %v", got)
	}
	if _, ok := o.Value().Get("b"); ok {
		t.Fatal("removed key still present")
	}
}

// Test_Join_releasesRemoved checks the unlink half: an inner incremental dropped from
// the map stops being computed, which is what keeps a join over a large collection
// from doing work for entries nobody is reading.
func Test_Join_releasesRemoved(t *testing.T) {
	ctx := context.Background()
	g := incr.New(incr.OptGraphMaxHeight(64))

	base := incr.Var(g, 1)
	var recomputes int
	inner := incr.Map(g, base, func(x int) int { recomputes++; return x * 2 })

	outer := pmap.New[string, incr.Incr[int]]().Set("k", inner)
	v := incr.Var(g, outer)
	o := incr.MustObserve(g, Join(g, v))

	if err := g.Stabilize(ctx); err != nil {
		t.Fatal(err)
	}
	if got, _ := o.Value().Get("k"); got != 2 {
		t.Fatalf("initial value %d", got)
	}
	if recomputes != 1 {
		t.Fatalf("inner recomputed %d times, want 1", recomputes)
	}

	// while linked, a change to its input recomputes it
	base.Set(2)
	if err := g.Stabilize(ctx); err != nil {
		t.Fatal(err)
	}
	if recomputes != 2 {
		t.Fatalf("inner recomputed %d times, want 2", recomputes)
	}

	// once unlinked it should not be recomputed at all
	v.Set(pmap.New[string, incr.Incr[int]]())
	if err := g.Stabilize(ctx); err != nil {
		t.Fatal(err)
	}
	before := recomputes
	base.Set(3)
	if err := g.Stabilize(ctx); err != nil {
		t.Fatal(err)
	}
	if recomputes != before {
		t.Fatalf("an unlinked inner incremental recomputed: %d then %d", before, recomputes)
	}
	if o.Value().Len() != 0 {
		t.Fatalf("output should be empty, has %d", o.Value().Len())
	}
}

// Test_Join_work asserts the cost claim: an inner value change rewrites one entry
// however many entries the map holds.
func Test_Join_work(t *testing.T) {
	ctx := context.Background()
	for _, size := range []int{64, 2048} {
		g := incr.New(incr.OptGraphMaxHeight(256))
		vars := make([]incr.VarIncr[int], size)
		outer := pmap.New[int, incr.Incr[int]]()
		var recomputes int
		for i := range size {
			vars[i] = incr.Var(g, i)
			// count how many inner nodes recompute per pass
			outer = outer.Set(i, incr.Map(g, vars[i], func(x int) int { recomputes++; return x }))
		}
		v := incr.Var(g, outer)
		o := incr.MustObserve(g, Join(g, v))
		if err := g.Stabilize(ctx); err != nil {
			t.Fatal(err)
		}
		if o.Value().Len() != size {
			t.Fatalf("size %d: output has %d entries", size, o.Value().Len())
		}

		recomputes = 0
		vars[size/2].Set(-1)
		if err := g.Stabilize(ctx); err != nil {
			t.Fatal(err)
		}
		if recomputes != 1 {
			t.Errorf("size %d: changing one inner input recomputed %d inner nodes, want 1", size, recomputes)
		}
		if got, _ := o.Value().Get(size / 2); got != -1 {
			t.Errorf("size %d: changed entry is %d, want -1", size, got)
		}
	}
}
