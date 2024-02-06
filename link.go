package incr

import "fmt"

// Link is a helper for setting up node parent child relationships.
//
// A child is a node that has "parent" or input nodes. The parents then
// have the given node as a "child" or node that uses them as inputs.
//
// An error is returned if the provided inputs to the child node
// would produce a cycle.
func Link(child INode, inputs ...INode) {
	_ = link(child, inputs...)
}

func link(child INode, inputs ...INode) error {
	child.Node().addParents(inputs...)
	for _, input := range inputs {
		input.Node().addChildren(child)
	}
	if err := propagateHeightChange(child.Node().ID(), child); err != nil {
		return err
	}
	for _, input := range inputs {
		if err := propagateHeightChange(input.Node().ID(), input); err != nil {
			return err
		}
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

func propagateHeightChange(originalChildID Identifier, in INode) error {
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
			if c.Node().ID() == originalChildID {
				return fmt.Errorf("cycle detected at %v", originalChildID)
			}
			if err := propagateHeightChange(originalChildID, c); err != nil {
				return err
			}
		}
	}
	return nil
}
