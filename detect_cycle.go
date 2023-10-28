package incr

import "fmt"

// DetectCycle determines if adding a given input to a given
// child would cause a graph cycle.
//
// It is a low-level utility function that should be used
// in special cases; the vast majority of direct use cases
// for the incremental library cannot create graph cycles.
func DetectCycle(child, parent INode) error {
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

// NOTE(wc): this is a more comprehensive, but memory intensive, cycle
// detection method that uses a "seen" identifier set and walks up the graph
// from the parent as if the child had been added as a child of the parent
// through an injected `GetChildren` method.
//
// In practice, according to the test suite, it's not more correct
// than the fast method but I'm keeping it here in case we need to revert.

// func detectCycleSlow(child, parent INode) error {
// 	getChildren := func(n INode) []INode {
// 		if n.Node().id == parent.Node().id {
// 			return append(parent.Node().Children(), child)
// 		}
// 		return n.Node().Children()
// 	}
// 	seen := make(set[Identifier])
// 	if err := cycleSeen(child, getChildren, seen); err != nil {
// 		return fmt.Errorf("%w; adding %v as a child of %v would cause a cycle", err, child, parent)
// 	}
// 	return nil
// }

// func cycleSeen(n INode, getChildren func(INode) []INode, seen set[Identifier]) error {
// 	if seen.has(n.Node().id) {
// 		return fmt.Errorf("cycle detected at %v", n)
// 	}
// 	seen.add(n.Node().id)
// 	children := getChildren(n)
// 	for _, c := range children {
// 		if err := cycleSeen(c, getChildren, seen.copy()); err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }
