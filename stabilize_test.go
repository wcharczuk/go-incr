package incr

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"
)

func Test_Stabilize(t *testing.T) {
	ctx := testContext()

	v0 := Var("foo")
	v1 := Var("bar")
	m0 := Apply2(v0.Read(), v1.Read(), func(a, b string) string {
		return a + " " + b
	})

	err := Stabilize(ctx, m0)
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

	err = Stabilize(ctx, m0)
	ItsNil(t, err)

	ItsEqual(t, 2, v0.Node().changedAt)
	ItsEqual(t, 1, v1.Node().changedAt)
	ItsEqual(t, 2, m0.Node().changedAt)

	ItsEqual(t, 2, v0.Node().recomputedAt)
	ItsEqual(t, 1, v1.Node().recomputedAt)
	ItsEqual(t, 2, m0.Node().recomputedAt)

	ItsEqual(t, "not foo bar", m0.Value())
}

func Test_Stabilize_many(t *testing.T) {
	ctx := testContext()

	v0 := Var("foo")
	v1 := Var("bar")
	m0 := Apply2(v0.Read(), v1.Read(), func(a, b string) string {
		return a + " " + b
	})

	err := Stabilize(ctx, m0, v0, v1)
	ItsNil(t, err)
	ItsEqual(t, "foo bar", m0.Value())
}

func Test_Stabilize_error(t *testing.T) {
	ctx := testContext()

	m0 := Func(func(_ context.Context) (string, error) {
		return "", fmt.Errorf("this is just a test")
	})
	err := Stabilize(ctx, m0)
	ItsNotNil(t, err)
	ItsEqual(t, "this is just a test", err.Error())
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

	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		if err := Stabilize(ctx, m0); err != nil {
			errs <- err
		}
	}()
	go func() {
		defer wg.Done()
		if err := Stabilize(ctx, m0); err != nil {
			errs <- err
		}
	}()
	err := <-errs
	ItsEqual(t, errAlreadyStabilizing, err)
	close(hold)
	wg.Wait()
	ItsEqual(t, "ok!", m0.Value())
}

func Test_Stabilize_updateHandlers(t *testing.T) {
	ctx := testContext()

	v0 := Var("foo")
	v1 := Var("bar")
	m0 := Apply2(v0.Read(), v1.Read(), func(a, b string) string {
		return a + " " + b
	})

	var updates int
	m0.Node().OnUpdate(func(_ context.Context) {
		updates++
	})

	err := Stabilize(ctx, m0)
	ItsNil(t, err)
	ItsEqual(t, 1, updates)

	v0.Set("not foo")
	err = Stabilize(ctx, m0)
	ItsNil(t, err)
	ItsEqual(t, 2, updates)
}

func Test_Stabilize_unevenHeights(t *testing.T) {
	ctx := testContext()

	v0 := Var("foo")
	v1 := Var("bar")
	m0 := Apply2[string, string](v0, v1, func(a, b string) string {
		return a + " " + b
	})
	r0 := Return("moo")
	m1 := Apply2(r0, m0, func(a, b string) string {
		return a + " != " + b
	})

	err := Stabilize(ctx, m1)
	ItsNil(t, err)
	ItsEqual(t, "moo != foo bar", m1.Value())

	v0.Set("not foo")
	err = Stabilize(ctx, m1)
	ItsNil(t, err)
	ItsEqual(t, "moo != not foo bar", m1.Value())
}

func Test_Stabilize_chain(t *testing.T) {
	ctx := testContext()

	v0 := Var(".")

	var maps []Incr[string]
	var previous Incr[string] = v0.Read()
	for x := 0; x < 100; x++ {
		m := Apply(previous, func(v0 string) string {
			return v0 + "."
		})
		maps = append(maps, m)
		previous = m
	}

	err := Stabilize(ctx, maps[len(maps)-1])
	ItsNil(t, err)
	ItsEqual(t, strings.Repeat(".", 101), maps[len(maps)-1].Value())

	ItsEqual(t, 101, v0.Node().g.numNodes)
	ItsEqual(t, 101, v0.Node().g.numNodesChanged)
	ItsEqual(t, 101, v0.Node().g.numNodesRecomputed)
}

func Test_Stabilize_setDuringStabilization(t *testing.T) {
	ctx := testContext()
	v0 := Var("foo")

	called := make(chan struct{})
	wait := make(chan struct{})
	m0 := Apply(v0.Read(), func(v string) string {
		close(called)
		<-wait
		return v
	})

	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = Stabilize(ctx, m0)
	}()
	<-called
	v0.Set("not-foo")
	ItsEqual(t, "foo", v0.Value())
	close(wait)
	<-done
	ItsEqual(t, "not-foo", v0.Value())
	ItsEqual(t, m0.Node().g.stabilizationNum, v0.Node().setAt)
	ItsEqual(t, 1, len(m0.Node().g.recomputeHeap.lookup))
}

