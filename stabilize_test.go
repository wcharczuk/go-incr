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
	. "github.com/wcharczuk/go-incr/testutil"
)

func Test_Stabilize(t *testing.T) {
	ctx := testContext()
	g := New()

	v0 := Var(g, "foo")
	v1 := Var(g, "bar")
	m0 := Map2(g, v0, v1, func(a, b string) string {
		return a + " " + b
	})

	_ = Observe(g, m0)

	err := g.Stabilize(ctx)
	Nil(t, err)

	Equal(t, 0, v0.Node().setAt)
	Equal(t, 1, v0.Node().changedAt)
	Equal(t, 0, v1.Node().setAt)
	Equal(t, 1, v1.Node().changedAt)
	Equal(t, 1, m0.Node().changedAt)
	Equal(t, 1, v0.Node().recomputedAt)
	Equal(t, 1, v1.Node().recomputedAt)
	Equal(t, 1, m0.Node().recomputedAt)

	Equal(t, "foo bar", m0.Value())

	v0.Set("not foo")
	Equal(t, 2, v0.Node().setAt)
	Equal(t, 0, v1.Node().setAt)

	err = g.Stabilize(ctx)
	Nil(t, err)

	Equal(t, 2, v0.Node().changedAt)
	Equal(t, 1, v1.Node().changedAt)
	Equal(t, 2, m0.Node().changedAt)

	Equal(t, 2, v0.Node().recomputedAt)
	Equal(t, 1, v1.Node().recomputedAt)
	Equal(t, 2, m0.Node().recomputedAt)

	Equal(t, "not foo bar", m0.Value())
}

func Test_Stabilize_error(t *testing.T) {
	ctx := testContext()
	g := New()

	m0 := Func(g, func(_ context.Context) (string, error) {
		return "", fmt.Errorf("this is just a test")
	})

	_ = Observe(g, m0)

	err := g.Stabilize(ctx)
	NotNil(t, err)
	Equal(t, "this is just a test", err.Error())
}

func Test_Stabilize_errorHandler(t *testing.T) {
	ctx := testContext()
	g := New()

	m0 := Func(g, func(_ context.Context) (string, error) {
		return "", fmt.Errorf("this is just a test")
	})
	var gotError error
	m0.Node().OnError(func(ctx context.Context, err error) {
		BlueDye(ctx, t)
		gotError = err
	})

	_ = Observe(g, m0)

	err := g.Stabilize(ctx)
	NotNil(t, err)
	Equal(t, "this is just a test", err.Error())
	Equal(t, "this is just a test", gotError.Error())
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

	_ = Observe(g, m0)

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
	Equal(t, ErrAlreadyStabilizing, err)
	close(hold)
	wg.Wait()
	Equal(t, "ok!", m0.Value())
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

	_ = Observe(g, m0)

	err := g.Stabilize(ctx)
	Nil(t, err)
	Equal(t, 1, updates)

	v0.Set("not foo")
	err = g.Stabilize(ctx)
	Nil(t, err)
	Equal(t, 2, updates)
}

func Test_Stabilize_observedHandlers(t *testing.T) {
	ctx := testContext()
	g := New()

	v0 := Var(g, "foo")
	v1 := Var(g, "bar")
	m0 := Map2(g, v0, v1, func(a, b string) string {
		return a + " " + b
	})

	var observes int
	m0.Node().OnObserved(func(on IObserver) {
		testutil.NotNil(t, on)
		observes++
	})

	_ = Observe(g, m0)

	err := g.Stabilize(ctx)
	Nil(t, err)
	Equal(t, 1, observes)

	v0.Set("not foo")
	err = g.Stabilize(ctx)
	Nil(t, err)
	Equal(t, 1, observes)

	_ = Observe(g, m0)
	Equal(t, 2, observes)
}

