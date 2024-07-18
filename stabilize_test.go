package incr

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_Stabilize(t *testing.T) {
	ctx := testContext()
	g := New()

	v0 := Var(g, "foo")
	v1 := Var(g, "bar")
	m0 := Map2(g, v0, v1, func(a, b string) string {
		return a + " " + b
	})

	_ = MustObserve(g, m0)

	err := g.Stabilize(ctx)
	testutil.Nil(t, err)

	testutil.Equal(t, "foo bar", m0.Value())

	testutil.Equal(t, 0, v0.Node().setAt)
	testutil.Equal(t, 0, v0.Node().changedAt, "vars only are recomputed after the first set")
	testutil.Equal(t, 0, v1.Node().setAt)
	testutil.Equal(t, 0, v1.Node().changedAt)
	testutil.Equal(t, 1, m0.Node().changedAt)
	testutil.Equal(t, 0, v0.Node().recomputedAt)
	testutil.Equal(t, 0, v1.Node().recomputedAt)
	testutil.Equal(t, 1, m0.Node().recomputedAt)

	v0.Set("not foo")
	testutil.Equal(t, 2, v0.Node().setAt)
	testutil.Equal(t, 0, v1.Node().setAt)

	err = g.Stabilize(ctx)
	testutil.Nil(t, err)

	testutil.Equal(t, 2, v0.Node().changedAt)
	testutil.Equal(t, 0, v1.Node().changedAt)
	testutil.Equal(t, 2, m0.Node().changedAt)

	testutil.Equal(t, 2, v0.Node().recomputedAt)
	testutil.Equal(t, 0, v1.Node().recomputedAt)
	testutil.Equal(t, 2, m0.Node().recomputedAt)

	testutil.Equal(t, "not foo bar", m0.Value())
}

func Test_Stabilize_error_noClear(t *testing.T) {
	ctx := testContext()
	g := New(
		OptGraphClearRecomputeHeapOnError(false),
	)

	var didCallAbortedHandler bool
	v0 := Var(g, "hello")
	m0 := Map(g, v0, ident)
	m1 := Map(g, m0, ident)
	m1.Node().OnAborted(func(_ context.Context, err error) {
		didCallAbortedHandler = true
	})

	f0 := Func(g, func(_ context.Context) (string, error) {
		return "", fmt.Errorf("this is just a test")
	})

	_ = MustObserve(g, f0)
	_ = MustObserve(g, m1)

	testutil.Equal(t, true, g.recomputeHeap.has(m1))
	testutil.Equal(t, true, g.recomputeHeap.has(f0))

	err := g.Stabilize(ctx)
	testutil.NotNil(t, err)
	testutil.Equal(t, "this is just a test", err.Error())

	testutil.Equal(t, true, g.recomputeHeap.has(m1), "we should not clear the recompute heap on error")
	testutil.Equal(t, false, g.recomputeHeap.has(f0))
	testutil.Equal(t, false, didCallAbortedHandler)
}

func Test_Stabilize_error_shouldClear(t *testing.T) {
	ctx := testContext()
	g := New(
		OptGraphClearRecomputeHeapOnError(true),
	)

	var didCallAbortedHandler bool
	v0 := Var(g, "hello")
	m0 := Map(g, v0, ident)
	m1 := Map(g, m0, ident)
	m1.Node().OnAborted(func(_ context.Context, err error) {
		didCallAbortedHandler = true
	})

	f0 := Func(g, func(_ context.Context) (string, error) {
		return "", fmt.Errorf("this is just a test")
	})

	_ = MustObserve(g, f0)
	_ = MustObserve(g, m1)

	testutil.Equal(t, true, g.recomputeHeap.has(m1))
	testutil.Equal(t, true, g.recomputeHeap.has(f0))

	err := g.Stabilize(ctx)
	testutil.NotNil(t, err)
	testutil.Equal(t, "this is just a test", err.Error())

	testutil.Equal(t, false, g.recomputeHeap.has(m1), "we should clear the recompute heap on error")
	testutil.Equal(t, false, g.recomputeHeap.has(f0))
	testutil.Equal(t, true, didCallAbortedHandler)
}

func Test_Stabilize_errorHandler(t *testing.T) {
	ctx := testContext()
	g := New()

	m0 := Func(g, func(_ context.Context) (string, error) {
		return "", fmt.Errorf("this is just a test")
	})
	var gotError error
	m0.Node().OnError(func(ctx context.Context, err error) {
		testutil.BlueDye(ctx, t)
		gotError = err
	})

	_ = MustObserve(g, m0)

	err := g.Stabilize(ctx)
	testutil.NotNil(t, err)
	testutil.Equal(t, "this is just a test", err.Error())
	testutil.Equal(t, "this is just a test", gotError.Error())
}

func Test_Stabilize_alreadyStabilizing(t *testing.T) {
	ctx := testContext()

	// deadlocks. deadlocks everywhere.
	hold := make(chan struct{})
	errs := make(chan error)

	g := New()
	m0 := Func(g, func(_ context.Context) (string, error) {
		<-hold
		return "ok!", nil
	})

	_ = MustObserve(g, m0)

	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		if err := g.Stabilize(ctx); err != nil {
			errs <- err
		}
	}()
	go func() {
		defer wg.Done()
		if err := g.Stabilize(ctx); err != nil {
			errs <- err
		}
	}()
	err := <-errs
	testutil.Equal(t, ErrAlreadyStabilizing, err)
	close(hold)
	wg.Wait()
	testutil.Equal(t, "ok!", m0.Value())
}