func Test_Stabilize_onUpdate(t *testing.T) {
	ctx := testContext()

	var didCallUpdateHandler0, didCallUpdateHandler1 bool
	v0 := Var("hello")
	v1 := Var("world")
	m0 := Apply2(v0.Read(), v1.Read(), concat)
	m0.Node().OnUpdate(func(context.Context) {
		didCallUpdateHandler0 = true
	})
	m0.Node().OnUpdate(func(context.Context) {
		didCallUpdateHandler1 = true
	})

	err := Stabilize(ctx, m0)
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
	b := Apply(a.Read(), edge("b"))
	c := Apply(b, edge("c"))
	d := Apply(c, edge("d"))
	f := Apply(a.Read(), edge("f"))
	e := Apply(f, edge("e"))

	z := Apply2(d, e, func(v0, v1 string) string {
		return v0 + "+" + v1 + "->z"
	})

	err := Stabilize(ctx, z)
	ItsNil(t, err)
	ItsEqual(t, 1, z.Node().numRecomputes)
	ItsEqual(t, "a->b->c->d+a->f->e->z", z.Value())

	a.Set("!a")

	err = Stabilize(ctx, z)
	ItsNil(t, err)
	ItsEqual(t, "!a->b->c->d+!a->f->e->z", z.Value())
	ItsEqual(t, 2, z.Node().numRecomputes)
}

func Test_Stabilize_doubleVarSet_singleUpdate(t *testing.T) {
	ctx := testContext()

	a := Var("a")
	b := Var("b")
	m := Apply2(a.Read(), b.Read(), func(v0, v1 string) string {
		return v0 + " " + v1
	})

	_ = Stabilize(ctx, m)
	ItsEqual(t, "a b", m.Value())

	a.Set("aa")
	ItsEqual(t, 1, a.Node().g.recomputeHeap.Len())

	a.Set("aaa")
	ItsEqual(t, 1, a.Node().g.recomputeHeap.Len())

	_ = Stabilize(ctx, m)
	ItsEqual(t, "aaa b", m.Value())
}

func Test_Stabilize_verifyPartial(t *testing.T) {
	ctx := testContext()

	v0 := Var("foo")
	c0 := Return("bar")
	v1 := Var("moo")
	c1 := Return("baz")

	m0 := Apply2(v0.Read(), c0, func(a, b string) string {
		return a + " " + b
	})
	co0 := Cutoff(m0, func(n, o string) bool {
		return len(n) == len(o)
	})
	m1 := Apply2(v1.Read(), c1, func(a, b string) string {
		return a + " != " + b
	})
	co1 := Cutoff(m1, func(n, o string) bool {
		return len(n) == len(o)
	})

	sw := Var(true)
	mi := ApplyIf(co0, co1, sw)

	err := Stabilize(ctx, mi)
	ItsNil(t, err)
	ItsEqual(t, "foo bar", mi.Value())

	v0.Set("Foo")

	err = Stabilize(ctx, mi)
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
	output := Apply(
		i.Read(),
		func(entries []Entry) (output []string) {
			for _, e := range entries {
				if e.Time.Sub(now) > 2*time.Second {
					output = append(output, e.Entry)
				}
			}
			return
		},
	)

	err := Stabilize(
		ctx,
		output,
	)
	ItsNil(t, err)
	ItsEqual(t, 2, len(output.Value()))

	data = append(data, Entry{
		"5", now.Add(5 * time.Second),
	})
	err = Stabilize(
		ctx,
		output,
	)
	ItsNil(t, err)
	ItsEqual(t, 2, len(output.Value()))

	i.Set(data)
	err = Stabilize(
		context.Background(),
		output,
	)
	ItsNil(t, err)
	ItsEqual(t, 3, len(output.Value()))
}

func Test_Stabilize_bind(t *testing.T) {
	ctx := testContext()

	sw := Var(false)
	i0 := Return("foo")
	i1 := Return("bar")

	b := Bind(sw.Read(), func(_ context.Context, swv bool) (Incr[string], error) {
		if swv {
			return i0, nil
		}
		return i1, nil
	})

	err := Stabilize(ctx, b)
	ItsNil(t, err)

	ItsNil(t, i0.Node().g, "i0 should not be in the graph after the first stabilization")
	ItsNotNil(t, i1.Node().g, "i1 should be in the graph after the first stabilization")

	ItsEqual(t, "bar", b.Value())

	sw.Set(true)
	err = Stabilize(ctx, b)
	ItsNil(t, err)

	ItsNil(t, i1.Node().g, "i0 should be in the graph after the second stabilization")
	ItsNotNil(t, i0.Node().g, "i1 should not be in the graph after the second stabilization")

	ItsEqual(t, "foo", b.Value())
}

