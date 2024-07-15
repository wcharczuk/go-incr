package incr

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_ParallelStabilize(t *testing.T) {
	ctx := testContext()
	g := New()

	v0 := Var(g, "foo")
	v1 := Var(g, "bar")
	m0 := Map2(g, v0, v1, func(a, b string) string {
		return a + " " + b
	})

	_ = MustObserve(g, m0)

	err := g.ParallelStabilize(ctx)
	testutil.Nil(t, err)

	testutil.Equal(t, 0, v0.Node().setAt)
	testutil.Equal(t, 0, v0.Node().changedAt)
	testutil.Equal(t, 0, v1.Node().setAt)
	testutil.Equal(t, 0, v1.Node().changedAt)
	testutil.Equal(t, 1, m0.Node().changedAt)
	testutil.Equal(t, 0, v0.Node().recomputedAt)
	testutil.Equal(t, 0, v1.Node().recomputedAt)
	testutil.Equal(t, 1, m0.Node().recomputedAt)

	testutil.Equal(t, "foo bar", m0.Value())

	v0.Set("not foo")
	testutil.Equal(t, 2, v0.Node().setAt)
	testutil.Equal(t, 0, v1.Node().setAt)

	err = g.ParallelStabilize(ctx)
	testutil.Nil(t, err)

	testutil.Equal(t, 2, v0.Node().changedAt)
	testutil.Equal(t, 0, v1.Node().changedAt)
	testutil.Equal(t, 2, m0.Node().changedAt)

	testutil.Equal(t, 2, v0.Node().recomputedAt)
	testutil.Equal(t, 0, v1.Node().recomputedAt)
	testutil.Equal(t, 2, m0.Node().recomputedAt)

	testutil.Equal(t, "not foo bar", m0.Value())
}

func Test_ParallelStabilize_alreadyStabilizing(t *testing.T) {
	ctx := testContext()

	graph := New()
	graph.status = StatusStabilizing

	err := graph.ParallelStabilize(ctx)
	testutil.NotNil(t, err)
}

func Test_ParallelStabilize_jsDocs(t *testing.T) {
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

	err := g.ParallelStabilize(ctx)
	testutil.Nil(t, err)
	testutil.Equal(t, 2, len(output.Value()))

	data = append(data, Entry{
		"5", now.Add(5 * time.Second),
	})
	err = g.ParallelStabilize(ctx)
	testutil.Nil(t, err)
	testutil.Equal(t, 2, len(output.Value()))

	i.Set(data)
	err = g.ParallelStabilize(ctx)
	testutil.Nil(t, err)
	testutil.Equal(t, 3, len(output.Value()))
}

func Test_ParallelStabilize_error(t *testing.T) {
	ctx := testContext()
	g := New()

	v0 := Var(g, "hello")
	m0 := Map(g, v0, ident)
	m1 := Map(g, m0, ident)

	f0 := Func(g, func(ctx context.Context) (string, error) {
		return "", fmt.Errorf("this is only a test")
	})

	_ = MustObserve(g, f0)
	_ = MustObserve(g, m1)

	testutil.Equal(t, true, g.recomputeHeap.has(m1))
	testutil.Equal(t, true, g.recomputeHeap.has(f0))

	err := g.ParallelStabilize(ctx)
	testutil.NotNil(t, err)

	testutil.Equal(t, false, g.recomputeHeap.has(m1), "we should clear the recompute heap on error")
	testutil.Equal(t, false, g.recomputeHeap.has(f0))
}

func Test_ParallelStabilize_Always(t *testing.T) {
	ctx := testContext()
	g := New()

	v := Var(g, "foo")
	m0 := Map(g, v, ident)
	a := Always(g, m0)
	m1 := Map(g, a, ident)

	var updates int
	m1.Node().OnUpdate(func(_ context.Context) {
		updates++
	})

	o := MustObserve(g, m1)

	_ = g.ParallelStabilize(ctx)

	testutil.Equal(t, "foo", o.Value())
	testutil.Equal(t, 1, updates)

	_ = g.ParallelStabilize(ctx)

	testutil.Equal(t, "foo", o.Value())
	testutil.Equal(t, 2, updates)

	v.Set("bar")

	_ = g.ParallelStabilize(ctx)

	testutil.Equal(t, "bar", o.Value())
	testutil.Equal(t, 3, updates)
}

func Test_ParallelStabilize_always_cutoff(t *testing.T) {
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

	err := g.ParallelStabilize(ctx)
	testutil.Nil(t, err)
	testutil.Equal(t, "test-1", o.Value())

	err = g.ParallelStabilize(ctx)
	testutil.Nil(t, err)
	testutil.Equal(t, "test-1", o.Value())

	modtime = 2

	err = g.ParallelStabilize(ctx)
	testutil.Nil(t, err)
	testutil.Equal(t, "test-2", o.Value())

	err = g.ParallelStabilize(ctx)
	testutil.Nil(t, err)
	testutil.Equal(t, "test-2", o.Value())
}

