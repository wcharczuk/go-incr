package incr

func remove[A INode](nodes []A, id Identifier) []A {
	var output []A
	for _, n := range nodes {
		if n.Node().id != id {
			output = append(output, n)
		}
	}
	return output
}

func removeFirst[A INode](nodes []A) (node INode, out []A) {
	if len(nodes) == 0 {
		return
	}
	node = nodes[0]
	out = nodes[1:]
	return
}

func hasKey[A INode](nodes []A, id Identifier) bool {
	for _, n := range nodes {
		if n.Node().id == id {
			return true
		}
	}
	return false
}
