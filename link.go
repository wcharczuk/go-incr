package incr

// Link is a helper for setting up node parent child relationships.
//
// A child is a node that has "parent" or input nodes. The parents then
// have the given node as a "child" or node that uses them as inputs.
//
// An error is returned if the provided inputs to the child node
// would produce a cycle.
<<<<<<< HEAD
func Link(child INode, parents ...INode) {
	child.Node().addParents(parents...)
	if graph := graphFromScope(child); graph != nil {
		for _, parent := range parents {
			graph.addChild(child, parent)
=======
func Link(child INode, parents ...INode) error {
	graph := graphFromScope(child)
	wasNecessary := graph.isNecessary(child)
	child.Node().addParents(parents...)
	for _, parent := range parents {
		parent.Node().addChildren(child)
	}
	if !wasNecessary {
		if err := graph.becameNecessary(child); err != nil {
			return err
		}
	}
	for _, parent := range parents {
		if err := graph.adjustHeightsHeap.adjustHeights(graph.recomputeHeap, child, parent); err != nil {
			return err
>>>>>>> main
		}
	}
	return nil
}

func graphFromScope(n INode) *Graph {
	return n.Node().createdIn.scopeGraph()
}

func heightFromScope(n INode) int {
	return n.Node().createdIn.scopeHeight()
}
