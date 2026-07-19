package pmap

import (
	"cmp"
	"slices"
)

// FromGoMap builds a [Map] from a builtin map.
//
// This costs O(n log n) and is the boundary between the two representations: use
// it once where data enters an incremental computation, not on every update. An
// incremental map that is rebuilt from a builtin map each pass shares no structure
// with its predecessor, so [Map.SymmetricDiff] would have nothing to skip and the
// point of the structure would be lost.
func FromGoMap[K cmp.Ordered, V any](in map[K]V) Map[K, V] {
	var out Map[K, V]
	for _, key := range sortedKeys(in) {
		out = out.Set(key, in[key])
	}
	return out
}

// ToGoMap copies a [Map] into a builtin map, for handing results to code that does
// not know about this package. Costs O(n).
func ToGoMap[K cmp.Ordered, V any](in Map[K, V]) map[K]V {
	out := make(map[K]V, in.Len())
	for key, value := range in.All() {
		out[key] = value
	}
	return out
}

// SetAll returns a map with every entry of in applied on top of the receiver.
//
// This is a Set per entry, so it costs the same as the caller's own loop would. What it adds
// is a defined order -- entries are applied in key order -- so the resulting tree shape is
// reproducible rather than depending on Go's map iteration.
func (m Map[K, V]) SetAll(in map[K]V) Map[K, V] {
	out := m
	for _, key := range sortedKeys(in) {
		out = out.Set(key, in[key])
	}
	return out
}

// DeleteAll returns a map without any of the given keys.
func (m Map[K, V]) DeleteAll(keys ...K) Map[K, V] {
	out := m
	for _, key := range keys {
		out = out.Delete(key)
	}
	return out
}

// sortedKeys returns a map's keys in order, so that building a tree from a builtin
// map does not depend on Go's randomized iteration order. Two maps with the same
// contents then produce trees of the same shape, which keeps behavior
// reproducible.
func sortedKeys[K cmp.Ordered, V any](in map[K]V) []K {
	keys := make([]K, 0, len(in))
	for key := range in {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	return keys
}
