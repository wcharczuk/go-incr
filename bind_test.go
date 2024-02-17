package incr

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_Bind_basic(t *testing.T) {
	ctx := testContext()
	g := New()

	bindVar := Var(g, "a")
	bindVar.Node().SetLabel("bindVar")

	av := Var(g, "a-value")
	av.Node().SetLabel("av")
	a0 := Map(g, av, ident)
	a0.Node().SetLabel("a0")
	a1 := Map(g, a0, ident)
	a1.Node().SetLabel("a1")

	bv := Var(g, "b-value")
	bv.Node().SetLabel("bv")
	b0 := Map(g, bv, ident)
	b0.Node().SetLabel("b0")
	b1 := Map(g, b0, ident)
	b1.Node().SetLabel("b1")
	b2 := Map(g, b1, ident)
	b2.Node().SetLabel("b2")

	bind := Bind(g, bindVar, func(_ Scope, which string) Incr[string] {
		if which == "a" {
			return a1
		}
		if which == "b" {
			return b2
		}
		return nil
	})

	bind.Node().SetLabel("bind")

	testutil.Matches(t, "bind\\[(.*)\\]:bind", bind.String())

	s0 := Return(g, "hello")
	s0.Node().SetLabel("s0")
	s1 := Map(g, s0, ident)
	s1.Node().SetLabel("s1")

	o := Map2(g, bind, s1, concat)
	o.Node().SetLabel("o")
	_ = MustObserve(g, o)

	var err error
	testutil.Equal(t, 0, bindVar.Node().height)
	testutil.Equal(t, 0, s0.Node().height)
	testutil.Equal(t, 1, s1.Node().height)

	testutil.Equal(t, 2, bind.Node().height)
	testutil.Equal(t, 3, o.Node().height)

	testutil.Equal(t, true, g.Has(bindVar))
	testutil.Equal(t, true, g.Has(s0))
	testutil.Equal(t, true, g.Has(s1))
	testutil.Equal(t, true, g.Has(bind))
	testutil.Equal(t, true, g.Has(o))

	testutil.Equal(t, false, g.Has(av))
	testutil.Equal(t, false, g.Has(a0))
	testutil.Equal(t, false, g.Has(a1))

	testutil.Equal(t, false, g.Has(bv))
	testutil.Equal(t, false, g.Has(b0))
	testutil.Equal(t, false, g.Has(b1))
	testutil.Equal(t, false, g.Has(b2))

	err = g.Stabilize(ctx)
	testutil.Nil(t, err)

	_ = dumpDot(g, homedir("bind_basic_00.png"))

	testutil.Equal(t, 1, bind.Node().recomputedAt)
	testutil.Equal(t, 1, bind.Node().changedAt)
	testutil.Equal(t, 1, a1.Node().changedAt)

	testutil.Equal(t, 0, bindVar.Node().height)
	testutil.Equal(t, 0, s0.Node().height)
	testutil.Equal(t, 1, s1.Node().height)

	testutil.Equal(t, 0, av.Node().height)
	testutil.Equal(t, 1, a0.Node().height)
	testutil.Equal(t, 2, a1.Node().height)

	testutil.Equal(t, 3, bind.Node().height)
	testutil.Equal(t, 4, o.Node().height)

	testutil.Equal(t, true, g.Has(bindVar))
	testutil.Equal(t, true, g.Has(s0))
	testutil.Equal(t, true, g.Has(s1))
	testutil.Equal(t, true, g.Has(bind))
	testutil.Equal(t, true, g.Has(o))

	testutil.Equal(t, true, g.Has(av))
	testutil.Equal(t, true, g.Has(a0))
	testutil.Equal(t, true, g.Has(a1))

	testutil.Equal(t, false, g.Has(bv))
	testutil.Equal(t, false, g.Has(b0))
	testutil.Equal(t, false, g.Has(b1))
	testutil.Equal(t, false, g.Has(b2))

	testutil.Equal(t, "a-value", av.Value())
	testutil.Equal(t, "a-value", bind.Value())
	testutil.Equal(t, "a-valuehello", o.Value())

	bindVar.Set("b")
	err = g.Stabilize(ctx)
	testutil.Nil(t, err)

	err = dumpDot(g, homedir("bind_basic_01.png"))
	testutil.Nil(t, err)

	testutil.Equal(t, 0, bindVar.Node().height)
	testutil.Equal(t, 0, s0.Node().height)
	testutil.Equal(t, 1, s1.Node().height)

	testutil.Equal(t, HeightUnset, av.Node().height)
	testutil.Equal(t, HeightUnset, a0.Node().height)
	testutil.Equal(t, HeightUnset, a1.Node().height)

	testutil.Equal(t, 0, bv.Node().height)
	testutil.Equal(t, 1, b0.Node().height)
	testutil.Equal(t, 2, b1.Node().height)
	testutil.Equal(t, 3, b2.Node().height)

	testutil.Equal(t, 4, bind.Node().height)
	testutil.Equal(t, 5, o.Node().height)

	testutil.Equal(t, true, g.Has(bindVar))
	testutil.Equal(t, true, g.Has(s0))
	testutil.Equal(t, true, g.Has(s1))
	testutil.Equal(t, true, g.Has(bind))
	testutil.Equal(t, true, g.Has(o))

	testutil.Equal(t, false, g.Has(av))
	testutil.Equal(t, false, g.Has(a0))
	testutil.Equal(t, false, g.Has(a1))

	testutil.Equal(t, true, g.Has(bv))
	testutil.Equal(t, true, g.Has(b0))
	testutil.Equal(t, true, g.Has(b1))
	testutil.Equal(t, true, g.Has(b2))

	testutil.Equal(t, "a-value", av.Value())
	testutil.Equal(t, "b-value", bv.Value())
	testutil.Equal(t, "b-valuehello", o.Value())

	bindVar.Set("neither")
	err = g.Stabilize(ctx)
	testutil.Nil(t, err)

	err = dumpDot(g, homedir("bind_basic_02.png"))
	testutil.Nil(t, err)

	testutil.Equal(t, 0, bindVar.Node().height)
	testutil.Equal(t, 0, s0.Node().height)
	testutil.Equal(t, 1, s1.Node().height)

	testutil.Equal(t, HeightUnset, av.Node().height)
	testutil.Equal(t, HeightUnset, a0.Node().height)
	testutil.Equal(t, HeightUnset, a1.Node().height)

	testutil.Equal(t, HeightUnset, bv.Node().height)
	testutil.Equal(t, HeightUnset, b0.Node().height)
	testutil.Equal(t, HeightUnset, b1.Node().height)
	testutil.Equal(t, HeightUnset, b2.Node().height)

	testutil.Equal(t, 4, bind.Node().height)
	testutil.Equal(t, 5, o.Node().height)

	testutil.Equal(t, true, g.Has(bindVar))
	testutil.Equal(t, true, g.Has(s0))
	testutil.Equal(t, true, g.Has(s1))
	testutil.Equal(t, true, g.Has(bind))
	testutil.Equal(t, true, g.Has(o))

	testutil.Equal(t, false, g.Has(av))
	testutil.Equal(t, false, g.Has(a0))
	testutil.Equal(t, false, g.Has(a1))

	testutil.Equal(t, false, g.Has(bv))
	testutil.Equal(t, false, g.Has(b0))
	testutil.Equal(t, false, g.Has(b1))
	testutil.Equal(t, false, g.Has(b2))

	testutil.Equal(t, "a-value", av.Value())
	testutil.Equal(t, "b-value", bv.Value())
	testutil.Equal(t, "hello", o.Value())
}

