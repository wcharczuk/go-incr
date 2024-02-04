package incr

// Link is a helper for setting up node parent child relationships.
//
// A child is a node that has "parent" or input nodes. The parents then
// have the given node as a "child" or node that uses them as inputs.
//
// An error is returned if the provided inputs to the child node
// would produce a cycle.
func Link(child INode, inputs ...INode) {
	_ = link(child, false /*detectCycles*/, inputs...)
}

func link(child INode, detectCycles bool, inputs ...INode) error {
	if detectCycles {
		for _, parent := range inputs {
			if err := DetectCycleIfLinked(child, parent); err != nil {
				return err
			}
		}
	}

	child.Node().addParents(inputs...)
	for _, input := range inputs {
		input.Node().addChildren(child)
	}
	propagateHeightChange(child)
	for _, input := range inputs {
		propagateHeightChange(input)
	}
	return nil
}

func maxHeightOf[A INode](nodes ...A) (max int) {
	for _, n := range nodes {
		if n.Node().height > max {
			max = n.Node().height
		}
	}
	return
}

func propagateHeightChange(in INode) {
	n := in.Node()
	oldHeight := n.height
	maxParentHeight := maxHeightOf(n.Parents()...)
	if n.height == 0 || maxParentHeight >= n.height {
		n.height = maxParentHeight + 1
	}
	if n.height != oldHeight {
		if n.graph != nil {
			n.graph.adjustHeightsList.Push(in)
		}
		for _, c := range n.Children() {
			propagateHeightChange(c)
		}
	}
}
