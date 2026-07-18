package incr

// This file provides identifier lookup for a node's edge lists.
//
// A node's parent, child and observer lists are all unordered sets that are
// searched by identifier when an edge is removed. A linear search over a short
// contiguous slice is the right implementation for almost every node, but these
// lists have no fixed bound: one var can feed thousands of computations, and a
// mapN can take thousands of inputs. Removing every edge of such a node -- which
// is what tearing down a graph does -- then costs O(n) per edge and O(n^2)
// overall, and it is teardown rather than steady-state work, so a user hits it
// without having asked for anything unusual.
//
// Past a threshold each list gains a map from identifier to positions, which
// makes removal O(1). The list itself stays an ordinary slice, so every reader is
// unchanged.
//
// The positions are a list because a node can legitimately appear in one of these
// lists more than once, when it takes another node as several of its inputs. It
// holds a single entry in nearly every case.

// edgeIndexThreshold is the list length past which an index is maintained. Below
// it a scan over contiguous memory beats a map lookup and costs nothing.
const edgeIndexThreshold = 64

// edgeIndexBuild indexes an existing list by identifier.
func edgeIndexBuild[A INode](list []A) map[Identifier][]int {
	index := make(map[Identifier][]int, len(list))
	for position, item := range list {
		id := item.Node().id
		index[id] = append(index[id], position)
	}
	return index
}

// edgeIndexAppend adds an item to a list, keeping any index in step and building
// one if the list has grown past the threshold.
func edgeIndexAppend[A INode](list []A, index map[Identifier][]int, item A) ([]A, map[Identifier][]int) {
	list = append(list, item)
	if index != nil {
		id := item.Node().id
		index[id] = append(index[id], len(list)-1)
		return list, index
	}
	if len(list) > edgeIndexThreshold {
		return list, edgeIndexBuild(list)
	}
	return list, nil
}

// edgeIndexRemove removes every entry with the given identifier.
//
// Order is not preserved: the last entry is moved into each hole, which is what
// makes removal O(1). All three lists this serves are unordered -- a node's
// positional inputs are declared by its kind through [IParents], not by these
// lists -- so callers must not depend on their order.
func edgeIndexRemove[A INode](list []A, index map[Identifier][]int, id Identifier) ([]A, map[Identifier][]int, A) {
	var removed A
	if index == nil {
		var found A
		list, found = remove(list, id)
		return list, nil, found
	}
	positions := index[id]
	if len(positions) == 0 {
		return list, index, removed
	}
	// Work from the highest position down, so that moving the last entry into a
	// hole cannot disturb a position still to be removed.
	for k := len(positions) - 1; k >= 0; k-- {
		position := positions[k]
		removed = list[position]
		last := len(list) - 1
		if position != last {
			moved := list[last]
			list[position] = moved
			// correct the moved entry's recorded position
			movedPositions := index[moved.Node().id]
			for j, p := range movedPositions {
				if p == last {
					movedPositions[j] = position
					break
				}
			}
		}
		var zero A
		list[last] = zero
		list = list[:last]
	}
	delete(index, id)
	// Release the index once the list is comfortably small again, with enough
	// hysteresis that hovering around the threshold does not rebuild repeatedly.
	if len(list) <= edgeIndexThreshold/2 {
		return list, nil, removed
	}
	return list, index, removed
}