func Test_Stabilize_updateHandlers(t *testing.T) {
	ctx := testContext()
	g := New()

	v0 := Var(g, "foo")
	v1 := Var(g, "bar")
	m0 := Map2(g, v0, v1, func(a, b string) string {
		return a + " " + b
	})

	var updates int
	m0.Node().OnUpdate(func(_ context.Context) {
		updates++
	})

	_ = MustObserve(g, m0)

	err := g.Stabilize(ctx)
	testutil.Nil(t, err)
	testutil.Equal(t, 1, updates)

	v0.Set("not foo")
	err = g.Stabilize(ctx)
	testutil.Nil(t, err)
	testutil.Equal(t, 2, updates)
}

func Test_Stabilize_unevenHeights(t *testing.T) {
	ctx := testContext()
	g := New()

	v0 := Var(g, "foo")
	v1 := Var(g, "bar")
	m0 := Map2(g, v0, v1, func(a, b string) string {
		return a + " " + b
	})
	r0 := Return(g, "moo")
	m1 := Map2(g, r0, m0, func(a, b string) string {
		return a + " != " + b
	})

	_ = MustObserve(g, m1)

	err := g.Stabilize(ctx)
	testutil.Nil(t, err)
	testutil.Equal(t, "moo != foo bar", m1.Value())

	v0.Set("not foo")
	err = g.Stabilize(ctx)
	testutil.Nil(t, err)
	testutil.Equal(t, "moo != not foo bar", m1.Value())
}

func Test_Stabilize_chain(t *testing.T) {
	ctx := testContext()
	g := New()

	v0 := Var(g, ".")

	var maps []Incr[string]
	var previous Incr[string] = v0
	for x := 0; x < 100; x++ {
		m := Map(g, previous, func(v0 string) string {
			return v0 + "."
		})
		maps = append(maps, m)
		previous = m
	}

	o := MustObserve(g, maps[len(maps)-1])

	err := g.Stabilize(ctx)
	testutil.Nil(t, err)
	testutil.Equal(t, strings.Repeat(".", 101), o.Value())

	testutil.Equal(t, 102, g.numNodes, "should include the observer!")
	testutil.Equal(t, 100, g.numNodesChanged, "should _not_ include the observer!")
	testutil.Equal(t, 100, g.numNodesRecomputed, "should _not_ include the observer!")
}

func Test_Stabilize_setDuringStabilization(t *testing.T) {
	ctx := testContext()
	g := New()
	v0 := Var(g, "foo")

	called := make(chan struct{})
	wait := make(chan struct{})
	m0 := Map(g, v0, func(v string) string {
		close(called)
		<-wait
		return v
	})

	_ = MustObserve(g, m0)

	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = g.Stabilize(ctx)
	}()

	<-called

	// we're now stabilizing
	v0.Set("not-foo")
	testutil.Equal(t, "foo", v0.Value())

	close(wait)
	<-done

	// we're now _done_ stabilizing
	testutil.Equal(t, "not-foo", v0.Value())
	testutil.Equal(t, g.stabilizationNum, v0.Node().setAt)
	testutil.Equal(t, 1, g.recomputeHeap.numItems)
}

func Test_Stabilize_onUpdate(t *testing.T) {
	ctx := testContext()
	g := New()

	var didCallUpdateHandler0, didCallUpdateHandler1 bool
	v0 := Var(g, "hello")
	v1 := Var(g, "world")
	m0 := Map2(g, v0, v1, concat)
	m0.Node().OnUpdate(func(context.Context) {
		didCallUpdateHandler0 = true
	})
	m0.Node().OnUpdate(func(context.Context) {
		didCallUpdateHandler1 = true
	})

	_ = MustObserve(g, m0)

	err := g.Stabilize(ctx)
	testutil.Nil(t, err)
	testutil.Equal(t, "helloworld", m0.Value())
	testutil.Equal(t, true, didCallUpdateHandler0)
	testutil.Equal(t, true, didCallUpdateHandler1)
}

func Test_Stabilize_recombinant_singleUpdate(t *testing.T) {
	ctx := testContext()
	g := New()

	// a -> b -> c -> d -> z
	//   -> f -> e -> [z]
	// assert that [z] updates (1) time if we change [a]

	edge := func(l string) func(string) string {
		return func(v string) string {
			return v + "->" + l
		}
	}

	a := Var(g, "a")
	b := Map(g, a, edge("b"))
	c := Map(g, b, edge("c"))
	d := Map(g, c, edge("d"))
	f := Map(g, a, edge("f"))
	e := Map(g, f, edge("e"))

	z := Map2(g, d, e, func(v0, v1 string) string {
		return v0 + "+" + v1 + "->z"
	})

	_ = MustObserve(g, z)

	err := g.Stabilize(ctx)
	testutil.Nil(t, err)
	testutil.Equal(t, 1, z.Node().numRecomputes)
	testutil.Equal(t, "a->b->c->d+a->f->e->z", z.Value())

	a.Set("!a")

	err = g.Stabilize(ctx)
	testutil.Nil(t, err)
	testutil.Equal(t, "!a->b->c->d+!a->f->e->z", z.Value())
	testutil.Equal(t, 2, z.Node().numRecomputes)
}

