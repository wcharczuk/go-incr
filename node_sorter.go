package incr

func nodeSorter(a, b INode) int {
	if a.Node().height == b.Node().height {
		aID := a.Node().ID().String()
		bID := b.Node().ID().String()
		if aID == bID {
			return 0
		} else if aID > bID {
			return -1
		}
		return 1
	} else if a.Node().height > b.Node().height {
		return -1
	}
	return 1
}