func Test_Bind_scopes(t *testing.T) {
	/*
		   {[
				let t1 = map ... in
		   		bind t2 ~f:(fun _ ->
		   			let t3 = map ... in
		   			map2 t1 t3 ~f:(...))
		   ]}

		   In this example, [t1] is created outside of [bind t2], whereas [t3] is created by the
		   right-hand side of [bind t2].  So, [t3] depends on [t2] (and has a greater height),
		   whereas [t1] does not.  And, in a stabilization in which [t2] changes, we are
		   guaranteed to not recompute the old [t3], but we have no such guarantee about [t1].
		   Furthermore, when [t2] changes, the old [t3] will be invalidated, whereas [t1] will
		   not.

	*/

	g := New()
	ctx := testContext()

	t1 := Map(g, Return(g, "t1"), mapAppend("-mapped"))
	t1.Node().SetLabel("t1")

	t2 := Bind(g, Return(g, "t2"), func(bs Scope, _ string) Incr[string] {
		rt3 := Return(bs, "t3")
		t3 := Map(bs, rt3, mapAppend("-mapped"))
		t3.Node().SetLabel("t3")
		r := Map2(bs, t1, t3, concat)
		r.Node().SetLabel("t2-map")
		return r
	})
	t2.Node().SetLabel("t2")

	o := MustObserve(g, t2)

	err := g.Stabilize(ctx)
	testutil.Nil(t, err)

	testutil.Equal(t, true, t1.Node().createdIn.isTopScope())
	testutil.Equal(t, true, g.Has(t1))
	testutil.Equal(t, "t1-mappedt3-mapped", o.Value())
	testutil.NotNil(t, t1.Node().createdIn)
	testutil.Equal(t, true, t1.Node().createdIn.isTopScope())
}

