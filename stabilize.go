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
// in the graph, the full graph will be discovered, and initialized
// on the first call to stabilize, and evaluated subsequently each pass.
//
// The general process for stabilization (after initialization) is to
// color the graph in a first pass, and then recmopute all dirty nodes
// on the second pass.
func Stabilize(ctx context.Context, s Stabilizer) error {
	if s.Node().initializedAt == 0 {
		tracePrintln(ctx, "stabilize; initializing")
		if err := initialize(ctx, s); err != nil {
			return err
		}
	}

	s.Node().gs.generation++

	tracePrintf(ctx, "stabilize; generation %d", s.Node().gs.generation)

	// we stabilize using (2) passes currently
	// one marks the graph as dirty
	// the second stabilizes the values.
	// this is not super efficient
	// and i'm sure the canonical implementation does
	// something more clever.

	var err error
	var nn *Node
	for _, n := range s.Node().gs.recomputeHeap {
		nn = n.Node()
		if nn.changedAt > nn.recomputedAt {
			// here the node is dirty, mark the graph below this node as changed
			// optionally -- allow this propagation to be cutoff
			// there also is an opportunity here to collect "dirty" nodes
			// without having to iterate through all of the (potentially unchanged) nodes
			propagateChangedAt(ctx, nn.changedAt, n)
		}
	}

	for _, n := range s.Node().gs.recomputeHeap {
		nn = n.Node()
		if nn.changedAt > nn.recomputedAt {
			if err = n.Stabilize(ctx); err != nil {
				return err
			}
			nn.recomputedAt = nn.gs.generation
		}
	}
	return nil
}

func propagateChangedAt(ctx context.Context, changedAt uint64, s Stabilizer) {
	sn := s.Node()
	sn.changedAt = changedAt
	for _, c := range sn.children {
		propagateChangedAt(ctx, changedAt, c)

	}
}

// initialize starts the initialization cycle.
//
// it creates the graph state for the graph, discovers and initialized
// all nodes, and then establishes the recompute heap based on node heights.
func initialize(ctx context.Context, s Stabilizer) error {
	tracePrintf(ctx, "initialize; initializing graph at: %s", s.String())
	gs := newGraphState()
	if err := discoverNodes(ctx, gs, s); err != nil {
		return err
	}
	tracePrintf(ctx, "initialize; discovered nodes: %d", len(gs.nodeLookup))
	roots := getRoots(ctx, gs)
	tracePrintf(ctx, "initialize; discovered roots: %d", len(roots))
	for _, r := range roots {
		establishHeights(ctx, gs, r, 1)
	}
	buildRecomputeHeap(ctx, gs)
	return nil
}

// discoverNodes fully traverses the graph and collects
// seen nodes in the graph state node lookup map.
func discoverNodes(ctx context.Context, gs *graphState, s Stabilizer) error {
	sn := s.Node()
	if _, ok := gs.nodeLookup[sn.id]; ok {
		return nil
	}

	sn.gs = gs
	sn.initializedAt = 1
	gs.nodeLookup[sn.id] = s

	var err error
	if typed, ok := s.(Initializer); ok {
		if err = typed.Initialize(ctx); err != nil {
			return err
		}
	}

	for _, p := range sn.parents {
		if err = discoverNodes(ctx, gs, p); err != nil {
			return err
		}
	}
	for _, c := range sn.children {
		if err = discoverNodes(ctx, gs, c); err != nil {
			return err
		}
	}
	return nil
}

// getRoots iterates through the graphState and finds any
// nodes with zero parents (indicating they're roots).
func getRoots(ctx context.Context, gs *graphState) (output []Stabilizer) {
	for _, n := range gs.nodeLookup {
		if len(n.Node().parents) == 0 {
			output = append(output, n)
		}
	}
	return
}

func establishHeights(ctx context.Context, gs *graphState, s Stabilizer, currentHeight int) {
	if s.Node().height < currentHeight {
		s.Node().height = currentHeight
	}
	for _, c := range s.Node().children {
		establishHeights(ctx, gs, c, currentHeight+1)
	}
}

func initializeNode(ctx context.Context, gs *graphState, s Stabilizer) error {
	sn := s.Node()
	sn.initializedAt = gs.generation
	if typed, ok := s.(Initializer); ok {
		if err := typed.Initialize(ctx); err != nil {
			return err
		}
	}
	return nil
}

func buildRecomputeHeap(ctx context.Context, gs *graphState) {
	for _, n := range gs.nodeLookup {
		gs.addRecomputeHeap(n)
	}
}
