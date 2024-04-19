package incr

func remove[A INode](nodes []A, id Identifier) []A {
	output := make([]A, 0, len(nodes))
	for _, n := range nodes {
		if n.Node().id != id {
			output = append(output, n)
		}
	}
	return output
}