func Test_Bind_necessary(t *testing.T) {
	/*
		We need to unobserve nodes such that a node isn't fully unobserved as long as there is a valid
		path from an observer to that node *somewhere*.
	*/

	ctx := testContext()
	g := New()

	root := Var(g, "hello")
	root.Node().SetLabel("root")

	ma00 := Map(g, root, ident)
	ma00.Node().SetLabel("ma00")
	mb00 := Map(g, Return(g, "not-hello"), ident)
	mb00.Node().SetLabel("mb00")

	b0v := Var(g, "a")
	b0v.Node().SetLabel("b0v")
	b0 := Bind(g, b0v, func(_ Scope, v string) Incr[string] {
		if v == "a" {
			return ma00
		}
		return mb00
	})
	b0.Node().SetLabel("b0")

	ma01 := Map(g, root, ident)
	ma01.Node().SetLabel("ma01")
	mb01 := Map(g, Return(g, "not-hello"), ident)
	mb01.Node().SetLabel("mb01")
	b1v := Var(g, "a")
	b1v.Node().SetLabel("b1v")
	b1 := Bind(g, b1v, func(_ Scope, v string) Incr[string] {
		if v == "a" {
			return ma01
		}
		return mb01
	})
	b1.Node().SetLabel("b1")
	m2 := Map2(g, b0, b1, concat)
	m2.Node().SetLabel("join")
	o := MustObserve(g, m2)

	err := g.Stabilize(ctx)
	testutil.Nil(t, err)
	testutil.Equal(t, "hellohello", o.Value())

	_ = dumpDot(g, homedir("bind_necessary_00.png"))

	b0v.Set("b")
	err = g.Stabilize(ctx)
	testutil.Nil(t, err)

	_ = dumpDot(g, homedir("bind_necessary_01.png"))

	testutil.Equal(t, true, g.Has(root))
}

func Test_Bind_unbindConflict(t *testing.T) {
	/*
		We need to unbind such that if another node is returning
		a shared node we don't unbind that node's "reference" to the node.
	*/
	g := New()
	root := Var(g, "hello")
	root.Node().SetLabel("root")

	ma := Map(g, root, ident)
	ma.Node().SetLabel("ma")
	mb := Map(g, Return(g, "not-hello"), ident)
	mb.Node().SetLabel("mb")

	b0v := Var(g, "a")
	b0v.Node().SetLabel("b0v")
	b0 := Bind(g, b0v, func(_ Scope, v string) Incr[string] {
		if v == "a" {
			return ma
		}
		return mb
	})
	b0.Node().SetLabel("b0")

	b1v := Var(g, "a")
	b1v.Node().SetLabel("b1v")
	b1 := Bind(g, b1v, func(_ Scope, v string) Incr[string] {
		if v == "a" {
			return ma
		}
		return mb
	})
	b1.Node().SetLabel("b1")
	m2 := Map2(g, b0, b1, concat)
	m2.Node().SetLabel("join")
	o := MustObserve(g, m2)

	err := g.Stabilize(testContext())
	testutil.Nil(t, err)
	testutil.Equal(t, "hellohello", o.Value())

	_ = dumpDot(g, homedir("bind_unbind_confict_00.png"))

	b0v.Set("b")
	err = g.Stabilize(testContext())
	testutil.Nil(t, err)

	_ = dumpDot(g, homedir("bind_unbind_conflict_01.png"))

	testutil.Equal(t, true, g.Has(ma))
}

func Test_Bind_rebind(t *testing.T) {
	/*
		The goal here is to stress the narrow case that
		if you stabilize a bind twice, but the returned rhs
		is the same, nothing really should happen.
	*/
	ctx := testContext()
	g := New()

	bindVar := Var(g, "a")
	bindVar.Node().SetLabel("bindVar")

	av := Var(g, "a-value")
	av.Node().SetLabel("av")
	a0 := Map(g, av, ident)
	a0.Node().SetLabel("a0")
	a1 := Map(g, a0, ident)
	a1.Node().SetLabel("a1")

	bv := Var(g, "b-value")
	bv.Node().SetLabel("bv")
	b0 := Map(g, bv, ident)
	b0.Node().SetLabel("b0")
	b1 := Map(g, b0, ident)
	b1.Node().SetLabel("b1")
	b2 := Map(g, b1, ident)
	b2.Node().SetLabel("b2")

	bind := Bind(g, bindVar, func(_ Scope, which string) Incr[string] {
		if which == "a" {
			return a1
		}
		if which == "b" {
			return b2
		}
		return nil
	})
	bind.Node().SetLabel("bind")
	s0 := Return(g, "hello")
	s0.Node().SetLabel("s0")
	s1 := Map(g, s0, ident)
	s1.Node().SetLabel("s1")
	o := Map2(g, bind, s1, func(a, b string) string {
		return concat(a, b)
	})
	o.Node().SetLabel("o")

	var childRecomputes int
	o.Node().OnUpdate(func(_ context.Context) {
		childRecomputes++
	})

	_ = MustObserve(g, o)

	var err error
	err = g.Stabilize(ctx)
	testutil.Nil(t, err)
	testutil.Equal(t, 1, bind.Node().recomputedAt)
	testutil.Equal(t, 1, childRecomputes)

	testutil.Equal(t, "a-value", av.Value())
	testutil.Equal(t, "b-value", bv.Value())
	testutil.Equal(t, "a-valuehello", o.Value())

	bindVar.Set("a")

	err = g.Stabilize(ctx)
	testutil.Nil(t, err)
	testutil.Equal(t, 2, bind.Node().recomputedAt)
	testutil.Equal(t, 2, childRecomputes)

	testutil.Equal(t, "a-value", av.Value())
	testutil.Equal(t, "b-value", bv.Value())
	testutil.Equal(t, "a-valuehello", o.Value())
}

