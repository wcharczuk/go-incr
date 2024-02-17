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

	// Set sets the var value.
	//
	// The value is realized on the next stabilization pass.
	//
	// This will invalidate any nodes that reference this variable.
	Set(T)
}

// Assert interface implementations.
var (
	_ VarIncr[string]      = (*varIncr[string])(nil)
	_ IShouldBeInvalidated = (*varIncr[string])(nil)
	_ IStale               = (*varIncr[string])(nil)
	_ IStabilize           = (*varIncr[string])(nil)
	_ fmt.Stringer         = (*varIncr[string])(nil)
)

// VarIncr is a type that can represent a Var incremental.
type varIncr[T any] struct {
	n                           *Node
	setAt                       uint64
	value                       T
	setDuringStabilizationValue T
	setDuringStabilization      bool
}

// Stale implements IStale.
func (vn *varIncr[T]) Stale() bool {
	return vn.setAt > vn.n.recomputedAt
}

func (vn *varIncr[T]) ShouldBeInvalidated() bool {
	return false
}

func (vn *varIncr[T]) Set(v T) {
	graph := GraphForNode(vn)
	if atomic.LoadInt32(&graph.status) == StatusStabilizing {
		vn.setDuringStabilizationValue = v
		vn.setDuringStabilization = true

		graph.setDuringStabilizationMu.Lock()
		graph.setDuringStabilization[vn.Node().id] = vn
		graph.setDuringStabilizationMu.Unlock()
		return
	}
	vn.value = v
	if vn.n.isNecessary() {
		graph.SetStale(vn)
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
