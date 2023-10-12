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

	v0 := Var("foo")
	v1 := Var("bar")
	m0 := Map2(v0, v1, func(a, b string) string {
		return a + " " + b
	})

	graph := New()
	_ = Observe(graph, m0)

	err := graph.Stabilize(ctx)
	ItsNil(t, err)

	ItsEqual(t, 0, v0.Node().setAt)
	ItsEqual(t, 1, v0.Node().changedAt)
	ItsEqual(t, 0, v1.Node().setAt)
	ItsEqual(t, 1, v1.Node().changedAt)
	ItsEqual(t, 1, m0.Node().changedAt)
	ItsEqual(t, 1, v0.Node().recomputedAt)
	ItsEqual(t, 1, v1.Node().recomputedAt)
	ItsEqual(t, 1, m0.Node().recomputedAt)

	ItsEqual(t, "foo bar", m0.Value())

	v0.Set("not foo")
	ItsEqual(t, 2, v0.Node().setAt)
	ItsEqual(t, 0, v1.Node().setAt)

	err = graph.Stabilize(ctx)
	ItsNil(t, err)

	ItsEqual(t, 2, v0.Node().changedAt)
	ItsEqual(t, 1, v1.Node().changedAt)
	ItsEqual(t, 2, m0.Node().changedAt)

	ItsEqual(t, 2, v0.Node().recomputedAt)
	ItsEqual(t, 1, v1.Node().recomputedAt)
	ItsEqual(t, 2, m0.Node().recomputedAt)

	ItsEqual(t, "not foo bar", m0.Value())
}

func Test_Stabilize_error(t *testing.T) {
	ctx := testContext()

	m0 := Func(func(_ context.Context) (string, error) {
		return "", fmt.Errorf("this is just a test")
	})

	graph := New()
	_ = Observe(graph, m0)

	err := graph.Stabilize(ctx)
	ItsNotNil(t, err)
	ItsEqual(t, "this is just a test", err.Error())
}

func Test_Stabilize_errorHandler(t *testing.T) {
	ctx := testContext()

	m0 := Func(func(_ context.Context) (string, error) {
		return "", fmt.Errorf("this is just a test")
	})
	var gotError error
	m0.Node().OnError(func(ctx context.Context, err error) {
		ItsBlueDye(ctx, t)
		gotError = err
	})

	graph := New()
	_ = Observe(graph, m0)

	err := graph.Stabilize(ctx)
	ItsNotNil(t, err)
	ItsEqual(t, "this is just a test", err.Error())
	ItsEqual(t, "this is just a test", gotError.Error())
}

func Test_Stabilize_alreadyStabilizing(t *testing.T) {
	ctx := testContext()

	// deadlocks. deadlocks everywhere.
	hold := make(chan struct{})
	errs := make(chan error)
	m0 := Func(func(_ context.Context) (string, error) {
		<-hold
		return "ok!", nil
	})

	graph := New()
	_ = Observe(graph, m0)

	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		if err := graph.Stabilize(ctx); err != nil {
			errs <- err
		}
	}()
	go func() {
		defer wg.Done()
		if err := graph.Stabilize(ctx); err != nil {
			errs <- err
		}
	}()
	err := <-errs
	ItsEqual(t, ErrAlreadyStabilizing, err)
	close(hold)
	wg.Wait()
	ItsEqual(t, "ok!", m0.Value())
}

func Test_Stabilize_updateHandlers(t *testing.T) {
	ctx := testContext()

	v0 := Var("foo")
	v1 := Var("bar")
	m0 := Map2(v0, v1, func(a, b string) string {
		return a + " " + b
	})

	var updates int
	m0.Node().OnUpdate(func(_ context.Context) {
		updates++
	})

	graph := New()
	_ = Observe(graph, m0)

	err := graph.Stabilize(ctx)
	ItsNil(t, err)
	ItsEqual(t, 1, updates)

	v0.Set("not foo")
	err = graph.Stabilize(ctx)
	ItsNil(t, err)
	ItsEqual(t, 2, updates)
}

func Test_Stabilize_observedHandlers(t *testing.T) {
	ctx := testContext()

	v0 := Var("foo")
	v1 := Var("bar")
	m0 := Map2(v0, v1, func(a, b string) string {
		return a + " " + b
	})

	var observes int
	m0.Node().OnObserved(func(IObserver) {
		observes++
	})

	graph := New()
	_ = Observe(graph, m0)

	err := graph.Stabilize(ctx)
	ItsNil(t, err)
	ItsEqual(t, 1, observes)

	v0.Set("not foo")
	err = graph.Stabilize(ctx)
	ItsNil(t, err)
	ItsEqual(t, 1, observes)

	_ = Observe(graph, m0)
	ItsEqual(t, 2, observes)
}

