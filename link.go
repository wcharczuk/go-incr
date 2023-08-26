package incr

// Link is a helper for setting up node parent child relationships.
//
// A parent is a node that has "child" or input nodes. The children then
// have the given node as a "parent" or node that takes them as inputs.
//
// The reverse of this is `Unlink` on the parent node, which
// removes the inputs as "children" of the "parent" node.
func Link(parent INode, inputs ...INode) {
	parent.Node().AddChildren(inputs...)
	for _, gnp := range inputs {
		gnp.Node().AddParents(parent)
	}
}

// Unlink removes the parent child association
// between two nodes.
func Unlink(parent, input INode) {
	parent.Node().RemoveChild(input.Node().id)
	input.Node().RemoveParent(parent.Node().id)
}