func Test_Bind_error(t *testing.T) {
	ctx := testContext()
	g := New()

	v0 := Var(g, "a")
	bind := BindContext(g, v0, func(_ context.Context, _ Scope, which string) (Incr[string], error) {
		return nil, fmt.Errorf("this is just a test")
	})
	bind.Node().SetLabel("bind")
	o := Map(g, bind, ident)

	_ = MustObserve(g, o)
	err := g.Stabilize(ctx)
	testutil.NotNil(t, err)
	testutil.Equal(t, "this is just a test", err.Error())
}

func Test_Bind_nested(t *testing.T) {

	// a -> c
	// a -> b -> c
	// a -> c

	ctx := testContext()
	g := New()

	a0 := createDynamicMaps(g, "a0")
	a1 := createDynamicMaps(g, "a1")

	bv, b := createDynamicBind(g, "b", a0, a1)
	cv, c := createDynamicBind(g, "c", a0, b)

	rfinal := Return(g, "final")
	rfinal.Node().SetLabel("return - final")
	final := Map2(g, c, rfinal, func(c, rf string) string {
		return c + "->" + rf
	})
	final.Node().SetLabel("final")

	o := MustObserve(g, final)

	err := g.Stabilize(ctx)
	testutil.Nil(t, err)
	err = dumpDot(g, homedir("bind_nested_00.png"))
	testutil.Nil(t, err)
	testutil.Nil(t, g.recomputeHeap.sanityCheck())
	testutil.Equal(t, "a0-0+a0-1->c->final", o.Value())

	cv.Set("b")
	_ = g.Stabilize(ctx)

	err = dumpDot(g, homedir("bind_nested_01.png"))
	testutil.Nil(t, err)

	testutil.Nil(t, g.recomputeHeap.sanityCheck())
	testutil.Equal(t, "a0-0+a0-1->b->c->final", o.Value())

	bv.Set("b")
	_ = g.Stabilize(ctx)

	err = dumpDot(g, homedir("bind_nested_02.png"))
	testutil.Nil(t, err)

	testutil.Nil(t, g.recomputeHeap.sanityCheck())
	testutil.Equal(t, "a1-0+a1-1->b->c->final", o.Value())

	bv.Set("a")
	_ = g.Stabilize(ctx)

	testutil.Nil(t, g.recomputeHeap.sanityCheck())
	testutil.Equal(t, "a0-0+a0-1->b->c->final", o.Value())

	cv.Set("a")
	_ = g.Stabilize(ctx)

	testutil.Nil(t, g.recomputeHeap.sanityCheck())
	testutil.Equal(t, "a0-0+a0-1->c->final", o.Value())
}

func Test_Bind_nested_unlinksBind(t *testing.T) {
	ctx := testContext()
	g := New()

	a00v := Var(g, "a00")
	a00v.Node().SetLabel("a00v")
	a00 := Bind(g, a00v, func(bs Scope, _ string) Incr[string] {
		return Return(bs, "a00")
	})
	a00.Node().SetLabel("a00")

	a01v := Var(g, "a01")
	a01v.Node().SetLabel("a01v")
	a01 := Bind(g, a01v, func(_ Scope, _ string) Incr[string] {
		return a00
	})
	a01.Node().SetLabel("a01")

	b00v := Var(g, "b00")
	b00v.Node().SetLabel("b11v")
	b00 := Bind(g, b00v, func(bs Scope, _ string) Incr[string] {
		return Return(bs, "b00")
	})
	b00.Node().SetLabel("b00")

	b01v := Var(g, "b01")
	b01v.Node().SetLabel("b01v")
	b01 := Bind(g, b01v, func(bs Scope, _ string) Incr[string] {
		return b00
	})
	b01.Node().SetLabel("b01")

	bindv := Var(g, "a")
	bindv.Node().SetLabel("bv")
	bind := Bind(g, bindv, func(bs Scope, vv string) Incr[string] {
		if vv == "a" {
			return a01
		}
		return b01
	})
	bind.Node().SetLabel("b")

	o := MustObserve(g, bind)

	err := g.Stabilize(ctx)
	testutil.Nil(t, err)
	testutil.Nil(t, dumpDot(g, homedir("bind_unobserve_00_base.png")))
	testutil.Equal(t, "a00", o.Value())

	testutil.Equal(t, true, g.Has(a00))
	testutil.Equal(t, true, g.Has(a01))
	testutil.Equal(t, false, g.Has(b00))
	testutil.Equal(t, false, g.Has(b01))

	bindv.Set("b")
	err = g.Stabilize(ctx)
	testutil.Nil(t, err)
	testutil.Nil(t, dumpDot(g, homedir("bind_unobserve_01_switch_b.png")))
	testutil.Equal(t, "b00", o.Value())

	testutil.Equal(t, false, g.Has(a00))
	testutil.Equal(t, false, g.Has(a01))
	testutil.Equal(t, true, g.Has(b00))
	testutil.Equal(t, true, g.Has(b01))

	bindv.Set("a")

	err = g.Stabilize(ctx)
	testutil.Nil(t, err)
	testutil.Nil(t, dumpDot(g, homedir("bind_unobserve_02_switch_a.png")))
	testutil.Equal(t, "a00", a01.Value())
	testutil.Equal(t, "a00", o.Value())

	testutil.Equal(t, true, g.Has(a00))
	testutil.Equal(t, true, g.Has(a01))

	testutil.Equal(t, false, g.Has(b00))
	testutil.Equal(t, false, g.Has(b01))
}