func Test_Stabilize_unobservedHandlers(t *testing.T) {
	ctx := testContext()

	v0 := Var("foo")
	v1 := Var("bar")
	m0 := Map2(v0, v1, func(a, b string) string {
		return a + " " + b
	})

	var observes, unobserves int
	m0.Node().OnObserved(func(IObserver) {
		observes++
	})
	m0.Node().OnUnobserved(func(IObserver) {
		unobserves++
	})

	graph := New()
	o0 := Observe(graph, m0)

	err := graph.Stabilize(ctx)
	ItsNil(t, err)
	ItsEqual(t, 1, observes)
	ItsEqual(t, 0, unobserves)

	v0.Set("not foo")
	err = graph.Stabilize(ctx)
	ItsNil(t, err)
	ItsEqual(t, 1, observes)

	_ = Observe(graph, m0)
	ItsEqual(t, 2, observes)
	ItsEqual(t, 0, unobserves)

	o0.Unobserve()
	ItsEqual(t, 2, observes)
	ItsEqual(t, 1, unobserves)
}

func Test_Stabilize_unevenHeights(t *testing.T) {
	ctx := testContext()

	v0 := Var("foo")
	v1 := Var("bar")
	m0 := Map2[string, string](v0, v1, func(a, b string) string {
		return a + " " + b
	})
	r0 := Return("moo")
	m1 := Map2(r0, m0, func(a, b string) string {
		return a + " != " + b
	})

	graph := New()
	_ = Observe(graph, m1)

	err := graph.Stabilize(ctx)
	ItsNil(t, err)
	ItsEqual(t, "moo != foo bar", m1.Value())

	v0.Set("not foo")
	err = graph.Stabilize(ctx)
	ItsNil(t, err)
	ItsEqual(t, "moo != not foo bar", m1.Value())
}

func Test_Stabilize_chain(t *testing.T) {
	ctx := testContext()

	v0 := Var(".")

	var maps []Incr[string]
	var previous Incr[string] = v0
	for x := 0; x < 100; x++ {
		m := Map(previous, func(v0 string) string {
			return v0 + "."
		})
		maps = append(maps, m)
		previous = m
	}

	graph := New()
	o := Observe(graph, maps[len(maps)-1])

	err := graph.Stabilize(ctx)
	ItsNil(t, err)
	ItsEqual(t, strings.Repeat(".", 101), o.Value())

	ItsEqual(t, 102, graph.numNodes, "should include the observer!")
	ItsEqual(t, 102, graph.numNodesChanged)
	ItsEqual(t, 102, graph.numNodesRecomputed)
}

func Test_Stabilize_setDuringStabilization(t *testing.T) {
	ctx := testContext()
	v0 := Var("foo")

	called := make(chan struct{})
	wait := make(chan struct{})
	m0 := Map(v0, func(v string) string {
		close(called)
		<-wait
		return v
	})

	graph := New()
	_ = Observe(graph, m0)

	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = graph.Stabilize(ctx)
	}()
	<-called
	v0.Set("not-foo")
	ItsEqual(t, "foo", v0.Value())
	close(wait)
	<-done
	ItsEqual(t, "not-foo", v0.Value())
	ItsEqual(t, graph.stabilizationNum, v0.Node().setAt)
	ItsEqual(t, 1, len(graph.recomputeHeap.lookup))
}

func Test_Stabilize_onUpdate(t *testing.T) {
	ctx := testContext()

	var didCallUpdateHandler0, didCallUpdateHandler1 bool
	v0 := Var("hello")
	v1 := Var("world")
	m0 := Map2(v0, v1, concat)
	m0.Node().OnUpdate(func(context.Context) {
		didCallUpdateHandler0 = true
	})
	m0.Node().OnUpdate(func(context.Context) {
		didCallUpdateHandler1 = true
	})

	graph := New()
	_ = Observe(graph, m0)

	err := graph.Stabilize(ctx)
	ItsNil(t, err)
	ItsEqual(t, "helloworld", m0.Value())
	ItsEqual(t, true, didCallUpdateHandler0)
	ItsEqual(t, true, didCallUpdateHandler1)
}

