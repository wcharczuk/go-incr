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
	m0 := Map2[string, string](v0, v1, func(a, b string) string {
		return a + " " + b
	})

	err := Stabilize(ctx, m0)
	ItsNil(t, err)

	ItsEqual(t, 0, v0.n.setAt)
	ItsEqual(t, 0, v0.n.changedAt)
	ItsEqual(t, 0, v1.n.setAt)
	ItsEqual(t, 0, v1.n.changedAt)
	ItsEqual(t, 0, m0.(*map2Node[string, string, string]).n.changedAt)
	ItsEqual(t, 0, v0.n.recomputedAt)
	ItsEqual(t, 0, v1.n.recomputedAt)
	ItsEqual(t, 0, m0.(*map2Node[string, string, string]).n.recomputedAt)

	ItsEqual(t, "foo bar", m0.Value())

	v0.Set("not foo")
	ItsEqual(t, 1, v0.n.setAt)
	ItsEqual(t, 0, v1.n.setAt)

	err = Stabilize(ctx, m0)
	ItsNil(t, err)

	ItsEqual(t, 1, v0.n.changedAt)
	ItsEqual(t, 0, v1.n.changedAt)
	ItsEqual(t, 1, m0.(*map2Node[string, string, string]).n.changedAt)

	ItsEqual(t, 1, v0.n.recomputedAt)
	ItsEqual(t, 0, v1.n.recomputedAt)
	ItsEqual(t, 1, m0.(*map2Node[string, string, string]).n.recomputedAt)

	ItsEqual(t, "not foo bar", m0.Value())
}

func Test_Stabilize_updateHandlers(t *testing.T) {
	ctx := testContext()

	v0 := Var("foo")
	v1 := Var("bar")
	m0 := Map2[string, string](v0, v1, func(a, b string) string {
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
	m0 := Map2[string, string](v0, v1, func(a, b string) string {
		return a + " " + b
	})
	r0 := Return("moo")
	m1 := Map2(r0, m0, func(a, b string) string {
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

func Test_Stabilize_verifyPartial(t *testing.T) {
	ctx := testContext()

	v0 := Var("foo")
	c0 := Return("bar")
	v1 := Var("moo")
	c1 := Return("baz")

	m0 := Map2[string](v0, c0, func(a, b string) string {
		return a + " " + b
	})
	co0 := Cutoff(m0, func(n, o string) bool {
		return len(n) == len(o)
	})
	m1 := Map2[string](v1, c1, func(a, b string) string {
		return a + " != " + b
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
	output := Map[[]Entry](
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

	b := Bind[bool](sw, func(swv bool) Incr[string] {
		if swv {
			return i0
		}
		return i1
	})

	err := Stabilize(ctx, b)
	ItsNil(t, err)

	ItsEqual(t, "bar", b.Value())

	sw.Set(true)
	err = Stabilize(ctx, b)
	ItsNil(t, err)

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

func Test_Stabilize_map3(t *testing.T) {
	ctx := testContext()

	c0 := Return(1)
	c1 := Return(2)
	c2 := Return(3)
	m3 := Map3(c0, c1, c2, func(a, b, c int) int {
		return a + b + c
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
