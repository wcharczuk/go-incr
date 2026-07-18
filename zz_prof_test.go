package incr

import (
	"context"
	"testing"
)

func Test_zz_prof(t *testing.T) {
	const d = 15
	g := New(OptGraphMaxHeight(16 * d))
	sel := Var(g, 0)
	b := Bind(g, sel, func(bs Scope, which int) Incr[int] {
		var cur Incr[int] = Return(bs, which)
		for i := 0; i < d; i++ {
			cur = Map3(bs, cur, cur, cur, func(a, b, c int) int { return a + b - c })
			cur = Cutoff(bs, cur, func(p, n int) bool { return false })
		}
		return cur
	})
	MustObserve(g, b)
	ctx := context.Background()
	if err := g.Stabilize(ctx); err != nil {
		t.Fatal(err)
	}
	sel.Set(1)
	if err := g.Stabilize(ctx); err != nil {
		t.Fatal(err)
	}
}