func Test_Stabilize_recombinant_singleUpdate(t *testing.T) {
	ctx := testContext()

	// a -> b -> c -> d -> z
	//   -> f -> e -> [z]
	// assert that [z] updates (1) time if we change [a]

	edge := func(l string) func(string) string {
		return func(v string) string {
			return v + "->" + l
		}
	}

	a := Var("a")
	b := Map(a, edge("b"))
	c := Map(b, edge("c"))
	d := Map(c, edge("d"))
	f := Map(a, edge("f"))
	e := Map(f, edge("e"))

	z := Map2(d, e, func(v0, v1 string) string {
		return v0 + "+" + v1 + "->z"
	})

	graph := New()
	_ = Observe(graph, z)

	err := graph.Stabilize(ctx)
	ItsNil(t, err)
	ItsEqual(t, 1, z.Node().numRecomputes)
	ItsEqual(t, "a->b->c->d+a->f->e->z", z.Value())

	a.Set("!a")

	err = graph.Stabilize(ctx)
	ItsNil(t, err)
	ItsEqual(t, "!a->b->c->d+!a->f->e->z", z.Value())
	ItsEqual(t, 2, z.Node().numRecomputes)
}

func Test_Stabilize_doubleVarSet_singleUpdate(t *testing.T) {
	ctx := testContext()

	a := Var("a")
	b := Var("b")
	m := Map2(a, b, func(v0, v1 string) string {
		return v0 + " " + v1
	})

	graph := New()
	_ = Observe(graph, m)

	_ = graph.Stabilize(ctx)
	ItsEqual(t, "a b", m.Value())

	a.Set("aa")
	ItsEqual(t, 1, graph.recomputeHeap.Len())

	a.Set("aaa")
	ItsEqual(t, 1, graph.recomputeHeap.Len())

	_ = graph.Stabilize(ctx)
	ItsEqual(t, "aaa b", m.Value())
}

func Test_Stabilize_verifyPartial(t *testing.T) {
	ctx := testContext()

	v0 := Var("foo")
	c0 := Return("bar")
	v1 := Var("moo")
	c1 := Return("baz")

	m0 := Map2(v0, c0, func(a, b string) string {
		return a + " " + b
	})
	co0 := Cutoff(m0, func(n, o string) bool {
		return len(n) == len(o)
	})
	m1 := Map2(v1, c1, func(a, b string) string {
		return a + " != " + b
	})
	co1 := Cutoff(m1, func(n, o string) bool {
		return len(n) == len(o)
	})

	sw := Var(true)
	mi := MapIf(co0, co1, sw)

	graph := New()
	_ = Observe(graph, mi)

	err := graph.Stabilize(ctx)
	ItsNil(t, err)
	ItsEqual(t, "foo bar", mi.Value())

	v0.Set("Foo")

	err = graph.Stabilize(ctx)
	ItsNil(t, err)
	ItsEqual(t, "foo bar", mi.Value())
}

