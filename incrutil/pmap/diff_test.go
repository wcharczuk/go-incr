package pmap

import (
	"fmt"
	"maps"
	"math"
	"math/rand"
	"slices"
	"testing"
	"time"
)

func collect[K interface{ ~int | ~string }, V any](m, other Map[K, V], equal func(a, b V) bool) []Change[K, V] {
	out := make([]Change[K, V], 0, 8)
	for change := range m.SymmetricDiff(other, equal) {
		out = append(out, change)
	}
	return out
}

func intsEqual(a, b int) bool { return a == b }

func Test_SymmetricDiff(t *testing.T) {
	base := New[string, int]().Set("a", 1).Set("b", 2).Set("c", 3)

	if got := collect(base, base, intsEqual); len(got) != 0 {
		t.Fatalf("a map differs from itself: %v", got)
	}

	changed := base.Set("d", 4).Delete("a").Set("b", 20)
	got := collect(base, changed, intsEqual)
	want := []Change[string, int]{
		{Kind: ChangeRemoved, Key: "a", Old: 1},
		{Kind: ChangeUpdated, Key: "b", Old: 2, New: 20},
		{Kind: ChangeAdded, Key: "d", New: 4},
	}
	if len(got) != len(want) {
		t.Fatalf("got %d changes, want %d: %v", len(got), len(want), got)
	}
	for i := range want {
		if got[i].Kind != want[i].Kind || got[i].Key != want[i].Key {
			t.Fatalf("change %d: got %v %v, want %v %v", i, got[i].Kind, got[i].Key, want[i].Kind, want[i].Key)
		}
		if want[i].Kind != ChangeAdded && got[i].Old != want[i].Old {
			t.Fatalf("change %d old: got %d, want %d", i, got[i].Old, want[i].Old)
		}
		if want[i].Kind != ChangeRemoved && got[i].New != want[i].New {
			t.Fatalf("change %d new: got %d, want %d", i, got[i].New, want[i].New)
		}
	}
}

// Test_SymmetricDiff_nilEqual covers omitting the equality function, which reports
// structural changes only.
func Test_SymmetricDiff_nilEqual(t *testing.T) {
	base := New[string, int]().Set("a", 1).Set("b", 2)
	rebound := base.Set("a", 99)
	if got := collect(base, rebound, nil); len(got) != 0 {
		t.Fatalf("expected no changes without an equality function, got %v", got)
	}
	if got := collect(base, rebound, intsEqual); len(got) != 1 {
		t.Fatalf("expected one change with an equality function, got %v", got)
	}
}

// Test_SymmetricDiff_againstReference checks the diff against one computed by brute
// force over many random mutations, in both directions.
func Test_SymmetricDiff_againstReference(t *testing.T) {
	rng := rand.New(rand.NewSource(7))

	for trial := range 200 {
		oldRef := map[int]int{}
		var older Map[int, int]
		for range rng.Intn(60) {
			key := rng.Intn(80)
			oldRef[key] = rng.Intn(10)
			older = older.Set(key, oldRef[key])
		}
		newRef := maps.Clone(oldRef)
		newer := older
		for range rng.Intn(30) {
			key := rng.Intn(80)
			switch rng.Intn(3) {
			case 0:
				delete(newRef, key)
				newer = newer.Delete(key)
			default:
				value := rng.Intn(10)
				newRef[key] = value
				newer = newer.Set(key, value)
			}
		}

		// brute force expectation
		var wantKeys []int
		for k, ov := range oldRef {
			if nv, ok := newRef[k]; !ok || nv != ov {
				wantKeys = append(wantKeys, k)
			}
		}
		for k := range newRef {
			if _, ok := oldRef[k]; !ok {
				wantKeys = append(wantKeys, k)
			}
		}
		slices.Sort(wantKeys)

		gotKeys := make([]int, 0, len(wantKeys))
		for _, change := range collect(older, newer, intsEqual) {
			gotKeys = append(gotKeys, change.Key)
			switch change.Kind {
			case ChangeAdded:
				if _, ok := oldRef[change.Key]; ok {
					t.Fatalf("trial %d: key %d reported added but was present", trial, change.Key)
				}
			case ChangeRemoved:
				if _, ok := newRef[change.Key]; ok {
					t.Fatalf("trial %d: key %d reported removed but is present", trial, change.Key)
				}
			case ChangeUpdated:
				if change.Old == change.New {
					t.Fatalf("trial %d: key %d reported updated with equal values", trial, change.Key)
				}
			}
		}
		if !slices.Equal(gotKeys, wantKeys) {
			t.Fatalf("trial %d: diff keys mismatch\n got %v\nwant %v", trial, gotKeys, wantKeys)
		}
	}
}