func Test_Stabilize_unobservedHandlers(t *testing.T) {
	ctx := testContext()
	g := New()

	v0 := Var(g, "foo")
	v1 := Var(g, "bar")
	m0 := Map2(g, v0, v1, func(a, b string) string {
		return a + " " + b
	})

	var observes, unobserves int
	m0.Node().OnObserved(func(IObserver) {
		observes++
	})
	m0.Node().OnUnobserved(func(IObserver) {
		unobserves++
	})

	o0 := Observe(g, m0)

	err := g.Stabilize(ctx)
	Nil(t, err)
	Equal(t, 1, observes)
	Equal(t, 0, unobserves)

	v0.Set("not foo")
	err = g.Stabilize(ctx)
	Nil(t, err)
	Equal(t, 1, observes)

	_ = Observe(g, m0)
	Equal(t, 2, observes)
	Equal(t, 0, unobserves)

	o0.Unobserve(ctx)
	Equal(t, 2, observes)
	Equal(t, 1, unobserves)
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

	_ = Observe(g, m1)

	err := g.Stabilize(ctx)
	Nil(t, err)
	Equal(t, "moo != foo bar", m1.Value())

	v0.Set("not foo")
	err = g.Stabilize(ctx)
	Nil(t, err)
	Equal(t, "moo != not foo bar", m1.Value())
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

	o := Observe(g, maps[len(maps)-1])

	err := g.Stabilize(ctx)
	Nil(t, err)
	Equal(t, strings.Repeat(".", 101), o.Value())

	Equal(t, 102, g.numNodes, "should include the observer!")
	Equal(t, 102, g.numNodesChanged)
	Equal(t, 102, g.numNodesRecomputed)
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

	_ = Observe(g, m0)

	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = g.Stabilize(ctx)
	}()

	<-called

	// we're now stabilizing
	v0.Set("not-foo")
	Equal(t, "foo", v0.Value())

	close(wait)
	<-done

	// we're now _done_ stabilizing
	Equal(t, "not-foo", v0.Value())
	Equal(t, g.stabilizationNum, v0.Node().setAt)
	Equal(t, 1, len(g.recomputeHeap.lookup))
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

	_ = Observe(g, m0)

	err := g.Stabilize(ctx)
	Nil(t, err)
	Equal(t, "helloworld", m0.Value())
	Equal(t, true, didCallUpdateHandler0)
	Equal(t, true, didCallUpdateHandler1)
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

	_ = Observe(g, z)

	err := g.Stabilize(ctx)
	Nil(t, err)
	Equal(t, 1, z.Node().numRecomputes)
	Equal(t, "a->b->c->d+a->f->e->z", z.Value())

	a.Set("!a")

	err = g.Stabilize(ctx)
	Nil(t, err)
	Equal(t, "!a->b->c->d+!a->f->e->z", z.Value())
	Equal(t, 2, z.Node().numRecomputes)
}