func Test_Bind_nested_bindCreatesBind(t *testing.T) {
	ctx := testContext()
	g := New()

	cv := Var(g, "a")
	cv.Node().SetLabel("cv")
	bv := Var(g, "a")
	bv.Node().SetLabel("bv")
	c := BindContext[string](g, cv, func(_ context.Context, bs Scope, _ string) (Incr[string], error) {
		a0 := createDynamicMaps(bs, "a0")
		a1 := createDynamicMaps(bs, "a1")
		bind := BindContext(bs, bv, func(_ context.Context, bs Scope, which string) (Incr[string], error) {
			switch which {
			case "a":
				return Map(bs, a0, func(v string) string {
					return v + "->" + "b"
				}), nil
			case "b":
				return Map(bs, a1, func(v string) string {
					return v + "->" + "b"
				}), nil
			default:
				return nil, fmt.Errorf("invalid bind node selector: %v", which)
			}
		})
		bind.Node().SetLabel(fmt.Sprintf("bind - %s", "b"))
		return bind, nil
	})
	c.Node().SetLabel("c")
	final := Map2(g, c, Return(g, "final"), func(a, b string) string {
		return a + "->" + b
	})

	o := MustObserve(g, final)

	err := g.Stabilize(ctx)
	testutil.Nil(t, err)

	testutil.Nil(t, g.recomputeHeap.sanityCheck())
	testutil.Equal(t, "a0-0+a0-1->b->final", o.Value())

	cv.Set("b")
	err = g.Stabilize(ctx)
	testutil.Nil(t, err)

	testutil.Nil(t, g.recomputeHeap.sanityCheck())
	testutil.Equal(t, "a0-0+a0-1->b->final", o.Value())

	bv.Set("b")
	err = g.Stabilize(ctx)
	testutil.Nil(t, err)

	testutil.Nil(t, g.recomputeHeap.sanityCheck())
	testutil.Nil(t, g.recomputeHeap.sanityCheck())
	testutil.Equal(t, "a1-0+a1-1->b->final", o.Value())

	bv.Set("a")
	err = g.Stabilize(testContext())
	testutil.Nil(t, err)

	testutil.Nil(t, g.recomputeHeap.sanityCheck())
	testutil.Equal(t, "a0-0+a0-1->b->final", o.Value())

	cv.Set("a")
	err = g.Stabilize(testContext())
	testutil.Nil(t, err)

	testutil.Nil(t, g.recomputeHeap.sanityCheck())
	testutil.Equal(t, "a0-0+a0-1->b->final", o.Value())
}

func Test_Bind_nested_bindHeightsChange(t *testing.T) {
	ctx := testContext()
	g := New()

	driver01var := Var(g, "a")
	driver01var.Node().SetLabel("driver01var")
	driver01 := Bind(g, driver01var, func(bs Scope, _ string) Incr[string] {
		r := Return(bs, "driver01")
		r.Node().SetLabel("driver01return")
		return r
	})
	driver01.Node().SetLabel("driver01")

	driver02var := Var(g, "a")
	driver02var.Node().SetLabel("driver02var")
	driver02 := Bind(g, driver02var, func(_ Scope, _ string) Incr[string] {
		return driver01
	})
	driver02.Node().SetLabel("driver02")

	m2 := Map2(g, driver01, driver02, concat)
	m2.Node().SetLabel("m2")
	o := MustObserve(g, m2)
	o.Node().SetLabel("observem2")

	err := g.Stabilize(ctx)
	testutil.Nil(t, err)
	testutil.Nil(t, dumpDot(g, homedir("bind_user_00.png")))
	testutil.Equal(t, "driver01driver01", o.Value())
}

func Test_Bind_regression(t *testing.T) {
	ctx := testContext()

	graph, o := makeRegressionGraph(ctx)
	_ = graph.Stabilize(ctx)
	_ = dumpDot(graph, homedir("bind_regression.png"))

	testutil.NotNil(t, o.Value())
	testutil.Equal(t, 24, *o.Value())
}

