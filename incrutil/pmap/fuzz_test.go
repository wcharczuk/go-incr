package pmap

import (
	"slices"
	"testing"
)

// checkStructure verifies the invariants the tree relies on but never checks at runtime:
// that keys are ordered, that the cached height and size at every node agree with the
// subtrees beneath it, and that the AVL balance condition holds.
//
// The cached values are the interesting part. Len, Nth and Rank all trust size, and
// balance trusts height; if a rotation ever fails to recompute one of them the tree keeps
// working and quietly returns wrong answers, which is exactly the failure a fuzzer should
// be able to find.
func checkStructure[K cmp2[K], V any](t *testing.T, n *node[K, V]) (size int, height int) {
	t.Helper()
	if n == nil {
		return 0, 0
	}
	leftSize, leftHeight := checkStructure[K, V](t, n.left)
	rightSize, rightHeight := checkStructure[K, V](t, n.right)

	if n.left != nil && !(n.left.key < n.key) {
		t.Fatalf("left child %v is not less than %v", n.left.key, n.key)
	}
	if n.right != nil && !(n.key < n.right.key) {
		t.Fatalf("right child %v is not greater than %v", n.right.key, n.key)
	}

	size = leftSize + rightSize + 1
	height = max(leftHeight, rightHeight) + 1
	if n.size != size {
		t.Fatalf("node %v caches size %d, subtrees hold %d", n.key, n.size, size)
	}
	if n.height != height {
		t.Fatalf("node %v caches height %d, subtrees are %d deep", n.key, n.height, height)
	}
	if d := leftHeight - rightHeight; d < -2 || d > 2 {
		t.Fatalf("node %v is unbalanced: left %d, right %d", n.key, leftHeight, rightHeight)
	}
	return size, height
}

// cmp2 restates the constraint locally so checkStructure can be generic over it.
type cmp2[T any] interface {
	~int | ~string
}

// FuzzMap drives a map through arbitrary sequences of writes and deletes and checks it
// against a plain Go map, so that any input order which breaks rebalancing shows up as a
// disagreement rather than as a wrong answer in a caller.
func FuzzMap(f *testing.F) {
	f.Add([]byte{1, 5, 1, 3, 1, 9, 2, 5})
	f.Add([]byte{1, 1, 1, 2, 1, 3, 1, 4, 1, 5, 2, 3, 2, 1})
	// an ascending run is the worst case for an unbalanced tree
	ascending := make([]byte, 0, 128)
	for i := range 64 {
		ascending = append(ascending, 1, byte(i))
	}
	f.Add(ascending)

	f.Fuzz(func(t *testing.T, data []byte) {
		m := New[int, int]()
		reference := map[int]int{}

		for i := 0; i+1 < len(data); i += 2 {
			key := int(data[i+1])
			switch data[i] % 3 {
			case 0, 1:
				value := key * 7
				m = m.Set(key, value)
				reference[key] = value
			case 2:
				m = m.Delete(key)
				delete(reference, key)
			}
		}

		checkStructure[int, int](t, m.root)

		if m.Len() != len(reference) {
			t.Fatalf("Len is %d, reference holds %d", m.Len(), len(reference))
		}
		for key, want := range reference {
			got, ok := m.Get(key)
			if !ok {
				t.Fatalf("key %d missing", key)
			}
			if got != want {
				t.Fatalf("key %d is %d, want %d", key, got, want)
			}
		}

		// iteration order, and the positional accessors that read the cached sizes
		want := slices.Sorted(maps2Keys(reference))
		got := slices.Collect(m.Keys())
		if !slices.Equal(want, got) {
			t.Fatalf("iteration gave %v, want %v", got, want)
		}
		for i, key := range want {
			nthKey, _, ok := m.Nth(i)
			if !ok || nthKey != key {
				t.Fatalf("Nth(%d) = %v (ok=%v), want %v", i, nthKey, ok, key)
			}
			rank, present := m.Rank(key)
			if !present || rank != i {
				t.Fatalf("Rank(%v) = %d (present=%v), want %d", key, rank, present, i)
			}
		}
		if _, _, ok := m.Nth(len(want)); ok {
			t.Fatal("Nth past the end reported ok")
		}
	})
}

// FuzzSymmetricDiff checks the diff against brute force comparison. The diff is the whole
// reason this map exists, and it prunes shared subtrees by pointer identity -- so a bug in
// the pruning shows up as a change that is silently never reported.
func FuzzSymmetricDiff(f *testing.F) {
	f.Add([]byte{1, 2, 3}, []byte{1, 2, 4})
	f.Add([]byte{}, []byte{1, 2, 3})
	f.Add([]byte{1, 2, 3}, []byte{})

	f.Fuzz(func(t *testing.T, before, after []byte) {
		buildFrom := func(data []byte) (Map[int, int], map[int]int) {
			m := New[int, int]()
			reference := map[int]int{}
			for i, b := range data {
				key, value := int(b), i
				m = m.Set(key, value)
				reference[key] = value
			}
			return m, reference
		}
		beforeMap, beforeRef := buildFrom(before)
		afterMap, afterRef := buildFrom(after)

		// a shared ancestor is what makes the pruning apply at all; without one the diff
		// walks both trees whole and the interesting path is never taken
		shared := beforeMap.Set(1000, 1000)
		afterShared := afterMap.Set(1000, 1000)

		for _, tc := range []struct {
			name string
			a, b Map[int, int]
			aRef map[int]int
			bRef map[int]int
		}{
			{"plain", beforeMap, afterMap, beforeRef, afterRef},
			{"sharing", shared, afterShared, withKey(beforeRef, 1000, 1000), withKey(afterRef, 1000, 1000)},
		} {
			wantChanges := map[int]string{}
			for key, value := range tc.aRef {
				if next, ok := tc.bRef[key]; !ok {
					wantChanges[key] = ChangeRemoved.String()
				} else if next != value {
					wantChanges[key] = ChangeUpdated.String()
				}
			}
			for key := range tc.bRef {
				if _, ok := tc.aRef[key]; !ok {
					wantChanges[key] = ChangeAdded.String()
				}
			}

			gotChanges := map[int]string{}
			for change := range tc.a.SymmetricDiff(tc.b, intsEqualFuzz) {
				kind := change.Kind.String()
				if previous, seen := gotChanges[change.Key]; seen {
					t.Fatalf("%s: key %d reported twice (%s then %s)", tc.name, change.Key, previous, kind)
				}
				gotChanges[change.Key] = kind
			}

			if len(gotChanges) != len(wantChanges) {
				t.Fatalf("%s: diff reported %d changes, want %d\ngot:  %v\nwant: %v",
					tc.name, len(gotChanges), len(wantChanges), gotChanges, wantChanges)
			}
			for key, want := range wantChanges {
				if got := gotChanges[key]; got != want {
					t.Fatalf("%s: key %d reported as %q, want %q", tc.name, key, got, want)
				}
			}
		}
	})
}

func intsEqualFuzz(a, b int) bool { return a == b }

func withKey(base map[int]int, key, value int) map[int]int {
	out := make(map[int]int, len(base)+1)
	for k, v := range base {
		out[k] = v
	}
	out[key] = value
	return out
}

func maps2Keys(m map[int]int) func(func(int) bool) {
	return func(yield func(int) bool) {
		for key := range m {
			if !yield(key) {
				return
			}
		}
	}
}
