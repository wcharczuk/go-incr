package incr

// Unlink removes the parent child association
// between two nodes.
func Unlink(child, input INode) {
	child.Node().removeParent(input.Node().id)
	input.Node().removeChild(child.Node().id)

	if graph := graphFromAnyScope(child, input); graph != nil {
		graph.checkIfUnnecessary(input)
	}
}

func graphFromAnyScope(nodes ...INode) *Graph {
	for _, n := range nodes {
		if graph := graphFromScope(n); graph != nil {
			return graph
		}
	}
	return nil
}
