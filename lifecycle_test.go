package incr

import (
	"context"
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

// Test_lifecycle_observeAndUnobserve covers the transitions a node goes through as
// it enters and leaves the observed part of the graph.
//
// A node can stop being part of the computation without its value ever changing, so
// none of this is visible through update handlers; this is the only way a caller
// holding a resource on a node's behalf can learn to release it.
func Test_lifecycle_observeAndUnobserve(t *testing.T) {
	ctx := context.Background()
	g := New()

	v := Var(g, 1)
	m := Map(g, v, func(x int) int { return x + 1 })

	var necessary, unnecessary, invalidated int
	m.Node().OnBecameNecessary(func() { necessary++ })
	m.Node().OnBecameUnnecessary(func() { unnecessary++ })
	m.Node().OnInvalidated(func() { invalidated++ })

	// registering handlers must not itself make anything happen
	testutil.Equal(t, 0, necessary)

	o := MustObserve(g, m)
	testutil.Equal(t, 1, necessary, "observing should make the node necessary")
	testutil.Equal(t, 0, unnecessary)

	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 2, o.Value())
	// stabilizing changes values, not necessity
	testutil.Equal(t, 1, necessary)

	o.Unobserve(ctx)
	testutil.Equal(t, 1, unnecessary, "unobserving the only observer should release the node")
	testutil.Equal(t, 0, invalidated, "unobserving is not invalidation")

	// and observing again should report it necessary a second time
	o2 := MustObserve(g, m)
	testutil.Equal(t, 2, necessary)
	testutil.Nil(t, g.Stabilize(ctx))
	o2.Unobserve(ctx)
	testutil.Equal(t, 2, unnecessary)
}

// Test_lifecycle_invalidatedByBind covers the transition that has no other signal
// at all: a bind rewriting its right-hand side invalidates the nodes it previously
// created.
func Test_lifecycle_invalidatedByBind(t *testing.T) {
	ctx := context.Background()
	g := New(OptGraphMaxHeight(64))

	sel := Var(g, 0)
	var invalidated int
	b := Bind(g, sel, func(bs Scope, which int) Incr[int] {
		inner := Map(bs, Return(bs, which), func(x int) int { return x + 1 })
		inner.Node().OnInvalidated(func() { invalidated++ })
		return inner
	})
	o := MustObserve(g, b)

	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 1, o.Value())
	testutil.Equal(t, 0, invalidated, "nothing is invalidated on the first pass")

	// each rebuild discards the previous right-hand side
	for i := 1; i <= 3; i++ {
		sel.Set(i)
		testutil.Nil(t, g.Stabilize(ctx))
		testutil.Equal(t, i+1, o.Value())
		testutil.Equal(t, i, invalidated, "each rebuild should invalidate the previous subgraph")
	}
}

// Test_lifecycle_releasesResource is the motivating case written out: a node holding
// something that has to be closed.
func Test_lifecycle_releasesResource(t *testing.T) {
	ctx := context.Background()
	g := New()

	var open, closed int
	v := Var(g, 1)
	m := Map(g, v, func(x int) int { return x * 2 })
	m.Node().OnBecameNecessary(func() { open++ })
	m.Node().OnBecameUnnecessary(func() { closed++ })

	o := MustObserve(g, m)
	testutil.Nil(t, g.Stabilize(ctx))
	v.Set(5)
	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 10, o.Value())

	testutil.Equal(t, 1, open)
	testutil.Equal(t, 0, closed, "the resource should stay open while the node is in use")

	o.Unobserve(ctx)
	testutil.Equal(t, 1, closed, "the resource should be released once nothing needs the node")
}

// Test_lifecycle_noHandlersNoCost checks that a node without lifecycle handlers does
// not allocate the side struct that holds them, since every node would otherwise pay
// for a feature almost none use.
func Test_lifecycle_noHandlersNoCost(t *testing.T) {
	g := New()
	v := Var(g, 1)
	m := Map(g, v, func(x int) int { return x })
	MustObserve(g, m)
	testutil.Nil(t, g.Stabilize(context.Background()))

	testutil.Nil(t, m.Node().ext, "a node with no optional fields should not allocate nodeExtra")
	testutil.Nil(t, m.Node().becameUnnecessaryHandlers())
	testutil.Nil(t, m.Node().invalidatedHandlers())
	testutil.Nil(t, m.Node().becameNecessaryHandlers())
}
