package incr

// Link is a helper for setting up node parent child relationships.
//
// A child is a node that has "parent" or input nodes. The parents then
// have the given node as a "child" or node that uses them as inputs.
//
// The reverse of this is `Unlink` on the child node, which
// removes the inputs as "parents" of the "child" node.
func Link(child INode, inputs ...INode) {
	for _, p := range inputs {
		if err := DetectCycleIfLinked(child, p); err != nil {
			if child.Node().graph != nil {
				child.Node().graph.maybeAddObservedNode(p)
			} else {
				p.Node().graph.maybeAddObservedNode(child)
			}
			child.Node().addParents(p)
			p.Node().addChildren(child)
			_ = dumpDot(child.Node().graph, homedir("bind_if_cycle.png"))
			panic(err)
		}
	}
	child.Node().addParents(inputs...)
	for _, input := range inputs {
		input.Node().addChildren(child)
	}
}
