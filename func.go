package incr

import "context"

// Func returns an incr that wraps a function that returns a value
// for a given context.
func Func[A any](fn func(context.Context) (A, error)) Incr[A] {
	f := &funcIncr[A]{
		fn: fn,
	}
	f.n = NewNode(f)
	return f
}

type funcIncr[A any] struct {
	n     *Node
	fn    func(context.Context) (A, error)
	value A
}

func (f *funcIncr[A]) Value() A {
	return f.value
}

func (f *funcIncr[A]) Stale() bool { return true }

func (f *funcIncr[A]) Stabilize(ctx context.Context) error {
	value, err := f.fn(ctx)
	if err != nil {
		return err
	}
	f.value = value
	return nil
}

func (f *funcIncr[A]) Node() *Node { return f.n }
