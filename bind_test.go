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

func Test_Bind_nestedUnlinksBind(t *testing.T) {
	ctx := testContext()
	g := New()

	b01 := Bind(Var("a"), func(_ string) Incr[string] {
		return Return("b01")
	})
	b01.Node().SetLabel("b01")
	b00 := Bind(Var("a"), func(_ string) Incr[string] {
		return b01
	})
	b00.Node().SetLabel("b00")

	b11 := Bind(Var("a"), func(_ string) Incr[string] {
		return Return("b11")
	})
	b11.Node().SetLabel("b11")
	b10 := Bind(Var("a"), func(_ string) Incr[string] {
		return b11
	})
	b10.Node().SetLabel("b10")

	bv := Var("a")
	b := Bind(bv, func(vv string) Incr[string] {
		if vv == "a" {
			return b00
		}
		return b10
	})
	b.Node().SetLabel("b")

	o := Observe(g, b)

	err := g.Stabilize(ctx)
	testutil.ItsNil(t, err)
	testutil.ItsNil(t, dumpDot(g, homedir("bind_unobserve_00.png")))
	testutil.ItsEqual(t, "b01", o.Value())

	testutil.ItsEqual(t, 1, bv.Node().height)
	testutil.ItsEqual(t, 2, b.Node().height)
	testutil.ItsEqual(t, 3, o.Node().height)
	testutil.ItsEqual(t, 2, b01.Node().height)
	testutil.ItsEqual(t, 2, b00.Node().height)

	bv.Set("b")

	err = g.Stabilize(ctx)
	testutil.ItsNil(t, err)
	testutil.ItsNil(t, dumpDot(g, homedir("bind_unobserve_01.png")))
	testutil.ItsEqual(t, "b11", o.Value())

	testutil.ItsEqual(t, 1, bv.Node().height)
	testutil.ItsEqual(t, 2, b.Node().height)
	testutil.ItsEqual(t, 3, o.Node().height)
	testutil.ItsEqual(t, 2, b11.Node().height)
	testutil.ItsEqual(t, 2, b10.Node().height)

	testutil.ItsEqual(t, false, g.IsObserving(b00))
	testutil.ItsEqual(t, false, o.Node().HasParent(b00.Node().ID()))
	testutil.ItsEqual(t, false, g.IsObserving(b01))
	testutil.ItsEqual(t, false, o.Node().HasParent(b01.Node().ID()))

	testutil.ItsEqual(t, true, g.IsObserving(b10))
	testutil.ItsEqual(t, true, o.Node().HasParent(b10.Node().ID()))
	testutil.ItsEqual(t, true, g.IsObserving(b11))
	testutil.ItsEqual(t, true, o.Node().HasParent(b11.Node().ID()))

	bv.Set("a")

	err = g.Stabilize(ctx)
	testutil.ItsNil(t, err)
	testutil.ItsNil(t, dumpDot(g, homedir("bind_unobserve_02.png")))
	testutil.ItsEqual(t, "b01", b00.Value())
	testutil.ItsEqual(t, "b01", o.Value())

	testutil.ItsEqual(t, true, g.IsObserving(b00))
	testutil.ItsEqual(t, true, o.Node().HasParent(b00.Node().ID()))
	testutil.ItsEqual(t, true, g.IsObserving(b01))
	testutil.ItsEqual(t, true, o.Node().HasParent(b01.Node().ID()))

	testutil.ItsEqual(t, false, g.IsObserving(b10))
	testutil.ItsEqual(t, false, o.Node().HasParent(b10.Node().ID()))
	testutil.ItsEqual(t, false, g.IsObserving(b11))
	testutil.ItsEqual(t, false, o.Node().HasParent(b11.Node().ID()))
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

	driver01var := Var("a")
	driver01var.Node().SetLabel("driver01var")
	driver01 := Bind(driver01var, func(_ string) Incr[string] {
		r := Return("driver01")
		r.Node().SetLabel("driver01return")
		return r
	})
	driver01.Node().SetLabel("driver01")

	driver02var := Var("a")
	driver02var.Node().SetLabel("driver02var")
	driver02 := Bind(driver02var, func(_ string) Incr[string] {
		return driver01
	})
	driver02.Node().SetLabel("driver02")

	m2 := Map2(driver01, driver02, concat)
	m2.Node().SetLabel("m2")
	o := Observe(g, m2)
	o.Node().SetLabel("observem2")

	err := g.Stabilize(ctx)
	testutil.ItsNil(t, err)
	testutil.ItsNil(t, dumpDot(g, homedir("bind_user_00.png")))
	testutil.ItsEqual(t, "driver01driver01", o.Value())
}

func Test_Bind_regression(t *testing.T) {
	cache := make(map[string]Incr[*int])

	fakeFormula := Var("fakeformula")
	var f func(t int) Incr[*int]
	f = func(t int) Incr[*int] {
		key := fmt.Sprintf("f-%d", t)
		if _, ok := cache[key]; ok {
			return cache[key]
		}
		r := Bind(fakeFormula, func(formula string) Incr[*int] {
			key := fmt.Sprintf("map-f-%d", t)
			if _, ok := cache[key]; ok {
				return cache[key]
			}
			if t == 0 {
				out := 0
				return Return(&out)
			}
			bindOutput := Map(f(t-1), func(r *int) *int {
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

	g := func(t int) Incr[*int] {
		key := fmt.Sprintf("g-%d", t)
		if _, ok := cache[key]; ok {
			return cache[key]
		}
		r := Bind(fakeFormula, func(formula string) Incr[*int] {
			output := f(t)
			return output
		})
		r.Node().SetLabel(key)
		cache[key] = r
		return r
	}

	h := func(t int) Incr[*int] {
		b := Bind(fakeFormula, func(formula string) Incr[*int] {
			var m2 Incr[*int]
			m2 = Map2(g(t), Return(10), func(l *int, r int) *int {
				if l == nil {
					return nil
				}
				out := *l * r
				return &out
			})
			m2.Node().SetLabel("h-m2")
			return m2
		})
		b.Node().SetLabel("h")
		return b
	}

	o := Map3(f(2), g(2), h(2), func(first *int, second *int, third *int) *int {
		if first == nil || second == nil || third == nil {
			return nil
		}
		out := *first + *second + *third
		return &out
	})
	o.Node().SetLabel("map3-final")

	graph := New()
	_ = Observe(graph, o)

	ctx := testContext()
	_ = graph.Stabilize(ctx)
	_ = dumpDot(graph, homedir("bind_regression.png"))

	testutil.ItsNotNil(t, o.Value())
	testutil.ItsEqual(t, 24, *o.Value())
}
