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
