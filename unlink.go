package incr

// Unlink removes the parent child association
// between two nodes.
func Unlink(child, input INode) {
	child.Node().removeParent(input.Node().id)
	input.Node().removeChild(child.Node().id)
}
