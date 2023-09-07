package incr

import (
	"context"
	"fmt"
	"testing"
	"time"

	. "github.com/wcharczuk/go-incr/testutil"
)

func Test_ParallelStabilize(t *testing.T) {
	ctx := testContext()

	v0 := Var("foo")
	v1 := Var("bar")
	m0 := Map2(v0, v1, func(a, b string) string {
		return a + " " + b
	})

	graph := New()
	_ = MustObserve(graph, m0)

	err := graph.ParallelStabilize(ctx)
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

	err = graph.ParallelStabilize(ctx)
	ItsNil(t, err)

	ItsEqual(t, 2, v0.Node().changedAt)
	ItsEqual(t, 1, v1.Node().changedAt)
	ItsEqual(t, 2, m0.Node().changedAt)

	ItsEqual(t, 2, v0.Node().recomputedAt)
	ItsEqual(t, 1, v1.Node().recomputedAt)
	ItsEqual(t, 2, m0.Node().recomputedAt)

	ItsEqual(t, "not foo bar", m0.Value())
}

func Test_ParallelStabilize_alreadyStabilizing(t *testing.T) {
	ctx := testContext()

	graph := New()
	graph.status = StatusStabilizing

	err := graph.ParallelStabilize(ctx)
	ItsNotNil(t, err)
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
	_ = MustObserve(graph, output)

	err := graph.ParallelStabilize(ctx)
	ItsNil(t, err)
	ItsEqual(t, 2, len(output.Value()))

	data = append(data, Entry{
		"5", now.Add(5 * time.Second),
	})
	err = graph.ParallelStabilize(ctx)
	ItsNil(t, err)
	ItsEqual(t, 2, len(output.Value()))

	i.Set(data)
	err = graph.ParallelStabilize(ctx)
	ItsNil(t, err)
	ItsEqual(t, 3, len(output.Value()))
}

func Test_ParallelStabilize_error(t *testing.T) {
	ctx := testContext()

	v0 := Var("foo")
	m0 := MapContext(v0, func(ctx context.Context, a string) (string, error) {
		return "", fmt.Errorf("this is only a test")
	})

	graph := New()
	_ = MustObserve(graph, m0)

	err := graph.ParallelStabilize(ctx)
	ItsNotNil(t, err)
}

func Test_parallelBatch(t *testing.T) {
	pb := new(parallelBatch)

	var values = make(chan string, 1)
	pb.Go(func() error {
		values <- "hello"
		return nil
	})
	err := pb.Wait()
	ItsNil(t, err)
	got := <-values
	ItsEqual(t, "hello", got)
}

func Test_parallelBatch_error(t *testing.T) {
	pb := new(parallelBatch)

	pb.Go(func() error {
		return fmt.Errorf("this is a test")
	})
	err := pb.Wait()
	ItsNotNil(t, err)
}

func Test_parallelBatch_SetLimit(t *testing.T) {
	pb := new(parallelBatch)

	pb.SetLimit(4)
	ItsEqual(t, 0, len(pb.sem))
	ItsEqual(t, 4, cap(pb.sem))

	pb.SetLimit(-1)
	ItsNil(t, pb.sem)

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

	ItsNotNil(t, recovered)
}
