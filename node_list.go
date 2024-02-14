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

func hasKey[A INode](nodes []A, id Identifier) bool {
	for _, n := range nodes {
		if n.Node().id == id {
			return true
		}
	}
	return false
}

func mapHasKey[K comparable, V any](m map[K]V, k K) (ok bool) {
	_, ok = m[k]
	return
}

func popMap[K comparable, V any](m map[K]V) (v V, ok bool) {
	var k K
	for k, v = range m {
		ok = true
		break
	}
	delete(m, k)
	return
}