func Test_Stabilize_doubleVarSet_singleUpdate(t *testing.T) {
	ctx := testContext()
	g := New()

	a := Var(g, "a")
	b := Var(g, "b")
	m := Map2(g, a, b, func(v0, v1 string) string {
		return v0 + " " + v1
	})

	_ = Observe(g, m)

	_ = g.Stabilize(ctx)
	Equal(t, "a b", m.Value())

	a.Set("aa")
	Equal(t, 1, g.recomputeHeap.len())

	a.Set("aaa")
	Equal(t, 1, g.recomputeHeap.len())

	_ = g.Stabilize(ctx)
	Equal(t, "aaa b", m.Value())
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

	_ = Observe(g, mi)

	err := g.Stabilize(ctx)
	Nil(t, err)
	Equal(t, "foo bar", mi.Value())

	v0.Set("Foo")

	err = g.Stabilize(ctx)
	Nil(t, err)
	Equal(t, "foo bar", mi.Value())
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

	_ = Observe(g, output)

	err := g.Stabilize(
		ctx,
	)
	Nil(t, err)
	Equal(t, 2, len(output.Value()))

	data = append(data, Entry{
		"5", now.Add(5 * time.Second),
	})
	err = g.Stabilize(
		ctx,
	)
	Nil(t, err)
	Equal(t, 2, len(output.Value()))

	i.Set(data)
	err = g.Stabilize(
		context.Background(),
	)
	Nil(t, err)
	Equal(t, 3, len(output.Value()))
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

	_ = Observe(g, mb)

	Equal(t, true, g.IsObserving(sw))

	err := g.Stabilize(ctx)
	Nil(t, err)

	Equal(t, false, g.IsObserving(i0))
	Equal(t, false, g.IsObserving(m0))
	Nil(t, i0.Node().graph, "i0 should not be in the graph after the first stabilization")
	Nil(t, m0.Node().graph, "m0 should not be in the graph after the first stabilization")

	Equal(t, true, g.IsObserving(i1))
	Equal(t, true, g.IsObserving(m1))
	NotNil(t, i1.Node().graph, "i1 should be in the graph after the first stabilization")
	NotNil(t, m1.Node().graph, "m1 should be in the graph after the first stabilization")

	Equal(t, "bar-loo-baz", mb.Value())

	sw.Set(true)
	Equal(t, true, g.recomputeHeap.has(sw))

	err = g.Stabilize(ctx)
	Nil(t, err)

	Equal(t, true, g.IsObserving(i0))
	Equal(t, true, g.IsObserving(m0))
	NotNil(t, i0.Node().graph, "i0 should be in the graph after the second stabilization")
	NotNil(t, m0.Node().graph, "m0 should be in the graph after the second stabilization")

	Equal(t, false, g.IsObserving(i1))
	Equal(t, false, g.IsObserving(m1))
	Nil(t, i1.Node().graph, "i1 should not be in the graph after the second stabilization")
	Nil(t, m1.Node().graph, "m1 should not be in the graph after the second stabilization")

	Equal(t, "foo-moo-baz", mb.Value())
}

func Test_Stabilize_BindIf(t *testing.T) {
	ctx := testContext()
	g := New()

	sw := Var(g, false)
	i0 := Return(g, "foo")
	i1 := Return(g, "bar")

	b := BindIf(g, sw, func(ctx context.Context, bs Scope, swv bool) (Incr[string], error) {
		BlueDye(ctx, t)
		if swv {
			return i0, nil
		}
		return i1, nil
	})

	_ = Observe(g, b)

	err := g.Stabilize(ctx)
	Nil(t, err)

	Nil(t, i0.Node().graph, "i0 should not be in the graph after the first stabilization")
	NotNil(t, i1.Node().graph, "i1 should be in the graph after the first stabilization")

	Equal(t, "bar", b.Value())

	sw.Set(true)
	err = g.Stabilize(ctx)
	Nil(t, err)

	Nil(t, i1.Node().graph, "i0 should be in the graph after the third stabilization")
	NotNil(t, i0.Node().graph, "i1 should not be in the graph after the third stabilization")

	Equal(t, "foo", b.Value())
}

func Test_Stabilize_Bind2(t *testing.T) {
	ctx := testContext()
	g := New()

	v0 := Var(g, "a")
	v1 := Var(g, "b")

	b2 := Bind2(g, v0, v1, func(bs Scope, a, b string) Incr[string] {
		return Return(bs, a+b)
	})

	Equal(t, "bind2", b2.Node().Kind())

	o := Observe(g, b2)
	err := g.Stabilize(ctx)
	NoError(t, err)
	Equal(t, "ab", o.Value())

	v0.Set("xa")

	err = g.Stabilize(ctx)
	NoError(t, err)
	Equal(t, "xab", o.Value())

	v1.Set("xb")

	err = g.Stabilize(ctx)
	NoError(t, err)
	Equal(t, "xaxb", o.Value())
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
	Equal(t, "bind3", b3.Node().Kind())

	o := Observe(g, b3)
	err := g.Stabilize(ctx)
	NoError(t, err)
	Equal(t, "abc", o.Value())

	v0.Set("xa")

	err = g.Stabilize(ctx)
	NoError(t, err)
	Equal(t, "xabc", o.Value())

	v1.Set("xb")

	err = g.Stabilize(ctx)
	NoError(t, err)
	Equal(t, "xaxbc", o.Value())

	v2.Set("xc")

	err = g.Stabilize(ctx)
	NoError(t, err)
	Equal(t, "xaxbxc", o.Value())
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
	Equal(t, "bind4", b4.Node().Kind())

	o := Observe(g, b4)
	err := g.Stabilize(ctx)
	NoError(t, err)
	Equal(t, "abcd", o.Value())

	v0.Set("xa")

	err = g.Stabilize(ctx)
	NoError(t, err)
	Equal(t, "xabcd", o.Value())

	v1.Set("xb")

	err = g.Stabilize(ctx)
	NoError(t, err)
	Equal(t, "xaxbcd", o.Value())

	v2.Set("xc")

	err = g.Stabilize(ctx)
	NoError(t, err)
	Equal(t, "xaxbxcd", o.Value())

	v3.Set("xd")

	err = g.Stabilize(ctx)
	NoError(t, err)
	Equal(t, "xaxbxcxd", o.Value())
}