func Test_Bind_regression_parallel(t *testing.T) {
	ctx := testContext()

	graph, o := makeRegressionGraph(ctx)
	_ = graph.ParallelStabilize(ctx)
	_ = dumpDot(graph, homedir("bind_regression.png"))

	testutil.NotNil(t, o.Value())
	testutil.Equal(t, 24, *o.Value())
}

func makeRegressionGraph(ctx context.Context) (*Graph, ObserveIncr[*int]) {
	cacheMu := sync.Mutex{}
	cache := make(map[string]Incr[*int])

	getCached := func(key string) (out Incr[*int], ok bool) {
		cacheMu.Lock()
		out, ok = cache[key]
		cacheMu.Unlock()
		return
	}
	setCached := func(key string, i Incr[*int]) {
		cacheMu.Lock()
		cache[key] = i
		cacheMu.Unlock()
	}
	graph := New()
	fakeFormula := Var(graph, "fakeformula")
	fakeFormula.Node().SetLabel("fakeformula")
	var f func(Scope, int) Incr[*int]
	f = func(bs Scope, t int) Incr[*int] {
		key := fmt.Sprintf("f-%d", t)
		if cached, ok := getCached(key); ok {
			return WithinScope(bs, cached)
		}
		r := Bind(bs, fakeFormula, func(bs Scope, formula string) Incr[*int] {
			key := fmt.Sprintf("map-f-%d", t)
			if cached, ok := getCached(key); ok {
				return WithinScope(bs, cached)
			}
			if t == 0 {
				out := 0
				r := Return(bs, &out)
				r.Node().SetLabel("f-0")
				return r
			}
			bindOutput := Map(bs, f(bs, t-1), func(r *int) *int {
				out := *r + 1
				return &out
			})
			bindOutput.Node().SetLabel(fmt.Sprintf("map-f-%d", t))
			setCached(key, bindOutput)
			return bindOutput
		})
		r.Node().SetLabel(key)
		setCached(key, r)
		return r
	}

	g := func(bs Scope, t int) Incr[*int] {
		key := fmt.Sprintf("g-%d", t)
		if cached, ok := getCached(key); ok {
			return WithinScope(bs, cached)
		}
		r := Bind(bs, fakeFormula, func(bs Scope, formula string) Incr[*int] {
			output := f(bs, t)
			return output
		})
		r.Node().SetLabel(key)
		setCached(key, r)
		return r
	}

	h := func(bs Scope, t int) Incr[*int] {
		b := Bind(bs, fakeFormula, func(bs Scope, formula string) Incr[*int] {
			hr := Return(bs, 10)
			hr.Node().SetLabel(fmt.Sprintf("h-r-%d", t))
			hm2 := Map2(bs, g(bs, t), hr, func(l *int, r int) *int {
				if l == nil {
					return nil
				}
				out := *l * r
				return &out
			})
			hm2.Node().SetLabel("h-m2")
			return hm2
		})
		b.Node().SetLabel(fmt.Sprintf("h-%d", 2))
		return b
	}

	o := Map3(graph, f(graph, 2), g(graph, 2), h(graph, 2), func(first *int, second *int, third *int) *int {
		if first == nil || second == nil || third == nil {
			return nil
		}
		out := *first + *second + *third
		return &out
	})
	o.Node().SetLabel("map3-final")
	return graph, MustObserve(graph, o)
}

func Test_Bind_regression_neseted(t *testing.T) {
	graph := New()
	fakeFormula := Var(graph, "fakeFormula")
	cache := make(map[string]Incr[*int])

	var f func(Scope, int) Incr[*int]
	f = func(bs Scope, t int) Incr[*int] {
		key := fmt.Sprintf("f-%d", t)
		if _, ok := cache[key]; ok {
			return WithinScope(bs, cache[key])
		}
		r := Bind(bs, fakeFormula, func(bs Scope, formula string) Incr[*int] {
			key := fmt.Sprintf("map-f-%d", t)
			if _, ok := cache[key]; ok {
				return WithinScope(bs, cache[key])
			}
			if t == 0 {
				out := 0
				r := Return(bs, &out)
				r.Node().SetLabel("f-0")
				return r
			}
			bindOutput := Map(bs, f(bs, t-1), func(r *int) *int {
				out := *r + 1
				return &out
			})
			bindOutput.Node().SetLabel(fmt.Sprintf("map-f-%d", t))
			cache[key] = bindOutput
			return bindOutput
		})
		r.Node().SetLabel(key)
		cache[key] = r
		return r
	}

	b := func(bs Scope, t int) Incr[*int] {
		key := fmt.Sprintf("b-%d", t)
		if _, ok := cache[key]; ok {
			return cache[key]
		}

		r := Bind(bs, fakeFormula, func(bs Scope, formula string) Incr[*int] {
			return Map2(bs, f(bs, t), f(bs, t+2), func(l *int, r *int) *int {
				out := *l * *r
				return &out
			})
		})
		cache[key] = r
		return r
	}

	o := b(graph, 5)

	_ = MustObserve(graph, o)
	ctx := testContext()
	err := graph.Stabilize(ctx)

	testutil.Nil(t, err)
	testutil.NotNil(t, o.Value())
	testutil.Equal(t, 35, *o.Value())
}

