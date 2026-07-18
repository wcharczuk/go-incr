package pmap

import (
	"math/rand"
	"slices"
	"testing"
)

func Test_Nth_and_Rank(t *testing.T) {
	var m Map[int, string]
	keys := []int{10, 20, 30, 40, 50}
	for _, key := range keys {
		m = m.Set(key, string(rune('a'+key/10)))
	}

	for index, wantKey := range keys {
		key, value, ok := m.Nth(index)
		if !ok || key != wantKey {
			t.Fatalf("Nth(%d) = %d, %q, %v; want %d", index, key, value, ok, wantKey)
		}
		rank, present := m.Rank(wantKey)
		if !present || rank != index {
			t.Fatalf("Rank(%d) = %d, %v; want %d, true", wantKey, rank, present, index)
		}
	}

	// out of range in both directions
	if _, _, ok := m.Nth(-1); ok {
		t.Fatal("Nth(-1) should not report a value")
	}
	if _, _, ok := m.Nth(len(keys)); ok {
		t.Fatal("Nth past the end should not report a value")
	}

	// an absent key still has a meaningful rank: where it would go
	if rank, present := m.Rank(35); present || rank != 3 {
		t.Fatalf("Rank(35) = %d, %v; want 3, false", rank, present)
	}
	if rank, present := m.Rank(5); present || rank != 0 {
		t.Fatalf("Rank(5) = %d, %v; want 0, false", rank, present)
	}
	if rank, present := m.Rank(99); present || rank != len(keys) {
		t.Fatalf("Rank(99) = %d, %v; want %d, false", rank, present, len(keys))
	}
}

// Test_Nth_and_Rank_againstReference checks both against a sorted slice over a randomly
// built map, since the subtree sizes they rely on are maintained during rebalancing.
func Test_Nth_and_Rank_againstReference(t *testing.T) {
	rng := rand.New(rand.NewSource(29))
	reference := map[int]struct{}{}
	var m Map[int, int]

	for range 600 {
		key := rng.Intn(200)
		if rng.Intn(4) == 0 {
			delete(reference, key)
			m = m.Delete(key)
		} else {
			reference[key] = struct{}{}
			m = m.Set(key, key)
		}
	}

	sorted := make([]int, 0, len(reference))
	for key := range reference {
		sorted = append(sorted, key)
	}
	slices.Sort(sorted)

	for index, wantKey := range sorted {
		key, _, ok := m.Nth(index)
		if !ok || key != wantKey {
			t.Fatalf("Nth(%d) = %d, %v; want %d", index, key, ok, wantKey)
		}
		if rank, present := m.Rank(wantKey); !present || rank != index {
			t.Fatalf("Rank(%d) = %d, %v; want %d", wantKey, rank, present, index)
		}
	}

	// Nth and Rank invert each other across the whole map
	for index := range sorted {
		key, _, _ := m.Nth(index)
		rank, _ := m.Rank(key)
		if rank != index {
			t.Fatalf("Rank(Nth(%d)) = %d", index, rank)
		}
	}
}