func Test_Stabilize_cutoff(t *testing.T) {
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

	_ = Observe(g, output)

	_ = g.Stabilize(
		ctx,
	)
	Equal(t, 13.14, output.Value())
	Equal(t, 3.14, cutoff.Value())

	input.Set(3.15)

	_ = g.Stabilize(
		ctx,
	)
	Equal(t, 3.14, cutoff.Value())
	Equal(t, 13.14, output.Value())

	input.Set(3.26) // differs by 0.11, which is > 0.1

	_ = g.Stabilize(
		ctx,
	)
	Equal(t, 3.26, cutoff.Value())
	Equal(t, 13.26, output.Value())

	_ = g.Stabilize(
		ctx,
	)
	Equal(t, 13.26, output.Value())
}

func Test_Stabilize_cutoffContext(t *testing.T) {
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

	_ = Observe(g, output)

	_ = g.Stabilize(
		ctx,
	)
	Equal(t, 13.14, output.Value())
	Equal(t, 3.14, cutoff.Value())

	input.Set(3.15)

	_ = g.Stabilize(
		ctx,
	)
	Equal(t, 3.14, cutoff.Value())
	Equal(t, 13.14, output.Value())

	input.Set(3.26) // differs by 0.11, which is > 0.1

	_ = g.Stabilize(
		ctx,
	)
	Equal(t, 3.26, cutoff.Value())
	Equal(t, 13.26, output.Value())

	_ = g.Stabilize(
		ctx,
	)
	Equal(t, 13.26, output.Value())
}

func Test_Stabilize_cutoffContext_error(t *testing.T) {
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

	_ = Observe(g, output)

	err := g.Stabilize(
		ctx,
	)
	NotNil(t, err)
	Equal(t, 1, errors)
	Equal(t, 0, output.Value())

	input.Set(3.15)

	err = g.Stabilize(
		ctx,
	)
	NotNil(t, err)
	Equal(t, 2, errors)
	Equal(t, 0, output.Value())
}

func Test_Stabilize_cutoff2(t *testing.T) {
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

	_ = Observe(g, output)

	_ = g.Stabilize(
		ctx,
	)
	Equal(t, 13.14, output.Value())
	Equal(t, 3.14, cutoff.Value())

	input.Set(3.15)

	_ = g.Stabilize(
		ctx,
	)
	Equal(t, 3.14, cutoff.Value())
	Equal(t, 13.14, output.Value())

	input.Set(3.26) // differs by 0.11, which is > 0.1

	_ = g.Stabilize(
		ctx,
	)
	Equal(t, 3.26, cutoff.Value())
	Equal(t, 13.26, output.Value())

	_ = g.Stabilize(
		ctx,
	)
	Equal(t, 13.26, output.Value())

	epsilon.Set(0.5)
	input.Set(3.37) // differs by 0.11, which is < 0.5

	_ = g.Stabilize(
		ctx,
	)
	Equal(t, 3.26, cutoff.Value())
	Equal(t, 13.26, output.Value())

	_ = g.Stabilize(
		ctx,
	)
	Equal(t, 13.26, output.Value())
}