func Test_Stabilize_bind2(t *testing.T) {
	ctx := testContext()

	sw0 := Var(false)
	sw1 := Var(false)
	i0 := Return("foo")
	i1 := Return("bar")

	b := Bind2(sw0.Read(), sw1.Read(), func(_ context.Context, swv0, swv1 bool) (Incr[string], error) {
		if swv0 && swv1 {
			return i0, nil
		}
		return i1, nil
	})

	err := Stabilize(ctx, b)
	ItsNil(t, err)

	ItsNil(t, i0.Node().g, "i0 should not be in the graph after the first stabilization")
	ItsNotNil(t, i1.Node().g, "i1 should be in the graph after the first stabilization")

	ItsEqual(t, "bar", b.Value())

	sw0.Set(true)
	err = Stabilize(ctx, b)
	ItsNil(t, err)

	ItsNil(t, i0.Node().g, "i0 should not be in the graph after the second stabilization")
	ItsNotNil(t, i1.Node().g, "i1 should be in the graph after the second stabilization")

	ItsEqual(t, "bar", b.Value())

	sw1.Set(true)
	err = Stabilize(ctx, b)
	ItsNil(t, err)

	ItsNil(t, i1.Node().g, "i0 should be in the graph after the third stabilization")
	ItsNotNil(t, i0.Node().g, "i1 should not be in the graph after the third stabilization")

	ItsEqual(t, "foo", b.Value())
}

func Test_Stabilize_bind3(t *testing.T) {
	ctx := testContext()

	sw0 := Var(false)
	sw1 := Var(false)
	sw2 := Var(false)

	i0 := Return("foo")
	i1 := Return("bar")

	b := Bind3(sw0.Read(), sw1.Read(), sw2.Read(), func(_ context.Context, swv0, swv1, swv2 bool) (Incr[string], error) {
		if swv0 && swv1 && swv2 {
			return i0, nil
		}
		return i1, nil
	})

	err := Stabilize(ctx, b)
	ItsNil(t, err)
	ItsNil(t, i0.Node().g, "i0 should not be in the graph after the first stabilization")
	ItsNotNil(t, i1.Node().g, "i1 should be in the graph after the first stabilization")
	ItsEqual(t, "bar", b.Value())

	sw0.Set(true)

	err = Stabilize(ctx, b)
	ItsNil(t, err)
	ItsNil(t, i0.Node().g, "i0 should not be in the graph after the second stabilization")
	ItsNotNil(t, i1.Node().g, "i1 should be in the graph after the second stabilization")
	ItsEqual(t, "bar", b.Value())

	sw1.Set(true)

	err = Stabilize(ctx, b)
	ItsNil(t, err)
	ItsNil(t, i0.Node().g, "i0 should not be in the graph after the third stabilization")
	ItsNotNil(t, i1.Node().g, "i1 should be in the graph after the third stabilization")
	ItsEqual(t, "bar", b.Value())
	ItsNil(t, err)

	sw2.Set(true)

	err = Stabilize(ctx, b)
	ItsNil(t, err)
	ItsNil(t, i1.Node().g, "i0 should be in the graph after the fourth stabilization")
	ItsNotNil(t, i0.Node().g, "i1 should not be in the graph after the fourth stabilization")
	ItsEqual(t, "foo", b.Value())
}

func Test_Stabilize_bindIf(t *testing.T) {
	ctx := testContext()

	sw := Var(false)
	i0 := Return("foo")
	i1 := Return("bar")

	b := BindIf(sw, func(ctx context.Context, swv bool) (Incr[string], error) {
		itsBlueDye(ctx, t)
		if swv {
			return i0, nil
		}
		return i1, nil
	})

	err := Stabilize(ctx, b)
	ItsNil(t, err)

	ItsNil(t, i0.Node().g, "i0 should not be in the graph after the first stabilization")
	ItsNotNil(t, i1.Node().g, "i1 should be in the graph after the first stabilization")

	ItsEqual(t, "bar", b.Value())

	sw.Set(true)
	err = Stabilize(ctx, b)
	ItsNil(t, err)

	ItsNil(t, i1.Node().g, "i0 should be in the graph after the third stabilization")
	ItsNotNil(t, i0.Node().g, "i1 should not be in the graph after the third stabilization")

	ItsEqual(t, "foo", b.Value())
}

