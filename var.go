package incr

import (
	"context"
	"fmt"
)

// Var returns a new var node.
//
// It will include an extra method `Set` above what you
// typically find on Incr[A].
func Var[T any](t T) VarIncr[T] {
	return &varIncr[T]{
		n:  NewNode(),
		nv: t,
	}
}

type VarIncr[T any] interface {
	Incr[T]
	Set(T)
	Read() Incr[T]
}

// Assert interface implementations.
var (
	_ Incr[string] = (*varIncr[string])(nil)
	_ GraphNode    = (*varIncr[string])(nil)
	_ Stabilizer   = (*varIncr[string])(nil)
	_ fmt.Stringer = (*varIncr[string])(nil)
)

// VarIncr is a type that can represent a Var incremental.
type varIncr[T any] struct {
	n  *Node
	v  T
	nv T
}

// Set sets the var value.
//
// The value is realized on the next stabilization pass.
//
// This will invalidate any nodes that reference this variable.
func (vn *varIncr[T]) Set(v T) {
	vn.nv = v
	vn.n.setAt = vn.n.gs.sn
	vn.n.gs.rh.add(vn)
}

// Node implements Incr[A].
func (vn *varIncr[T]) Node() *Node { return vn.n }

// Value implements Incr[A].
func (vn *varIncr[T]) Value() T { return vn.v }

// Stabilize implements Incr[A].
func (vn *varIncr[T]) Stabilize(ctx context.Context) error {
	vn.v = vn.nv
	return nil
}

// Read returns the var as a plain incr.
//
// This helps with some random type inference issues.
func (vn *varIncr[T]) Read() Incr[T] { return vn }

// String implements fmt.Striger.
func (vn *varIncr[T]) String() string {
	return "var[" + vn.n.id.Short() + "]"
}