func Test_Stabilize_cutoff2Context_error(t *testing.T) {
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

	_ = Observe(g, output)

	err := g.Stabilize(
		ctx,
	)
	NotNil(t, err)
	Equal(t, 1, errors)
	Equal(t, 0, output.Value())

	input.Set(3.15)

	err = g.Stabilize(
		ctx,
	)
	NotNil(t, err)
	Equal(t, 2, errors)
	Equal(t, 0, output.Value())
}

func Test_Stabilize_watch(t *testing.T) {
	ctx := testContext()
	g := New()

	v0 := Var(g, 1)
	v1 := Var(g, 1)
	m0 := Map2(g, v0, v1, add)
	w0 := Watch(g, m0)

	_ = Observe(g, w0)

	_ = g.Stabilize(ctx)

	Equal(t, 1, len(w0.Values()))
	Equal(t, 2, w0.Values()[0])
	Equal(t, 2, w0.Value())

	v0.Set(2)

	_ = g.Stabilize(ctx)

	Equal(t, 2, len(w0.Values()))
	Equal(t, 2, w0.Values()[0])
	Equal(t, 3, w0.Values()[1])
}

func Test_Stabilize_Map(t *testing.T) {
	ctx := testContext()
	g := New()

	c0 := Return(g, 1)
	m := Map(g, c0, func(a int) int {
		return a + 10
	})

	_ = Observe(g, m)

	_ = g.Stabilize(ctx)
	Equal(t, 11, m.Value())
}

func Test_Stabilize_MapContext(t *testing.T) {
	ctx := testContext()
	g := New()

	c0 := Return(g, 1)
	m := MapContext(g, c0, func(ictx context.Context, a int) (int, error) {
		BlueDye(ictx, t)
		return a + 10, nil
	})

	_ = Observe(g, m)

	_ = g.Stabilize(ctx)
	Equal(t, 11, m.Value())
}

func Test_Stabilize_Map2(t *testing.T) {
	ctx := testContext()
	g := New()

	c0 := Return(g, 1)
	c1 := Return(g, 2)
	m2 := Map2(g, c0, c1, func(a, b int) int {
		return a + b
	})

	_ = Observe(g, m2)

	_ = g.Stabilize(ctx)
	Equal(t, 3, m2.Value())
}

func Test_Stabilize_Map2Context(t *testing.T) {
	ctx := testContext()
	g := New()

	c0 := Return(g, 1)
	c1 := Return(g, 2)
	m2 := Map2Context(g, c0, c1, func(ictx context.Context, a, b int) (int, error) {
		BlueDye(ctx, t)
		return a + b, nil
	})

	_ = Observe(g, m2)

	_ = g.Stabilize(ctx)
	Equal(t, 3, m2.Value())
}

func Test_Stabilize_Map2Context_error(t *testing.T) {
	ctx := testContext()
	g := New()

	c0 := Return(g, 1)
	c1 := Return(g, 2)
	m2 := Map2Context(g, c0, c1, func(ictx context.Context, a, b int) (int, error) {
		BlueDye(ctx, t)
		return a + b, fmt.Errorf("this is just a test")
	})

	_ = Observe(g, m2)

	err := g.Stabilize(ctx)
	NotNil(t, err)
	Equal(t, 0, m2.Value())
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

	_ = Observe(g, m3)

	_ = g.Stabilize(ctx)
	Equal(t, 6, m3.Value())
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

	_ = Observe(g, m3)

	_ = g.Stabilize(ctx)
	Equal(t, 10, m3.Value())
}

func Test_Stabilize_Map3Context(t *testing.T) {
	ctx := testContext()
	g := New()

	c0 := Return(g, 1)
	c1 := Return(g, 2)
	c2 := Return(g, 3)
	m3 := Map3Context(g, c0, c1, c2, func(ictx context.Context, a, b, c int) (int, error) {
		BlueDye(ictx, t)
		return a + b + c, nil
	})

	_ = Observe(g, m3)

	_ = g.Stabilize(ctx)
	Equal(t, 6, m3.Value())
}

