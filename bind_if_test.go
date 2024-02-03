package incr

import (
	"context"
	"fmt"
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_BindIf_regression(t *testing.T) {
	ctx := testContext()
	threshold := 2
	cache := make(map[string]Incr[*int])
	fakeFormula := Var(ctx, "fakeformula")

	var f func(context.Context, int) Incr[*int]
	f = func(ctx context.Context, t int) Incr[*int] {
		key := fmt.Sprintf("f-%d", t)
		if _, ok := cache[key]; ok {
			return WithBindScope(ctx, cache[key])
		}
		r := Bind(ctx, fakeFormula, func(ctx context.Context, formula string) Incr[*int] {
			if t <= 0 {
				out := 0
				v := Return(ctx, &out)
				v.Node().SetLabel(fmt.Sprintf("return-%s", key))
				return v
			}
			m := Map(ctx, f(ctx, t-1), func(r *int) *int {
				if r == nil {
					return nil
				}
				out := *r + 1
				return &out
			})
			m.Node().SetLabel(fmt.Sprintf("map-%s", key))
			return m
		})
		r.Node().SetLabel(key)
		cache[key] = r
		return r
	}

	f_range := func(ctx context.Context, start, end int) Incr[[]*int] {
		incrs := make([]Incr[*int], end-start+1)
		for i := range incrs {
			incrs[i] = f(ctx, start+i)
		}
		mn := MapN(ctx, func(vals ...*int) []*int {
			return vals
		}, incrs...)
		mn.Node().SetLabel(fmt.Sprintf("f_range_%d_%d", start, end))
		return mn
	}

	// f_range(0,4) = [0, 1, 2, 3]
	predicateIncr := Map(ctx, f_range(ctx, 0, 4), func(vals []*int) bool {
		res := false
		for _, val := range vals {
			if val == nil {
				continue
			}
			res = *val >= threshold || res
		}
		return res
	})
	predicateIncr.Node().SetLabel("predicate")

	o := BindIf(ctx, predicateIncr, func(ctx context.Context, b bool) (Incr[*int], error) {
		if b {
			return f(ctx, 2), nil
		}
		return f(ctx, 0), nil
	})
	o.Node().SetLabel("bind-if")

	graph := New()
	_ = Observe(ctx, graph, o)

	err := graph.Stabilize(ctx)
	testutil.ItsNil(t, err)
	testutil.ItsNotNil(t, o.Value())
	testutil.ItsEqual(t, 2, *o.Value())
}