func Test_Stabilize_doubleVarSet_singleUpdate(t *testing.T) {
	ctx := testContext()
	g := New()

	a := Var(g, "a")
	b := Var(g, "b")
	m := Map2(g, a, b, func(v0, v1 string) string {
		return v0 + " " + v1
	})

	_ = MustObserve(g, m)

	_ = g.Stabilize(ctx)
	testutil.Equal(t, "a b", m.Value())

	a.Set("aa")
	testutil.Equal(t, 1, g.recomputeHeap.len())

	a.Set("aaa")
	testutil.Equal(t, 1, g.recomputeHeap.len())

	_ = g.Stabilize(ctx)
	testutil.Equal(t, "aaa b", m.Value())
}

func Test_Stabilize_verifyPartial(t *testing.T) {
	ctx := testContext()
	g := New()

	v0 := Var(g, "foo")
	c0 := Return(g, "bar")
	v1 := Var(g, "moo")
	c1 := Return(g, "baz")

	m0 := Map2(g, v0, c0, func(a, b string) string {
		return a + " " + b
	})
	co0 := Cutoff(g, m0, func(n, o string) bool {
		return len(n) == len(o)
	})
	m1 := Map2(g, v1, c1, func(a, b string) string {
		return a + " != " + b
	})
	co1 := Cutoff(g, m1, func(n, o string) bool {
		return len(n) == len(o)
	})

	sw := Var(g, true)
	mi := MapIf(g, co0, co1, sw)

	_ = MustObserve(g, mi)

	err := g.Stabilize(ctx)
	testutil.Nil(t, err)
	testutil.Equal(t, "foo bar", mi.Value())

	v0.Set("Foo")

	err = g.Stabilize(ctx)
	testutil.Nil(t, err)
	testutil.Equal(t, "foo bar", mi.Value())
}

func Test_Stabilize_jsDocs(t *testing.T) {
	ctx := testContext()
	g := New()

	type Entry struct {
		Entry string
		Time  time.Time
	}

	now := time.Date(2022, 05, 04, 12, 11, 10, 9, time.UTC)

	data := []Entry{
		{"0", now},
		{"1", now.Add(time.Second)},
		{"2", now.Add(2 * time.Second)},
		{"3", now.Add(3 * time.Second)},
		{"4", now.Add(4 * time.Second)},
	}

	i := Var(g, data)
	output := Map(
		g,
		i,
		func(entries []Entry) (output []string) {
			for _, e := range entries {
				if e.Time.Sub(now) > 2*time.Second {
					output = append(output, e.Entry)
				}
			}
			return
		},
	)

	_ = MustObserve(g, output)

	err := g.Stabilize(
		ctx,
	)
	testutil.Nil(t, err)
	testutil.Equal(t, 2, len(output.Value()))

	data = append(data, Entry{
		"5", now.Add(5 * time.Second),
	})
	err = g.Stabilize(
		ctx,
	)
	testutil.Nil(t, err)
	testutil.Equal(t, 2, len(output.Value()))

	i.Set(data)
	err = g.Stabilize(
		context.Background(),
	)
	testutil.Nil(t, err)
	testutil.Equal(t, 3, len(output.Value()))
}

func Test_Stabilize_Bind(t *testing.T) {
	ctx := testContext()
	g := New()

	sw := Var(g, false)
	i0 := Return(g, "foo")
	i0.Node().SetLabel("i0")
	m0 := Map(g, i0, func(v0 string) string { return v0 + "-moo" })
	m0.Node().SetLabel("m0")
	i1 := Return(g, "bar")
	i1.Node().SetLabel("i1")
	m1 := Map(g, i1, func(v0 string) string { return v0 + "-loo" })
	m1.Node().SetLabel("m1")
	b := Bind(g, sw, func(_ Scope, swv bool) Incr[string] {
		if swv {
			return m0
		}
		return m1
	})
	mb := Map(g, b, func(v string) string {
		return v + "-baz"
	})
	mb.Node().SetLabel("mb")

	_ = MustObserve(g, mb)

	testutil.Equal(t, true, g.Has(sw))

	err := g.Stabilize(ctx)
	testutil.Nil(t, err)

	testutil.Equal(t, false, g.Has(i0))
	testutil.Equal(t, false, g.Has(m0))

	testutil.Equal(t, true, g.Has(i1))
	testutil.Equal(t, true, g.Has(m1))

	testutil.Equal(t, true, i1.Node().isNecessary())
	testutil.Equal(t, true, m1.Node().isNecessary())

	testutil.Equal(t, "bar-loo-baz", mb.Value())

	sw.Set(true)
	testutil.Equal(t, true, g.recomputeHeap.has(sw))

	err = g.Stabilize(ctx)
	testutil.Nil(t, err)

	testutil.Equal(t, true, g.Has(i0))
	testutil.Equal(t, true, g.Has(m0))

	testutil.Equal(t, true, i0.Node().isNecessary())
	testutil.Equal(t, true, m0.Node().isNecessary())

	testutil.Equal(t, false, g.Has(i1))
	testutil.Equal(t, false, g.Has(m1))

	testutil.Equal(t, "foo-moo-baz", mb.Value())
}

