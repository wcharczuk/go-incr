package incr

import (
	"context"
	"fmt"
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_Bind_basic(t *testing.T) {
	ctx := testContext()

	bindVar := Var(ctx, "a")
	bindVar.Node().SetLabel("bindVar")

	av := Var(ctx, "a-value")
	av.Node().SetLabel("av")
	a0 := Map(ctx, av, ident)
	a0.Node().SetLabel("a0")
	a1 := Map(ctx, a0, ident)
	a1.Node().SetLabel("a1")

	bv := Var(ctx, "b-value")
	bv.Node().SetLabel("bv")
	b0 := Map(ctx, bv, ident)
	b0.Node().SetLabel("b0")
	b1 := Map(ctx, b0, ident)
	b1.Node().SetLabel("b1")
	b2 := Map(ctx, b1, ident)
	b2.Node().SetLabel("b2")

	bind := Bind(ctx, bindVar, func(_ context.Context, which string) Incr[string] {
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

	s0 := Return(ctx, "hello")
	s0.Node().SetLabel("s0")
	s1 := Map(ctx, s0, ident)
	s1.Node().SetLabel("s1")

	o := Map2(ctx, bind, s1, concat)
	o.Node().SetLabel("o")

	g := New()
	_ = Observe(ctx, g, o)

	var err error

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

	err = g.Stabilize(ctx)
	testutil.ItsNil(t, err)

	err = dumpDot(g, homedir("bind_basic_00.png"))
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
	testutil.ItsEqual(t, "a-value", bind.Value())
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

	t1 := Map(ctx, Return(ctx, "t1"), mapAppend("-mapped"))
	t1.Node().SetLabel("t1")
	t2 := Bind(ctx, Return(ctx, "t2"), func(ctx context.Context, _ string) Incr[string] {
		t3 := Map(ctx, Return(ctx, "t3"), mapAppend("-mapped"))
		t3.Node().SetLabel("t3")
		r := Map2(ctx, t1, t3, concat)
		return r
	})
	t2.Node().SetLabel("t2")

	g := New()
	o := Observe(ctx, g, t2)

	err := g.Stabilize(ctx)
	testutil.ItsNil(t, err)
	testutil.ItsEqual(t, "t1-mappedt3-mapped", o.Value())
	testutil.ItsNil(t, t1.Node().createdIn, "t1 should have an empty scope list as it was created outside a bind")
}

func Test_Bind_rebind(t *testing.T) {
	ctx := testContext()

	bindVar := Var(ctx, "a")
	bindVar.Node().SetLabel("bindVar")

	av := Var(ctx, "a-value")
	av.Node().SetLabel("av")
	a0 := Map(ctx, av, ident)
	a0.Node().SetLabel("a0")
	a1 := Map(ctx, a0, ident)
	a1.Node().SetLabel("a1")

	bv := Var(ctx, "b-value")
	bv.Node().SetLabel("bv")
	b0 := Map(ctx, bv, ident)
	b0.Node().SetLabel("b0")
	b1 := Map(ctx, b0, ident)
	b1.Node().SetLabel("b1")
	b2 := Map(ctx, b1, ident)
	b2.Node().SetLabel("b2")

	bind := Bind(ctx, bindVar, func(_ context.Context, which string) Incr[string] {
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

	s0 := Return(ctx, "hello")
	s0.Node().SetLabel("s0")
	s1 := Map(ctx, s0, ident)
	s1.Node().SetLabel("s1")

	o := Map2(ctx, bind, s1, concat)
	o.Node().SetLabel("o")

	g := New()
	_ = Observe(ctx, g, o)

	var err error

	err = g.Stabilize(ctx)
	testutil.ItsNil(t, err)
	testutil.ItsEqual(t, 1, bind.Node().boundAt)

	testutil.ItsEqual(t, "a-value", av.Value())
	testutil.ItsEqual(t, "b-value", bv.Value())
	testutil.ItsEqual(t, "a-valuehello", o.Value())

	bindVar.Set("a")

	err = g.Stabilize(ctx)
	testutil.ItsNil(t, err)
	testutil.ItsEqual(t, 1, bind.Node().boundAt)

	testutil.ItsEqual(t, "a-value", av.Value())
	testutil.ItsEqual(t, "b-value", bv.Value())
	testutil.ItsEqual(t, "a-valuehello", o.Value())
}

func Test_Bind_error(t *testing.T) {
	ctx := testContext()

	v0 := Var(ctx, "a")
	bind := BindContext(ctx, v0, func(_ context.Context, which string) (Incr[string], error) {
		return nil, fmt.Errorf("this is just a test")
	})
	bind.Node().SetLabel("bind")
	var gotError error
	bind.Node().OnError(func(_ context.Context, err error) {
		gotError = err
	})

	o := Map(ctx, bind, ident)

	g := New()
	_ = Observe(ctx, g, o)
	err := g.Stabilize(ctx)
	testutil.ItsNotNil(t, err)
	testutil.ItsEqual(t, "this is just a test", err.Error())
	testutil.ItsEqual(t, "this is just a test", gotError.Error())
}

func Test_Bind_nested(t *testing.T) {
	ctx := testContext()

	// a -> c
	// a -> b -> c
	// a -> c

	a0 := createDynamicMaps(ctx, "a0")
	a1 := createDynamicMaps(ctx, "a1")
	bv, b := createDynamicBind(ctx, "b", a0, a1)
	cv, c := createDynamicBind(ctx, "c", a0, b)
	final := Map2(ctx, c, Return(ctx, "final"), func(a, b string) string {
		return a + "->" + b
	})
	final.Node().SetLabel("final")

	g := New()
	o := Observe(ctx, g, final)

	err := g.Stabilize(ctx)
	testutil.ItsNil(t, err)
	err = dumpDot(g, homedir("bind_nested_00.png"))
	testutil.ItsNil(t, err)
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

func Test_Bind_nestedUnlinksBind(t *testing.T) {
	ctx := testContext()
	g := New()
	b01v := Var(ctx, "a")
	b01v.Node().SetLabel("b01v")
	b01 := Bind(ctx, b01v, func(ctx context.Context, _ string) Incr[string] {
		return Return(ctx, "b01")
	})
	b01.Node().SetLabel("b01")

	b00v := Var(ctx, "a")
	b00v.Node().SetLabel("b00v")
	b00 := Bind(ctx, b00v, func(_ context.Context, _ string) Incr[string] {
		return b01
	})
	b00.Node().SetLabel("b00")

	b11v := Var(ctx, "a")
	b11v.Node().SetLabel("b11v")
	b11 := Bind(ctx, b11v, func(ctx context.Context, _ string) Incr[string] {
		return Return(ctx, "b11")
	})
	b11.Node().SetLabel("b11")

	b10v := Var(ctx, "a")
	b10v.Node().SetLabel("b10v")
	b10 := Bind(ctx, b10v, func(_ context.Context, _ string) Incr[string] {
		return b11
	})
	b10.Node().SetLabel("b10")

	bv := Var(ctx, "a")
	bv.Node().SetLabel("bv")
	b := Bind(ctx, bv, func(_ context.Context, vv string) Incr[string] {
		if vv == "a" {
			return b00
		}
		return b10
	})
	b.Node().SetLabel("b")

	o := Observe(ctx, g, b)

	err := g.Stabilize(ctx)
	testutil.ItsNil(t, err)
	testutil.ItsNil(t, dumpDot(g, homedir("bind_unobserve_00_base.png")))
	testutil.ItsEqual(t, "b01", o.Value())

	bv.Set("b")

	err = g.Stabilize(ctx)
	testutil.ItsNil(t, err)
	testutil.ItsNil(t, dumpDot(g, homedir("bind_unobserve_01_switch_b.png")))
	testutil.ItsEqual(t, "b11", o.Value())

	testutil.ItsEqual(t, false, g.IsObserving(b00))
	testutil.ItsEqual(t, false, g.IsObserving(b01))

	testutil.ItsEqual(t, true, g.IsObserving(b10))
	testutil.ItsEqual(t, true, g.IsObserving(b11))

	bv.Set("a")

	err = g.Stabilize(ctx)
	testutil.ItsNil(t, err)
	testutil.ItsNil(t, dumpDot(g, homedir("bind_unobserve_02_switch_a.png")))
	testutil.ItsEqual(t, "b01", b00.Value())
	testutil.ItsEqual(t, "b01", o.Value())

	testutil.ItsEqual(t, true, g.IsObserving(b00))
	testutil.ItsEqual(t, true, g.IsObserving(b01))

	testutil.ItsEqual(t, false, g.IsObserving(b10))
	testutil.ItsEqual(t, false, g.IsObserving(b11))
}

func Test_Bind_nested_bindCreatesBind(t *testing.T) {
	ctx := testContext()

	cv := Var(ctx, "a")
	cv.Node().SetLabel("cv")
	bv := Var(ctx, "a")
	bv.Node().SetLabel("bv")
	c := BindContext[string](ctx, cv, func(ctx context.Context, _ string) (Incr[string], error) {
		a0 := createDynamicMaps(ctx, "a0")
		a1 := createDynamicMaps(ctx, "a1")
		bind := BindContext(ctx, bv, func(ctx context.Context, which string) (Incr[string], error) {
			switch which {
			case "a":
				return Map(ctx, a0, func(v string) string {
					return v + "->" + "b"
				}), nil
			case "b":
				return Map(ctx, a1, func(v string) string {
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
	final := Map2(ctx, c, Return(ctx, "final"), func(a, b string) string {
		return a + "->" + b
	})

	g := New()
	o := Observe(ctx, g, final)

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

func Test_Bind_nested_bindHeightsChange(t *testing.T) {
	ctx := testContext()
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

	driver01var := Var(ctx, "a")
	driver01var.Node().SetLabel("driver01var")
	driver01 := Bind(ctx, driver01var, func(ctx context.Context, _ string) Incr[string] {
		r := Return(ctx, "driver01")
		r.Node().SetLabel("driver01return")
		return r
	})
	driver01.Node().SetLabel("driver01")

	driver02var := Var(ctx, "a")
	driver02var.Node().SetLabel("driver02var")
	driver02 := Bind(ctx, driver02var, func(_ context.Context, _ string) Incr[string] {
		return driver01
	})
	driver02.Node().SetLabel("driver02")

	m2 := Map2(ctx, driver01, driver02, concat)
	m2.Node().SetLabel("m2")
	o := Observe(ctx, g, m2)
	o.Node().SetLabel("observem2")

	err := g.Stabilize(ctx)
	testutil.ItsNil(t, err)
	testutil.ItsNil(t, dumpDot(g, homedir("bind_user_00.png")))
	testutil.ItsEqual(t, "driver01driver01", o.Value())
}

func Test_Bind_regression(t *testing.T) {
	ctx := testContext()

	graph, o := makeRegressionGraph(ctx)
	_ = graph.Stabilize(ctx)
	_ = dumpDot(graph, homedir("bind_regression.png"))

	testutil.ItsNotNil(t, o.Value())
	testutil.ItsEqual(t, 24, *o.Value())
}

func Test_Bind_regression_parallel(t *testing.T) {
	ctx := testContext()

	graph, o := makeRegressionGraph(ctx)
	_ = graph.ParallelStabilize(ctx)
	_ = dumpDot(graph, homedir("bind_regression.png"))

	testutil.ItsNotNil(t, o.Value())
	testutil.ItsEqual(t, 24, *o.Value())
}

func makeRegressionGraph(ctx context.Context) (*Graph, ObserveIncr[*int]) {
	cache := make(map[string]Incr[*int])

	fakeFormula := Var(ctx, "fakeformula")
	fakeFormula.Node().SetLabel("fakeformula")
	var f func(context.Context, int) Incr[*int]
	f = func(ctx context.Context, t int) Incr[*int] {
		key := fmt.Sprintf("f-%d", t)
		if _, ok := cache[key]; ok {
			return WithBindScope(ctx, cache[key])
		}
		r := Bind(ctx, fakeFormula, func(ctx context.Context, formula string) Incr[*int] {
			key := fmt.Sprintf("map-f-%d", t)
			if _, ok := cache[key]; ok {
				return WithBindScope(ctx, cache[key])
			}
			if t == 0 {
				out := 0
				r := Return(ctx, &out)
				r.Node().SetLabel("f-0")
				return r
			}
			bindOutput := Map(ctx, f(ctx, t-1), func(r *int) *int {
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

	g := func(ctx context.Context, t int) Incr[*int] {
		key := fmt.Sprintf("g-%d", t)
		if _, ok := cache[key]; ok {
			return WithBindScope(ctx, cache[key])
		}
		r := Bind(ctx, fakeFormula, func(ctx context.Context, formula string) Incr[*int] {
			output := f(ctx, t)
			return output
		})
		r.Node().SetLabel(key)
		cache[key] = r
		return r
	}

	h := func(ctx context.Context, t int) Incr[*int] {
		b := Bind(ctx, fakeFormula, func(ctx context.Context, formula string) Incr[*int] {
			hr := Return(ctx, 10)
			hr.Node().SetLabel(fmt.Sprintf("h-r-%d", t))
			hm2 := Map2(ctx, g(ctx, t), hr, func(l *int, r int) *int {
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

	o := Map3(ctx, f(ctx, 2), g(ctx, 2), h(ctx, 2), func(first *int, second *int, third *int) *int {
		if first == nil || second == nil || third == nil {
			return nil
		}
		out := *first + *second + *third
		return &out
	})
	o.Node().SetLabel("map3-final")

	graph := New()
	return graph, Observe(ctx, graph, o)
}

func Test_bindChange_value(t *testing.T) {
	bc := &bindChangeIncr[string, string]{
		rhs: nil,
	}

	testutil.ItsEqual(t, "", bc.Value())

	bc.rhs = &returnIncr[string]{
		v: "hello",
	}

	testutil.ItsEqual(t, "hello", bc.Value())
}
