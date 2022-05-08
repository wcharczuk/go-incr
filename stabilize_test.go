package incr

import (
	"context"
	"testing"
	"time"
)

func Test_Stabilize(t *testing.T) {
	ctx := testContext()

	v0 := Var("foo")
	v1 := Var("bar")
	m0 := Map2(v0.Read(), v1.Read(), func(_ context.Context, a, b string) (string, error) {
		return a + " " + b, nil
	})

	err := Stabilize(ctx, m0)
	ItsNil(t, err)

	ItsEqual(t, 0, v0.Node().setAt)
	ItsEqual(t, 0, v0.Node().changedAt)
	ItsEqual(t, 0, v1.Node().setAt)
	ItsEqual(t, 0, v1.Node().changedAt)
	ItsEqual(t, 0, m0.Node().changedAt)
	ItsEqual(t, 0, v0.Node().recomputedAt)
	ItsEqual(t, 0, v1.Node().recomputedAt)
	ItsEqual(t, 0, m0.Node().recomputedAt)

	ItsEqual(t, "foo bar", m0.Value())

	v0.Set("not foo")
	ItsEqual(t, 1, v0.Node().setAt)
	ItsEqual(t, 0, v1.Node().setAt)

	err = Stabilize(ctx, m0)
	ItsNil(t, err)

	ItsEqual(t, 1, v0.Node().changedAt)
	ItsEqual(t, 0, v1.Node().changedAt)
	ItsEqual(t, 1, m0.Node().changedAt)

	ItsEqual(t, 1, v0.Node().recomputedAt)
	ItsEqual(t, 0, v1.Node().recomputedAt)
	ItsEqual(t, 1, m0.Node().recomputedAt)

	ItsEqual(t, "not foo bar", m0.Value())
}

