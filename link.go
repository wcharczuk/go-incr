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

	if err := propagateHeightChange(child); err != nil {
		return err
	}
	for _, input := range inputs {
		if err := propagateHeightChange(input); err != nil {
			return err
		}
	}

	return nil
}

func maxHeightOfParents(n INode) (max int) {
	for _, p := range n.Node().parents {
		if p.Node().height > max {
			max = p.Node().height
		}
	}
	return
}

func propagateHeightChange(in INode) error {
	return propagateHeightChangeRecursive(in.Node().id, in)
}

func propagateHeightChangeRecursive(originalChildID Identifier, in INode) error {
	n := in.Node()
	oldHeight := n.height
	maxParentHeight := maxHeightOfParents(in)
	if n.height == 0 || maxParentHeight >= n.height {
		n.height = maxParentHeight + 1
	}
	if n.height != oldHeight {
		if n.graph != nil {
			n.graph.adjustHeightsHeap.Add(in)
		}
		for _, c := range n.Children() {
			if c.Node().ID() == originalChildID {
				return fmt.Errorf("cycle detected at %v", originalChildID)
			}
			if err := propagateHeightChangeRecursive(originalChildID, c); err != nil {
				return err
			}
		}
	}
	return nil
}
