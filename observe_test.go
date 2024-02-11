package incr

import (
	"context"
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_Observe_Unobserve(t *testing.T) {
	ctx := testContext()
	g := New()

	v0 := Var(g, "hello 0")
	m0 := Map(g, v0, ident)

	v1 := Var(g, "hello 1")
	m1 := Map(g, v1, ident)

	o0 := Observe(g, m0)
	o1 := Observe(g, m1)

	testutil.Equal(t, 6, g.numNodes)

	testutil.Equal(t, true, g.IsObserving(m0))
	testutil.Equal(t, true, g.IsObserving(m1))

	testutil.Equal(t, "", o0.Value())
	testutil.Equal(t, "", o1.Value())

	err := g.Stabilize(context.TODO())
	testutil.Nil(t, err)

	testutil.Equal(t, "hello 0", o0.Value())
	testutil.Equal(t, "hello 1", o1.Value())

	o1.Unobserve(ctx)

	testutil.Equal(t, len(g.observed), g.numNodes-1, "we don't observe the observer but we do track it!")
	testutil.Nil(t, o1.Node().graph)

	// should take effect immediately because there is only (1) observer.
	testutil.Equal(t, true, g.IsObserving(m0))
	testutil.Equal(t, false, g.IsObserving(m1))

	v0.Set("not hello 0")
	v1.Set("not hello 1")
	err = g.Stabilize(context.TODO())
	testutil.Nil(t, err)

	testutil.Equal(t, "not hello 0", o0.Value())
	testutil.Equal(t, "", o1.Value())
}

func Test_Observe_Unobserve_multiple(t *testing.T) {
	ctx := testContext()
	g := New()

	v0 := Var(g, "hello 0")
	m0 := Map(g, v0, ident)

	v1 := Var(g, "hello 1")
	m1 := Map(g, v1, ident)

	o0 := Observe(g, m0)
	o1 := Observe(g, m1)
	o11 := Observe(g, m1)

	testutil.Equal(t, true, g.IsObserving(v0))
	testutil.Equal(t, true, g.IsObserving(m0))
	testutil.Equal(t, true, g.IsObserving(v1))
	testutil.Equal(t, true, g.IsObserving(m1))

	testutil.Equal(t, 1, len(v0.Node().Observers()))
	testutil.Equal(t, 1, len(m0.Node().Observers()))
	testutil.Equal(t, 2, len(v1.Node().Observers()))
	testutil.Equal(t, 2, len(m1.Node().Observers()))

	testutil.Equal(t, "", o0.Value())
	testutil.Equal(t, "", o1.Value())
	testutil.Equal(t, "", o11.Value())

	err := g.Stabilize(context.TODO())
	testutil.Nil(t, err)

	testutil.Equal(t, "hello 0", o0.Value())
	testutil.Equal(t, "hello 1", o1.Value())
	testutil.Equal(t, "hello 1", o11.Value())

	o1.Unobserve(ctx)

	testutil.Equal(t, len(g.observed), g.numNodes-2, "we should have (1) less observer after unobserve!")
	testutil.Nil(t, o1.Node().graph)

	testutil.Equal(t, 0, len(o1.Node().parents))
	testutil.Equal(t, 0, len(o1.Node().children))
	testutil.None(t, m1.Node().Children(), func(n INode) bool {
		return n.Node().ID() == o1.Node().ID()
	})

	testutil.Equal(t, true, g.IsObserving(m0))
	testutil.Equal(t, true, g.IsObserving(m1))

	testutil.Equal(t, 1, len(v0.Node().Observers()))
	testutil.Equal(t, 1, len(m0.Node().Observers()))
	testutil.Equal(t, 1, len(v1.Node().Observers()))
	testutil.Equal(t, 1, len(m1.Node().Observers()))

	v0.Set("not hello 0")
	v1.Set("not hello 1")
	err = g.Stabilize(ctx)
	testutil.Nil(t, err)

	testutil.Equal(t, "not hello 0", o0.Value())
	testutil.Equal(t, "", o1.Value())
	testutil.Equal(t, "not hello 1", o11.Value())
}

func Test_Observer_Unobserve_reobserve(t *testing.T) {
	ctx := testContext()
	g := New()
	v0 := Var(g, "hello")
	m0 := Map(g, v0, ident)
	o0 := Observe(g, m0)

	_ = g.Stabilize(context.TODO())
	testutil.Equal(t, "hello", o0.Value())

	o0.Unobserve(ctx)

	_ = g.Stabilize(context.TODO())
	testutil.Equal(t, false, g.IsObserving(m0))
	// strictly, the value shouldn't change ...
	testutil.Equal(t, "hello", m0.Value())

	o1 := Observe(g, m0)
	_ = g.Stabilize(context.TODO())
	testutil.Equal(t, "hello", o1.Value())
}
