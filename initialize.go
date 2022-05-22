package incr

import (
	"context"
	"fmt"
)

// Initialize starts the initialization cycle.
//
// it creates the graph state for the graph, discovers and initialized
// all nodes, and then establishes the recompute heap based on node heights.
func Initialize(ctx context.Context, nodes ...INode) {
	for _, n := range nodes {
		if n.Node().g != nil {
			continue
		}
		discoverAllNodes(ctx, newGraph(), n)
	}
}

func discoverAllNodes(ctx context.Context, g *graph, gn INode) {
	discoverNode(ctx, g, gn)
	gnn := gn.Node()
	for _, c := range gnn.children {
		if c.Node().g != nil {
			continue
		}
		discoverAllNodes(ctx, g, c)
	}
	for _, p := range gnn.parents {
		if p.Node().g != nil {
			continue
		}
		discoverAllNodes(ctx, g, p)
	}
}

func discoverNode(ctx context.Context, g *graph, gn INode) {
	gnn := gn.Node()
	gnn.g = g
	gnn.detectCutoff(gn)
	gnn.detectStabilize(gn)
	gnn.height = gnn.calculateHeight()
	g.numNodes++
	g.recomputeHeap.Add(gn)
	return
}

// undiscoverAllNodes removes a node and all its parents
// from a given graph.
//
// NOTE: you _must_ unlink it first or you'll just blow away the whole graph.
func undiscoverAllNodes(ctx context.Context, g *graph, gn INode) {
	undiscoverNode(ctx, g, gn)
	gnn := gn.Node()
	for _, c := range gnn.children {
		if c.Node().g == nil {
			continue
		}
		undiscoverAllNodes(ctx, g, c)
	}
	for _, p := range gnn.parents {
		if p.Node().g == nil {
			continue
		}
		undiscoverAllNodes(ctx, g, p)
	}
}

func undiscoverNode(ctx context.Context, g *graph, gn INode) {
	println("undiscover", fmt.Sprint(gn))
	gnn := gn.Node()
	gnn.g = nil
	g.numNodes--
	g.recomputeHeap.Remove(gn)
}
