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

	itsEqual(t, 1, len(b.getNode().parents))
	itsEqual(t, 1, len(a.getNode().children))

	itsNil(t, a.Stabilize(context.TODO()))
	itsNil(t, b.Stabilize(context.TODO()))

	itsEqual(t, 6, b.Value())
}