func Test_ParallelStabilize_always_cutoff_error(t *testing.T) {
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

	err := g.ParallelStabilize(ctx)
	testutil.NotNil(t, err)
	testutil.Equal(t, "", o.Value())

	testutil.Equal(t, 1, g.recomputeHeap.len(), "we should clear the recompute heap on error")
}

func Test_ParallelStabilize_printsErrors(t *testing.T) {
	ctx := context.Background()
	g := New()

	outBuf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)
	ctx = WithTracingOutputs(ctx, outBuf, errBuf)

	v0 := Var(g, "hello")
	gonnaPanic := MapContext(g, v0, func(_ context.Context, _ string) (string, error) {
		return "", fmt.Errorf("this is only a test")
	})
	_ = MustObserve(g, gonnaPanic)

	err := g.ParallelStabilize(ctx)
	testutil.NotNil(t, err)
	testutil.NotEqual(t, 0, len(outBuf.String()))
	testutil.NotEqual(t, 0, len(errBuf.String()))
	testutil.Equal(t, true, strings.Contains(errBuf.String(), "this is only a test"))
}

func Test_ParallelStabilize_preRequisite_heightsAreParallel(t *testing.T) {
	g := New()

	v0 := Var(g, "0")
	v1 := Var(g, "1")
	v2 := Var(g, "2")
	v3 := Var(g, "3")
	v4 := Var(g, "4")
	v5 := Var(g, "5")
	v6 := Var(g, "6")
	v7 := Var(g, "7")

	m00 := Map2(g, v0, v1, concat)
	m01 := Map2(g, v2, v3, concat)
	m02 := Map2(g, v4, v5, concat)
	m03 := Map2(g, v6, v7, concat)

	m10 := Map2(g, m00, m01, concat)
	m11 := Map2(g, m02, m03, concat)

	m20 := Map2(g, m10, m11, concat)

	o, err := Observe(g, m20)
	testutil.NoError(t, err)

	testutil.Equal(t, 0, v0.Node().height)
	testutil.Equal(t, HeightUnset, v0.Node().heightInRecomputeHeap)
	testutil.Equal(t, 0, v1.Node().height)
	testutil.Equal(t, HeightUnset, v1.Node().heightInRecomputeHeap)
	testutil.Equal(t, 0, v2.Node().height)
	testutil.Equal(t, HeightUnset, v2.Node().heightInRecomputeHeap)
	testutil.Equal(t, 0, v3.Node().height)
	testutil.Equal(t, HeightUnset, v3.Node().heightInRecomputeHeap)
	testutil.Equal(t, 0, v4.Node().height)
	testutil.Equal(t, HeightUnset, v4.Node().heightInRecomputeHeap)
	testutil.Equal(t, 0, v5.Node().height)
	testutil.Equal(t, HeightUnset, v5.Node().heightInRecomputeHeap)
	testutil.Equal(t, 0, v6.Node().height)
	testutil.Equal(t, HeightUnset, v6.Node().heightInRecomputeHeap)
	testutil.Equal(t, 0, v7.Node().height)
	testutil.Equal(t, HeightUnset, v7.Node().heightInRecomputeHeap)

	testutil.Equal(t, 1, m00.Node().height)
	testutil.Equal(t, 1, m00.Node().heightInRecomputeHeap)
	testutil.Equal(t, 1, m01.Node().height)
	testutil.Equal(t, 1, m01.Node().heightInRecomputeHeap)
	testutil.Equal(t, 1, m02.Node().height)
	testutil.Equal(t, 1, m02.Node().heightInRecomputeHeap)
	testutil.Equal(t, 1, m03.Node().height)
	testutil.Equal(t, 1, m03.Node().heightInRecomputeHeap)

	testutil.Equal(t, 2, m10.Node().height)
	testutil.Equal(t, 2, m10.Node().heightInRecomputeHeap)
	testutil.Equal(t, 2, m11.Node().height)
	testutil.Equal(t, 2, m11.Node().heightInRecomputeHeap)

	testutil.Equal(t, 3, m20.Node().height)
	testutil.Equal(t, 3, m20.Node().heightInRecomputeHeap)

	_ = g.ParallelStabilize(testContext())
	testutil.NotEqual(t, "", o.Value())
}

func Test_ParallelStabilize_alwaysInRecomputeHeapOnError(t *testing.T) {
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

	err := g.ParallelStabilize(testContext())
	testutil.Error(t, err)
	testutil.Equal(t, "this is only a test", err.Error())
}
