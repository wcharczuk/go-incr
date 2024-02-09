package incr

// Unlink removes the parent child association
// between two nodes.
func Unlink(child, input INode) {
	unlinkWithoutAdjustingHeights(child, input)
	_ = propagateHeightChange(input)
	_ = propagateHeightChange(child)
}

func unlinkWithoutAdjustingHeights(child, input INode) {
	child.Node().removeParent(input.Node().id)
	input.Node().removeChild(child.Node().id)
}
