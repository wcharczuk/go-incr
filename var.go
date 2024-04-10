package incr

import (
	"context"
	"fmt"
	"sync/atomic"
)

// Var returns a new var node.
//
// [Var] nodes are special nodes in incremental as they let you input data
// into a computation, and specifically change data between stabilization passes.
//
// [Var] nodes include a method [Var.Set] that let you update the value after the initial
// construction. Calling [Var.Set] will mark the [Var] node stale, as well any of the nodes that
// take the [Var] node as an input (i.e. the [Var] node's children).
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
	// Calling [Set] will invalidate any nodes that reference this variable.
	Set(T)
}

var (
	_ VarIncr[string]      = (*varIncr[string])(nil)
	_ IShouldBeInvalidated = (*varIncr[string])(nil)
	_ IStale               = (*varIncr[string])(nil)
	_ IStabilize           = (*varIncr[string])(nil)
	_ fmt.Stringer         = (*varIncr[string])(nil)
)

type varIncr[T any] struct {
	n                           *Node
	setAt                       uint64
	value                       T
	setDuringStabilizationValue T
	setDuringStabilization      bool
}

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

func (vn *varIncr[T]) Node() *Node { return vn.n }

func (vn *varIncr[T]) Value() T { return vn.value }

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

func (vn *varIncr[T]) String() string {
	return vn.n.String()
}
