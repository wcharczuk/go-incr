package incr

// Link is a helper for setting up node parent child relationships.
//
// A child is a node that has "parent" or input nodes. The parents then
// have the given node as a "child" or node that uses them as inputs.
//
// An error is returned if the provided inputs to the child node
// would produce a cycle.
func Link(child INode, parents ...INode) {
	child.Node().addParents(parents...)
	if graph := graphFromScope(child); graph != nil {
		for _, parent := range parents {
			graph.addChild(child, parent)
		}
	}
}

func graphFromScope(n INode) *Graph {
	return n.Node().createdIn.scopeGraph()
}

func heightFromScope(n INode) int {
	return n.Node().createdIn.scopeHeight()
}
