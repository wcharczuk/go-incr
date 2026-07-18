package incr

import "context"

// ArrayFold folds a fixed set of inputs in order, starting from an initial value.
//
// The fold runs over every input on each pass, so a single input changing costs O(n).
// It exists for the shape rather than the speed: where an accumulation over inputs
// reads more clearly than a [MapN] whose function does the folding itself. When the cost
// matters, [UnorderedArrayFold] adjusts an accumulator in constant time and
// [ReduceBalanced] recomputes one path; see the package documentation.
func ArrayFold[A, B any](scope Scope, initial B, fold func(B, A) B, inputs ...Incr[A]) Incr[B] {
	return MapN(scope, func(values ...A) B {
		acc := initial
		for _, value := range values {
			acc = fold(acc, value)
		}
		return acc
	}, inputs...)
}

// ForAll reports whether every input is true.
//
// Built on [ReduceBalanced], so one input changing costs O(log n) rather than reading
// them all. Returns true for no inputs, as an empty conjunction does.
func ForAll(scope Scope, inputs ...Incr[bool]) Incr[bool] {
	if len(inputs) == 0 {
		return Return(scope, true)
	}
	return ReduceBalanced(scope, func(a, b bool) bool { return a && b }, inputs...)
}

// Exists reports whether any input is true.
//
// Built on [ReduceBalanced], so one input changing costs O(log n). Returns false for no
// inputs, as an empty disjunction does.
func Exists(scope Scope, inputs ...Incr[bool]) Incr[bool] {
	if len(inputs) == 0 {
		return Return(scope, false)
	}
	return ReduceBalanced(scope, func(a, b bool) bool { return a || b }, inputs...)
}

// DependOn returns an incremental with the value of input, which also depends on another
// node.
//
// The dependency is made necessary and its changes cause this node to recompute, but its
// value is discarded. That is useful for ordering an effect against a computation, and
// for keeping a node alive because something else needs it to have run, without the
// awkwardness of threading its value through a computation that does not want it.
func DependOn[A any](scope Scope, input Incr[A], dependency INode) Incr[A] {
	return Map2(scope, input, asUnit(scope, dependency), func(value A, _ struct{}) A {
		return value
	})
}

// asUnit adapts any node to one with a value the caller can ignore, so that it can be an
// input to a combinator that does not care what it produces.
func asUnit(scope Scope, node INode) Incr[struct{}] {
	u := &unitIncr{input: node}
	u.n = NewNode(KindUnit)
	u.parents[0] = node
	return WithinScope(scope, u)
}

var (
	_ Incr[struct{}] = (*unitIncr)(nil)
	_ IStabilize     = (*unitIncr)(nil)
	_ IParents       = (*unitIncr)(nil)
)

type unitIncr struct {
	n       *Node
	input   INode
	parents [1]INode
}

func (u *unitIncr) Parents() []INode { return u.parents[:] }

func (u *unitIncr) Node() *Node { return u.n }

func (u *unitIncr) Value() struct{} { return struct{}{} }

// Stabilize does nothing: this node exists to carry a dependency edge, and its value is
// the same on every pass. It still recomputes when its input changes, which is what
// propagates that change onward.
func (u *unitIncr) Stabilize(_ context.Context) error { return nil }

func (u *unitIncr) String() string { return u.n.String() }
