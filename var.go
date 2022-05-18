package incr

import (
	"context"
	"fmt"
	"sync/atomic"
)

// Var returns a new var node.
//
// It will include an extra method `Set` above what you
// typically find on Incr[A], as well as a `Read` method
// that helps integrate into subcomputations.
func Var[T any](t T) VarIncr[T] {
	return &varIncr[T]{
		n:  NewNode(),
		nv: t,
	}
}

// VarIncr is a type that implements a variable.
type VarIncr[T any] interface {
	Incr[T]
	Set(T)
	Read() Incr[T]
}

// Assert interface implementations.
var (
	_ Incr[string]    = (*varIncr[string])(nil)
	_ VarIncr[string] = (*varIncr[string])(nil)
	_ INode           = (*varIncr[string])(nil)
	_ IStabilize      = (*varIncr[string])(nil)
	_ fmt.Stringer    = (*varIncr[string])(nil)
)

// VarIncr is a type that can represent a Var incremental.
type varIncr[T any] struct {
	n  *Node
	v  T
	nv T
	uv T
}

// Set sets the var value.
//
// The value is realized on the next stabilization pass.
//
// This will invalidate any nodes that reference this variable.
func (vn *varIncr[T]) Set(v T) {
	// if the node is "stabilizing"
	// what should we do? just hold the new value
	// until stabilization is done?
	if atomic.LoadInt32(&vn.n.g.status) == StatusStabilizing {
		vn.uv = v
		vn.n.g.setDuringStabilization.Push(vn.n.id, vn)
		return
	}
	vn.nv = v
	SetStale(vn)
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
	return FormatNode(vn.n, "var")
}