func Test_Stabilize_updateHandlers(t *testing.T) {
	ctx := testContext()

	v0 := Var("foo")
	v1 := Var("bar")
	m0 := Map2(v0.Read(), v1.Read(), func(_ context.Context, a, b string) (string, error) {
		return a + " " + b, nil
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
	m0 := Map2[string, string](v0, v1, func(_ context.Context, a, b string) (string, error) {
		return a + " " + b, nil
	})
	r0 := Return("moo")
	m1 := Map2(r0, m0, func(_ context.Context, a, b string) (string, error) {
		return a + " != " + b, nil
	})

	err := Stabilize(ctx, m1)
	ItsNil(t, err)
	ItsEqual(t, "moo != foo bar", m1.Value())

	v0.Set("not foo")
	err = Stabilize(ctx, m1)
	ItsNil(t, err)
	ItsEqual(t, "moo != not foo bar", m1.Value())
}

func Test_Stabilize_recombinant_singleUpdate(t *testing.T) {
	ctx := testContext()

	// a -> b -> c -> d -> z
	//   -> f -> e -> [z]
	// assert that [z] updates (1) time if we change [a]

	edge := func(l string) func(context.Context, string) (string, error) {
		return func(_ context.Context, v string) (string, error) {
			return v + "->" + l, nil
		}
	}

	a := Var("a")
	b := Map(a.Read(), edge("b"))
	c := Map(b, edge("c"))
	d := Map(c, edge("d"))
	f := Map(a.Read(), edge("f"))
	e := Map(f, edge("e"))

	z := Map2(d, e, func(_ context.Context, v0, v1 string) (string, error) {
		return v0 + "+" + v1 + "->z", nil
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

func Test_Stabilize_verifyPartial(t *testing.T) {
	ctx := testContext()

	v0 := Var("foo")
	c0 := Return("bar")
	v1 := Var("moo")
	c1 := Return("baz")

	m0 := Map2(v0.Read(), c0, func(_ context.Context, a, b string) (string, error) {
		return a + " " + b, nil
	})
	co0 := Cutoff(m0, func(n, o string) bool {
		return len(n) == len(o)
	})
	m1 := Map2(v1.Read(), c1, func(_ context.Context, a, b string) (string, error) {
		return a + " != " + b, nil
	})
	co1 := Cutoff(m1, func(n, o string) bool {
		return len(n) == len(o)
	})

	sw := Var(true)
	mi := MapIf(co0, co1, sw)

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
	output := Map(
		i.Read(),
		func(_ context.Context, entries []Entry) (output []string, err error) {
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

	ItsNil(t, i0.Node().gs, "i0 should not be in the graph after the first stabilization")
	ItsNotNil(t, i1.Node().gs, "i1 should be in the graph after the first stabilization")

	ItsEqual(t, "bar", b.Value())

	sw.Set(true)
	err = Stabilize(ctx, b)
	ItsNil(t, err)

	ItsNil(t, i1.Node().gs, "i0 should be in the graph after the second stabilization")
	ItsNotNil(t, i0.Node().gs, "i1 should not be in the graph after the second stabilization")

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

	ItsNil(t, i0.Node().gs, "i0 should not be in the graph after the first stabilization")
	ItsNotNil(t, i1.Node().gs, "i1 should be in the graph after the first stabilization")

	ItsEqual(t, "bar", b.Value())

	sw0.Set(true)
	err = Stabilize(ctx, b)
	ItsNil(t, err)

	ItsNil(t, i0.Node().gs, "i0 should not be in the graph after the second stabilization")
	ItsNotNil(t, i1.Node().gs, "i1 should be in the graph after the second stabilization")

	ItsEqual(t, "bar", b.Value())

	sw1.Set(true)
	err = Stabilize(ctx, b)
	ItsNil(t, err)

	ItsNil(t, i1.Node().gs, "i0 should be in the graph after the third stabilization")
	ItsNotNil(t, i0.Node().gs, "i1 should not be in the graph after the third stabilization")

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
	ItsNil(t, i0.Node().gs, "i0 should not be in the graph after the first stabilization")
	ItsNotNil(t, i1.Node().gs, "i1 should be in the graph after the first stabilization")
	ItsEqual(t, "bar", b.Value())

	sw0.Set(true)

	err = Stabilize(ctx, b)
	ItsNil(t, err)
	ItsNil(t, i0.Node().gs, "i0 should not be in the graph after the second stabilization")
	ItsNotNil(t, i1.Node().gs, "i1 should be in the graph after the second stabilization")
	ItsEqual(t, "bar", b.Value())

	sw1.Set(true)

	err = Stabilize(ctx, b)
	ItsNil(t, err)
	ItsNil(t, i0.Node().gs, "i0 should not be in the graph after the third stabilization")
	ItsNotNil(t, i1.Node().gs, "i1 should be in the graph after the third stabilization")
	ItsEqual(t, "bar", b.Value())
	ItsNil(t, err)

	sw2.Set(true)

	err = Stabilize(ctx, b)
	ItsNil(t, err)
	ItsNil(t, i1.Node().gs, "i0 should be in the graph after the fourth stabilization")
	ItsNotNil(t, i0.Node().gs, "i1 should not be in the graph after the fourth stabilization")
	ItsEqual(t, "foo", b.Value())
}

func Test_Stabilize_bindIf(t *testing.T) {
	ctx := testContext()

	sw := Var(false)
	i0 := Return("foo")
	i1 := Return("bar")

	b := BindIf(i0, i1, sw)

	err := Stabilize(ctx, b)
	ItsNil(t, err)

	ItsNil(t, i0.Node().gs, "i0 should not be in the graph after the first stabilization")
	ItsNotNil(t, i1.Node().gs, "i1 should be in the graph after the first stabilization")

	ItsEqual(t, "bar", b.Value())

	sw.Set(true)
	err = Stabilize(ctx, b)
	ItsNil(t, err)

	ItsNil(t, i1.Node().gs, "i0 should be in the graph after the third stabilization")
	ItsNotNil(t, i0.Node().gs, "i1 should not be in the graph after the third stabilization")

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
	m0 := Map2[int, int](v0, v1, add[int])
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

func Test_Stabilize_map(t *testing.T) {
	ctx := testContext()

	c0 := Return(1)
	m := Map(c0, func(_ context.Context, a int) (int, error) {
		return a + 10, nil
	})
	_ = Stabilize(ctx, m)
	ItsEqual(t, 11, m.Value())
}

func Test_Stabilize_map2(t *testing.T) {
	ctx := testContext()

	c0 := Return(1)
	c1 := Return(2)
	m2 := Map2(c0, c1, func(_ context.Context, a, b int) (int, error) {
		return a + b, nil
	})
	_ = Stabilize(ctx, m2)
	ItsEqual(t, 3, m2.Value())
}

func Test_Stabilize_map3(t *testing.T) {
	ctx := testContext()

	c0 := Return(1)
	c1 := Return(2)
	c2 := Return(3)
	m3 := Map3(c0, c1, c2, func(_ context.Context, a, b, c int) (int, error) {
		return a + b + c, nil
	})

	_ = Stabilize(ctx, m3)
	ItsEqual(t, 6, m3.Value())
}

func Test_Stabilize_mapIf(t *testing.T) {
	ctx := testContext()

	c0 := Return(1)
	c1 := Return(2)
	v0 := Var(false)
	mi0 := MapIf(c0, c1, v0)

	_ = Stabilize(ctx, mi0)
	ItsEqual(t, 2, mi0.Value())

	v0.Set(true)

	_ = Stabilize(ctx, mi0)
	ItsEqual(t, 1, mi0.Value())

	_ = Stabilize(ctx, mi0)
	ItsEqual(t, 1, mi0.Value())
}
