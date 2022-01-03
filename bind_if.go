package incr

import "context"

// BindIf returns A or B based on a C incremental, but will return the underlying
// incremental forms of them, computing the values of each.
func BindIf[A comparable](i0 Incr[A], i1 Incr[A], c Incr[bool]) Incr[A] {
	bi := &bindIfIncr[A]{
		i0: i0,
		i1: i1,
		c:  c,
	}
	bi.n = newNode(bi,
		optNodeChildOf(i0),
		optNodeChildOf(i1),
		optNodeChildOf(c),
	)
	return bi
}

type bindIfIncr[A comparable] struct {
	n     *node
	i0    Incr[A]
	i1    Incr[A]
	c     Incr[bool]
	value Incr[A]
}

func (bii *bindIfIncr[A]) Value() A {
	return bii.value.Value()
}

func (bii *bindIfIncr[A]) Stabilize(ctx context.Context) error {
	if bii.c.Value() {
		bii.value = bii.i0
	} else {
		bii.value = bii.i1
	}
	return nil
}

func (bii *bindIfIncr[A]) getValue() any {
	return bii.Value()
}

func (bii *bindIfIncr[A]) getNode() *node {
	return bii.n
}
