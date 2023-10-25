package incr

import "fmt"

// DetectCycle determines if adding a given input to a given
// child would cause a graph cycle.
//
// It is a low-level utility function that should be used
// in special cases; the vast majority of direct use cases
// for the incremental library cannot create graph cycles.
func DetectCycle(child, input INode) error {
	getChildren := func(n INode) []INode {
		if n.Node().id == input.Node().id {
			return append(input.Node().Children(), child)
		}
		return n.Node().Children()
	}
	seen := make(set[Identifier])
	if err := cycleSeen(child, getChildren, seen); err != nil {
		return fmt.Errorf("linking %v and %v would cause a cycle: %w", child, input, err)
	}
	return nil
}

func cycleSeen(n INode, getChildren func(INode) []INode, seen set[Identifier]) error {
	if seen.has(n.Node().id) {
		return fmt.Errorf("cycle detected at %v", n)
	}
	seen.add(n.Node().id)
	children := getChildren(n)
	for _, c := range children {
		if err := cycleSeen(c, getChildren, seen.copy()); err != nil {
			return err
		}
	}
	return nil
}