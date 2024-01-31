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

	v0 := Var("foo")
	v1 := Var("bar")
	m0 := Map2(v0, v1, func(a, b string) string {
		return a + " " + b
	})

	graph := New()
	_ = Observe(graph, m0)

	err := graph.ParallelStabilize(ctx)
	testutil.ItsNil(t, err)

	testutil.ItsEqual(t, 0, v0.Node().setAt)
	testutil.ItsEqual(t, 1, v0.Node().changedAt)
	testutil.ItsEqual(t, 0, v1.Node().setAt)
	testutil.ItsEqual(t, 1, v1.Node().changedAt)
	testutil.ItsEqual(t, 1, m0.Node().changedAt)
	testutil.ItsEqual(t, 1, v0.Node().recomputedAt)
	testutil.ItsEqual(t, 1, v1.Node().recomputedAt)
	testutil.ItsEqual(t, 1, m0.Node().recomputedAt)

	testutil.ItsEqual(t, "foo bar", m0.Value())

	v0.Set("not foo")
	testutil.ItsEqual(t, 2, v0.Node().setAt)
	testutil.ItsEqual(t, 0, v1.Node().setAt)

	err = graph.ParallelStabilize(ctx)
	testutil.ItsNil(t, err)

	testutil.ItsEqual(t, 2, v0.Node().changedAt)
	testutil.ItsEqual(t, 1, v1.Node().changedAt)
	testutil.ItsEqual(t, 2, m0.Node().changedAt)

	testutil.ItsEqual(t, 2, v0.Node().recomputedAt)
	testutil.ItsEqual(t, 1, v1.Node().recomputedAt)
	testutil.ItsEqual(t, 2, m0.Node().recomputedAt)

	testutil.ItsEqual(t, "not foo bar", m0.Value())
}

func Test_ParallelStabilize_alreadyStabilizing(t *testing.T) {
	ctx := testContext()

	graph := New()
	graph.status = StatusStabilizing

	err := graph.ParallelStabilize(ctx)
	testutil.ItsNotNil(t, err)
}

func Test_ParallelStabilize_jsDocs(t *testing.T) {
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

	err := graph.ParallelStabilize(ctx)
	testutil.ItsNil(t, err)
	testutil.ItsEqual(t, 2, len(output.Value()))

	data = append(data, Entry{
		"5", now.Add(5 * time.Second),
	})
	err = graph.ParallelStabilize(ctx)
	testutil.ItsNil(t, err)
	testutil.ItsEqual(t, 2, len(output.Value()))

	i.Set(data)
	err = graph.ParallelStabilize(ctx)
	testutil.ItsNil(t, err)
	testutil.ItsEqual(t, 3, len(output.Value()))
}

func Test_ParallelStabilize_error(t *testing.T) {
	ctx := testContext()

	v0 := Var("foo")
	m0 := MapContext(v0, func(ctx context.Context, a string) (string, error) {
		return "", fmt.Errorf("this is only a test")
	})

	graph := New()
	_ = Observe(graph, m0)

	err := graph.ParallelStabilize(ctx)
	testutil.ItsNotNil(t, err)
}

func Test_parallelBatch(t *testing.T) {
	pb := new(parallelBatch)

	var values = make(chan string, 1)
	pb.Go(func() error {
		values <- "hello"
		return nil
	})
	err := pb.Wait()
	testutil.ItsNil(t, err)
	got := <-values
	testutil.ItsEqual(t, "hello", got)
}

func Test_parallelBatch_error(t *testing.T) {
	pb := new(parallelBatch)

	pb.Go(func() error {
		return fmt.Errorf("this is a test")
	})
	err := pb.Wait()
	testutil.ItsNotNil(t, err)
}

func Test_parallelBatch_SetLimit(t *testing.T) {
	pb := new(parallelBatch)

	pb.SetLimit(4)
	testutil.ItsEqual(t, 0, len(pb.sem))
	testutil.ItsEqual(t, 4, cap(pb.sem))

	pb.SetLimit(-1)
	testutil.ItsNil(t, pb.sem)

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

	testutil.ItsNotNil(t, recovered)
}

func Test_ParallelStabilize_Always(t *testing.T) {
	v := Var("foo")
	m0 := Map(v, ident)
	a := Always(m0)
	m1 := Map(a, ident)

	var updates int
	m1.Node().OnUpdate(func(_ context.Context) {
		updates++
	})

	g := New()
	o := Observe(g, m1)

	_ = g.ParallelStabilize(context.TODO())

	testutil.ItsEqual(t, "foo", o.Value())
	testutil.ItsEqual(t, 1, updates)

	_ = g.ParallelStabilize(context.TODO())

	testutil.ItsEqual(t, "foo", o.Value())
	testutil.ItsEqual(t, 2, updates)

	v.Set("bar")

	_ = g.ParallelStabilize(context.TODO())

	testutil.ItsEqual(t, "bar", o.Value())
	testutil.ItsEqual(t, 3, updates)
}

func Test_ParallelStabilize_always_cutoff(t *testing.T) {
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

	err := g.ParallelStabilize(ctx)
	testutil.ItsNil(t, err)
	testutil.ItsEqual(t, "test-1", o.Value())

	err = g.ParallelStabilize(ctx)
	testutil.ItsNil(t, err)
	testutil.ItsEqual(t, "test-1", o.Value())

	modtime = 2

	err = g.ParallelStabilize(ctx)
	testutil.ItsNil(t, err)
	testutil.ItsEqual(t, "test-2", o.Value())

	err = g.ParallelStabilize(ctx)
	testutil.ItsNil(t, err)
	testutil.ItsEqual(t, "test-2", o.Value())
}

func Test_ParallelStabilize_always_cutoff_error(t *testing.T) {
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

	err := g.ParallelStabilize(ctx)
	testutil.ItsNotNil(t, err)
	testutil.ItsEqual(t, "", o.Value())

	testutil.ItsEqual(t, 3, g.recomputeHeap.Len())
}

func Test_ParallelStabilize_inconsistent_error(t *testing.T) {
	g := New()

	v0 := newHeightIncr(2)
	v1 := newHeightIncr(2)

	g.recomputeHeap.lookup[v0.n.id] = &recomputeHeapItem{node: v0, height: 2}
	g.recomputeHeap.lookup[v1.n.id] = &recomputeHeapItem{node: v1, height: 2}

	g.recomputeHeap.minHeight = 1

	err := g.ParallelStabilize(testContext())
	testutil.ItsNotNil(t, err)
}

func Test_ParallelStabilize_recoversPanics(t *testing.T) {
	g := New()

	v0 := Var("hello")
	gonnaPanic := Map(v0, func(_ string) string {
		panic("help!")
	})
	_ = Observe(g, gonnaPanic)
	err := g.ParallelStabilize(testContext())
	testutil.ItsNotNil(t, err)
}

func Test_ParallelStabilize_printsErrors(t *testing.T) {
	g := New()

	outBuf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)
	ctx := WithTracingOutputs(context.Background(), outBuf, errBuf)

	v0 := Var("hello")
	gonnaPanic := MapContext(v0, func(_ context.Context, _ string) (string, error) {
		return "", fmt.Errorf("this is only a test")
	})
	_ = Observe(g, gonnaPanic)

	err := g.ParallelStabilize(ctx)
	testutil.ItsNotNil(t, err)
	testutil.ItsNotEqual(t, 0, len(outBuf.String()))
	testutil.ItsNotEqual(t, 0, len(errBuf.String()))
	testutil.ItsEqual(t, true, strings.Contains(errBuf.String(), "this is only a test"))
}