func Test_Stabilize_Map3Context_error(t *testing.T) {
	ctx := testContext()
	g := New()

	c0 := Return(g, 1)
	c1 := Return(g, 2)
	c2 := Return(g, 3)
	m3 := Map3Context(g, c0, c1, c2, func(ictx context.Context, a, b, c int) (int, error) {
		BlueDye(ictx, t)
		return a + b + c, fmt.Errorf("this is just a test")
	})

	_ = Observe(g, m3)

	err := g.Stabilize(ctx)
	NotNil(t, err)
	Equal(t, 0, m3.Value())
}

func Test_Stabilize_MapIf(t *testing.T) {
	ctx := testContext()
	g := New()

	c0 := Return(g, 1)
	c1 := Return(g, 2)
	v0 := Var(g, false)
	mi0 := MapIf(g, c0, c1, v0)

	_ = Observe(g, mi0)

	_ = g.Stabilize(ctx)
	Equal(t, 2, mi0.Value())

	v0.Set(true)

	_ = g.Stabilize(ctx)
	Equal(t, 1, mi0.Value())

	_ = g.Stabilize(ctx)
	Equal(t, 1, mi0.Value())
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

	_ = Observe(g, mn)

	_ = g.Stabilize(ctx)
	Equal(t, 6, mn.Value())
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
	mn.AddInput(c0)
	mn.AddInput(c1)
	mn.AddInput(c2)

	_ = Observe(g, mn)

	_ = g.Stabilize(ctx)
	Equal(t, 6, mn.Value())
}

