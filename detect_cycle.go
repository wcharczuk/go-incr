package incr

import "fmt"

// DetectCycle determines if adding a given input to a given
// child would cause a graph cycle.
//
// It is a low-level utility function that should be used
// in special cases; the vast majority of direct use cases
// for the incremental library cannot create graph cycles.
func DetectCycle(child, input INode) error {
	seen := make(set[Identifier])

	getParents := func(n INode) []INode {
		if n.Node().id == child.Node().id {
			return append(child.Node().Parents(), input)
		}
		return n.Node().Parents()
	}
	if cycleSeen(child, getParents, seen) {
		return fmt.Errorf("linking %v and %v would cause a cycle", child, input)
	}
	return nil
}

func cycleSeen(n INode, getParents func(INode) []INode, seen set[Identifier]) bool {
	if seen.has(n.Node().id) {
		return true
	}

	seen.add(n.Node().id)
	parents := getParents(n)
	for _, p := range parents {
		if cycleSeen(p, getParents, seen) {
			return true
		}
	}
	return false

}
