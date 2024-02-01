package incrutil

import (
	"fmt"
	"testing"

	"github.com/wcharczuk/go-incr"
	"github.com/wcharczuk/go-incr/testutil"
)

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

	mv := incr.Var(ctx, m)
	mda := DiffMapByKeysAdded(mv)
	mf := incr.FoldMap(ctx, mda, 0, func(key string, val, accum int) int {
		return accum + val
	})

	graph := incr.New()
	_ = incr.Observe(graph, mf)

	_ = graph.Stabilize(ctx)
	testutil.ItsEqual(t, 21, mf.Value())

	m["seven"] = 7
	m["eight"] = 8

	mv.Set(m)

	_ = graph.Stabilize(ctx)
	testutil.ItsEqual(t, 36, mf.Value())

	m["nine"] = 9

	mv.Set(m)

	_ = graph.Stabilize(ctx)
	testutil.ItsEqual(t, 45, mf.Value())
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

	mv := incr.Var(ctx, m)
	mdr := DiffMapByKeysRemoved(mv)
	mf := incr.FoldMap(ctx, mdr, 0, func(key string, val, accum int) int {
		return accum + val
	})

	graph := incr.New()
	_ = incr.Observe(graph, mf)

	_ = graph.Stabilize(ctx)
	testutil.ItsEqual(t, 0, mf.Value())

	delete(m, "two")
	delete(m, "five")

	mv.Set(m)

	_ = graph.Stabilize(ctx)
	testutil.ItsEqual(t, 7, mf.Value())
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

	mv := incr.Var(ctx, m)
	mda, mdr := DiffMapByKeys(mv)
	mfa := incr.FoldMap(ctx, mda, 0, func(key string, val, accum int) int {
		return accum + val
	})
	mfr := incr.FoldMap(ctx, mdr, 0, func(key string, val, accum int) int {
		return accum + val
	})

	graph := incr.New()
	_ = incr.Observe(graph, mfa)
	_ = incr.Observe(graph, mfr)

	_ = graph.Stabilize(ctx)
	_ = graph.Stabilize(ctx)
	testutil.ItsEqual(t, 21, mfa.Value())
	testutil.ItsEqual(t, 0, mfr.Value())

	delete(m, "two")
	delete(m, "five")
	m["seven"] = 7
	m["eight"] = 8

	mv.Set(m)

	_ = graph.Stabilize(ctx)
	_ = graph.Stabilize(ctx)
	testutil.ItsEqual(t, 36, mfa.Value())
	testutil.ItsEqual(t, 7, mfr.Value())
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
	mv := incr.Var(ctx, m)
	mf := incr.FoldLeft(ctx, DiffSliceByIndicesAdded(mv), "", func(accum string, val int) string {
		return accum + fmt.Sprint(val)
	})

	graph := incr.New()
	_ = incr.Observe(graph, mf)

	_ = graph.Stabilize(ctx)
	testutil.ItsEqual(t, "123456", mf.Value())

	m = append(m, 7, 8, 9)
	mv.Set(m)

	_ = graph.Stabilize(ctx)
	testutil.ItsEqual(t, "123456789", mf.Value())
}