func Test_Stabilize_jsDocs(t *testing.T) {
	ctx := testContext()

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

	i := Var(data)
	output := Map(
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

	graph := New()
	_ = Observe(graph, output)

	err := graph.Stabilize(
		ctx,
	)
	ItsNil(t, err)
	ItsEqual(t, 2, len(output.Value()))

	data = append(data, Entry{
		"5", now.Add(5 * time.Second),
	})
	err = graph.Stabilize(
		ctx,
	)
	ItsNil(t, err)
	ItsEqual(t, 2, len(output.Value()))

	i.Set(data)
	err = graph.Stabilize(
		context.Background(),
	)
	ItsNil(t, err)
	ItsEqual(t, 3, len(output.Value()))
}

func Test_Stabilize_bind(t *testing.T) {
	ctx := testContext()

	sw := Var(false)
	i0 := Return("foo")
	i0.Node().SetLabel("i0")
	m0 := Map(i0, func(v0 string) string { return v0 + "-moo" })
	m0.Node().SetLabel("m0")
	i1 := Return("bar")
	i1.Node().SetLabel("i1")
	m1 := Map(i1, func(v0 string) string { return v0 + "-loo" })
	m1.Node().SetLabel("m1")
	b := Bind(sw, func(swv bool) Incr[string] {
		if swv {
			return m0
		}
		return m1
	})
	mb := Map[string](b, func(v string) string {
		return v + "-baz"
	})
	mb.Node().SetLabel("mb")

	ItsEqual(t, 0, len(i0.Node().parents))
	ItsEqual(t, 1, len(i0.Node().children))
	ItsEqual(t, 1, len(m0.Node().parents))
	ItsEqual(t, 0, len(m0.Node().children))

	ItsEqual(t, 0, len(i1.Node().parents))
	ItsEqual(t, 1, len(i1.Node().children))
	ItsEqual(t, 1, len(m1.Node().parents))
	ItsEqual(t, 0, len(m1.Node().children))

	graph := New()
	_ = Observe(graph, mb)

	ItsEqual(t, true, graph.IsObserving(sw))

	err := graph.Stabilize(ctx)
	ItsNil(t, err)

	ItsEqual(t, 0, len(i0.Node().parents))
	ItsEqual(t, 1, len(i0.Node().children))
	ItsEqual(t, 1, len(m0.Node().parents))
	ItsEqual(t, 0, len(m0.Node().children), "children should not include mb (util we flip the var and rebind)")

	ItsEqual(t, 0, len(i1.Node().parents))
	ItsEqual(t, 1, len(i1.Node().children))
	ItsEqual(t, 1, len(m1.Node().parents), "parents should include i1")
	ItsEqual(t, 1, len(m1.Node().children), "children should include mb")

	ItsEqual(t, false, graph.IsObserving(i0))
	ItsEqual(t, false, graph.IsObserving(m0))
	ItsNil(t, i0.Node().graph, "i0 should not be in the graph after the first stabilization")
	ItsNil(t, m0.Node().graph, "m0 should not be in the graph after the first stabilization")

	ItsEqual(t, true, graph.IsObserving(i1))
	ItsEqual(t, true, graph.IsObserving(m1))
	ItsNotNil(t, i1.Node().graph, "i1 should be in the graph after the first stabilization")
	ItsNotNil(t, m1.Node().graph, "m1 should be in the graph after the first stabilization")

	ItsEqual(t, "bar-loo-baz", mb.Value())

	sw.Set(true)
	ItsEqual(t, true, graph.recomputeHeap.Has(sw))

	err = graph.Stabilize(ctx)
	ItsNil(t, err)

	ItsEqual(t, 0, len(i0.Node().parents))
	ItsEqual(t, 1, len(i0.Node().children))
	ItsEqual(t, 1, len(m0.Node().parents))
	ItsEqual(t, 1, len(m0.Node().children))

	ItsEqual(t, 0, len(i1.Node().parents))
	ItsEqual(t, 1, len(i1.Node().children))
	ItsEqual(t, 1, len(m1.Node().parents), "m1 parents should include the input")
	ItsEqual(t, 0, len(m1.Node().children), "m1 should be swapped out for m0 at this point and have no children")

	ItsEqual(t, true, graph.IsObserving(i0))
	ItsEqual(t, true, graph.IsObserving(m0))
	ItsNotNil(t, i0.Node().graph, "i0 should be in the graph after the second stabilization")
	ItsNotNil(t, m0.Node().graph, "m0 should be in the graph after the second stabilization")

	ItsEqual(t, false, graph.IsObserving(i1))
	ItsEqual(t, false, graph.IsObserving(m1))
	ItsNil(t, i1.Node().graph, "i1 should not be in the graph after the second stabilization")
	ItsNil(t, m1.Node().graph, "m1 should not be in the graph after the second stabilization")

	ItsEqual(t, "foo-moo-baz", mb.Value())
}

func Test_Stabilize_bindIf(t *testing.T) {
	ctx := testContext()

	sw := Var(false)
	i0 := Return("foo")
	i1 := Return("bar")

	b := BindIf(sw, func(ctx context.Context, swv bool) (Incr[string], error) {
		testutil.ItsBlueDye(ctx, t)
		if swv {
			return i0, nil
		}
		return i1, nil
	})

	graph := New()
	_ = Observe(graph, b)

	err := graph.Stabilize(ctx)
	ItsNil(t, err)

	ItsNil(t, i0.Node().graph, "i0 should not be in the graph after the first stabilization")
	ItsNotNil(t, i1.Node().graph, "i1 should be in the graph after the first stabilization")

	ItsEqual(t, "bar", b.Value())

	sw.Set(true)
	err = graph.Stabilize(ctx)
	ItsNil(t, err)

	ItsNil(t, i1.Node().graph, "i0 should be in the graph after the third stabilization")
	ItsNotNil(t, i0.Node().graph, "i1 should not be in the graph after the third stabilization")

	ItsEqual(t, "foo", b.Value())
}

func Test_Stabilize_cutoff(t *testing.T) {
	ctx := testContext()

	input := Var(3.14)
	cutoff := Cutoff[float64](
		input,
		epsilon(0.1),
	)
	output := Map2(
		cutoff,
		Return(10.0),
		add[float64],
	)

	graph := New()
	_ = Observe(graph, output)

	_ = graph.Stabilize(
		ctx,
	)
	ItsEqual(t, 13.14, output.Value())
	ItsEqual(t, 3.14, cutoff.Value())

	input.Set(3.15)

	_ = graph.Stabilize(
		ctx,
	)
	ItsEqual(t, 3.14, cutoff.Value())
	ItsEqual(t, 13.14, output.Value())

	input.Set(3.26) // differs by 0.11, which is > 0.1

	_ = graph.Stabilize(
		ctx,
	)
	ItsEqual(t, 3.26, cutoff.Value())
	ItsEqual(t, 13.26, output.Value())

	_ = graph.Stabilize(
		ctx,
	)
	ItsEqual(t, 13.26, output.Value())
}

type MathTypes interface {
	~int | ~int64 | ~float32 | ~float64
}

func Test_Stabilize_cutoffContext(t *testing.T) {
	ctx := testContext()
	input := Var(3.14)

	cutoff := CutoffContext[float64](
		input,
		epsilonContext(t, 0.1),
	)

	output := Map2(
		cutoff,
		Return(10.0),
		add[float64],
	)

	graph := New()
	_ = Observe(graph, output)

	_ = graph.Stabilize(
		ctx,
	)
	ItsEqual(t, 13.14, output.Value())
	ItsEqual(t, 3.14, cutoff.Value())

	input.Set(3.15)

	_ = graph.Stabilize(
		ctx,
	)
	ItsEqual(t, 3.14, cutoff.Value())
	ItsEqual(t, 13.14, output.Value())

	input.Set(3.26) // differs by 0.11, which is > 0.1

	_ = graph.Stabilize(
		ctx,
	)
	ItsEqual(t, 3.26, cutoff.Value())
	ItsEqual(t, 13.26, output.Value())

	_ = graph.Stabilize(
		ctx,
	)
	ItsEqual(t, 13.26, output.Value())
}

func Test_Stabilize_cutoffContext_error(t *testing.T) {
	ctx := testContext()
	input := Var(3.14)

	cutoff := CutoffContext[float64](
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
		cutoff,
		Return(10.0),
		add[float64],
	)

	graph := New()
	_ = Observe(graph, output)

	err := graph.Stabilize(
		ctx,
	)
	testutil.ItsNotNil(t, err)
	testutil.ItsEqual(t, 1, errors)
	testutil.ItsEqual(t, 0, output.Value())

	input.Set(3.15)

	err = graph.Stabilize(
		ctx,
	)
	testutil.ItsNotNil(t, err)
	testutil.ItsEqual(t, 2, errors)
	testutil.ItsEqual(t, 0, output.Value())
}

func epsilonFn[A, B MathTypes](eps A, oldv, newv B) bool {
	if oldv > newv {
		return oldv-newv <= B(eps)
	}
	return newv-oldv <= B(eps)
}

func Test_Stabilize_cutoff2(t *testing.T) {
	ctx := testContext()

	epsilon := Var(0.1)
	input := Var(3.14)
	cutoff := Cutoff2[float64, float64](
		epsilon,
		input,
		epsilonFn,
	)
	output := Map2(
		cutoff,
		Return(10.0),
		add[float64],
	)

	graph := New()
	_ = Observe(graph, output)

	_ = graph.Stabilize(
		ctx,
	)
	ItsEqual(t, 13.14, output.Value())
	ItsEqual(t, 3.14, cutoff.Value())

	input.Set(3.15)

	_ = graph.Stabilize(
		ctx,
	)
	ItsEqual(t, 3.14, cutoff.Value())
	ItsEqual(t, 13.14, output.Value())

	input.Set(3.26) // differs by 0.11, which is > 0.1

	_ = graph.Stabilize(
		ctx,
	)
	ItsEqual(t, 3.26, cutoff.Value())
	ItsEqual(t, 13.26, output.Value())

	_ = graph.Stabilize(
		ctx,
	)
	ItsEqual(t, 13.26, output.Value())

	epsilon.Set(0.5)
	input.Set(3.37) // differs by 0.11, which is < 0.5

	_ = graph.Stabilize(
		ctx,
	)
	ItsEqual(t, 3.26, cutoff.Value())
	ItsEqual(t, 13.26, output.Value())

	_ = graph.Stabilize(
		ctx,
	)
	ItsEqual(t, 13.26, output.Value())
}

func Test_Stabilize_cutoff2Context_error(t *testing.T) {
	ctx := testContext()
	epsilon := Var(0.1)
	input := Var(3.14)

	cutoff := Cutoff2Context[float64, float64](
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
		cutoff,
		Return(10.0),
		add[float64],
	)

	graph := New()
	_ = Observe(graph, output)

	err := graph.Stabilize(
		ctx,
	)
	testutil.ItsNotNil(t, err)
	testutil.ItsEqual(t, 1, errors)
	testutil.ItsEqual(t, 0, output.Value())

	input.Set(3.15)

	err = graph.Stabilize(
		ctx,
	)
	testutil.ItsNotNil(t, err)
	testutil.ItsEqual(t, 2, errors)
	testutil.ItsEqual(t, 0, output.Value())
}

func Test_Stabilize_watch(t *testing.T) {
	ctx := testContext()

	v0 := Var(1)
	v1 := Var(1)
	m0 := Map2(v0, v1, add)
	w0 := Watch(m0)

	graph := New()
	_ = Observe(graph, w0)

	_ = graph.Stabilize(ctx)

	ItsEqual(t, 1, len(w0.Values()))
	ItsEqual(t, 2, w0.Values()[0])
	ItsEqual(t, 2, w0.Value())

	v0.Set(2)

	_ = graph.Stabilize(ctx)

	ItsEqual(t, 2, len(w0.Values()))
	ItsEqual(t, 2, w0.Values()[0])
	ItsEqual(t, 3, w0.Values()[1])
}

func Test_Stabilize_Map(t *testing.T) {
	ctx := testContext()

	c0 := Return(1)
	m := Map(c0, func(a int) int {
		return a + 10
	})

	graph := New()
	_ = Observe(graph, m)

	_ = graph.Stabilize(ctx)
	ItsEqual(t, 11, m.Value())
}

func Test_Stabilize_MapContext(t *testing.T) {
	ctx := testContext()

	c0 := Return(1)
	m := MapContext(c0, func(ictx context.Context, a int) (int, error) {
		testutil.ItsBlueDye(ictx, t)
		return a + 10, nil
	})

	graph := New()
	_ = Observe(graph, m)

	_ = graph.Stabilize(ctx)
	ItsEqual(t, 11, m.Value())
}

func Test_Stabilize_Map2(t *testing.T) {
	ctx := testContext()

	c0 := Return(1)
	c1 := Return(2)
	m2 := Map2(c0, c1, func(a, b int) int {
		return a + b
	})

	graph := New()
	_ = Observe(graph, m2)

	_ = graph.Stabilize(ctx)
	ItsEqual(t, 3, m2.Value())
}

func Test_Stabilize_Map2Context(t *testing.T) {
	ctx := testContext()

	c0 := Return(1)
	c1 := Return(2)
	m2 := Map2Context(c0, c1, func(ictx context.Context, a, b int) (int, error) {
		testutil.ItsBlueDye(ctx, t)
		return a + b, nil
	})

	graph := New()
	_ = Observe(graph, m2)

	_ = graph.Stabilize(ctx)
	ItsEqual(t, 3, m2.Value())
}

func Test_Stabilize_Map2Context_error(t *testing.T) {
	ctx := testContext()

	c0 := Return(1)
	c1 := Return(2)
	m2 := Map2Context(c0, c1, func(ictx context.Context, a, b int) (int, error) {
		testutil.ItsBlueDye(ctx, t)
		return a + b, fmt.Errorf("this is just a test")
	})

	graph := New()
	_ = Observe(graph, m2)

	err := graph.Stabilize(ctx)
	ItsNotNil(t, err)
	ItsEqual(t, 0, m2.Value())
}

func Test_Stabilize_Map3(t *testing.T) {
	ctx := testContext()

	c0 := Return(1)
	c1 := Return(2)
	c2 := Return(3)
	m3 := Map3(c0, c1, c2, func(a, b, c int) int {
		return a + b + c
	})

	graph := New()
	_ = Observe(graph, m3)

	_ = graph.Stabilize(ctx)
	ItsEqual(t, 6, m3.Value())
}

func Test_Stabilize_Map3Context(t *testing.T) {
	ctx := testContext()

	c0 := Return(1)
	c1 := Return(2)
	c2 := Return(3)
	m3 := Map3Context(c0, c1, c2, func(ictx context.Context, a, b, c int) (int, error) {
		testutil.ItsBlueDye(ictx, t)
		return a + b + c, nil
	})

	graph := New()
	_ = Observe(graph, m3)

	_ = graph.Stabilize(ctx)
	ItsEqual(t, 6, m3.Value())
}

func Test_Stabilize_Map3Context_error(t *testing.T) {
	ctx := testContext()

	c0 := Return(1)
	c1 := Return(2)
	c2 := Return(3)
	m3 := Map3Context(c0, c1, c2, func(ictx context.Context, a, b, c int) (int, error) {
		testutil.ItsBlueDye(ictx, t)
		return a + b + c, fmt.Errorf("this is just a test")
	})

	graph := New()
	_ = Observe(graph, m3)

	err := graph.Stabilize(ctx)
	ItsNotNil(t, err)
	ItsEqual(t, 0, m3.Value())
}

func Test_Stabilize_MapIf(t *testing.T) {
	ctx := testContext()

	c0 := Return(1)
	c1 := Return(2)
	v0 := Var(false)
	mi0 := MapIf(c0, c1, v0)

	graph := New()
	_ = Observe(graph, mi0)

	_ = graph.Stabilize(ctx)
	ItsEqual(t, 2, mi0.Value())

	v0.Set(true)

	_ = graph.Stabilize(ctx)
	ItsEqual(t, 1, mi0.Value())

	_ = graph.Stabilize(ctx)
	ItsEqual(t, 1, mi0.Value())
}

func Test_Stabilize_MapN(t *testing.T) {
	ctx := testContext()

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

	c0 := Return(1)
	c1 := Return(2)
	c2 := Return(3)
	mn := MapN(sum, c0, c1, c2)

	graph := New()
	_ = Observe(graph, mn)

	_ = graph.Stabilize(ctx)
	ItsEqual(t, 6, mn.Value())
}

func Test_Stabilize_MapN_AddInput(t *testing.T) {
	ctx := testContext()

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

	c0 := Return(1)
	c1 := Return(2)
	c2 := Return(3)
	mn := MapN(sum)
	mn.AddInput(c0)
	mn.AddInput(c1)
	mn.AddInput(c2)

	graph := New()
	_ = Observe(graph, mn)

	_ = graph.Stabilize(ctx)
	ItsEqual(t, 6, mn.Value())
}

func Test_Stabilize_MapNContext(t *testing.T) {
	ctx := testContext()

	sum := func(ctx context.Context, values ...int) (output int, err error) {
		ItsBlueDye(ctx, t)
		if len(values) == 0 {
			return
		}
		output = values[0]
		for _, value := range values[1:] {
			output += value
		}
		return
	}

	c0 := Return(1)
	c1 := Return(2)
	c2 := Return(3)
	mn := MapNContext(sum, c0, c1, c2)

	graph := New()
	_ = Observe(graph, mn)

	_ = graph.Stabilize(ctx)
	ItsEqual(t, 6, mn.Value())
}

func Test_Stabilize_MapNContext_error(t *testing.T) {
	ctx := testContext()

	sum := func(ctx context.Context, values ...int) (output int, err error) {
		ItsBlueDye(ctx, t)
		for _, value := range values {
			output += value
		}
		err = fmt.Errorf("this is just a test")
		return
	}

	c0 := Return(1)
	c1 := Return(2)
	c2 := Return(3)
	mn := MapNContext(sum, c0, c1, c2)

	graph := New()
	_ = Observe(graph, mn)

	err := graph.Stabilize(ctx)
	ItsNotNil(t, err)
	ItsEqual(t, 0, mn.Value())
}

func Test_Stabilize_func(t *testing.T) {
	ctx := testContext()

	value := "hello"
	f := Func(func(ictx context.Context) (string, error) {
		testutil.ItsBlueDye(ictx, t)
		return value, nil
	})
	m := MapContext(f, func(ictx context.Context, v string) (string, error) {
		testutil.ItsBlueDye(ctx, t)
		return v + " world!", nil
	})

	graph := New()
	_ = Observe(graph, m)

	_ = graph.Stabilize(ctx)
	ItsEqual(t, "hello world!", m.Value())

	value = "not hello"

	_ = graph.Stabilize(ctx)
	ItsEqual(t, "hello world!", m.Value())

	// mark the func node as stale
	// not sure a better way to do this automatically?
	graph.SetStale(f)

	_ = graph.Stabilize(ctx)
	ItsEqual(t, "not hello world!", m.Value())
}

func Test_Stabilize_foldMap(t *testing.T) {
	ctx := testContext()

	m := map[string]int{
		"one":   1,
		"two":   2,
		"three": 3,
		"four":  4,
		"five":  5,
		"six":   6,
	}
	mf := FoldMap(Return(m), 0, func(key string, val, accum int) int {
		return accum + val
	})

	graph := New()
	_ = Observe(graph, mf)

	_ = graph.Stabilize(ctx)
	ItsEqual(t, 21, mf.Value())
}

func Test_Stabilize_foldLeft(t *testing.T) {
	ctx := testContext()

	m := []int{
		1,
		2,
		3,
		4,
		5,
		6,
	}
	mf := FoldLeft(Return(m), "", func(accum string, val int) string {
		return accum + fmt.Sprint(val)
	})

	graph := New()
	_ = Observe(graph, mf)

	_ = graph.Stabilize(ctx)
	ItsEqual(t, "123456", mf.Value())
}

func Test_Stabilize_foldRight(t *testing.T) {
	ctx := testContext()

	m := []int{
		1,
		2,
		3,
		4,
		5,
		6,
	}
	mf := FoldRight(Return(m), "", func(val int, accum string) string {
		return accum + fmt.Sprint(val)
	})

	graph := New()
	_ = Observe(graph, mf)

	_ = graph.Stabilize(ctx)
	ItsEqual(t, "654321", mf.Value())

	graph.SetStale(mf)

	_ = graph.Stabilize(ctx)
	ItsEqual(t, "654321654321", mf.Value())
}

func Test_Stabilize_freeze(t *testing.T) {
	ctx := testContext()

	v0 := Var("hello")
	fv := Freeze(v0)

	graph := New()
	_ = Observe(graph, fv)

	err := graph.Stabilize(ctx)
	ItsNil(t, err)
	ItsEqual(t, "hello", v0.Value())
	ItsEqual(t, "hello", fv.Value())

	v0.Set("not-hello")

	err = graph.Stabilize(ctx)
	ItsNil(t, err)
	ItsEqual(t, "not-hello", v0.Value())
	ItsEqual(t, "hello", fv.Value())
}

func Test_Stabilize_always_cutoff(t *testing.T) {
	ctx := testContext()
	g := New()

	filename := Var("test")
	filenameAlways := Always(filename)
	modtime := 1
	statfile := Map(filenameAlways, func(s string) int { return modtime })
	statfileCutoff := Cutoff(statfile, func(ov, nv int) bool {
		return ov == nv
	})
	readFile := Map2(filename, statfileCutoff, func(p string, mt int) string {
		return fmt.Sprintf("%s-%d", p, mt)
	})
	o := Observe(g, readFile)

	err := g.Stabilize(ctx)
	testutil.ItsNil(t, err)
	testutil.ItsEqual(t, "test-1", o.Value())

	err = g.Stabilize(ctx)
	testutil.ItsNil(t, err)
	testutil.ItsEqual(t, "test-1", o.Value())

	modtime = 2

	err = g.Stabilize(ctx)
	testutil.ItsNil(t, err)
	testutil.ItsEqual(t, "test-2", o.Value())

	err = g.Stabilize(ctx)
	testutil.ItsNil(t, err)
	testutil.ItsEqual(t, "test-2", o.Value())
}

func Test_Stabilize_always_cutoff_error(t *testing.T) {
	ctx := testContext()
	g := New()

	filename := Var("test")
	filenameAlways := Always(filename)
	modtime := 1
	statfile := Map(filenameAlways, func(s string) int { return modtime })
	statfileCutoff := CutoffContext(statfile, func(_ context.Context, ov, nv int) (bool, error) {
		return false, fmt.Errorf("this is only a test")
	})
	readFile := Map2(filename, statfileCutoff, func(p string, mt int) string {
		return fmt.Sprintf("%s-%d", p, mt)
	})
	o := Observe(g, readFile)

	err := g.Stabilize(ctx)
	testutil.ItsNotNil(t, err)
	testutil.ItsEqual(t, "", o.Value())

	testutil.ItsEqual(t, 3, g.recomputeHeap.Len())
}

func Test_Stabilize_printsErrors(t *testing.T) {
	g := New()

	v0 := Var("hello")
	gonnaPanic := MapContext(v0, func(_ context.Context, _ string) (string, error) {
		return "", fmt.Errorf("this is only a test")
	})
	_ = Observe(g, gonnaPanic)

	ctx := context.Background()
	outBuf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)
	ctx = WithTracingOutputs(ctx, outBuf, errBuf)
	err := g.Stabilize(ctx)
	testutil.ItsNotNil(t, err)
	testutil.ItsNotEqual(t, 0, len(outBuf.String()))
	testutil.ItsNotEqual(t, 0, len(errBuf.String()))
}
