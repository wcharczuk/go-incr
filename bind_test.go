package incr

import (
	"context"
	"fmt"
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_Bind_basic(t *testing.T) {
	ctx := testContext()

	bindVar := Var(Root(), "a")
	bindVar.Node().SetLabel("bindVar")

	av := Var(Root(), "a-value")
	av.Node().SetLabel("av")
	a0 := Map(Root(), av, ident)
	a0.Node().SetLabel("a0")
	a1 := Map(Root(), a0, ident)
	a1.Node().SetLabel("a1")

	bv := Var(Root(), "b-value")
	bv.Node().SetLabel("bv")
	b0 := Map(Root(), bv, ident)
	b0.Node().SetLabel("b0")
	b1 := Map(Root(), b0, ident)
	b1.Node().SetLabel("b1")
	b2 := Map(Root(), b1, ident)
	b2.Node().SetLabel("b2")

	bind := Bind(Root(), bindVar, func(_ *BindScope, which string) Incr[string] {
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

	s0 := Return(Root(), "hello")
	s0.Node().SetLabel("s0")
	s1 := Map(Root(), s0, ident)
	s1.Node().SetLabel("s1")

	o := Map2(Root(), bind, s1, concat)
	o.Node().SetLabel("o")

	// we shouldn't have bind internals set up on construction
	testutil.Nil(t, bind.BindChange())
	testutil.Nil(t, bind.Bound())

	g := New()
	_ = Observe(Root(), g, o)

	// we shouldn't have bind internals set up after observation either
	testutil.Nil(t, bind.BindChange())
	testutil.Nil(t, bind.Bound())

	var err error
	testutil.Equal(t, 0, bindVar.Node().height)
	testutil.Equal(t, 0, s0.Node().height)
	testutil.Equal(t, 1, s1.Node().height)

	testutil.Equal(t, 1, bind.Node().height)
	testutil.Equal(t, 2, o.Node().height)

	testutil.Equal(t, true, g.IsObserving(bindVar))
	testutil.Equal(t, true, g.IsObserving(s0))
	testutil.Equal(t, true, g.IsObserving(s1))
	testutil.Equal(t, true, g.IsObserving(bind))
	testutil.Equal(t, true, g.IsObserving(o))

	testutil.Equal(t, false, g.IsObserving(av))
	testutil.Equal(t, false, g.IsObserving(a0))
	testutil.Equal(t, false, g.IsObserving(a1))

	testutil.Equal(t, false, g.IsObserving(bv))
	testutil.Equal(t, false, g.IsObserving(b0))
	testutil.Equal(t, false, g.IsObserving(b1))
	testutil.Equal(t, false, g.IsObserving(b2))

	err = g.Stabilize(ctx)
	testutil.Nil(t, err)

	err = dumpDot(g, homedir("bind_basic_00.png"))
	testutil.Nil(t, err)

	// we _should_ have bind internals set up after stabilization
	testutil.NotNil(t, bind.BindChange())
	testutil.NotNil(t, bind.Bound())

	bindChange := bind.BindChange()
	testutil.Equal(t, true, hasKey(bindChange.Node().parents, bindVar.Node().ID()))
	testutil.Equal(t, true, hasKey(bindChange.Node().children, a1.Node().ID()))
	testutil.Equal(t, false, hasKey(bindChange.Node().children, b1.Node().ID()))

	testutil.Equal(t, 1, bind.Node().boundAt)
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

	testutil.Equal(t, true, g.IsObserving(bindVar))
	testutil.Equal(t, true, g.IsObserving(s0))
	testutil.Equal(t, true, g.IsObserving(s1))
	testutil.Equal(t, true, g.IsObserving(bind))
	testutil.Equal(t, true, g.IsObserving(o))

	testutil.Equal(t, true, g.IsObserving(av))
	testutil.Equal(t, true, g.IsObserving(a0))
	testutil.Equal(t, true, g.IsObserving(a1))

	testutil.Equal(t, false, g.IsObserving(bv))
	testutil.Equal(t, false, g.IsObserving(b0))
	testutil.Equal(t, false, g.IsObserving(b1))
	testutil.Equal(t, false, g.IsObserving(b2))

	testutil.Equal(t, "a-value", av.Value())
	testutil.Equal(t, "a-value", bind.Value())
	testutil.Equal(t, "a-valuehello", o.Value())

	bindVar.Set("b")
	err = g.Stabilize(ctx)
	testutil.Nil(t, err)

	err = dumpDot(g, homedir("bind_basic_01.png"))
	testutil.Nil(t, err)

	bindChange = bind.BindChange()
	testutil.Equal(t, true, hasKey(bindChange.Node().parents, bindVar.Node().ID()))
	testutil.Equal(t, false, hasKey(bindChange.Node().children, a1.Node().ID()))
	testutil.Equal(t, true, hasKey(bindChange.Node().children, b2.Node().ID()))

	testutil.Equal(t, 0, bindVar.Node().height)
	testutil.Equal(t, 0, s0.Node().height)
	testutil.Equal(t, 1, s1.Node().height)

	testutil.Equal(t, 0, av.Node().height)
	testutil.Equal(t, 0, a0.Node().height)
	testutil.Equal(t, 0, a1.Node().height)

	testutil.Equal(t, 0, bv.Node().height)
	testutil.Equal(t, 1, b0.Node().height)
	testutil.Equal(t, 2, b1.Node().height)
	testutil.Equal(t, 3, b2.Node().height)

	testutil.Equal(t, 4, bind.Node().height)
	testutil.Equal(t, 5, o.Node().height)

	testutil.Equal(t, true, g.IsObserving(bindVar))
	testutil.Equal(t, true, g.IsObserving(s0))
	testutil.Equal(t, true, g.IsObserving(s1))
	testutil.Equal(t, true, g.IsObserving(bind))
	testutil.Equal(t, true, g.IsObserving(o))

	testutil.Equal(t, false, g.IsObserving(av), "if we switch to b, we should unobserve the 'a' tree")
	testutil.Equal(t, false, g.IsObserving(a0))
	testutil.Equal(t, false, g.IsObserving(a1))

	testutil.Equal(t, true, g.IsObserving(bv))
	testutil.Equal(t, true, g.IsObserving(b0))
	testutil.Equal(t, true, g.IsObserving(b1))
	testutil.Equal(t, true, g.IsObserving(b2))

	testutil.Equal(t, "a-value", av.Value())
	testutil.Equal(t, "b-value", bv.Value())
	testutil.Equal(t, "b-valuehello", o.Value())

	bindVar.Set("neither")
	err = g.Stabilize(ctx)
	testutil.Nil(t, err)

	err = dumpDot(g, homedir("bind_basic_02.png"))
	testutil.Nil(t, err)

	bindChange = bind.BindChange()
	testutil.Nil(t, bindChange)

	testutil.Equal(t, 0, bindVar.Node().height)
	testutil.Equal(t, 0, s0.Node().height)
	testutil.Equal(t, 1, s1.Node().height)

	testutil.Equal(t, 0, av.Node().height)
	testutil.Equal(t, 0, a0.Node().height)
	testutil.Equal(t, 0, a1.Node().height)

	testutil.Equal(t, 0, bv.Node().height)
	testutil.Equal(t, 0, b0.Node().height)
	testutil.Equal(t, 0, b1.Node().height)
	testutil.Equal(t, 0, b2.Node().height)

	testutil.Equal(t, 4, bind.Node().height)
	testutil.Equal(t, 5, o.Node().height)

	testutil.Equal(t, true, g.IsObserving(bindVar))
	testutil.Equal(t, true, g.IsObserving(s0))
	testutil.Equal(t, true, g.IsObserving(s1))
	testutil.Equal(t, true, g.IsObserving(bind))
	testutil.Equal(t, true, g.IsObserving(o))

	testutil.Equal(t, false, g.IsObserving(av))
	testutil.Equal(t, false, g.IsObserving(a0))
	testutil.Equal(t, false, g.IsObserving(a1))

	testutil.Equal(t, false, g.IsObserving(bv))
	testutil.Equal(t, false, g.IsObserving(b0))
	testutil.Equal(t, false, g.IsObserving(b1))
	testutil.Equal(t, false, g.IsObserving(b2))

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

	ctx := testContext()

	t1 := Map(Root(), Return(Root(), "t1"), mapAppend("-mapped"))
	t1.Node().SetLabel("t1")

	var rt3id, t3id, rid Identifier
	t2 := Bind(Root(), Return(Root(), "t2"), func(bs *BindScope, _ string) Incr[string] {
		rt3 := Return(bs, "t3")
		rt3id = rt3.Node().ID()
		t3 := Map(bs, rt3, mapAppend("-mapped"))
		t3.Node().SetLabel("t3")
		t3id = t3.Node().ID()
		r := Map2(bs, t1, t3, concat)
		rid = r.Node().ID()
		return r
	})
	t2.Node().SetLabel("t2")

	g := New()
	o := Observe(Root(), g, t2)

	err := g.Stabilize(ctx)
	testutil.Nil(t, err)
	testutil.Equal(t, "t1-mappedt3-mapped", o.Value())
	testutil.Nil(t, t1.Node().createdIn, "t1 should have an unset created_in as it was created outside a bind")

	scope := t2.(*bindIncr[string, string]).scope
	testutil.NotNil(t, scope)

	testutil.Equal(t, t2.Node().ID(), scope.bind.Node().ID())

	testutil.Equal(t, 3, len(scope.rhsNodes))

	testutil.Equal(t, true, hasKey(scope.rhsNodes, rt3id))
	testutil.Equal(t, true, hasKey(scope.rhsNodes, t3id))
	testutil.Equal(t, true, hasKey(scope.rhsNodes, rid))
}

func Test_Bind_necessary(t *testing.T) {
	/*
		We need to unobserve nodes such that a node isn't fully unobserved as long as there is a valid
		path from an observer to that node *somewhere*.
	*/
	root := Var(Root(), "hello")
	root.Node().SetLabel("root")

	ma00 := Map(Root(), root, ident)
	ma00.Node().SetLabel("ma00")
	mb00 := Map(Root(), Return(Root(), "not-hello"), ident)
	mb00.Node().SetLabel("mb00")

	b0v := Var(Root(), "a")
	b0v.Node().SetLabel("b0v")
	b0 := Bind(Root(), b0v, func(_ *BindScope, v string) Incr[string] {
		if v == "a" {
			return ma00
		}
		return mb00
	})
	b0.Node().SetLabel("b0")

	ma01 := Map(Root(), root, ident)
	ma01.Node().SetLabel("ma01")
	mb01 := Map(Root(), Return(Root(), "not-hello"), ident)
	mb01.Node().SetLabel("mb01")
	b1v := Var(Root(), "a")
	b1v.Node().SetLabel("b1v")
	b1 := Bind(Root(), b1v, func(_ *BindScope, v string) Incr[string] {
		if v == "a" {
			return ma01
		}
		return mb01
	})
	b1.Node().SetLabel("b1")
	m2 := Map2(Root(), b0, b1, concat)
	m2.Node().SetLabel("join")
	g := New()
	o := Observe(Root(), g, m2)

	testutil.Equal(t, true, hasKey(m2.Node().children, o.Node().id))

	err := g.Stabilize(testContext())
	testutil.Nil(t, err)
	testutil.Equal(t, "hellohello", o.Value())
	testutil.Equal(t, true, g.canReachObserver(root, o.Node().id))

	testutil.Equal(t, true, hasKey(m2.Node().children, o.Node().id))

	_ = dumpDot(g, homedir("bind_necessary_00.png"))

	b0v.Set("b")
	err = g.Stabilize(testContext())
	testutil.Nil(t, err)

	_ = dumpDot(g, homedir("bind_necessary_01.png"))

	testutil.Equal(t, true, g.canReachObserver(root, o.Node().id))
	testutil.Equal(t, true, g.IsObserving(root))
}

func Test_Bind_unbindConflict(t *testing.T) {
	/*
		We need to unbind such that if another node is returning
		a shared node we don't unbind that node's "reference" to the node.
	*/
	root := Var(Root(), "hello")
	root.Node().SetLabel("root")

	ma := Map(Root(), root, ident)
	ma.Node().SetLabel("ma")
	mb := Map(Root(), Return(Root(), "not-hello"), ident)
	mb.Node().SetLabel("mb")

	b0v := Var(Root(), "a")
	b0v.Node().SetLabel("b0v")
	b0 := Bind(Root(), b0v, func(_ *BindScope, v string) Incr[string] {
		if v == "a" {
			return ma
		}
		return mb
	})
	b0.Node().SetLabel("b0")

	b1v := Var(Root(), "a")
	b1v.Node().SetLabel("b1v")
	b1 := Bind(Root(), b1v, func(_ *BindScope, v string) Incr[string] {
		if v == "a" {
			return ma
		}
		return mb
	})
	b1.Node().SetLabel("b1")
	m2 := Map2(Root(), b0, b1, concat)
	m2.Node().SetLabel("join")
	g := New()
	o := Observe(Root(), g, m2)

	err := g.Stabilize(testContext())
	testutil.Nil(t, err)
	testutil.Equal(t, "hellohello", o.Value())
	testutil.Equal(t, true, g.canReachObserver(root, o.Node().id))

	_ = dumpDot(g, homedir("bind_unbind_confict_00.png"))

	b0v.Set("b")
	err = g.Stabilize(testContext())
	testutil.Nil(t, err)

	_ = dumpDot(g, homedir("bind_unbind_conflict_01.png"))

	testutil.Equal(t, true, g.IsObserving(ma))
}

func Test_Bind_rebind(t *testing.T) {
	/*
		The goal here is to stress the narrow case that
		if you stabilize a bind twice, but the returned rhs
		is the same, nothing really should happen.
	*/

	bindVar := Var(Root(), "a")
	bindVar.Node().SetLabel("bindVar")

	av := Var(Root(), "a-value")
	av.Node().SetLabel("av")
	a0 := Map(Root(), av, ident)
	a0.Node().SetLabel("a0")
	a1 := Map(Root(), a0, ident)
	a1.Node().SetLabel("a1")

	bv := Var(Root(), "b-value")
	bv.Node().SetLabel("bv")
	b0 := Map(Root(), bv, ident)
	b0.Node().SetLabel("b0")
	b1 := Map(Root(), b0, ident)
	b1.Node().SetLabel("b1")
	b2 := Map(Root(), b1, ident)
	b2.Node().SetLabel("b2")

	bind := Bind(Root(), bindVar, func(_ *BindScope, which string) Incr[string] {
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

	s0 := Return(Root(), "hello")
	s0.Node().SetLabel("s0")
	s1 := Map(Root(), s0, ident)
	s1.Node().SetLabel("s1")

	o := Map2(Root(), bind, s1, concat)
	o.Node().SetLabel("o")

	g := New()
	_ = Observe(Root(), g, o)

	var err error
	ctx := testContext()
	err = g.Stabilize(ctx)
	testutil.Nil(t, err)
	testutil.Equal(t, 1, bind.Node().boundAt)

	testutil.Equal(t, "a-value", av.Value())
	testutil.Equal(t, "b-value", bv.Value())
	testutil.Equal(t, "a-valuehello", o.Value())

	bindVar.Set("a")

	err = g.Stabilize(ctx)
	testutil.Nil(t, err)
	testutil.Equal(t, 1, bind.Node().boundAt)

	testutil.Equal(t, "a-value", av.Value())
	testutil.Equal(t, "b-value", bv.Value())
	testutil.Equal(t, "a-valuehello", o.Value())
}

func Test_Bind_error(t *testing.T) {
	v0 := Var(Root(), "a")
	bind := BindContext(Root(), v0, func(_ context.Context, _ *BindScope, which string) (Incr[string], error) {
		return nil, fmt.Errorf("this is just a test")
	})
	bind.Node().SetLabel("bind")
	var gotError error
	bind.Node().OnError(func(_ context.Context, err error) {
		gotError = err
	})

	o := Map(Root(), bind, ident)

	g := New()
	_ = Observe(Root(), g, o)
	ctx := testContext()
	err := g.Stabilize(ctx)
	testutil.NotNil(t, err)
	testutil.Equal(t, "this is just a test", err.Error())
	testutil.Equal(t, "this is just a test", gotError.Error())
}

func Test_Bind_nested(t *testing.T) {

	// a -> c
	// a -> b -> c
	// a -> c

	a0 := createDynamicMaps(Root(), "a0")
	a1 := createDynamicMaps(Root(), "a1")

	bv, b := createDynamicBind(Root(), "b", a0, a1)
	cv, c := createDynamicBind(Root(), "c", a0, b)

	rfinal := Return(Root(), "final")
	rfinal.Node().SetLabel("return - final")
	final := Map2(Root(), c, rfinal, func(c, rf string) string {
		return c + "->" + rf
	})
	final.Node().SetLabel("final")

	g := New()
	o := Observe(Root(), g, final)

	ctx := testContext()

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
	g := New()
	a00v := Var(Root(), "a00")
	a00v.Node().SetLabel("a00v")
	a00 := Bind(Root(), a00v, func(bs *BindScope, _ string) Incr[string] {
		return Return(bs, "a00")
	})
	a00.Node().SetLabel("a00")

	a01v := Var(Root(), "a01")
	a01v.Node().SetLabel("a01v")
	a01 := Bind(Root(), a01v, func(_ *BindScope, _ string) Incr[string] {
		return a00
	})
	a01.Node().SetLabel("a01")

	b00v := Var(Root(), "b00")
	b00v.Node().SetLabel("b11v")
	b00 := Bind(Root(), b00v, func(bs *BindScope, _ string) Incr[string] {
		return Return(bs, "b00")
	})
	b00.Node().SetLabel("b00")

	b01v := Var(Root(), "b01")
	b01v.Node().SetLabel("b01v")
	b01 := Bind(Root(), b01v, func(bs *BindScope, _ string) Incr[string] {
		return b00
	})
	b01.Node().SetLabel("b01")

	bindv := Var(Root(), "a")
	bindv.Node().SetLabel("bv")
	bind := Bind(Root(), bindv, func(bs *BindScope, vv string) Incr[string] {
		if vv == "a" {
			return a01
		}
		return b01
	})
	bind.Node().SetLabel("b")

	o := Observe(Root(), g, bind)

	ctx := testContext()
	err := g.Stabilize(ctx)
	testutil.Nil(t, err)
	testutil.Nil(t, dumpDot(g, homedir("bind_unobserve_00_base.png")))
	testutil.Equal(t, "a00", o.Value())

	testutil.Equal(t, true, g.IsObserving(a00))
	testutil.Equal(t, true, g.IsObserving(a01))
	testutil.Equal(t, false, g.IsObserving(b00))
	testutil.Equal(t, false, g.IsObserving(b01))

	bindv.Set("b")
	err = g.Stabilize(ctx)
	testutil.Nil(t, err)
	testutil.Nil(t, dumpDot(g, homedir("bind_unobserve_01_switch_b.png")))
	testutil.Equal(t, "b00", o.Value())

	testutil.Equal(t, false, g.IsObserving(a00))
	testutil.Equal(t, false, g.IsObserving(a01))
	testutil.Equal(t, true, g.IsObserving(b00))
	testutil.Equal(t, true, g.IsObserving(b01))

	bindv.Set("a")

	err = g.Stabilize(ctx)
	testutil.Nil(t, err)
	testutil.Nil(t, dumpDot(g, homedir("bind_unobserve_02_switch_a.png")))
	testutil.Equal(t, "a00", a01.Value())
	testutil.Equal(t, "a00", o.Value())

	testutil.Equal(t, true, g.IsObserving(a00))
	testutil.Equal(t, true, g.IsObserving(a01))

	testutil.Equal(t, false, g.IsObserving(b00))
	testutil.Equal(t, false, g.IsObserving(b01))
}

func Test_Bind_nested_bindCreatesBind(t *testing.T) {
	cv := Var(Root(), "a")
	cv.Node().SetLabel("cv")
	bv := Var(Root(), "a")
	bv.Node().SetLabel("bv")
	c := BindContext[string](Root(), cv, func(_ context.Context, scope *BindScope, _ string) (Incr[string], error) {
		a0 := createDynamicMaps(scope, "a0")
		a1 := createDynamicMaps(scope, "a1")
		bind := BindContext(Root(), bv, func(_ context.Context, bs *BindScope, which string) (Incr[string], error) {
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
	final := Map2(Root(), c, Return(Root(), "final"), func(a, b string) string {
		return a + "->" + b
	})

	g := New()
	o := Observe(Root(), g, final)

	err := g.Stabilize(testContext())
	testutil.Nil(t, err)

	testutil.Nil(t, g.recomputeHeap.sanityCheck())
	testutil.Equal(t, "a0-0+a0-1->b->final", o.Value())

	cv.Set("b")
	err = g.Stabilize(testContext())
	testutil.Nil(t, err)

	testutil.Nil(t, g.recomputeHeap.sanityCheck())
	testutil.Equal(t, "a0-0+a0-1->b->final", o.Value())

	bv.Set("b")
	err = g.Stabilize(testContext())
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
	g := New()

	/*
		User Notes:

		The problem is that a bind node, B starts out with height, h.
		Let’s say h=2. Then it gets recomputed and then it gets bound to C, which ends up h=5 after the discovery process. Great.

		Somewhere along the line in the computation because some other bind nodes, we have to link D to B through a Map.
		Guess what? D’s pseudo height is 3!. This is because B’s height is not in sync with what it is bound to (C in this case).
		B still thinks its height is 2.

		Then D gets recomputed because of its low height but the answer to B is not ready because C has not been stabilized yet.
	*/

	driver01var := Var(Root(), "a")
	driver01var.Node().SetLabel("driver01var")
	driver01 := Bind(Root(), driver01var, func(bs *BindScope, _ string) Incr[string] {
		r := Return(bs, "driver01")
		r.Node().SetLabel("driver01return")
		return r
	})
	driver01.Node().SetLabel("driver01")

	driver02var := Var(Root(), "a")
	driver02var.Node().SetLabel("driver02var")
	driver02 := Bind(Root(), driver02var, func(_ *BindScope, _ string) Incr[string] {
		return driver01
	})
	driver02.Node().SetLabel("driver02")

	m2 := Map2(Root(), driver01, driver02, concat)
	m2.Node().SetLabel("m2")
	o := Observe(Root(), g, m2)
	o.Node().SetLabel("observem2")

	ctx := testContext()
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
	cache := make(map[string]Incr[*int])

	fakeFormula := Var(Root(), "fakeformula")
	fakeFormula.Node().SetLabel("fakeformula")
	var f func(*BindScope, int) Incr[*int]
	f = func(bs *BindScope, t int) Incr[*int] {
		key := fmt.Sprintf("f-%d", t)
		if cached, ok := cache[key]; ok {
			return WithinBindScope(bs, cached)
		}
		r := Bind(bs, fakeFormula, func(bs *BindScope, formula string) Incr[*int] {
			key := fmt.Sprintf("map-f-%d", t)
			if cached, ok := cache[key]; ok {
				return WithinBindScope(bs, cached)
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

	g := func(bs *BindScope, t int) Incr[*int] {
		key := fmt.Sprintf("g-%d", t)
		if cached, ok := cache[key]; ok {
			return WithinBindScope(bs, cached)
		}
		r := Bind(bs, fakeFormula, func(bs *BindScope, formula string) Incr[*int] {
			output := f(bs, t)
			return output
		})
		r.Node().SetLabel(key)
		cache[key] = r
		return r
	}

	h := func(bs *BindScope, t int) Incr[*int] {
		b := Bind(bs, fakeFormula, func(bs *BindScope, formula string) Incr[*int] {
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

	o := Map3(Root(), f(Root(), 2), g(Root(), 2), h(Root(), 2), func(first *int, second *int, third *int) *int {
		if first == nil || second == nil || third == nil {
			return nil
		}
		out := *first + *second + *third
		return &out
	})
	o.Node().SetLabel("map3-final")

	graph := New()
	return graph, Observe(Root(), graph, o)
}

func Test_bindChange_value(t *testing.T) {
	bc := &bindChangeIncr[string, string]{
		rhs: nil,
	}

	testutil.Equal(t, "", bc.Value())

	bc.rhs = &returnIncr[string]{
		v: "hello",
	}

	testutil.Equal(t, "hello", bc.Value())
}

func Test_Bind_regression2(t *testing.T) {
	fakeFormula := Var(Root(), "fakeFormula")
	cache := make(map[string]Incr[*int])

	var f func(*BindScope, int) Incr[*int]
	f = func(bs *BindScope, t int) Incr[*int] {
		key := fmt.Sprintf("f-%d", t)
		if _, ok := cache[key]; ok {
			return WithinBindScope(bs, cache[key])
		}
		r := Bind(bs, fakeFormula, func(bs *BindScope, formula string) Incr[*int] {
			key := fmt.Sprintf("map-f-%d", t)
			if _, ok := cache[key]; ok {
				return WithinBindScope(bs, cache[key])
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

	b := func(bs *BindScope, t int) Incr[*int] {
		key := fmt.Sprintf("b-%d", t)
		if _, ok := cache[key]; ok {
			return cache[key]
		}

		r := Bind(bs, fakeFormula, func(bs *BindScope, formula string) Incr[*int] {
			return Map2(bs, f(bs, t), f(bs, t+2), func(l *int, r *int) *int {
				out := *l * *r
				return &out
			})
		})
		cache[key] = r
		return r
	}

	t.Run("b(t) = f(t) * f(t+2) where f(t) = f(t-1) + 1", func(t *testing.T) {
		o := b(Root(), 5)

		graph := New()
		_ = Observe(Root(), graph, o)
		ctx := testContext()
		err := graph.Stabilize(ctx)

		testutil.Nil(t, err)
		testutil.NotNil(t, o.Value())
		testutil.Equal(t, 35, *o.Value())
	})
}

func Test_Bind_unbindRegression(t *testing.T) {
	fakeFormula := Var(Root(), "fakeFormula")
	cache := make(map[string]Incr[*int])

	var m func(bs *BindScope, t int) Incr[*int]
	left_bound := 3
	right_bound := 9
	m = func(bs *BindScope, t int) Incr[*int] {
		key := fmt.Sprintf("m-%d", t)
		if _, ok := cache[key]; ok {
			return WithinBindScope(bs, cache[key])
		}

		r := Bind(bs, fakeFormula, func(bs *BindScope, formula string) Incr[*int] {
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
		graph := New()
		o := m(Root(), 9)
		_ = Observe(Root(), graph, o)
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

func Test_Bind_nested_amplification(t *testing.T) {
	ctx := testContext()
	depth := 4
	fakeFormula := Var(Root(), "fakeFormula")
	observed := make([]ObserveIncr[*int], depth)
	g := New(
		GraphMaxRecomputeHeapHeight(1024),
	)
	for i := 0; i < depth; i++ {
		o := makeSimpleNestedBindGraph(g, depth, fakeFormula)
		observed[i] = o
	}
	err := g.Stabilize(ctx)
	testutil.Nil(t, err)
	testutil.Equal(t, 65, g.numNodes)

	for _, o := range observed {
		testutil.NotNil(t, o.Value())
	}

	err = g.Stabilize(ctx)
	testutil.Nil(t, err)
	testutil.Equal(t, 65, g.numNodes)

	err = g.Stabilize(ctx)
	testutil.Nil(t, err)
	testutil.Equal(t, 65, g.numNodes)
}