func Test_Stabilize_BindIf(t *testing.T) {
	ctx := testContext()
	g := New()

	sw := Var(g, false)
	i0 := Return(g, "foo")
	i1 := Return(g, "bar")

	b := BindIf(g, sw, func(ctx context.Context, bs Scope, swv bool) (Incr[string], error) {
		testutil.BlueDye(ctx, t)
		if swv {
			return i0, nil
		}
		return i1, nil
	})

	_ = MustObserve(g, b)

	err := g.Stabilize(ctx)
	testutil.Nil(t, err)

	// Nil(t, i0.Node().graph, "i0 should not be in the graph after the first stabilization")
	// NotNil(t, i1.Node().graph, "i1 should be in the graph after the first stabilization")

	testutil.Equal(t, "bar", b.Value())

	sw.Set(true)
	err = g.Stabilize(ctx)
	testutil.Nil(t, err)

	// NotNil(t, i0.Node().graph, "i1 should not be in the graph after the third stabilization")

	testutil.Equal(t, "foo", b.Value())
}

func Test_Stabilize_Bind2(t *testing.T) {
	ctx := testContext()
	g := New()

	v0 := Var(g, "a")
	v1 := Var(g, "b")

	b2 := Bind2(g, v0, v1, func(bs Scope, a, b string) Incr[string] {
		return Return(bs, a+b)
	})

	testutil.Equal(t, "bind2", b2.Node().Kind())

	o := MustObserve(g, b2)
	err := g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, "ab", o.Value())

	v0.Set("xa")

	err = g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, "xab", o.Value())

	v1.Set("xb")

	err = g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, "xaxb", o.Value())
}

func Test_Stabilize_Bind3(t *testing.T) {
	ctx := testContext()
	g := New()

	v0 := Var(g, "a")
	v1 := Var(g, "b")
	v2 := Var(g, "c")

	b3 := Bind3(g, v0, v1, v2, func(bs Scope, a, b, c string) Incr[string] {
		return Return(bs, a+b+c)
	})
	testutil.Equal(t, "bind3", b3.Node().Kind())

	o := MustObserve(g, b3)
	err := g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, "abc", o.Value())

	v0.Set("xa")

	err = g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, "xabc", o.Value())

	v1.Set("xb")

	err = g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, "xaxbc", o.Value())

	v2.Set("xc")

	err = g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, "xaxbxc", o.Value())
}

func Test_Stabilize_Bind4(t *testing.T) {
	ctx := testContext()
	g := New()

	v0 := Var(g, "a")
	v1 := Var(g, "b")
	v2 := Var(g, "c")
	v3 := Var(g, "d")

	b4 := Bind4(g, v0, v1, v2, v3, func(bs Scope, a, b, c, d string) Incr[string] {
		return Return(bs, a+b+c+d)
	})
	testutil.Equal(t, "bind4", b4.Node().Kind())

	o := MustObserve(g, b4)
	err := g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, "abcd", o.Value())

	v0.Set("xa")

	err = g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, "xabcd", o.Value())

	v1.Set("xb")

	err = g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, "xaxbcd", o.Value())

	v2.Set("xc")

	err = g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, "xaxbxcd", o.Value())

	v3.Set("xd")

	err = g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, "xaxbxcxd", o.Value())
}

func Test_Stabilize_Cutoff(t *testing.T) {
	ctx := testContext()
	g := New()

	input := Var(g, 3.14)
	cutoff := Cutoff(
		g,
		input,
		epsilon(0.1),
	)
	output := Map2(
		g,
		cutoff,
		Return(g, 10.0),
		add[float64],
	)

	_ = MustObserve(g, output)

	_ = g.Stabilize(
		ctx,
	)
	testutil.Equal(t, 13.14, output.Value())
	testutil.Equal(t, 3.14, cutoff.Value())

	input.Set(3.15)

	_ = g.Stabilize(
		ctx,
	)
	testutil.Equal(t, 3.14, cutoff.Value())
	testutil.Equal(t, 13.14, output.Value())

	input.Set(3.26) // differs by 0.11, which is > 0.1

	_ = g.Stabilize(
		ctx,
	)
	testutil.Equal(t, 3.26, cutoff.Value())
	testutil.Equal(t, 13.26, output.Value())

	_ = g.Stabilize(
		ctx,
	)
	testutil.Equal(t, 13.26, output.Value())
}

func Test_Stabilize_CutoffContext(t *testing.T) {
	ctx := testContext()
	g := New()
	input := Var(g, 3.14)

	cutoff := CutoffContext(
		g,
		input,
		epsilonContext(t, 0.1),
	)

	output := Map2(
		g,
		cutoff,
		Return(g, 10.0),
		add[float64],
	)

	_ = MustObserve(g, output)

	_ = g.Stabilize(
		ctx,
	)
	testutil.Equal(t, 13.14, output.Value())
	testutil.Equal(t, 3.14, cutoff.Value())

	input.Set(3.15)

	_ = g.Stabilize(
		ctx,
	)
	testutil.Equal(t, 3.14, cutoff.Value())
	testutil.Equal(t, 13.14, output.Value())

	input.Set(3.26) // differs by 0.11, which is > 0.1

	_ = g.Stabilize(
		ctx,
	)
	testutil.Equal(t, 3.26, cutoff.Value())
	testutil.Equal(t, 13.26, output.Value())

	_ = g.Stabilize(
		ctx,
	)
	testutil.Equal(t, 13.26, output.Value())
}