func Test_Stabilize_cutoff(t *testing.T) {
	ctx := testContext()
	input := Var(3.14)

	cutoff := Cutoff[float64](
		input,
		epsilon(0.1),
	)

	output := Apply2(
		cutoff,
		Return(10.0),
		add[float64],
	)

	_ = Stabilize(
		ctx,
		output,
	)
	ItsEqual(t, 13.14, output.Value())
	ItsEqual(t, 3.14, cutoff.Value())

	input.Set(3.15)

	_ = Stabilize(
		ctx,
		output,
	)
	ItsEqual(t, 3.14, cutoff.Value())
	ItsEqual(t, 13.14, output.Value())

	input.Set(3.26) // differs by 0.11, which is > 0.1

	_ = Stabilize(
		ctx,
		output,
	)
	ItsEqual(t, 3.26, cutoff.Value())
	ItsEqual(t, 13.26, output.Value())

	_ = Stabilize(
		ctx,
		output,
	)
	ItsEqual(t, 13.26, output.Value())
}

func Test_Stabilize_watch(t *testing.T) {
	ctx := testContext()

	v0 := Var(1)
	v1 := Var(1)
	m0 := Apply2[int, int](v0, v1, add[int])
	w0 := Watch(m0)

	_ = Stabilize(ctx, w0)

	ItsEqual(t, 1, len(w0.Values()))
	ItsEqual(t, 2, w0.Values()[0])

	v0.Set(2)

	_ = Stabilize(ctx, w0)

	ItsEqual(t, 2, len(w0.Values()))
	ItsEqual(t, 2, w0.Values()[0])
	ItsEqual(t, 3, w0.Values()[1])
}

func Test_Stabilize_apply(t *testing.T) {
	ctx := testContext()

	c0 := Return(1)
	m := Apply(c0, func(a int) int {
		return a + 10
	})
	_ = Stabilize(ctx, m)
	ItsEqual(t, 11, m.Value())
}

func Test_Stabilize_applyContext(t *testing.T) {
	ctx := testContext()

	c0 := Return(1)
	m := ApplyContext(c0, func(ictx context.Context, a int) (int, error) {
		itsBlueDye(ictx, t)
		return a + 10, nil
	})
	_ = Stabilize(ctx, m)
	ItsEqual(t, 11, m.Value())
}

func Test_Stabilize_apply2(t *testing.T) {
	ctx := testContext()

	c0 := Return(1)
	c1 := Return(2)
	m2 := Apply2(c0, c1, func(a, b int) int {
		return a + b
	})
	_ = Stabilize(ctx, m2)
	ItsEqual(t, 3, m2.Value())
}

func Test_Stabilize_apply2Context(t *testing.T) {
	ctx := testContext()

	c0 := Return(1)
	c1 := Return(2)
	m2 := Apply2Context(c0, c1, func(ictx context.Context, a, b int) (int, error) {
		itsBlueDye(ctx, t)
		return a + b, nil
	})
	_ = Stabilize(ctx, m2)
	ItsEqual(t, 3, m2.Value())
}

func Test_Stabilize_apply3(t *testing.T) {
	ctx := testContext()

	c0 := Return(1)
	c1 := Return(2)
	c2 := Return(3)
	m3 := Apply3(c0, c1, c2, func(a, b, c int) int {
		return a + b + c
	})

	_ = Stabilize(ctx, m3)
	ItsEqual(t, 6, m3.Value())
}

func Test_Stabilize_apply3Context(t *testing.T) {
	ctx := testContext()

	c0 := Return(1)
	c1 := Return(2)
	c2 := Return(3)
	m3 := Apply3Context(c0, c1, c2, func(ictx context.Context, a, b, c int) (int, error) {
		itsBlueDye(ictx, t)
		return a + b + c, nil
	})

	_ = Stabilize(ctx, m3)
	ItsEqual(t, 6, m3.Value())
}

func Test_Stabilize_applyIf(t *testing.T) {
	ctx := testContext()

	c0 := Return(1)
	c1 := Return(2)
	v0 := Var(false)
	mi0 := ApplyIf(c0, c1, v0)

	_ = Stabilize(ctx, mi0)
	ItsEqual(t, 2, mi0.Value())

	v0.Set(true)

	_ = Stabilize(ctx, mi0)
	ItsEqual(t, 1, mi0.Value())

	_ = Stabilize(ctx, mi0)
	ItsEqual(t, 1, mi0.Value())
}

