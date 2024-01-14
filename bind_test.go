package incr

import (
	"context"
	"fmt"
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_Bind_basic(t *testing.T) {
	ctx := testContext()

	bindVar := Var("a")
	bindVar.Node().SetLabel("bindVar")

	av := Var("a-value")
	av.Node().SetLabel("av")
	a0 := Map(av, ident)
	a0.Node().SetLabel("a0")
	a1 := Map(a0, ident)
	a1.Node().SetLabel("a1")

	bv := Var("b-value")
	bv.Node().SetLabel("bv")
	b0 := Map(bv, ident)
	b0.Node().SetLabel("b0")
	b1 := Map(b0, ident)
	b1.Node().SetLabel("b1")
	b2 := Map(b1, ident)
	b2.Node().SetLabel("b2")

	bind := Bind(bindVar, func(which string) Incr[string] {
		if which == "a" {
			return a1
		}
		if which == "b" {
			return b2
		}
		return nil
	})

	bind.Node().SetLabel("bind")

	testutil.ItMatches(t, "bind\\[(.*)\\]:bind", bind.String())

	s0 := Return("hello")
	s0.Node().SetLabel("s0")
	s1 := Map(s0, ident)
	s1.Node().SetLabel("s1")

	o := Map2(bind, s1, concat)
	o.Node().SetLabel("o")

	g := New()
	_ = Observe(g, o)

	testutil.ItsEqual(t, 1, bindVar.Node().height)
	testutil.ItsEqual(t, 1, s0.Node().height)
	testutil.ItsEqual(t, 2, s1.Node().height)

	testutil.ItsEqual(t, 2, bind.Node().height)
	testutil.ItsEqual(t, 3, o.Node().height)

	testutil.ItsEqual(t, true, g.IsObserving(bindVar))
	testutil.ItsEqual(t, true, g.IsObserving(s0))
	testutil.ItsEqual(t, true, g.IsObserving(s1))
	testutil.ItsEqual(t, true, g.IsObserving(bind))
	testutil.ItsEqual(t, true, g.IsObserving(o))

	testutil.ItsEqual(t, false, g.IsObserving(av))
	testutil.ItsEqual(t, false, g.IsObserving(a0))
	testutil.ItsEqual(t, false, g.IsObserving(a1))

	testutil.ItsEqual(t, false, g.IsObserving(bv))
	testutil.ItsEqual(t, false, g.IsObserving(b0))
	testutil.ItsEqual(t, false, g.IsObserving(b1))
	testutil.ItsEqual(t, false, g.IsObserving(b2))

	err := g.Stabilize(ctx)
	testutil.ItsNil(t, err)

	testutil.ItsEqual(t, 1, bind.Node().boundAt)
	testutil.ItsEqual(t, 1, bind.Node().changedAt)
	testutil.ItsEqual(t, 1, a1.Node().changedAt)

	testutil.ItsEqual(t, 1, bindVar.Node().height)
	testutil.ItsEqual(t, 1, s0.Node().height)
	testutil.ItsEqual(t, 2, s1.Node().height)

	testutil.ItsEqual(t, 1, av.Node().height)
	testutil.ItsEqual(t, 2, a0.Node().height)
	testutil.ItsEqual(t, 3, a1.Node().height)

	testutil.ItsEqual(t, 2, bind.Node().height)
	testutil.ItsEqual(t, 4, o.Node().height)

	testutil.ItsEqual(t, true, g.IsObserving(bindVar))
	testutil.ItsEqual(t, true, g.IsObserving(s0))
	testutil.ItsEqual(t, true, g.IsObserving(s1))
	testutil.ItsEqual(t, true, g.IsObserving(bind))
	testutil.ItsEqual(t, true, g.IsObserving(o))

	testutil.ItsEqual(t, true, g.IsObserving(av))
	testutil.ItsEqual(t, true, g.IsObserving(a0))
	testutil.ItsEqual(t, true, g.IsObserving(a1))

	testutil.ItsEqual(t, false, g.IsObserving(bv))
	testutil.ItsEqual(t, false, g.IsObserving(b0))
	testutil.ItsEqual(t, false, g.IsObserving(b1))
	testutil.ItsEqual(t, false, g.IsObserving(b2))

	testutil.ItsEqual(t, "a-value", av.Value())
	testutil.ItsEqual(t, "a-valuehello", o.Value())

	bindVar.Set("b")
	err = g.Stabilize(ctx)
	testutil.ItsNil(t, err)

	testutil.ItsEqual(t, 1, bindVar.Node().height)
	testutil.ItsEqual(t, 1, s0.Node().height)
	testutil.ItsEqual(t, 2, s1.Node().height)

	testutil.ItsEqual(t, 1, av.Node().height)
	testutil.ItsEqual(t, 2, a0.Node().height)
	testutil.ItsEqual(t, 3, a1.Node().height)

	testutil.ItsEqual(t, 1, bv.Node().height)
	testutil.ItsEqual(t, 2, b0.Node().height)
	testutil.ItsEqual(t, 3, b1.Node().height)
	testutil.ItsEqual(t, 4, b2.Node().height)

	testutil.ItsEqual(t, 2, bind.Node().height)
	testutil.ItsEqual(t, 5, o.Node().height)

	testutil.ItsEqual(t, true, g.IsObserving(bindVar))
	testutil.ItsEqual(t, true, g.IsObserving(s0))
	testutil.ItsEqual(t, true, g.IsObserving(s1))
	testutil.ItsEqual(t, true, g.IsObserving(bind))
	testutil.ItsEqual(t, true, g.IsObserving(o))

	testutil.ItsEqual(t, false, g.IsObserving(av), "if we switch to b, we should unobserve the 'a' tree")
	testutil.ItsEqual(t, false, g.IsObserving(a0))
	testutil.ItsEqual(t, false, g.IsObserving(a1))

	testutil.ItsEqual(t, true, g.IsObserving(bv))
	testutil.ItsEqual(t, true, g.IsObserving(b0))
	testutil.ItsEqual(t, true, g.IsObserving(b1))
	testutil.ItsEqual(t, true, g.IsObserving(b2))

	testutil.ItsEqual(t, "a-value", av.Value())
	testutil.ItsEqual(t, "b-value", bv.Value())
	testutil.ItsEqual(t, "b-valuehello", o.Value())

	bindVar.Set("neither")
	err = g.Stabilize(ctx)
	testutil.ItsNil(t, err)

	testutil.ItsEqual(t, 1, bindVar.Node().height)
	testutil.ItsEqual(t, 1, s0.Node().height)
	testutil.ItsEqual(t, 2, s1.Node().height)

	testutil.ItsEqual(t, 1, av.Node().height)
	testutil.ItsEqual(t, 2, a0.Node().height)
	testutil.ItsEqual(t, 3, a1.Node().height)

	testutil.ItsEqual(t, 1, bv.Node().height)
	testutil.ItsEqual(t, 2, b0.Node().height)
	testutil.ItsEqual(t, 3, b1.Node().height)
	testutil.ItsEqual(t, 4, b2.Node().height)

	testutil.ItsEqual(t, 2, bind.Node().height)
	testutil.ItsEqual(t, 5, o.Node().height)

	testutil.ItsEqual(t, true, g.IsObserving(bindVar))
	testutil.ItsEqual(t, true, g.IsObserving(s0))
	testutil.ItsEqual(t, true, g.IsObserving(s1))
	testutil.ItsEqual(t, true, g.IsObserving(bind))
	testutil.ItsEqual(t, true, g.IsObserving(o))

	testutil.ItsEqual(t, false, g.IsObserving(av))
	testutil.ItsEqual(t, false, g.IsObserving(a0))
	testutil.ItsEqual(t, false, g.IsObserving(a1))

	testutil.ItsEqual(t, false, g.IsObserving(bv))
	testutil.ItsEqual(t, false, g.IsObserving(b0))
	testutil.ItsEqual(t, false, g.IsObserving(b1))
	testutil.ItsEqual(t, false, g.IsObserving(b2))

	testutil.ItsEqual(t, "a-value", av.Value())
	testutil.ItsEqual(t, "b-value", bv.Value())
	testutil.ItsEqual(t, "hello", o.Value())
}

func Test_Bind_error(t *testing.T) {
	ctx := testContext()

	v0 := Var("a")
	bind := BindContext(v0, func(_ context.Context, which string) (Incr[string], error) {
		return nil, fmt.Errorf("this is just a test")
	})
	bind.Node().SetLabel("bind")
	var gotError error
	bind.Node().OnError(func(_ context.Context, err error) {
		gotError = err
	})

	o := Map(bind, ident)

	g := New()
	_ = Observe(g, o)
	err := g.Stabilize(ctx)
	testutil.ItsEqual(t, "this is just a test", err.Error())
	testutil.ItsEqual(t, "this is just a test", gotError.Error())
}

func Test_Bind_nested(t *testing.T) {
	ctx := testContext()

	// a -> c
	// a -> b -> c
	// a -> c

	a0 := createDynamicMaps("a0")
	a1 := createDynamicMaps("a1")
	bv, b := createDynamicBind("b", a0, a1)
	cv, c := createDynamicBind("c", a0, b)
	final := Map2(c, Return("final"), func(a, b string) string {
		return a + "->" + b
	})

	g := New()
	o := Observe(g, final)

	_ = g.Stabilize(ctx)

	testutil.ItsNil(t, g.recomputeHeap.sanityCheck())
	testutil.ItsEqual(t, "a0-0+a0-1->c->final", o.Value())

	cv.Set("b")
	_ = g.Stabilize(ctx)

	testutil.ItsNil(t, g.recomputeHeap.sanityCheck())
	testutil.ItsEqual(t, "a0-0+a0-1->b->c->final", o.Value())

	bv.Set("b")
	_ = g.Stabilize(ctx)

	testutil.ItsNil(t, g.recomputeHeap.sanityCheck())
	testutil.ItsEqual(t, "a1-0+a1-1->b->c->final", o.Value())

	bv.Set("a")
	_ = g.Stabilize(ctx)

	testutil.ItsNil(t, g.recomputeHeap.sanityCheck())
	testutil.ItsEqual(t, "a0-0+a0-1->b->c->final", o.Value())

	cv.Set("a")
	_ = g.Stabilize(ctx)

	testutil.ItsNil(t, g.recomputeHeap.sanityCheck())
	testutil.ItsEqual(t, "a0-0+a0-1->c->final", o.Value())
}

func Test_Bind_nested_bindCreatesBind(t *testing.T) {
	ctx := testContext()

	cv := Var("a")
	cv.Node().SetLabel("cv")
	bv := Var("a")
	bv.Node().SetLabel("bv")
	c := BindContext[string](cv, func(ctx context.Context, _ string) (Incr[string], error) {
		a0 := createDynamicMaps("a0")
		a1 := createDynamicMaps("a1")
		bind := BindContext(bv, func(_ context.Context, which string) (Incr[string], error) {
			switch which {
			case "a":
				return Map(a0, func(v string) string {
					return v + "->" + "b"
				}), nil
			case "b":
				return Map(a1, func(v string) string {
					return v + "->" + "b"
				}), nil
			default:
				return nil, fmt.Errorf("invalid bind node selector: %v", which)
			}
		})
		bind.Node().SetLabel(fmt.Sprintf("bind - %s", "b"))
		TracePrintf(ctx, "returning new bind node")
		return bind, nil
	})
	c.Node().SetLabel("c")
	final := Map2(c, Return("final"), func(a, b string) string {
		return a + "->" + b
	})

	g := New()
	o := Observe(g, final)

	TracePrintln(ctx, "first stabilization")
	err := g.Stabilize(ctx)
	testutil.ItsNil(t, err)

	testutil.ItsNil(t, g.recomputeHeap.sanityCheck())
	testutil.ItsEqual(t, "a0-0+a0-1->b->final", o.Value())

	cv.Set("b")
	TracePrintln(ctx, "setting c to 'b'")
	err = g.Stabilize(ctx)
	testutil.ItsNil(t, err)

	testutil.ItsNil(t, g.recomputeHeap.sanityCheck())
	testutil.ItsEqual(t, "a0-0+a0-1->b->final", o.Value())

	bv.Set("b")
	TracePrintln(ctx, "setting b to 'b'")
	err = g.Stabilize(ctx)
	testutil.ItsNil(t, err)

	testutil.ItsNil(t, g.recomputeHeap.sanityCheck())
	testutil.ItsNil(t, g.recomputeHeap.sanityCheck())
	testutil.ItsEqual(t, "a1-0+a1-1->b->final", o.Value())

	bv.Set("a")
	TracePrintln(ctx, "setting b to 'a'")
	err = g.Stabilize(ctx)
	testutil.ItsNil(t, err)

	testutil.ItsNil(t, g.recomputeHeap.sanityCheck())
	testutil.ItsEqual(t, "a0-0+a0-1->b->final", o.Value())

	cv.Set("a")
	TracePrintln(ctx, "setting c to 'a'")
	err = g.Stabilize(ctx)
	testutil.ItsNil(t, err)

	testutil.ItsNil(t, g.recomputeHeap.sanityCheck())
	testutil.ItsEqual(t, "a0-0+a0-1->b->final", o.Value())
}

func Test_Bind_nested_bindUnobserved(t *testing.T) {
	ctx := testContext()

	v := Var("a")

	m0v := Var("foo")

	m0 := Map2(m0v, Return("bar"), concat)

	b := Bind(v, func(vv string) Incr[string] {
		return m0
	})

	g := New()
	o := Observe(g, b)

	err := g.Stabilize(ctx)
	testutil.ItsNil(t, err)
	testutil.ItsEqual(t, "foobar", o.Value())

	v.Set("b")
	m0v.Set("loo")
	err = g.Stabilize(ctx)
	testutil.ItsNil(t, err)
	testutil.ItsEqual(t, "loobar", o.Value())
}