func Test_Stabilize_CutoffContext_error(t *testing.T) {
	ctx := testContext()
	g := New()
	input := Var(g, 3.14)

	cutoff := CutoffContext(
		g,
		input,
		func(_ context.Context, _, _ float64) (bool, error) {
			return false, fmt.Errorf("this is just a test")
		},
	)

	var errors int
	cutoff.Node().OnError(func(_ context.Context, err error) {
		if err != nil {
			errors++
		}
	})

	output := Map2(
		g,
		cutoff,
		Return(g, 10.0),
		add[float64],
	)

	_ = MustObserve(g, output)

	err := g.Stabilize(
		ctx,
	)
	testutil.NotNil(t, err)
	testutil.Equal(t, 1, errors)
	testutil.Equal(t, 0, output.Value())

	input.Set(3.15)

	err = g.Stabilize(
		ctx,
	)
	testutil.NotNil(t, err)
	testutil.Equal(t, 2, errors)
	testutil.Equal(t, 0, output.Value())
}

func Test_Stabilize_Cutoff2(t *testing.T) {
	ctx := testContext()
	g := New()

	epsilon := Var(g, 0.1)
	input := Var(g, 3.14)
	cutoff := Cutoff2(
		g,
		epsilon,
		input,
		epsilonFn,
	)
	output := Map2(
		g,
		cutoff,
		Return(g, 10.0),
		add[float64],
	)

	_ = MustObserve(g, output)

	_ = g.Stabilize(
		ctx,
	)
	testutil.Equal(t, 13.14, output.Value())
	testutil.Equal(t, 3.14, cutoff.Value())

	input.Set(3.15)

	_ = g.Stabilize(
		ctx,
	)
	testutil.Equal(t, 3.14, cutoff.Value())
	testutil.Equal(t, 13.14, output.Value())

	input.Set(3.26) // differs by 0.11, which is > 0.1

	_ = g.Stabilize(
		ctx,
	)
	testutil.Equal(t, 3.26, cutoff.Value())
	testutil.Equal(t, 13.26, output.Value())

	_ = g.Stabilize(
		ctx,
	)
	testutil.Equal(t, 13.26, output.Value())

	epsilon.Set(0.5)
	input.Set(3.37) // differs by 0.11, which is < 0.5

	_ = g.Stabilize(
		ctx,
	)
	testutil.Equal(t, 3.26, cutoff.Value())
	testutil.Equal(t, 13.26, output.Value())

	_ = g.Stabilize(
		ctx,
	)
	testutil.Equal(t, 13.26, output.Value())
}

func Test_Stabilize_Cutoff2Context_error(t *testing.T) {
	ctx := testContext()
	g := New()
	epsilon := Var(g, 0.1)
	input := Var(g, 3.14)

	cutoff := Cutoff2Context(
		g,
		epsilon,
		input,
		func(_ context.Context, _, _, _ float64) (bool, error) {
			return false, fmt.Errorf("this is just a test")
		},
	)

	var errors int
	cutoff.Node().OnError(func(_ context.Context, err error) {
		if err != nil {
			errors++
		}
	})

	output := Map2(
		g,
		cutoff,
		Return(g, 10.0),
		add[float64],
	)

	_ = MustObserve(g, output)

	err := g.Stabilize(
		ctx,
	)
	testutil.NotNil(t, err)
	testutil.Equal(t, 1, errors)
	testutil.Equal(t, 0, output.Value())

	input.Set(3.15)

	err = g.Stabilize(
		ctx,
	)
	testutil.NotNil(t, err)
	testutil.Equal(t, 2, errors)
	testutil.Equal(t, 0, output.Value())
}

func Test_Stabilize_Watch(t *testing.T) {
	ctx := testContext()
	g := New()

	v0 := Var(g, 1)
	v1 := Var(g, 1)
	m0 := Map2(g, v0, v1, add)
	w0 := Watch(g, m0)

	_ = MustObserve(g, w0)

	_ = g.Stabilize(ctx)

	testutil.Equal(t, 1, len(w0.Values()))
	testutil.Equal(t, 2, w0.Values()[0])
	testutil.Equal(t, 2, w0.Value())

	v0.Set(2)

	_ = g.Stabilize(ctx)

	testutil.Equal(t, 2, len(w0.Values()))
	testutil.Equal(t, 2, w0.Values()[0])
	testutil.Equal(t, 3, w0.Values()[1])
}

func Test_Stabilize_Map(t *testing.T) {
	ctx := testContext()
	g := New()

	c0 := Return(g, 1)
	m := Map(g, c0, func(a int) int {
		return a + 10
	})

	_ = MustObserve(g, m)

	_ = g.Stabilize(ctx)
	testutil.Equal(t, 11, m.Value())
}

func Test_Stabilize_MapContext(t *testing.T) {
	ctx := testContext()
	g := New()

	c0 := Return(g, 1)
	m := MapContext(g, c0, func(ictx context.Context, a int) (int, error) {
		testutil.BlueDye(ictx, t)
		return a + 10, nil
	})

	_ = MustObserve(g, m)

	_ = g.Stabilize(ctx)
	testutil.Equal(t, 11, m.Value())
}

func Test_Stabilize_Map2(t *testing.T) {
	ctx := testContext()
	g := New()

	c0 := Return(g, 1)
	c1 := Return(g, 2)
	m2 := Map2(g, c0, c1, func(a, b int) int {
		return a + b
	})

	_ = MustObserve(g, m2)

	_ = g.Stabilize(ctx)
	testutil.Equal(t, 3, m2.Value())
}

