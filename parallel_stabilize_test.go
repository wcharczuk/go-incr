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

	_ = Observe(g, m0)

	err := g.ParallelStabilize(ctx)
	testutil.Nil(t, err)

	testutil.Equal(t, 0, v0.Node().setAt)
	testutil.Equal(t, 1, v0.Node().changedAt)
	testutil.Equal(t, 0, v1.Node().setAt)
	testutil.Equal(t, 1, v1.Node().changedAt)
	testutil.Equal(t, 1, m0.Node().changedAt)
	testutil.Equal(t, 1, v0.Node().recomputedAt)
	testutil.Equal(t, 1, v1.Node().recomputedAt)
	testutil.Equal(t, 1, m0.Node().recomputedAt)

	testutil.Equal(t, "foo bar", m0.Value())

	v0.Set("not foo")
	testutil.Equal(t, 2, v0.Node().setAt)
	testutil.Equal(t, 0, v1.Node().setAt)

	err = g.ParallelStabilize(ctx)
	testutil.Nil(t, err)

	testutil.Equal(t, 2, v0.Node().changedAt)
	testutil.Equal(t, 1, v1.Node().changedAt)
	testutil.Equal(t, 2, m0.Node().changedAt)

	testutil.Equal(t, 2, v0.Node().recomputedAt)
	testutil.Equal(t, 1, v1.Node().recomputedAt)
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

	_ = Observe(g, output)

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

	v0 := Var(g, "foo")
	m0 := MapContext(g, v0, func(ctx context.Context, a string) (string, error) {
		return "", fmt.Errorf("this is only a test")
	})

	_ = Observe(g, m0)

	err := g.ParallelStabilize(ctx)
	testutil.NotNil(t, err)
}

func Test_parallelBatch(t *testing.T) {
	pb := new(parallelBatch)

	var values = make(chan string, 1)
	pb.Go(func() error {
		values <- "hello"
		return nil
	})
	err := pb.Wait()
	testutil.Nil(t, err)
	got := <-values
	testutil.Equal(t, "hello", got)
}

func Test_parallelBatch_error(t *testing.T) {
	pb := new(parallelBatch)

	pb.Go(func() error {
		return fmt.Errorf("this is a test")
	})
	err := pb.Wait()
	testutil.NotNil(t, err)
}

func Test_parallelBatch_SetLimit(t *testing.T) {
	pb := new(parallelBatch)

	pb.SetLimit(4)
	testutil.Equal(t, 0, len(pb.sem))
	testutil.Equal(t, 4, cap(pb.sem))

	pb.SetLimit(-1)
	testutil.Nil(t, pb.sem)

	var recovered any
	func() {
		defer func() {
			recovered = recover()
		}()
		pb.SetLimit(4)
		pb.sem <- parallelBatchToken{}
		// this will panic hopefully
		pb.SetLimit(4)
	}()

	testutil.NotNil(t, recovered)
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

	o := Observe(g, m1)

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
	o := Observe(g, readFile)

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
	o := Observe(g, readFile)

	err := g.ParallelStabilize(ctx)
	testutil.NotNil(t, err)
	testutil.Equal(t, "", o.Value())

	testutil.Equal(t, 3, g.recomputeHeap.len())
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
	_ = Observe(g, gonnaPanic)

	err := g.ParallelStabilize(ctx)
	testutil.NotNil(t, err)
	testutil.NotEqual(t, 0, len(outBuf.String()))
	testutil.NotEqual(t, 0, len(errBuf.String()))
	testutil.Equal(t, true, strings.Contains(errBuf.String(), "this is only a test"))
}
