package incr

import (
	"context"
	"testing"
	"time"
)

func Test_Stabilize(t *testing.T) {
	v0 := Var("foo")
	v1 := Var("bar")
	m0 := Map2[string, string](v0, v1, func(a, b string) (string, error) {
		return a + " " + b, nil
	})

	err := Stabilize(context.TODO(), m0)
	ItsNil(t, err)

	ItsEqual(t, 1, v0.n.initializedAt)
	ItsEqual(t, 1, v1.n.initializedAt)
	ItsEqual(t, 1, m0.(*map2Node[string, string, string]).n.initializedAt)

	ItsEqual(t, 1, v0.n.changedAt)
	ItsEqual(t, 1, v1.n.changedAt)
	ItsEqual(t, 1, m0.(*map2Node[string, string, string]).n.changedAt)

	ItsEqual(t, 1, v0.n.recomputedAt)
	ItsEqual(t, 1, v1.n.recomputedAt)
	ItsEqual(t, 1, m0.(*map2Node[string, string, string]).n.recomputedAt)

	ItsEqual(t, "foo bar", m0.Value())

	v0.Set("not foo")

	err = Stabilize(context.TODO(), m0)
	ItsNil(t, err)

	ItsEqual(t, 1, v0.n.initializedAt)
	ItsEqual(t, 1, v1.n.initializedAt)
	ItsEqual(t, 1, m0.(*map2Node[string, string, string]).n.initializedAt)

	ItsEqual(t, 2, v0.n.changedAt)
	ItsEqual(t, 1, v1.n.changedAt)
	ItsEqual(t, 2, m0.(*map2Node[string, string, string]).n.changedAt)

	ItsEqual(t, 2, v0.n.recomputedAt)
	ItsEqual(t, 1, v1.n.recomputedAt)
	ItsEqual(t, 2, m0.(*map2Node[string, string, string]).n.recomputedAt)

	ItsEqual(t, "not foo bar", m0.Value())
}

func Test_Stabilize_unevenHeights(t *testing.T) {
	v0 := Var("foo")
	v1 := Var("bar")
	m0 := Map2[string, string](v0, v1, func(a, b string) (string, error) {
		return a + " " + b, nil
	})
	r0 := Return("moo")
	m1 := Map2(r0, m0, func(a, b string) (string, error) {
		return a + " != " + b, nil
	})

	err := Stabilize(context.TODO(), m1)
	ItsNil(t, err)
	ItsEqual(t, "moo != foo bar", m1.Value())

	v0.Set("not foo")
	err = Stabilize(context.TODO(), m1)
	ItsNil(t, err)
	ItsEqual(t, "moo != not foo bar", m1.Value())
}

func Test_Stabilize_jsDocs(t *testing.T) {
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
		func(entries []Entry) (output []string, err error) {
			for _, e := range entries {
				if e.Time.Sub(now) > 2*time.Second {
					output = append(output, e.Entry)
				}
			}
			return
		},
	)

	err := Stabilize(
		context.Background(),
		output,
	)
	ItsNil(t, err)
	ItsEqual(t, 2, len(output.Value()))

	data = append(data, Entry{
		"5", now.Add(5 * time.Second),
	})
	err = Stabilize(
		context.Background(),
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