func Test_Stabilize_Map2Context(t *testing.T) {
	ctx := testContext()
	g := New()

	c0 := Return(g, 1)
	c1 := Return(g, 2)
	m2 := Map2Context(g, c0, c1, func(ictx context.Context, a, b int) (int, error) {
		testutil.BlueDye(ctx, t)
		return a + b, nil
	})

	_ = MustObserve(g, m2)

	_ = g.Stabilize(ctx)
	testutil.Equal(t, 3, m2.Value())
}

func Test_Stabilize_Map2Context_error(t *testing.T) {
	ctx := testContext()
	g := New()

	c0 := Return(g, 1)
	c1 := Return(g, 2)
	m2 := Map2Context(g, c0, c1, func(ictx context.Context, a, b int) (int, error) {
		testutil.BlueDye(ctx, t)
		return a + b, fmt.Errorf("this is just a test")
	})

	_ = MustObserve(g, m2)

	err := g.Stabilize(ctx)
	testutil.NotNil(t, err)
	testutil.Equal(t, 0, m2.Value())
}

func Test_Stabilize_Map3(t *testing.T) {
	ctx := testContext()
	g := New()

	c0 := Return(g, 1)
	c1 := Return(g, 2)
	c2 := Return(g, 3)
	m3 := Map3(g, c0, c1, c2, func(a, b, c int) int {
		return a + b + c
	})

	_ = MustObserve(g, m3)

	_ = g.Stabilize(ctx)
	testutil.Equal(t, 6, m3.Value())
}

func Test_Stabilize_Map3Context(t *testing.T) {
	ctx := testContext()
	g := New()

	c0 := Return(g, 1)
	c1 := Return(g, 2)
	c2 := Return(g, 3)
	m3 := Map3Context(g, c0, c1, c2, func(ictx context.Context, a, b, c int) (int, error) {
		testutil.BlueDye(ictx, t)
		return a + b + c, nil
	})

	_ = MustObserve(g, m3)

	_ = g.Stabilize(ctx)
	testutil.Equal(t, 6, m3.Value())
}

func Test_Stabilize_Map3Context_error(t *testing.T) {
	ctx := testContext()
	g := New()

	c0 := Return(g, 1)
	c1 := Return(g, 2)
	c2 := Return(g, 3)
	m3 := Map3Context(g, c0, c1, c2, func(ictx context.Context, a, b, c int) (int, error) {
		testutil.BlueDye(ictx, t)
		return a + b + c, fmt.Errorf("this is just a test")
	})

	_ = MustObserve(g, m3)

	err := g.Stabilize(ctx)
	testutil.NotNil(t, err)
	testutil.Equal(t, 0, m3.Value())
}

func Test_Stabilize_Map4(t *testing.T) {
	ctx := testContext()
	g := New()

	c0 := Return(g, 1)
	c1 := Return(g, 2)
	c2 := Return(g, 3)
	c3 := Return(g, 4)
	m3 := Map4(g, c0, c1, c2, c3, func(a, b, c, d int) int {
		return a + b + c + d
	})

	_ = MustObserve(g, m3)

	_ = g.Stabilize(ctx)
	testutil.Equal(t, 10, m3.Value())
}

func Test_Stabilize_Map5(t *testing.T) {
	ctx := testContext()
	g := New()

	c0 := Return(g, 1)
	c1 := Return(g, 2)
	c2 := Return(g, 3)
	c3 := Return(g, 4)
	c4 := Return(g, 5)
	m3 := Map5(g, c0, c1, c2, c3, c4, func(a, b, c, d, e int) int {
		return a + b + c + d + e
	})

	_ = MustObserve(g, m3)

	_ = g.Stabilize(ctx)
	testutil.Equal(t, 15, m3.Value())
}

func Test_Stabilize_Map6(t *testing.T) {
	ctx := testContext()
	g := New()

	c0 := Return(g, 1)
	c1 := Return(g, 2)
	c2 := Return(g, 3)
	c3 := Return(g, 4)
	c4 := Return(g, 5)
	c5 := Return(g, 6)
	m3 := Map6(g, c0, c1, c2, c3, c4, c5, func(a, b, c, d, e, f int) int {
		return a + b + c + d + e + f
	})

	_ = MustObserve(g, m3)

	_ = g.Stabilize(ctx)
	testutil.Equal(t, 21, m3.Value())
}

func Test_Stabilize_Map7(t *testing.T) {
	ctx := testContext()
	g := New()

	c0 := Return(g, 1)
	c1 := Return(g, 2)
	c2 := Return(g, 3)
	c3 := Return(g, 4)
	c4 := Return(g, 5)
	c5 := Return(g, 6)
	c6 := Return(g, 7)
	m3 := Map7(g, c0, c1, c2, c3, c4, c5, c6, func(a, b, c, d, e, f, g int) int {
		return a + b + c + d + e + f + g
	})

	_ = MustObserve(g, m3)

	_ = g.Stabilize(ctx)
	testutil.Equal(t, 28, m3.Value())
}

func Test_Stabilize_Map8(t *testing.T) {
	ctx := testContext()
	g := New()

	c0 := Return(g, 1)
	c1 := Return(g, 2)
	c2 := Return(g, 3)
	c3 := Return(g, 4)
	c4 := Return(g, 5)
	c5 := Return(g, 6)
	c6 := Return(g, 7)
	c7 := Return(g, 8)
	m3 := Map8(g, c0, c1, c2, c3, c4, c5, c6, c7, func(a, b, c, d, e, f, g, h int) int {
		return a + b + c + d + e + f + g + h
	})

	_ = MustObserve(g, m3)

	_ = g.Stabilize(ctx)
	testutil.Equal(t, 36, m3.Value())
}

