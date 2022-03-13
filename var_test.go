package incr

import (
	"context"
	"testing"
)

func Test_Var(t *testing.T) {
	v := Var(1)
	itsNotNil(t, v.Node().changedAt)
	itsEqual(t, 0, v.Node().changedAt)

	err := v.Stabilize(context.TODO())
	itsNil(t, err)

	itsEqual(t, 1, v.Value())

	v.Set(2)
	err = v.Stabilize(context.TODO())
	itsNil(t, err)

	itsEqual(t, 2, v.Value())
}

func Test_Var_Watch(t *testing.T) {
	v := Var(1)
	w := v.Watch()

	itsEmpty(t, w.Values())

	itsNil(t, v.Stabilize(context.TODO()))
	itsNil(t, w.Stabilize(context.TODO()))

	itsEqual(t, 1, len(w.Values()))
	itsEqual(t, 1, w.Values()[0])

	itsNil(t, v.Stabilize(context.TODO()))
	itsNil(t, w.Stabilize(context.TODO()))

	itsEqual(t, 1, len(w.Values()))
	itsEqual(t, 1, w.Values()[0])

	v.Set(2)

	itsNil(t, v.Stabilize(context.TODO()))
	itsNil(t, w.Stabilize(context.TODO()))

	itsEqual(t, 2, len(w.Values()))
	itsEqual(t, 1, w.Values()[0])
	itsEqual(t, 2, w.Values()[1])
}
