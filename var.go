package incr

import "context"

// Var returns a new var node.
//
// It will include an extra method `Set` above what you
// typically find on Incr[A].
func Var[T any](t T) *VarIncr[T] {
	return &VarIncr[T]{
		n:  newNode(),
		nv: t,
	}
}

// Assert interface implementations.
var (
	_ Incr[string] = (*VarIncr[string])(nil)
	_ GraphNode    = (*VarIncr[string])(nil)
	_ Stabilizer   = (*VarIncr[string])(nil)
)

// VarIncr is a type that can represent a Var incremental.
type VarIncr[T any] struct {
	n  *Node
	v  T
	nv T
}

// Set sets the var value.
//
// The value is realized on the next stabilization pass.
//
// This will invalidate any nodes that reference this variable.
func (vn *VarIncr[T]) Set(v T) {
	vn.nv = v
	vn.n.setAt = vn.n.gs.sn
	vn.n.gs.rh.add(vn)
}

// Node implements Incr[A].
func (vn *VarIncr[T]) Node() *Node { return vn.n }

// Value implements Incr[A].
func (vn *VarIncr[T]) Value() T { return vn.v }

// Stabilize implements Incr[A].
func (vn *VarIncr[T]) Stabilize(ctx context.Context) error {
	vn.v = vn.nv
	return nil
}

// String implements fmt.Striger.
func (vn *VarIncr[T]) String() string {
	return "var[" + vn.n.id.Short() + "]"
}
