package incr

import (
	"context"
	"testing"
)

func Test_BindIf(t *testing.T) {
	a := Return("a")
	b := Return("b")
	c := Var(true)

	itsEqual(t, true, c.Value())

	bi := BindIf(a, b, c)

	itsEqual(t, 3, len(bi.getNode().parents))
	itsEqual(t, 1, len(a.getNode().children))
	itsEqual(t, 1, len(b.getNode().children))
	itsEqual(t, 1, len(c.getNode().children))

	value := bi.Value()
	itsEqual(t, "a", value)

	c.Set(false)
	_ = c.Stabilize(context.TODO())

	value = bi.Value()
	itsEqual(t, "b", value)
}
