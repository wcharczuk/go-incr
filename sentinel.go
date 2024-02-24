package incr

import "context"

// Sentinel returns a node that evaluates a staleness function for each stabilization.
//
// Put another way, you can attach sentinels to parts of a graph to automatically recompute
// nodes on the result of a function withouth having to mark those nodes explicitly stale.
//
// Sentinels are somewhat expensive as a result, and should only be used sparingly, and even
// then only in situations where the stalness function will return true infequently.
//
// The provided function should return `true` if we should recompute watched nodes.
// Returning `false` will stop recomputation from propagating past this node.
//
// The `watched` node will be recomputed if the provided function returns true.
// The watched node is associated as a child, but because the sentinel
// has no value, the node's existing parent values are passed to it as normal, and the sentinel
// does not provide a value to the watched node.
func Sentinel(scope Scope, fn func() bool, watched INode) SentinelIncr {
	return SentinelContext(scope, func(_ context.Context) (bool, error) {
		return fn(), nil
	}, watched)
}

// SentinelContext returns a node that evaluates a staleness function for each stabilization.
//
// You can attach sentinels to parts of a graph to automatically recompute nodes
// on the result of a function withouth having to mark those nodes explicitly stale.
//
// Sentinels are somewhat expensive as a result, and should only be used sparingly, and even
// then only in situations where the stalness function will return true infequently.
//
// The provided function should return `true` if we should recompute watched nodes.
// Returning `false` will stop recomputation from propagating past this node.
// If the provided function returns an error, the stabilization is stopped and depending
// on if the stabilization is parallel or serial the error will be returned after the
// height block is completed, or immediately respectively.
//
// The `watched` node will be recomputed if the provided function returns true.
// The watched node is associated as a child, but because the sentinel
// has no value, the node's existing parent values are passed to it as normal, and the sentinel
// does not provide a value to the watched node.
func SentinelContext(scope Scope, fn func(context.Context) (bool, error), watched INode) SentinelIncr {
	s := WithinScope(scope, &sentinelIncr{
		n:       NewNode("sentinel"),
		fn:      fn,
		watched: watched,
	})
	graph := scope.scopeGraph()
	_ = graph.watchNode(s, watched)
	return s
}

// SentinelIncr is a node that is recomputed always, but can cutoff
// the computation based on the result of a provided deligate.
type SentinelIncr interface {
	IAlways
	ICutoff
	IStale
	ISentinel
}

// ISentinel is a node that always cutsoff and has no children.
type ISentinel interface {
	INode
	Unwatch(context.Context)
}

type sentinelIncr struct {
	n       *Node
	fn      func(context.Context) (bool, error)
	watched INode
}

func (s *sentinelIncr) Node() *Node { return s.n }

func (s *sentinelIncr) Always() {}

func (s *sentinelIncr) Stale() bool { return true }

func (s *sentinelIncr) Cutoff(ctx context.Context) (bool, error) {
	isStale, err := s.fn(ctx)
	return !isStale, err
}

func (s *sentinelIncr) Unwatch(_ context.Context) {
	graph := s.n.createdIn.scopeGraph()
	graph.unwatchNode(s, s.watched)
	s.watched = nil
}

func (s *sentinelIncr) String() string {
	return s.n.String()
}