func Test_Stabilize_MapIf(t *testing.T) {
	ctx := testContext()
	g := New()

	c0 := Return(g, 1)
	c1 := Return(g, 2)
	v0 := Var(g, false)
	mi0 := MapIf(g, c0, c1, v0)

	_ = MustObserve(g, mi0)

	_ = g.Stabilize(ctx)
	testutil.Equal(t, 2, mi0.Value())

	v0.Set(true)

	_ = g.Stabilize(ctx)
	testutil.Equal(t, 1, mi0.Value())

	_ = g.Stabilize(ctx)
	testutil.Equal(t, 1, mi0.Value())
}

func Test_Stabilize_MapN(t *testing.T) {
	ctx := testContext()
	g := New()

	sum := func(values ...int) (output int) {
		if len(values) == 0 {
			return
		}
		output = values[0]
		for _, value := range values[1:] {
			output += value
		}
		return
	}

	c0 := Return(g, 1)
	c1 := Return(g, 2)
	c2 := Return(g, 3)
	mn := MapN(g, sum, c0, c1, c2)

	_ = MustObserve(g, mn)

	_ = g.Stabilize(ctx)
	testutil.Equal(t, 6, mn.Value())
}

func Test_Stabilize_MapN_AddInput(t *testing.T) {
	ctx := testContext()
	g := New()

	sum := func(values ...int) (output int) {
		if len(values) == 0 {
			return
		}
		output = values[0]
		for _, value := range values[1:] {
			output += value
		}
		return
	}

	c0 := Return(g, 1)
	c1 := Return(g, 2)
	c2 := Return(g, 3)
	mn := MapN(g, sum)
	var err error
	err = mn.AddInput(c0)
	testutil.NoError(t, err)
	err = mn.AddInput(c1)
	testutil.NoError(t, err)
	err = mn.AddInput(c2)
	testutil.NoError(t, err)

	_ = MustObserve(g, mn)

	_ = g.Stabilize(ctx)
	testutil.Equal(t, 6, mn.Value())
}

func Test_Stabilize_MapNContext(t *testing.T) {
	ctx := testContext()
	g := New()

	sum := func(ctx context.Context, values ...int) (output int, err error) {
		testutil.BlueDye(ctx, t)
		if len(values) == 0 {
			return
		}
		output = values[0]
		for _, value := range values[1:] {
			output += value
		}
		return
	}

	c0 := Return(g, 1)
	c1 := Return(g, 2)
	c2 := Return(g, 3)
	mn := MapNContext(g, sum, c0, c1, c2)

	_ = MustObserve(g, mn)

	_ = g.Stabilize(ctx)
	testutil.Equal(t, 6, mn.Value())
}

func Test_Stabilize_MapNContext_error(t *testing.T) {
	ctx := testContext()
	g := New()

	sum := func(ctx context.Context, values ...int) (output int, err error) {
		testutil.BlueDye(ctx, t)
		for _, value := range values {
			output += value
		}
		err = fmt.Errorf("this is just a test")
		return
	}

	c0 := Return(g, 1)
	c1 := Return(g, 2)
	c2 := Return(g, 3)
	mn := MapNContext(g, sum, c0, c1, c2)

	_ = MustObserve(g, mn)

	err := g.Stabilize(ctx)
	testutil.NotNil(t, err)
	testutil.Equal(t, 0, mn.Value())
}

func Test_Stabilize_Func(t *testing.T) {
	ctx := testContext()
	g := New()

	value := "hello"
	f := Func(g, func(ictx context.Context) (string, error) {
		testutil.BlueDye(ictx, t)
		return value, nil
	})
	m := MapContext(g, f, func(ictx context.Context, v string) (string, error) {
		testutil.BlueDye(ctx, t)
		return v + " world!", nil
	})

	_ = MustObserve(g, m)

	_ = g.Stabilize(ctx)
	testutil.Equal(t, "hello world!", m.Value())

	value = "not hello"

	_ = g.Stabilize(ctx)
	testutil.Equal(t, "hello world!", m.Value())

	// mark the func node as stale
	// not sure a better way to do this automatically?
	g.SetStale(f)

	_ = g.Stabilize(ctx)
	testutil.Equal(t, "not hello world!", m.Value())
}

func Test_Stabilize_Freeze(t *testing.T) {
	ctx := testContext()
	g := New()

	v0 := Var(g, "hello")
	fv := Freeze(g, v0)

	_ = MustObserve(g, fv)

	err := g.Stabilize(ctx)
	testutil.Nil(t, err)
	testutil.Equal(t, "hello", v0.Value())
	testutil.Equal(t, "hello", fv.Value())

	v0.Set("not-hello")

	err = g.Stabilize(ctx)
	testutil.Nil(t, err)
	testutil.Equal(t, "not-hello", v0.Value())
	testutil.Equal(t, "hello", fv.Value())
}

