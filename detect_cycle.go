package incr

import "fmt"

// DetectCycle detects a cycle in a graph rooted at a given node.
//
// It is useful to check for cycles _after_ a graph has been constructed.
func DetectCycle(root INode) error {
	seen := make(set[Identifier])
	return cycleSeen(root, seen)
}

// DetectCycleIfLinked determines if adding a given input to a given
// child would cause a graph cycle.
//
// It is a low-level utility function that should be used
// in special cases; the vast majority of direct use cases
// for the incremental library cannot create graph cycles.
func DetectCycleIfLinked(child, parent INode) error {
	getParents := func(n INode) []INode {
		if n.Node().ID() == child.Node().ID() {
			return append(child.Node().Parents(), parent)
		}
		return n.Node().Parents()
	}
	if detectCycleFast(child.Node().ID(), parent, getParents) {
		return fmt.Errorf("adding %v as child of %v would cause a cycle", child, parent)
	}
	return nil
}

func detectCycleFast(childID Identifier, startAt INode, getParents func(INode) []INode) bool {
	if startAt.Node().ID() == childID {
		return true
	}
	for _, p := range getParents(startAt) {
		if detectCycleFast(childID, p, getParents) {
			return true
		}
	}
	return false
}

func cycleSeen(n INode, seen set[Identifier]) error {
	if seen.has(n.Node().id) {
		return fmt.Errorf("cycle detected at %v", n)
	}
	seen.add(n.Node().id)
	parents := n.Node().Parents()
	for _, p := range parents {
		if err := cycleSeen(p, seen.copy()); err != nil {
			return err
		}
	}
	return nil
}
