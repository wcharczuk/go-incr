package pmap

import (
	"fmt"
	"math"
	"math/rand"
	"slices"
	"testing"
)

// checkInvariants verifies the tree is a balanced search tree with correct derived
// fields, which is what every other guarantee rests on.
func checkInvariants[K interface{ ~int | ~string }, V any](t *testing.T, m Map[K, V]) {
	t.Helper()
	var walk func(n *node[K, V]) (size, height int)
	walk = func(n *node[K, V]) (int, int) {
		if n == nil {
			return 0, 0
		}
		leftSize, leftHeight := walk(n.left)
		rightSize, rightHeight := walk(n.right)
		if n.left != nil && !(n.left.key < n.key) {
			t.Fatalf("left child key %v not below %v", n.left.key, n.key)
		}
		if n.right != nil && !(n.key < n.right.key) {
			t.Fatalf("right child key %v not above %v", n.right.key, n.key)
		}
		if diff := leftHeight - rightHeight; diff > 1 || diff < -1 {
			t.Fatalf("node %v unbalanced: subtree heights %d and %d", n.key, leftHeight, rightHeight)
		}
		size, height := leftSize+rightSize+1, max(leftHeight, rightHeight)+1
		if n.size != size {
			t.Fatalf("node %v records size %d, actual %d", n.key, n.size, size)
		}
		if n.height != height {
			t.Fatalf("node %v records height %d, actual %d", n.key, n.height, height)
		}
		return size, height
	}
	size, height := walk(m.root)
	if size != m.Len() {
		t.Fatalf("Len reports %d, tree holds %d", m.Len(), size)
	}
	// a balanced tree of n entries is no taller than ~1.45*log2(n+2)
	if size > 0 {
		limit := int(1.45*math.Log2(float64(size)+2)) + 1
		if height > limit {
			t.Fatalf("height %d exceeds the bound %d for %d entries", height, limit, size)
		}
	}
}

// Test_Map_againstReference exercises a long random sequence of sets and deletes
// against a builtin map, checking contents and invariants throughout.
//
// A persistent balanced tree has enough rebalancing cases that spot checks miss
// things; this is the test that actually establishes correctness.
func Test_Map_againstReference(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	reference := map[int]string{}
	var m Map[int, string]

	for step := range 4000 {
		key := rng.Intn(200)
		if rng.Intn(3) == 0 {
			delete(reference, key)
			m = m.Delete(key)
		} else {
			value := fmt.Sprintf("v%d-%d", key, step)
			reference[key] = value
			m = m.Set(key, value)
		}

		if m.Len() != len(reference) {
			t.Fatalf("step %d: Len %d, want %d", step, m.Len(), len(reference))
		}
		if step%200 == 0 {
			checkInvariants(t, m)
		}
	}

	checkInvariants(t, m)
	for key, want := range reference {
		got, ok := m.Get(key)
		if !ok || got != want {
			t.Fatalf("Get(%d) = %q, %v; want %q", key, got, ok, want)
		}
	}
	// iteration must be in key order and cover exactly the reference contents
	wantKeys := make([]int, 0, len(reference))
	for key := range reference {
		wantKeys = append(wantKeys, key)
	}
	slices.Sort(wantKeys)
	gotKeys := make([]int, 0, len(wantKeys))
	for key := range m.Keys() {
		gotKeys = append(gotKeys, key)
	}
	if !slices.Equal(gotKeys, wantKeys) {
		t.Fatalf("iteration order mismatch:\n got %v\nwant %v", gotKeys, wantKeys)
	}
}

// Test_Map_persistence checks that an update leaves the original map alone, which
// is what makes sharing safe.
func Test_Map_persistence(t *testing.T) {
	base := New[string, int]().Set("a", 1).Set("b", 2)
	added := base.Set("c", 3)
	removed := base.Delete("a")
	rebound := base.Set("a", 100)

	if base.Len() != 2 {
		t.Fatalf("base changed size to %d", base.Len())
	}
	if v, _ := base.Get("a"); v != 1 {
		t.Fatalf("base value for a changed to %d", v)
	}
	if added.Len() != 3 || removed.Len() != 1 {
		t.Fatalf("derived sizes wrong: %d and %d", added.Len(), removed.Len())
	}
	if v, _ := rebound.Get("a"); v != 100 {
		t.Fatalf("rebound value for a is %d", v)
	}
	if _, ok := removed.Get("a"); ok {
		t.Fatal("removed map still has a")
	}
}

// Test_Map_sharing checks that an update actually shares structure, since the diff
// depends on it. Rebinding one key of a large map must leave most subtrees
// pointer-identical.
func Test_Map_sharing(t *testing.T) {
	var m Map[int, int]
	for i := range 1024 {
		m = m.Set(i, i)
	}
	updated := m.Set(512, -1)

	count := func(n *node[int, int], seen map[*node[int, int]]bool) {
		var walk func(*node[int, int])
		walk = func(x *node[int, int]) {
			if x == nil {
				return
			}
			seen[x] = true
			walk(x.left)
			walk(x.right)
		}
		walk(n)
	}
	original := map[*node[int, int]]bool{}
	count(m.root, original)
	var fresh int
	var walk func(*node[int, int])
	walk = func(x *node[int, int]) {
		if x == nil {
			return
		}
		if !original[x] {
			fresh++
			walk(x.left)
			walk(x.right)
		}
	}
	walk(updated.root)
	// only the path to the changed key should be new; allow slack for rotations
	if fresh > 24 {
		t.Fatalf("rebinding one key of a 1024 entry map rebuilt %d nodes; expected the path only", fresh)
	}
}

func Test_Bridge(t *testing.T) {
	in := map[string]int{"a": 1, "b": 2, "c": 3}
	m := FromGoMap(in)
	if m.Len() != 3 {
		t.Fatalf("Len %d", m.Len())
	}
	out := ToGoMap(m)
	if len(out) != len(in) {
		t.Fatalf("round trip size %d, want %d", len(out), len(in))
	}
	for key, want := range in {
		if out[key] != want {
			t.Fatalf("round trip %q = %d, want %d", key, out[key], want)
		}
	}

	// FromGoMap must not depend on Go's randomized iteration order: the same
	// contents must produce the same tree shape every time.
	first := FromGoMap(in)
	for range 20 {
		if got := FromGoMap(in); got.root.treeHeight() != first.root.treeHeight() {
			t.Fatal("FromGoMap produced trees of differing shape for equal input")
		}
	}

	withMore := m.SetAll(map[string]int{"d": 4, "a": 10})
	if v, _ := withMore.Get("a"); v != 10 {
		t.Fatalf("SetAll did not rebind a, got %d", v)
	}
	if withMore.Len() != 4 {
		t.Fatalf("SetAll size %d", withMore.Len())
	}
	if fewer := withMore.DeleteAll("a", "d"); fewer.Len() != 2 {
		t.Fatalf("DeleteAll size %d", fewer.Len())
	}
}