// Test_SymmetricDiff_earlyExit checks that abandoning the iterator stops the walk,
// so a caller can look at the first change without paying for the rest.
func Test_SymmetricDiff_earlyExit(t *testing.T) {
	var older Map[int, int]
	for i := range 200 {
		older = older.Set(i, i)
	}
	newer := older
	for i := 0; i < 200; i += 2 {
		newer = newer.Set(i, -i)
	}
	var seen int
	for range older.SymmetricDiff(newer, intsEqual) {
		seen++
		break
	}
	if seen != 1 {
		t.Fatalf("iterator yielded %d changes after break", seen)
	}
}

// Test_SymmetricDiff_scaling is the assertion the whole representation exists for:
// diffing two maps related by a fixed number of updates must not get more
// expensive as the maps get larger.
//
// A builtin map cannot do this. Comparing two of them is O(n) whatever changed,
// because a hash table shares no structure with the map it was copied from.
func Test_SymmetricDiff_scaling(t *testing.T) {
	const updates = 8
	sizes := []int{1024, 8192, 65536}
	costs := make([]time.Duration, len(sizes))

	for i, size := range sizes {
		var older Map[int, int]
		for k := range size {
			older = older.Set(k, k)
		}
		newer := older
		for u := range updates {
			newer = newer.Set(u*(size/updates), -u-1)
		}

		// time enough repetitions to be stable
		iters := 200
		start := time.Now()
		for range iters {
			var count int
			for range older.SymmetricDiff(newer, intsEqual) {
				count++
			}
			if count != updates {
				t.Fatalf("size %d: diff reported %d changes, want %d", size, count, updates)
			}
		}
		costs[i] = time.Since(start) / time.Duration(iters)
		t.Logf("size %6d: diffing %d updates costs %v", size, updates, costs[i])
	}

	// cost may grow with log(size) as the paths lengthen, but must not track size
	worst := math.Inf(-1)
	for i := 1; i < len(sizes); i++ {
		e := math.Log(float64(costs[i])/float64(costs[i-1])) / math.Log(float64(sizes[i])/float64(sizes[i-1]))
		t.Logf("  %6d -> %6d: exponent %.2f", sizes[i-1], sizes[i], e)
		worst = math.Max(worst, e)
	}
	if worst > 0.45 {
		t.Errorf("diff cost tracks map size: exponent %.2f; expected roughly logarithmic", worst)
	}
}

// Test_SymmetricDiff_unrelatedMaps documents the case where the representation
// gives no advantage: two maps built independently share nothing, so the diff has
// to look at everything.
func Test_SymmetricDiff_unrelatedMaps(t *testing.T) {
	var a, b Map[int, int]
	for i := range 100 {
		a = a.Set(i, i)
		b = b.Set(i, i)
	}
	// same contents, no shared structure: correct, just not cheap
	if got := collect(a, b, intsEqual); len(got) != 0 {
		t.Fatalf("expected no differences between equal maps, got %d", len(got))
	}
}

func Test_ChangeKind_String(t *testing.T) {
	for kind, want := range map[ChangeKind]string{
		ChangeAdded:   "added",
		ChangeRemoved: "removed",
		ChangeUpdated: "updated",
		ChangeKind(9): "unknown",
	} {
		if got := kind.String(); got != want {
			t.Fatalf("ChangeKind(%d).String() = %q, want %q", kind, got, want)
		}
	}
	_ = fmt.Sprint(ChangeAdded)
}
