package incr

import (
	"context"
	"testing"
)

func Test_Bind(t *testing.T) {

	c0 := Map(
		Return(1),
		addConst(10),
	)

	c1 := Map(
		Return(5),
		addConst(100),
	)

	a := Var(1)
	b := Bind[int](a, func(v int) Incr[int] {
		if v > 5 {
			return c1
		}
		return c0
	})

	itsEqual(t, 1, len(b.Node().parents))
	itsEqual(t, 1, len(a.Node().children))

	itsNil(t, a.Stabilize(context.TODO(), 0))
	itsNil(t, b.Stabilize(context.TODO(), 0))

	itsEqual(t, 11, b.Value())

	a.Set(6)

	itsNil(t, a.Stabilize(context.TODO(), 0))
	itsNil(t, b.Stabilize(context.TODO(), 0))

	itsEqual(t, 105, b.Value())
}