func Test_Stabilize_Always_Cutoff(t *testing.T) {
	ctx := testContext()
	g := New()

	filename := Var(g, "test")
	filenameAlways := Always(g, filename)
	modtime := 1
	statfile := Map(g, filenameAlways, func(s string) int { return modtime })
	statfileCutoff := Cutoff(g, statfile, func(ov, nv int) bool {
		return ov == nv
	})
	readFile := Map2(g, filename, statfileCutoff, func(p string, mt int) string {
		return fmt.Sprintf("%s-%d", p, mt)
	})
	o := MustObserve(g, readFile)

	err := g.Stabilize(ctx)
	testutil.Nil(t, err)
	testutil.Equal(t, "test-1", o.Value())

	err = g.Stabilize(ctx)
	testutil.Nil(t, err)
	testutil.Equal(t, "test-1", o.Value())

	modtime = 2

	err = g.Stabilize(ctx)
	testutil.Nil(t, err)
	testutil.Equal(t, "test-2", o.Value())

	err = g.Stabilize(ctx)
	testutil.Nil(t, err)
	testutil.Equal(t, "test-2", o.Value())
}

func Test_Stabilize_Always_Cutoff_error(t *testing.T) {
	ctx := testContext()
	g := New()

	filename := Var(g, "test")
	filenameAlways := Always(g, filename)
	modtime := 1
	statfile := Map(g, filenameAlways, func(s string) int { return modtime })
	statfileCutoff := CutoffContext(g, statfile, func(_ context.Context, ov, nv int) (bool, error) {
		return false, fmt.Errorf("this is only a test")
	})
	readFile := Map2(g, filename, statfileCutoff, func(p string, mt int) string {
		return fmt.Sprintf("%s-%d", p, mt)
	})
	o := MustObserve(g, readFile)

	err := g.Stabilize(ctx)
	testutil.NotNil(t, err)
	testutil.Equal(t, "", o.Value())

	testutil.Equal(t, 1, g.recomputeHeap.len(), "we should clear the recompute heap on error")
}

func Test_Stabilize_printsErrors(t *testing.T) {
	g := New()

	outBuf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)

	ctx := WithTracingOutputs(context.Background(), outBuf, errBuf)

	v0 := Var(g, "hello")
	gonnaPanic := MapContext(g, v0, func(_ context.Context, _ string) (string, error) {
		return "", fmt.Errorf("this is only a test")
	})
	_ = MustObserve(g, gonnaPanic)

	err := g.Stabilize(ctx)
	testutil.NotNil(t, err)
	testutil.NotEqual(t, 0, len(outBuf.String()))
	testutil.NotEqual(t, 0, len(errBuf.String()))
	testutil.Equal(t, true, strings.Contains(errBuf.String(), "this is only a test"))
}

func Test_Stabilize_handlers(t *testing.T) {
	ctx := testContext()
	g := New()

	v0 := Var(g, "foo")
	v1 := Var(g, "bar")
	m0 := Map2(g, v0, v1, func(a, b string) string {
		return a + " " + b
	})

	var didCallStabilizationStart bool
	var didCallStabilizationEnd bool
	var startWasBlueDye bool
	var endWasBlueDye bool

	_ = MustObserve(g, m0)
	g.OnStabilizationStart(func(ictx context.Context) {
		startWasBlueDye = testutil.HasBlueDye(ctx)
		didCallStabilizationStart = true
	})
	g.OnStabilizationEnd(func(ictx context.Context, started time.Time, err error) {
		endWasBlueDye = testutil.HasBlueDye(ctx)
		didCallStabilizationEnd = true
	})
	err := g.Stabilize(ctx)
	testutil.Nil(t, err)
	testutil.Equal(t, true, didCallStabilizationStart)
	testutil.Equal(t, true, didCallStabilizationEnd)
	testutil.Equal(t, true, startWasBlueDye)
	testutil.Equal(t, true, endWasBlueDye)
}

func Test_Stabilize_Bind_jsCombination(t *testing.T) {
	ctx := testContext()
	g := New()

	v1 := Var(g, 1)
	v2 := Var(g, 2)
	v3 := Var(g, 3)
	v4 := Var(g, 4)

	o := MustObserve(g, Bind4(g, v1, v2, v3, v4, func(bs Scope, x1, x2, x3, x4 int) Incr[int] {
		return Bind3(bs, v2, v3, v3, func(bs Scope, y2, y3, y4 int) Incr[int] {
			return Bind2(bs, v4, v4, func(bs Scope, z3, z4 int) Incr[int] {
				return Bind(bs, v4, func(bs Scope, w4 int) Incr[int] {
					return Return(bs, x1+x2+x3+x4+y2+y3+y4+z3+z4+w4)
				})
			})
		})
	}))

	err := g.Stabilize(ctx)
	testutil.NoError(t, err)

	testutil.Equal(t, v1.Value()+(2*v2.Value())+(3*v3.Value())+(4*v4.Value()), o.Value())

	v1.Set(9)
	v2.Set(10)
	v3.Set(11)
	v4.Set(12)

	err = g.Stabilize(ctx)
	testutil.NoError(t, err)

	testutil.Equal(t, v1.Value()+(2*v2.Value())+(3*v3.Value())+(4*v4.Value()), o.Value())
}

func Test_Stabilize_alwaysInRecomputeHeapOnError(t *testing.T) {
	g := New()

	v0 := Var(g, "foo")
	coa := cutoffAlways(g, v0,
		func(_ context.Context, _ string) (bool, error) {
			return false, fmt.Errorf("this is only a test")
		},
		func(_ context.Context, i string) (string, error) {
			return i + "-bar", nil
		},
	)
	_, _ = Observe(g, coa)

	err := g.Stabilize(testContext())
	testutil.Error(t, err)
	testutil.Equal(t, "this is only a test", err.Error())
}
