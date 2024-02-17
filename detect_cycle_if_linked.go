package incr

import "fmt"

// DetectCycleIfLinked determines if adding a given input to a given
// child would cause a graph cycle.
//
// It is a low-level utility function that should be used
// in special cases; the vast majority of direct use cases
// for the incremental library cannot create graph cycles.
func DetectCycleIfLinked(child, parent INode) error {
	if child == nil || parent == nil {
		return nil
	}
	getParents := func(n INode) []INode {
		if typed, ok := n.(IParents); ok {
			return typed.Parents()
		}
		return nil
	}
	getParentsWithPossibleParent := func(n INode) []INode {
		if n.Node().ID() == child.Node().ID() {
			return append(getParents(n), parent)
		}
		return getParents(n)
	}
	if detectCycleFast(child.Node().ID(), parent /*startAt*/, getParentsWithPossibleParent) {
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
