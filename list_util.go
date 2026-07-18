package incr

// remove drops every node with a given id from nodes, returning the shortened
// list and the removed node if there was one.
//
// This compacts in place rather than building a new list. It is called for each
// parent, child, observer and sentinel edge that a bind unlinks when it rewrites
// its right-hand side, so allocating here made teardown one of the largest
// sources of garbage in bind-heavy graphs.
//
// Order is preserved, because a node's parent list is positional for
// multi-input nodes.
func remove[A INode](nodes []A, id Identifier) (output []A, removed A) {
	var zero A
	kept := 0
	for _, n := range nodes {
		if n.Node().id == id {
			removed = n
			continue
		}
		nodes[kept] = n
		kept++
	}
	// clear the vacated tail so the backing array does not retain nodes that are
	// no longer part of the graph.
	for index := kept; index < len(nodes); index++ {
		nodes[index] = zero
	}
	return nodes[:kept], removed
}