func Test_Bind_unbindRegression(t *testing.T) {
	graph := New()
	fakeFormula := Var(graph, "fakeFormula")
	cache := make(map[string]Incr[*int])

	var m func(bs Scope, t int) Incr[*int]
	left_bound := 3
	right_bound := 9
	m = func(bs Scope, t int) Incr[*int] {
		key := fmt.Sprintf("m-%d", t)
		if _, ok := cache[key]; ok {
			return WithinScope(bs, cache[key])
		}

		r := Bind(bs, fakeFormula, func(bs Scope, formula string) Incr[*int] {
			if t == 0 {
				out := 0
				r := Return(bs, &out)
				r.Node().SetLabel("m-0")
				return r
			}

			var bindOutput Incr[*int]
			offset := 1
			if t >= left_bound && t < right_bound {
				li := m(bs, t-1)
				bindOutput = Map2(bs, li, Return(bs, &offset), func(l *int, r *int) *int {
					if l == nil || r == nil {
						return nil
					}
					out := *l + *r
					return &out
				})
			} else {
				bindOutput = m(bs, t-1)
			}

			bindOutput.Node().SetLabel(fmt.Sprintf("%s-output", key))
			return bindOutput
		})

		r.Node().SetLabel(fmt.Sprintf("m(%d)", t))
		cache[key] = r
		return r
	}

	t.Run("m(0) = 0; if 3 <= t < 9, m(t) = m(t-1) + 1 else m(t) = m(t-1) - passes", func(t *testing.T) {
		o := m(graph, 9)
		_ = MustObserve(graph, o)
		ctx := testContext()
		err := graph.Stabilize(ctx)
		testutil.Nil(t, err)
		testutil.NotNil(t, o.Value())
		testutil.Equal(t, 6, *o.Value())

		graph.SetStale(fakeFormula)
		err = graph.Stabilize(ctx)
		testutil.Nil(t, err)

		_ = dumpDot(graph, homedir("bind_unbind_regression_00.png"))
		testutil.NotNil(t, o.Value())
		testutil.Equal(t, 6, *o.Value())
	})
}

func Test_Bind_boundChange_doesntCauseRebind(t *testing.T) {
	ctx := testContext()
	g := New()

	v0 := Var(g, "foo")
	v0.Node().SetLabel("v0")
	m0 := Map(g, v0, ident)
	m0.Node().SetLabel("m0")

	var bindUpdates int
	bv := Var(g, "a")
	b := Bind(g, bv, func(bs Scope, _ string) Incr[string] {
		bindUpdates++
		return m0
	})
	b.Node().SetLabel("b")

	m1 := Map(g, b, ident)
	m1.Node().SetLabel("m1")

	var m1Updates int
	m1.Node().OnUpdate(func(_ context.Context) {
		m1Updates++
	})

	o := MustObserve(g, m1)

	err := g.Stabilize(ctx)
	testutil.NoError(t, err)

	testutil.Equal(t, "foo", o.Value())
	testutil.Equal(t, 1, bindUpdates)
	testutil.Equal(t, 1, m1Updates)

	v0.Set("not-foo")

	err = g.Stabilize(ctx)
	testutil.NoError(t, err)

	testutil.Equal(t, 1, bindUpdates)
	testutil.Equal(t, 2, m1Updates)
	testutil.Equal(t, "not-foo", o.Value())
}

func Test_Bind_nestedScopeHasGraph(t *testing.T) {
	ctx := testContext()
	g := New()

	bv := Var(g, "a")
	var ibv00, ibv01 VarIncr[string]
	b := Bind(g, bv, func(bindScope Scope, bvv string) Incr[string] {
		ibv00 = Var(bindScope, bvv)
		return Bind(bindScope, ibv00, func(bindScope Scope, ibvv string) Incr[string] {
			ibv01 = Var(bindScope, ibvv)
			return Map2(bindScope, ibv00, ibv01, concat)
		})
	})
	ob := MustObserve(g, b)
	_ = g.Stabilize(ctx)
	testutil.Equal(t, "aa", ob.Value())

	testutil.NotNil(t, GraphForNode(ibv00))
	testutil.NotNil(t, GraphForNode(ibv01))

	testutil.Equal(t, false, ibv00.Node().createdIn.isTopScope())
	testutil.Equal(t, false, ibv01.Node().createdIn.isTopScope())

	ibv00CreatedInID := ibv00.Node().createdIn.(*bind[string, string]).main.n.id
	ibv01CreatedInID := ibv01.Node().createdIn.(*bind[string, string]).main.n.id
	testutil.NotEqual(t, ibv00CreatedInID, ibv01CreatedInID)
}

