package incr

// Unlink removes the parent child association
// between two nodes.
func Unlink(child, parent INode) {
	child.Node().removeParent(parent.Node().id)
	parent.Node().removeChild(child.Node().id)
	if graph := graphFromAnyScope(child, parent); graph != nil {
		graph.checkIfUnnecessary(parent)
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
