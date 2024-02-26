package mapi

func symmetricDiffAdded[M ~map[K]V, K comparable, V any](m0, m1 map[K]V) (added map[K]V) {
	added = make(map[K]V)
	var ok bool
	for k, v := range m1 {
		if _, ok = m0[k]; !ok {
			added[k] = v
		}
	}
	return
}

func symmetricDiffRemoved[M ~map[K]V, K comparable, V any](m0, m1 map[K]V) (removed map[K]V) {
	removed = make(map[K]V)
	var ok bool
	for k, v := range m0 {
		if _, ok = m1[k]; !ok {
			removed[k] = v
		}
	}
	return
}
