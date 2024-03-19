package mapi

// symmetricDiffAdded is a helper that compares two maps, and yields a new map with
// the keys and their associated values that were present in m1, but not in m0.
func symmetricDiffAdded[M ~map[K]V, K comparable, V any](m0, m1 map[K]V) (added map[K]V) {
	added = make(map[K]V, len(m1))
	var ok bool
	for k, v := range m1 {
		if _, ok = m0[k]; !ok {
			added[k] = v
		}
	}
	return
}

// symmetricDiffRemoved is a helper that compares two maps, and yields a new map with
// the keys and their associated values that were present in m0, but not in m1.
func symmetricDiffRemoved[M ~map[K]V, K comparable, V any](m0, m1 map[K]V) (removed map[K]V) {
	removed = make(map[K]V, len(m0))
	var ok bool
	for k, v := range m0 {
		if _, ok = m1[k]; !ok {
			removed[k] = v
		}
	}
	return
}
