package incr

import (
	"context"
	"fmt"
)

// Stabilizer is a type that can be stabilized.
type Stabilizer interface {
	fmt.Stringer
	Node() *Node
	Stabilize(context.Context) error
}

// Initializer is a type that can be initialized explicitly.
type Initializer interface {
	Initialize(context.Context) error
}

// Stabilize kicks off the full stabilization pass given an initial node
// connected to the graph.
//
// The node does not need to be an input, outpoot, root or leaf node
// in the graph, the full graph will be discovered and evaluated
// each pass.
func Stabilize(ctx context.Context, s Stabilizer) error {
	if s.Node().initializedAt == 0 {
		if err := initialize(ctx, s); err != nil {
			return err
		}
	}

	recomputeHeap := buildRecomputeHeap(ctx, s)

	var n Stabilizer
	var nn *Node
	var ok bool
	var err error
	for {
		n, ok = recomputeHeap.Pop()
		if !ok {
			break
		}

		nn = n.Node()
		if nn.changedAt > nn.recomputedAt {
			if err = n.Stabilize(ctx); err != nil {
				return err
			}
			nn.recomputedAt = nn.changedAt
		}
	}
	return nil
}

func initialize(ctx context.Context, s Stabilizer) error {
	sn := s.Node()
	if len(sn.parents) == 0 && sn.initializedAt == 0 {
		return initializeNode(ctx, s)
	}
	var err error
	for _, p := range sn.parents {
		if err = initialize(ctx, p); err != nil {
			return err
		}
	}
	return nil
}

func initializeNode(ctx context.Context, s Stabilizer) error {
	sn := s.Node()
	sn.initializedAt = 1

	if typed, ok := s.(Initializer); ok {
		if err := typed.Initialize(ctx); err != nil {
			return err
		}
	}

	var maxParentHeight int
	for _, p := range sn.parents {
		if p.Node().height > maxParentHeight {
			maxParentHeight = p.Node().height
		}
	}
	sn.height = maxParentHeight + 1

	// TODO(wc): future optimization to splice recompute heap into
	// the node metadata to save having to do this unless a bind node
	// redraws the graph
	var err error
	for _, c := range sn.children {
		if c.Node().initializedAt != 0 {
			continue
		}
		if err = initializeNode(ctx, c); err != nil {
			return err
		}
	}
	return nil
}

func buildRecomputeHeap(ctx context.Context, s Stabilizer) *Heap[Stabilizer] {
	seen := make(Set[nodeID])
	output := &Heap[Stabilizer]{
		LessFn: func(a, b Stabilizer) bool {
			return a.Node().height < b.Node().height
		},
	}
	buildRecomputeHeapVisit(seen, output, s)
	return output
}

func buildRecomputeHeapVisit(seen Set[nodeID], output *Heap[Stabilizer], s Stabilizer) {
	if seen.Has(s.Node().id) {
		return
	}
	seen.Set(s.Node().id)
	output.Push(s)
	for _, p := range s.Node().parents {
		buildRecomputeHeapVisit(seen, output, p)
	}
	for _, c := range s.Node().children {
		buildRecomputeHeapVisit(seen, output, c)
	}
}
