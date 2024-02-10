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

	mv := incr.Var(incr.Root(), m)
	mda := DiffMapByKeysAdded(incr.Root(), mv)
	mf := incr.FoldMap(incr.Root(), mda, 0, func(key string, val, accum int) int {
		return accum + val
	})

	graph := incr.New()
	_ = incr.Observe(incr.Root(), graph, mf)

	_ = graph.Stabilize(ctx)
	testutil.Equal(t, 21, mf.Value())

	m["seven"] = 7
	m["eight"] = 8

	mv.Set(m)

	_ = graph.Stabilize(ctx)
	testutil.Equal(t, 36, mf.Value())

	m["nine"] = 9

	mv.Set(m)

	_ = graph.Stabilize(ctx)
	testutil.Equal(t, 45, mf.Value())
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

	mv := incr.Var(incr.Root(), m)
	mdr := DiffMapByKeysRemoved(incr.Root(), mv)
	mf := incr.FoldMap(incr.Root(), mdr, 0, func(key string, val, accum int) int {
		return accum + val
	})

	graph := incr.New()
	_ = incr.Observe(incr.Root(), graph, mf)

	_ = graph.Stabilize(ctx)
	testutil.Equal(t, 0, mf.Value())

	delete(m, "two")
	delete(m, "five")

	mv.Set(m)

	_ = graph.Stabilize(ctx)
	testutil.Equal(t, 7, mf.Value())
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

	mv := incr.Var(incr.Root(), m)
	mda, mdr := DiffMapByKeys(incr.Root(), mv)
	mfa := incr.FoldMap(incr.Root(), mda, 0, func(key string, val, accum int) int {
		return accum + val
	})
	mfr := incr.FoldMap(incr.Root(), mdr, 0, func(key string, val, accum int) int {
		return accum + val
	})

	graph := incr.New()
	_ = incr.Observe(incr.Root(), graph, mfa)
	_ = incr.Observe(incr.Root(), graph, mfr)

	_ = graph.Stabilize(ctx)
	_ = graph.Stabilize(ctx)
	testutil.Equal(t, 21, mfa.Value())
	testutil.Equal(t, 0, mfr.Value())

	delete(m, "two")
	delete(m, "five")
	m["seven"] = 7
	m["eight"] = 8

	mv.Set(m)

	_ = graph.Stabilize(ctx)
	_ = graph.Stabilize(ctx)
	testutil.Equal(t, 36, mfa.Value())
	testutil.Equal(t, 7, mfr.Value())
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
	mv := incr.Var(incr.Root(), m)
	mf := incr.FoldLeft(incr.Root(), DiffSliceByIndicesAdded(incr.Root(), mv), "", func(accum string, val int) string {
		return accum + fmt.Sprint(val)
	})

	graph := incr.New()
	_ = incr.Observe(incr.Root(), graph, mf)

	_ = graph.Stabilize(ctx)
	testutil.Equal(t, "123456", mf.Value())

	m = append(m, 7, 8, 9)
	mv.Set(m)

	_ = graph.Stabilize(ctx)
	testutil.Equal(t, "123456789", mf.Value())
}