func Test_Stabilize_applyN(t *testing.T) {
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
	mn := ApplyN(func(_ context.Context, inputs ...int) (int, error) {
		return sum(inputs...), nil
	}, c0, c1, c2)

	_ = Stabilize(ctx, mn)
	ItsEqual(t, 6, mn.Value())
}

func Test_Stabilize_func(t *testing.T) {
	ctx := testContext()

	value := "hello"
	f := Func(func(ictx context.Context) (string, error) {
		itsBlueDye(ictx, t)
		return value, nil
	})
	m := ApplyContext(f, func(ictx context.Context, v string) (string, error) {
		itsBlueDye(ctx, t)
		return v + " world!", nil
	})
	_ = Stabilize(ctx, m)
	ItsEqual(t, "hello world!", m.Value())

	value = "not hello"

	_ = Stabilize(ctx, m)
	ItsEqual(t, "hello world!", m.Value())

	// mark the func node as stale
	// not sure a better way to do this automatically?
	SetStale(f)

	_ = Stabilize(ctx, m)
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
	_ = Stabilize(ctx, mf)
	ItsEqual(t, 21, mf.Value())
}

func Test_Stabilize_diffMapByKeysAdded(t *testing.T) {
	ctx := testContext()

	m := map[string]int{
		"one":   1,
		"two":   2,
		"three": 3,
		"four":  4,
		"five":  5,
		"six":   6,
	}

	mv := Var(m)
	mda := DiffMapByKeysAdded(mv.Read())
	mf := FoldMap(mda, 0, func(key string, val, accum int) int {
		return accum + val
	})
	_ = Stabilize(ctx, mf)
	ItsEqual(t, 21, mf.Value())

	m["seven"] = 7
	m["eight"] = 8

	mv.Set(m)

	_ = Stabilize(ctx, mf)
	ItsEqual(t, 36, mf.Value())

	m["nine"] = 9

	mv.Set(m)

	_ = Stabilize(ctx, mf)
	ItsEqual(t, 45, mf.Value())
}

func Test_Stabilize_diffMapByKeysRemoved(t *testing.T) {
	ctx := testContext()

	m := map[string]int{
		"one":   1,
		"two":   2,
		"three": 3,
		"four":  4,
		"five":  5,
		"six":   6,
	}

	mv := Var(m)
	mdr := DiffMapByKeysRemoved(mv.Read())
	mf := FoldMap(mdr, 0, func(key string, val, accum int) int {
		return accum + val
	})
	_ = Stabilize(ctx, mf)
	ItsEqual(t, 0, mf.Value())

	delete(m, "two")
	delete(m, "five")

	mv.Set(m)

	_ = Stabilize(ctx, mf)
	ItsEqual(t, 7, mf.Value())
}

func Test_Stabilize_diffMapByKeys(t *testing.T) {
	ctx := testContext()

	m := map[string]int{
		"one":   1,
		"two":   2,
		"three": 3,
		"four":  4,
		"five":  5,
		"six":   6,
	}

	mv := Var(m)
	mda, mdr := DiffMapByKeys(mv.Read())
	mfa := FoldMap(mda, 0, func(key string, val, accum int) int {
		return accum + val
	})
	mfr := FoldMap(mdr, 0, func(key string, val, accum int) int {
		return accum + val
	})
	_ = Stabilize(ctx, mfa, mfr)
	ItsEqual(t, 21, mfa.Value())
	ItsEqual(t, 0, mfr.Value())

	delete(m, "two")
	delete(m, "five")
	m["seven"] = 7
	m["eight"] = 8

	mv.Set(m)

	_ = Stabilize(ctx, mfa, mfr)
	ItsEqual(t, 36, mfa.Value())
	ItsEqual(t, 7, mfr.Value())
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
	_ = Stabilize(ctx, mf)
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
	_ = Stabilize(ctx, mf)
	ItsEqual(t, "654321", mf.Value())

	SetStale(mf)

	_ = Stabilize(ctx, mf)
	ItsEqual(t, "654321654321", mf.Value())
}

func Test_Stabilize_diffSlice(t *testing.T) {
	ctx := testContext()

	m := []int{
		1,
		2,
		3,
		4,
		5,
		6,
	}
	mv := Var(m)
	mf := FoldLeft(DiffSliceByIndicesAdded(mv.Read()), "", func(accum string, val int) string {
		return accum + fmt.Sprint(val)
	})

	_ = Stabilize(ctx, mf)
	ItsEqual(t, "123456", mf.Value())

	m = append(m, 7, 8, 9)
	mv.Set(m)

	_ = Stabilize(ctx, mf)
	ItsEqual(t, "123456789", mf.Value())
}