func Test_Bind_rebindCachedScope(t *testing.T) {
	ctx := testContext()
	g := New()

	var binnerar, binnerbr VarIncr[string]
	var binnera, binnerb Incr[string]
	bouterv := Var(g, "a")
	bouter := Bind(g, bouterv, func(bindScope Scope, which string) Incr[string] {
		if which == "a" {
			if binnera != nil {
				return binnera
			}
			binnera = Bind(bindScope, Return(g, ""), func(bindScope Scope, _ string) Incr[string] {
				if binnerar != nil {
					return binnerar
				}
				binnerar = Var(bindScope, "b-inner-a-r")
				return binnerar
			})
			return binnera
		}
		if binnerb != nil {
			return binnerb
		}
		binnerb = Bind(bindScope, Return(g, ""), func(bindScope Scope, _ string) Incr[string] {
			if binnerbr != nil {
				return binnerbr
			}
			binnerbr = Var(bindScope, "b-inner-b-r")
			return binnerbr
		})
		return binnerb
	})

	ob := MustObserve(g, bouter)
	_ = g.Stabilize(ctx)
	testutil.Equal(t, "b-inner-a-r", ob.Value())

	_ = g.Stabilize(ctx)
	testutil.Equal(t, "b-inner-a-r", ob.Value())

	bouterv.Set("b")

	_ = g.Stabilize(ctx)
	testutil.Equal(t, "b-inner-b-r", ob.Value())

	_ = g.Stabilize(ctx)
	testutil.Equal(t, "b-inner-b-r", ob.Value())

	testutil.NotNil(t, binnerar.Node().createdIn)
	binnerar.Set("throaway")
}

func Test_Bind_observedInner(t *testing.T) {
	ctx := testContext()
	g := New()

	bva := Var(g, "a")
	ba := Bind(g, bva, func(bindScope Scope, which string) Incr[string] {
		if which == "a" {
			return Return(bindScope, "bva-a-value")
		}
		return Return(bindScope, "bva-b-value")
	})
	oba := MustObserve(g, ba)

	bvb := Var(g, "a")
	bb := Bind(g, bvb, func(bindScope Scope, which string) Incr[string] {
		if which == "a" {
			return Return(bindScope, "bvb-a-value")
		}
		return Return(bindScope, "bvb-b-value")
	})
	obb := MustObserve(g, bb)

	bvo := Var(g, "a")
	bo := Bind(g, bvo, func(bindScope Scope, which string) Incr[string] {
		if which == "a" {
			return ba
		}
		return bb
	})

	obo := MustObserve(g, bo)

	_ = g.Stabilize(ctx)

	testutil.Equal(t, "bva-a-value", oba.Value())
	testutil.Equal(t, "bvb-a-value", obb.Value())
	testutil.Equal(t, "bva-a-value", obo.Value())

	bvo.Set("b")

	_ = g.Stabilize(ctx)
	testutil.Equal(t, "bva-a-value", oba.Value())
	testutil.Equal(t, "bvb-a-value", obb.Value())
	testutil.Equal(t, "bvb-a-value", obo.Value())

	bvb.Set("b")

	_ = g.Stabilize(ctx)
	testutil.Equal(t, "bva-a-value", oba.Value())
	testutil.Equal(t, "bvb-b-value", obb.Value())
	testutil.Equal(t, "bvb-b-value", obo.Value())
}

func Test_Bind_observedInner_cached(t *testing.T) {
	ctx := testContext()
	g := New()

	baa := Var(g, "ba-a-value")
	bab := Var(g, "ba-b-value")

	bva := Var(g, "a")
	ba := Bind(g, bva, func(bindScope Scope, which string) Incr[string] {
		if which == "a" {
			return baa
		}
		return bab
	})
	oba := MustObserve(g, ba)

	bba := Var(g, "bb-a-value")
	bbb := Var(g, "bb-b-value")

	bvb := Var(g, "a")
	bb := Bind(g, bvb, func(bindScope Scope, which string) Incr[string] {
		if which == "a" {
			return bba
		}
		return bbb
	})
	obb := MustObserve(g, bb)

	bvo := Var(g, "a")
	bo := Bind(g, bvo, func(bindScope Scope, which string) Incr[string] {
		if which == "a" {
			return ba
		}
		return bb
	})

	obo := MustObserve(g, bo)

	_ = g.Stabilize(ctx)

	testutil.Equal(t, "ba-a-value", oba.Value())
	testutil.Equal(t, "bb-a-value", obb.Value())
	testutil.Equal(t, "ba-a-value", obo.Value())

	bvo.Set("b")

	_ = g.Stabilize(ctx)
	testutil.Equal(t, "ba-a-value", oba.Value())
	testutil.Equal(t, "bb-a-value", obb.Value())
	testutil.Equal(t, "bb-a-value", obo.Value())

	bvb.Set("b")

	_ = g.Stabilize(ctx)
	testutil.Equal(t, "ba-a-value", oba.Value())
	testutil.Equal(t, "bb-b-value", obb.Value())
	testutil.Equal(t, "bb-b-value", obo.Value())
}
