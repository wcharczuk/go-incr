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
		if err := initialize(ctx, newGraphState(), s); err != nil {
			return err
		}
	}
	s.Node().gs.generation++

	var err error
	for _, n := range s.Node().gs.recomputeHeap {
		if n.Node().changedAt > n.Node().recomputedAt {
			if err = stabilizeChildren(ctx, s); err != nil {
				return err
			}
		}
	}
	return nil
}

func stabilizeChildren(ctx context.Context, s Stabilizer) error {
	if err := s.Stabilize(ctx); err != nil {
		return err
	}
	s.Node().recomputedAt = s.Node().gs.generation
	for _, c := range s.Node().children {
		if err := stabilizeChildren(ctx, c); err != nil {
			return err
		}
	}
	return nil
}

func initialize(ctx context.Context, gs *graphState, s Stabilizer) error {
	sn := s.Node()

	// we initialize starting at "roots" or nodes without parents
	// so that we can propagate the node neights correctly
	// to the child nodes
	if len(sn.parents) == 0 && sn.initializedAt == 0 {
		return initializeRoot(ctx, gs, s)
	}

	// traverse the graph up to the roots from
	// the initial node
	var err error
	for _, p := range sn.parents {
		if err = initialize(ctx, gs, p); err != nil {
			return err
		}
	}

	// we may need to search down through the
	// child nodes for other roots
	for _, c := range sn.children {
		if err = initialize(ctx, gs, c); err != nil {
			return err
		}
	}
	return nil
}

func initializeRoot(ctx context.Context, gs *graphState, s Stabilizer) error {
	sn := s.Node()
	sn.initializedAt = gs.generation
	sn.gs = gs

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

	gs.addNode(s)

	// TODO(wc): future optimization to splice recompute heap into
	// the node metadata to save having to do this unless a bind node
	// redraws the graph
	var err error
	for _, c := range sn.children {
		if c.Node().initializedAt != 0 {
			continue
		}
		if err = initializeRoot(ctx, gs, c); err != nil {
			return err
		}
	}
	return nil
}
