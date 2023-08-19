package incr

// Link is a common helper for setting up node relationships,
// specifically adding a set of "children" to a "parent" node.
//
// The reverse of this is `Unlink` on the parent node, which
// removes the inputs as "children" of the "parent" node.
func Link(parent INode, inputs ...INode) {
	parent.Node().addChildren(inputs...)
	for _, gnp := range inputs {
		gnp.Node().addParents(parent)
	}
}
