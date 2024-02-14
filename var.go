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
func Var[T any](scope Scope, t T) VarIncr[T] {
	return WithinScope(scope, &varIncr[T]{
		n:     NewNode("var"),
		value: t,
	})
}

// VarIncr is a graph node type that implements an incremental variable.
type VarIncr[T any] interface {
	Incr[T]
	IShouldBeInvalidated

	// Set sets the var value.
	//
	// The value is realized on the next stabilization pass.
	//
	// This will invalidate any nodes that reference this variable.
	Set(T)
}

// Assert interface implementations.
var (
	_ Incr[string]    = (*varIncr[string])(nil)
	_ IStale          = (*varIncr[string])(nil)
	_ VarIncr[string] = (*varIncr[string])(nil)
	_ INode           = (*varIncr[string])(nil)
	_ IStabilize      = (*varIncr[string])(nil)
	_ fmt.Stringer    = (*varIncr[string])(nil)
)

// VarIncr is a type that can represent a Var incremental.
type varIncr[T any] struct {
	n                           *Node
	setAt                       uint64
	value                       T
	setDuringStabilizationValue T
	setDuringStabilization      bool
}

func (vn *varIncr[T]) Parents() []INode { return nil }

// Stale implements IStale.
func (vn *varIncr[T]) Stale() bool {
	return vn.setAt > vn.n.recomputedAt
}

func (vn *varIncr[T]) ShouldBeInvalidated() bool {
	return false
}

func (vn *varIncr[T]) Set(v T) {
	if vn.n.graph != nil && atomic.LoadInt32(&vn.n.graph.status) == StatusStabilizing {
		vn.setDuringStabilizationValue = v
		vn.setDuringStabilization = true

		vn.n.graph.setDuringStabilizationMu.Lock()
		vn.n.graph.setDuringStabilization[vn.Node().id] = vn
		vn.n.graph.setDuringStabilizationMu.Unlock()
		return
	}

	vn.value = v
	if vn.n.graph != nil {
		vn.n.graph.SetStale(vn)
	}
}

// Node implements Incr[A].
func (vn *varIncr[T]) Node() *Node { return vn.n }

// Value implements Incr[A].
func (vn *varIncr[T]) Value() T { return vn.value }

// Stabilize implements Incr[A].
func (vn *varIncr[T]) Stabilize(ctx context.Context) error {
	if vn.setDuringStabilization {
		var zero T
		vn.value = vn.setDuringStabilizationValue
		vn.setDuringStabilizationValue = zero
		vn.setDuringStabilization = false
		return nil
	}
	return nil
}

// String implements fmt.Striger.
func (vn *varIncr[T]) String() string {
	return vn.n.String()
}
