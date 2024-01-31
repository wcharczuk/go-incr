package incr

import (
	"context"
	"fmt"
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_BindIf_regression(t *testing.T) {
	threshold := 2
	cache := make(map[string]Incr[*int])
	fakeFormula := Var("fakeformula")

	var f func(t int) Incr[*int]
	f = func(t int) Incr[*int] {
		key := fmt.Sprintf("f-%d", t)
		if _, ok := cache[key]; ok {
			return cache[key]
		}
		r := Bind(fakeFormula, func(formula string) Incr[*int] {
			if t <= 0 {
				out := 0
				return Return(&out)
			}
			return Map(f(t-1), func(r *int) *int {
				out := *r + 1
				return &out
			})
		})
		cache[key] = r
		return r
	}

	f_range := func(start, end int) Incr[[]*int] {
		incrs := make([]Incr[*int], end-start+1)
		for i := range incrs {
			incrs[i] = f(start + i)
		}
		return MapN[*int, []*int](func(vals ...*int) []*int {
			return vals
		}, incrs...)
	}

	// f_range(0,4) = [0, 1, 2, 3]
	predicateIncr := Map(f_range(0, 4), func(vals []*int) bool {
		res := false
		for _, val := range vals {
			res = *val >= threshold || res
		}
		return res
	})

	o := BindIf(predicateIncr, func(ctx context.Context, b bool) (Incr[*int], error) {
		if b {
			return f(2), nil
		} else {
			return f(0), nil
		}
	})

	graph := New()
	_ = Observe(graph, o)

	ctx := testContext()
	err := graph.Stabilize(ctx)
	testutil.ItsNil(t, err)
	testutil.ItsNotNil(t, o.Value())
	testutil.ItsEqual(t, 2, *o.Value())
}
