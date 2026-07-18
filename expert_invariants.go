package incr

import (
	"errors"
	"fmt"
)

// checkInvariants verifies the structural properties the graph relies on but does not
// enforce.
//
// This exists for callers assembling a graph through the expert interface rather than
// through the ordinary constructors -- reconstructing one from persisted state, say, or
// driving the machinery from another system. Those callers necessarily take on invariants
// that [Bind] and [Observe] would otherwise maintain for them, and the failure mode is
// quiet: a graph with a half-recorded edge or an inverted height computes plausible wrong
// answers, or leaks, rather than returning an error at the point of the mistake.
//
// Every problem found is reported, not just the first, since one bad assembly step tends
// to produce several.
func (graph *Graph) checkInvariants() error {
	var problems []error

	graph.nodesMu.Lock()
	nodes := make([]INode, 0, len(graph.nodes))
	for _, node := range graph.nodes {
		nodes = append(nodes, node)
	}
	// The node list is a slice with each node's position recorded on the node, so that
	// removal can splice in O(1). That makes the position part of the graph's structure: if
	// a node's recorded index stops pointing at itself, removing it later evicts a
	// different node and the graph silently loses one.
	for index, node := range graph.nodes {
		nn := node.Node()
		if nn.graphIndex != index {
			problems = append(problems, fmt.Errorf(
				"node %v is at position %d but records position %d",
				nn.id.Short(), index, nn.graphIndex))
		}
		if !nn.inGraph {
			problems = append(problems, fmt.Errorf(
				"node %v is in the graph's node list but is not marked as in the graph",
				nn.id.Short()))
		}
	}
	nodeCount := len(graph.nodes)
	graph.nodesMu.Unlock()

	graph.sentinelsMu.Lock()
	sentinelCount := len(graph.sentinels)
	graph.sentinelsMu.Unlock()

	graph.observersMu.Lock()
	observerCount := len(graph.observers)
	graph.observersMu.Unlock()

	// numNodes is maintained separately from the list and is what NumNodes reports. It
	// counts observers and sentinels too, both of which are tracked apart from ordinary
	// nodes, so the three have to be added up to compare against it.
	if counted := uint64(nodeCount + sentinelCount + observerCount); graph.numNodes != counted {
		problems = append(problems, fmt.Errorf(
			"graph counts %d nodes but holds %d nodes, %d observers and %d sentinels",
			graph.numNodes, nodeCount, observerCount, sentinelCount))
	}

	occurrences := func(list []INode, id Identifier) (count int) {
		for _, item := range list {
			if item.Node().id == id {
				count++
			}
		}
		return
	}

	for _, node := range nodes {
		nn := node.Node()

		// An edge is recorded on both nodes it joins, with the same multiplicity: a node
		// taking another as two of its inputs is two edges. If the two records disagree,
		// some step updated one side only, and nothing will ever remove the other half.
		for _, parent := range nn.parents {
			pn := parent.Node()
			forward := occurrences(nn.parents, pn.id)
			backward := occurrences(pn.children, nn.id)
			if forward != backward {
				problems = append(problems, fmt.Errorf(
					"edge asymmetry: %v lists %v as a parent %d time(s), but %v lists it as a child %d time(s)",
					nn, pn, forward, pn, backward))
			}
			// A child recomputes after its parents, which is what the heights are for.
			// Equal or inverted heights mean a pass can compute a node from a stale
			// input.
			if nn.height <= pn.height {
				problems = append(problems, fmt.Errorf(
					"height inversion: %v at height %d does not sit above its parent %v at height %d",
					nn, nn.height, pn, pn.height))
			}
		}
		for _, child := range nn.children {
			cn := child.Node()
			forward := occurrences(nn.children, cn.id)
			backward := occurrences(cn.parents, nn.id)
			if forward != backward {
				problems = append(problems, fmt.Errorf(
					"edge asymmetry: %v lists %v as a child %d time(s), but %v lists it as a parent %d time(s)",
					nn, cn, forward, cn, backward))
			}
		}

		// A node's recorded position in the recompute heap has to match its height, or
		// it will be pulled out in the wrong order.
		if nn.heightInRecomputeHeap != HeightUnset && nn.heightInRecomputeHeap != nn.height {
			problems = append(problems, fmt.Errorf(
				"%v is queued at height %d but has height %d",
				nn, nn.heightInRecomputeHeap, nn.height))
		}
	}

	graph.recomputeHeap.mu.Lock()
	heapErr := graph.recomputeHeap.sanityCheck()
	graph.recomputeHeap.mu.Unlock()
	if heapErr != nil {
		problems = append(problems, heapErr)
	}

	return errors.Join(problems...)
}
