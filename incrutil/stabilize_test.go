package incrutil

import (
	"fmt"
	"testing"

	"github.com/wcharczuk/go-incr"
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

	mv := incr.Var(m)
	mda := DiffMapByKeysAdded(mv.Read())
	mf := incr.FoldMap(mda, 0, func(key string, val, accum int) int {
		return accum + val
	})

	graph := incr.New()
	graph.AddNodes(mf)

	_ = graph.Stabilize(ctx)
	ItsEqual(t, 21, mf.Value())

	m["seven"] = 7
	m["eight"] = 8

	mv.Set(m)

	_ = graph.Stabilize(ctx)
	ItsEqual(t, 36, mf.Value())

	m["nine"] = 9

	mv.Set(m)

	_ = graph.Stabilize(ctx)
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

	mv := incr.Var(m)
	mdr := DiffMapByKeysRemoved(mv.Read())
	mf := incr.FoldMap(mdr, 0, func(key string, val, accum int) int {
		return accum + val
	})

	graph := incr.New()
	graph.AddNodes(mf)

	_ = graph.Stabilize(ctx)
	ItsEqual(t, 0, mf.Value())

	delete(m, "two")
	delete(m, "five")

	mv.Set(m)

	_ = graph.Stabilize(ctx)
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

	mv := incr.Var(m)
	mda, mdr := DiffMapByKeys(mv.Read())
	mfa := incr.FoldMap(mda, 0, func(key string, val, accum int) int {
		return accum + val
	})
	mfr := incr.FoldMap(mdr, 0, func(key string, val, accum int) int {
		return accum + val
	})

	graph := incr.New()
	graph.AddNodes(mfa, mfr)

	_ = graph.Stabilize(ctx)
	_ = graph.Stabilize(ctx)
	ItsEqual(t, 21, mfa.Value())
	ItsEqual(t, 0, mfr.Value())

	delete(m, "two")
	delete(m, "five")
	m["seven"] = 7
	m["eight"] = 8

	mv.Set(m)

	_ = graph.Stabilize(ctx)
	_ = graph.Stabilize(ctx)
	ItsEqual(t, 36, mfa.Value())
	ItsEqual(t, 7, mfr.Value())
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
	mv := incr.Var(m)
	mf := incr.FoldLeft(DiffSliceByIndicesAdded(mv.Read()), "", func(accum string, val int) string {
		return accum + fmt.Sprint(val)
	})

	graph := incr.New()
	graph.AddNodes(mv, mv)

	_ = graph.Stabilize(ctx)
	ItsEqual(t, "123456", mf.Value())

	m = append(m, 7, 8, 9)
	mv.Set(m)

	_ = graph.Stabilize(ctx)
	ItsEqual(t, "123456789", mf.Value())
}
