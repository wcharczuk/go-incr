package incr

// Unlink removes the parent child association
// between two nodes.
func Unlink(child, input INode) {
	if typed, ok := input.(IUnlink); ok {
		typed.Unlink()
	}

	child.Node().removeParent(input.Node().id)
	input.Node().removeChild(child.Node().id)
}
