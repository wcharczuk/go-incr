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
// The `watched` variadic list of nodes are the nodes that will be recomputed if the provided
// function returns true. The watched node list is associated as children, but because the sentinel
// has no value, the nodes existing parent values are passed to them as normal, and the sentinel
// does not provide a value to the watched node.
func Sentinel(scope Scope, fn func() bool, watched ...INode) SentinelIncr {
	return SentinelContext(scope, func(_ context.Context) (bool, error) {
		return fn(), nil
	}, watched...)
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
// The `watched` variadic list of nodes are the nodes that will be recomputed if the provided
// function returns true. The watched node list is associated as children, but because the sentinel
// has no value, the nodes existing parent values are passed to them as normal, and the sentinel
// does not provide a value to the watched node.
func SentinelContext(scope Scope, fn func(context.Context) (bool, error), watched ...INode) SentinelIncr {
	s := WithinScope(scope, &sentinelIncr{
		n:       NewNode("sentinel"),
		fn:      fn,
		watched: watched,
	})
	graph := scope.scopeGraph()
	for _, w := range watched {
		_ = graph.watchNode(s, w)
	}
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
	watched []INode
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
	for _, w := range s.watched {
		graph.unwatchNode(s, w)
	}
}

func (s *sentinelIncr) String() string {
	return s.n.String()
}
