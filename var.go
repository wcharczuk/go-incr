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
		n:     scope.newNode(KindVar),
		value: t,
	})
}

// VarEqual returns a var that ignores being set to the value it already holds.
//
// Setting a var normally marks it stale and propagates through everything downstream,
// including when the new value equals the old one. For inputs that are written
// repeatedly and often unchanged -- a poll result, a config reload, a replayed event --
// that is the dominant cost, and it is entirely avoidable.
//
// This is a separate constructor rather than the behavior of [Var] because deciding
// whether two values are equal requires them to be comparable, and an ordinary
// [Incr][A] holds any type at all. There is nowhere to put a graph-wide setting for it:
// the constraint has to appear where the type is known, which is here. For types with no
// ==, [VarEqualFunc] takes the comparison, and for cutting off further down the graph
// rather than at the source, see [CutoffEqual].
func VarEqual[T comparable](scope Scope, t T) VarIncr[T] {
	return VarEqualFunc(scope, t, func(a, b T) bool { return a == b })
}

// VarEqualFunc is [VarEqual] for types with no ==, taking the comparison.
func VarEqualFunc[T any](scope Scope, t T, equal func(a, b T) bool) VarIncr[T] {
	return WithinScope(scope, &varIncr[T]{
		n:     scope.newNode(KindVar),
		value: t,
		equal: equal,
	})
}

// VarIncr is a graph node type that implements an incremental variable.
type VarIncr[T any] interface {
	Incr[T]

	// Set sets the var value.
	//
	// Calling [Set] will invalidate any nodes that reference this variable.
	Set(T)
	// Update sets the var value from its current one.
	//
	// This saves reading the value out to change it, which matters because reading a
	// var directly bypasses the graph: during stabilization the value a [Set] will
	// take has not been applied yet, so computing the next value from what [Value]
	// returns can be reading a stale one.
	Update(func(T) T)
}

var (
	_ VarIncr[string]      = (*varIncr[string])(nil)
	_ IShouldBeInvalidated = (*varIncr[string])(nil)
	_ IStale               = (*varIncr[string])(nil)
	_ IStabilize           = (*varIncr[string])(nil)
	_ fmt.Stringer         = (*varIncr[string])(nil)
)

type varIncr[T any] struct {
	n     *Node
	setAt uint64
	value T
	// equal, when set, decides whether a [Set] is a change at all. It is nil for [Var]
	// and set by [VarEqual]; see there for why this cannot simply be the default.
	equal                       func(a, b T) bool
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
	// A var told to hold the value it already holds has not changed, and stopping here
	// is much cheaper than letting the graph work that out downstream: a cutoff node
	// still has to recompute this var and itself before deciding nothing happened,
	// whereas this costs one comparison and touches the graph not at all.
	if vn.equal != nil && !vn.setDuringStabilization && vn.equal(vn.value, v) {
		return
	}
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

func (vn *varIncr[T]) Update(fn func(T) T) {
	// read through the pending value if one is set, so that two updates within a
	// single stabilization compose rather than the second discarding the first
	current := vn.value
	if vn.setDuringStabilization {
		current = vn.setDuringStabilizationValue
	}
	vn.Set(fn(current))
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
