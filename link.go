package incr

// Link is a helper for setting up node parent child relationships.
//
// A child is a node that has "parent" or input nodes. The parents then
// have the given node as a "child" or node that uses them as inputs.
//
// The reverse of this is `Unlink` on the child node, which
// removes the inputs as "parents" of the "child" node.
func Link(child INode, inputs ...INode) {
	child.Node().addParents(inputs...)
	for _, input := range inputs {
		input.Node().addChildren(child)
	}
}

// Unlink removes the parent child association
// between two nodes.
func Unlink(child, input INode) {
	child.Node().removeParent(input.Node().id)
	input.Node().removeChild(child.Node().id)
}
