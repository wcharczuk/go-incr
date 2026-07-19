package incr

import (
	"context"
)

// Incr is a type that can be an incremental node in a computation graph.
type Incr[T any] interface {
	INode
	// Value returns the stabilized value of the node.
	//
	// In practice you do not want to access this value
	// directly, you almost always want to observe this
	// node with an [Observe] incremental which will
	// ensure the node is necessary and will be stabilized.
	Value() T
}

// INode is a node in the incremental graph.
type INode interface {
	// Node returns the underlying node metadata
	// for a given incremental node.
	//
	// The node metadata is used to track inputs,
	// and track dependencies, as well it's used
	// for internal state tracking within a [Graph].
	Node() *Node
}

// IParents is a type that has parents or inputs.
//
// This list is used to link the node into the graph
// but also to reconstruct the links if the node becomes
// unnecessary, and then necessary again after construction.
type IParents interface {
	// Parents provides a list of nodes that this node takes as inputs
	// for propagation of changes and dependency tracking.
	//
	// Almost all nodes outside [Var] and [Return] nodes will
	// take inputs and implement this interface.
	Parents() []INode
}

// IStabilize is a type that can be stabilized.
type IStabilize interface {
	// Stabilize instructs a node to compute a new value
	// based on the inferred status that its inputs are stale.
	//
	// An error returned by stabilize will halt computation, though
	// in practice specifically when computation is halted depends
	// if the graph is being stabilized serially or in parallel.
	Stabilize(context.Context) error
}

// IShouldBeInvalidateds a type that determines if a node should be invalidated or not.
type IShouldBeInvalidated interface {
	// ShouldBeInvalidated allows a node to return a hint to the graph
	// that it should be invalidated.
	ShouldBeInvalidated() bool
}

// IStale is a type that determines if it is stale or not.
type IStale interface {
	// Stale allows a node to opt into stabilization
	// regardless of the status of its inputs.
	Stale() bool
}

// ICutoff is a type that determines if changes should
// continue to propagate past this node or not.
type ICutoff interface {
	// Cutoff allows an incremental to be stabilized but
	// halt stabilization below itself, that is
	// if [ICutoff.Cutoff] returns true, the nodes that
	// take this node as an input will not be stabilized.
	Cutoff(context.Context) (bool, error)
}

// IAlways is a type that is opted into for recomputation each
// pass of stabilization.
type IAlways interface {
	// Always is used as a hint to the graph that this node should
	// always be stabilized regardless of the staleness of its
	// inputs, but also that these nodes should be added
	// to the recompute heap after each stabilization pass automatically.
	Always()
}

// ISentinel is a node that manages the staleness of a target node
// based on a predicate and can mark that target node for recomputation.
type ISentinel interface {
	INode
	// Unwatch removes the sentinel tracking of the given node.
	Unwatch(context.Context)
}

// IChildChanged is implemented by nodes that need to know which of their inputs
// changed, rather than re-reading all of them when they recompute.
//
// A node with many inputs that recomputes by reading them all costs O(inputs) per
// pass however few of them changed, which is what makes [MapN] linear per update.
// A node that is told which input changed, and what it changed from, can instead
// adjust an accumulator in constant time; see [UnorderedArrayFold].
//
// ChildChanged is called during the input's recompute, after the input has taken
// its new value and only when that value actually changed.
//
// How many times it is called for one changed input is not guaranteed. A node taking the
// same input in several slots hears about it once per slot when stabilizing in parallel, and
// once when stabilizing serially, because the serial path holds one dependent back and skips
// a repeat of it. An implementation therefore has to be idempotent within a pass --
// [UnorderedArrayFold] is, because it recomputes each affected slot's contribution from the
// input's current value rather than accumulating a delta per notification.
type IChildChanged interface {
	INode
	// ChildChanged reports that the given input has just taken a new value.
	ChildChanged(child INode)
}
