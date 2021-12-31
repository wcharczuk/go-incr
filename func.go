package incr

import "context"

// Func is a function that implements Incr[A], where `A` is the return
// type of the function.
//
// A `Func` incr is always stale, and will re-evaluate for each stabilization pass.
func Func[A any](fn func(context.Context) (A, error)) Incr[A] {
	fni := &funcIncr[A]{
		fn: fn,
	}
	fni.n = newNode(fni)
	return fni
}

type funcIncr[A any] struct {
	n     *node
	fn    func(context.Context) (A, error)
	value A
}

func (fni *funcIncr[A]) Value() A {
	return fni.value
}

func (fni *funcIncr[A]) Stabilize(ctx context.Context) (err error) {
	fni.value, err = fni.fn(ctx)
	return
}

func (fni *funcIncr[A]) Stale() bool { return true }

func (fni *funcIncr[A]) getNode() *node { return fni.n }