func Test_Stabilize_MapNContext(t *testing.T) {
	ctx := testContext()
	g := New()

	sum := func(ctx context.Context, values ...int) (output int, err error) {
		BlueDye(ctx, t)
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

	_ = Observe(g, mn)

	_ = g.Stabilize(ctx)
	Equal(t, 6, mn.Value())
}

func Test_Stabilize_MapNContext_error(t *testing.T) {
	ctx := testContext()
	g := New()

	sum := func(ctx context.Context, values ...int) (output int, err error) {
		BlueDye(ctx, t)
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

	_ = Observe(g, mn)

	err := g.Stabilize(ctx)
	NotNil(t, err)
	Equal(t, 0, mn.Value())
}

func Test_Stabilize_func(t *testing.T) {
	ctx := testContext()
	g := New()

	value := "hello"
	f := Func(g, func(ictx context.Context) (string, error) {
		BlueDye(ictx, t)
		return value, nil
	})
	m := MapContext(g, f, func(ictx context.Context, v string) (string, error) {
		BlueDye(ctx, t)
		return v + " world!", nil
	})

	_ = Observe(g, m)

	_ = g.Stabilize(ctx)
	Equal(t, "hello world!", m.Value())

	value = "not hello"

	_ = g.Stabilize(ctx)
	Equal(t, "hello world!", m.Value())

	// mark the func node as stale
	// not sure a better way to do this automatically?
	g.SetStale(f)

	_ = g.Stabilize(ctx)
	Equal(t, "not hello world!", m.Value())
}

func Test_Stabilize_foldMap(t *testing.T) {
	ctx := testContext()
	g := New()

	m := map[string]int{
		"one":   1,
		"two":   2,
		"three": 3,
		"four":  4,
		"five":  5,
		"six":   6,
	}
	mf := FoldMap(g, Return(g, m), 0, func(key string, val, accum int) int {
		return accum + val
	})

	_ = Observe(g, mf)

	_ = g.Stabilize(ctx)
	Equal(t, 21, mf.Value())
}

func Test_Stabilize_foldLeft(t *testing.T) {
	ctx := testContext()
	g := New()

	m := []int{
		1,
		2,
		3,
		4,
		5,
		6,
	}
	mf := FoldLeft(g, Return(g, m), "", func(accum string, val int) string {
		return accum + fmt.Sprint(val)
	})

	_ = Observe(g, mf)

	_ = g.Stabilize(ctx)
	Equal(t, "123456", mf.Value())
}

func Test_Stabilize_foldRight(t *testing.T) {
	ctx := testContext()
	g := New()

	m := []int{
		1,
		2,
		3,
		4,
		5,
		6,
	}
	mf := FoldRight(g, Return(g, m), "", func(val int, accum string) string {
		return accum + fmt.Sprint(val)
	})

	_ = Observe(g, mf)

	_ = g.Stabilize(ctx)
	Equal(t, "654321", mf.Value())

	g.SetStale(mf)

	_ = g.Stabilize(ctx)
	Equal(t, "654321654321", mf.Value())
}

func Test_Stabilize_freeze(t *testing.T) {
	ctx := testContext()
	g := New()

	v0 := Var(g, "hello")
	fv := Freeze(g, v0)

	_ = Observe(g, fv)

	err := g.Stabilize(ctx)
	Nil(t, err)
	Equal(t, "hello", v0.Value())
	Equal(t, "hello", fv.Value())

	v0.Set("not-hello")

	err = g.Stabilize(ctx)
	Nil(t, err)
	Equal(t, "not-hello", v0.Value())
	Equal(t, "hello", fv.Value())
}

func Test_Stabilize_always_cutoff(t *testing.T) {
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
	o := Observe(g, readFile)

	err := g.Stabilize(ctx)
	Nil(t, err)
	Equal(t, "test-1", o.Value())

	err = g.Stabilize(ctx)
	Nil(t, err)
	Equal(t, "test-1", o.Value())

	modtime = 2

	err = g.Stabilize(ctx)
	Nil(t, err)
	Equal(t, "test-2", o.Value())

	err = g.Stabilize(ctx)
	Nil(t, err)
	Equal(t, "test-2", o.Value())
}

func Test_Stabilize_always_cutoff_error(t *testing.T) {
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
	o := Observe(g, readFile)

	err := g.Stabilize(ctx)
	NotNil(t, err)
	Equal(t, "", o.Value())

	Equal(t, 3, g.recomputeHeap.len())
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
	_ = Observe(g, gonnaPanic)

	err := g.Stabilize(ctx)
	NotNil(t, err)
	NotEqual(t, 0, len(outBuf.String()))
	NotEqual(t, 0, len(errBuf.String()))
	Equal(t, true, strings.Contains(errBuf.String(), "this is only a test"))
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

	_ = Observe(g, m0)
	g.OnStabilizationStart(func(ictx context.Context) {
		startWasBlueDye = HasBlueDye(ctx)
		didCallStabilizationStart = true
	})
	g.OnStabilizationEnd(func(ictx context.Context, started time.Time, err error) {
		endWasBlueDye = HasBlueDye(ctx)
		didCallStabilizationEnd = true
	})
	err := g.Stabilize(ctx)
	Nil(t, err)
	Equal(t, true, didCallStabilizationStart)
	Equal(t, true, didCallStabilizationEnd)
	Equal(t, true, startWasBlueDye)
	Equal(t, true, endWasBlueDye)
}

func Test_Stabilize_bindCombination(t *testing.T) {
	ctx := testContext()
	g := New()

	v1 := Var(g, 1)
	v2 := Var(g, 2)
	v3 := Var(g, 3)
	v4 := Var(g, 4)

	o := Observe(g, Bind4(g, v1, v2, v3, v4, func(bs Scope, x1, x2, x3, x4 int) Incr[int] {
		return Bind3(bs, v2, v3, v3, func(bs Scope, y2, y3, y4 int) Incr[int] {
			return Bind2(bs, v4, v4, func(bs Scope, z3, z4 int) Incr[int] {
				return Bind(bs, v4, func(bs Scope, w4 int) Incr[int] {
					return Return(bs, x1+x2+x3+x4+y2+y3+y4+z3+z4+w4)
				})
			})
		})
	}))

	err := g.Stabilize(ctx)
	NoError(t, err)

	Equal(t, v1.Value()+(2*v2.Value())+(3*v3.Value())+(4*v4.Value()), o.Value())

	v1.Set(9)
	v2.Set(10)
	v3.Set(11)
	v4.Set(12)

	err = g.Stabilize(ctx)
	NoError(t, err)

	Equal(t, v1.Value()+(2*v2.Value())+(3*v3.Value())+(4*v4.Value()), o.Value())

}
