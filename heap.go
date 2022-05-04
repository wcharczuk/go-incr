package incr

// NewHeap returns a new heap.
func NewHeap[A any](lessfn func(A, A) bool, values ...A) *Heap[A] {
	h := &Heap[A]{
		LessFn: lessfn,
		Values: values,
	}
	h.Init()
	return h
}

// OkValue returns just the value from a (A,bool) return.
func OkValue[A any](v A, ok bool) A {
	return v
}

// Ok returns just the bool from a (A,bool) return.
func Ok[A any](_ A, ok bool) bool {
	return ok
}

// Heap is a generic container that satisfies the heap property.
//
// You should use the `NewHeap` constructor to create a Heap[A].
//
// A heap is a structure that organizes an array into a binary tree
// such that each element's children are less than the elements value
// in respec to the `LessFn` result.
//
// Most operations (Push, Pop etc.) are O(log n) where n = h.Len().
//
// Peeking the min element is O(1).
type Heap[A any] struct {
	LessFn func(A, A) bool
	Values []A
}

// MinLess returns a simple less function for ordered types.
func MinLess[A Ordered](i, j A) bool {
	return i < j
}

// MaxLess returns a simple less function for ordered types.
func MaxLess[A Ordered](i, j A) bool {
	return i > j
}

// Init establishes the heap invariants.
//
// Init is idempotent with respect to the heap invariants
// and may be called whenever the heap invariants may have been invalidated.
// The complexity is O(n) where n = h.Len().
func (h *Heap[A]) Init() {
	n := len(h.Values)
	for i := n/2 - 1; i >= 0; i-- {
		h.down(i, n)
	}
}

// Len returns the length, or number of items in the heap.
func (h *Heap[A]) Len() int {
	return len(h.Values)
}

// Push pushes values onto the heap.
func (h *Heap[A]) Push(v A) {
	h.Values = append(h.Values, v)
	h.up(len(h.Values) - 1)
}

// PushPop implements a combined push and pop action which will
// be faster in practice than calling Push and then Pop separately
// (or vice versa).
func (h *Heap[A]) PushPop(v A) (output A, ok bool) {
	if len(h.Values) == 0 {
		h.Values = append(h.Values, v)
		return
	}
	output = h.Values[0]
	ok = true
	h.Values[0] = v
	h.down(0, len(h.Values))
	return
}

// Peek returns the first (smallest according to LessFn) element in the heap.
func (h *Heap[A]) Peek() (output A, ok bool) {
	if len(h.Values) == 0 {
		return
	}
	output = h.Values[0]
	ok = true
	return
}

// Pop removes and returns the minimum element (according to `h.LessFn`) from the heap.
// The complexity is O(log n) where n = h.Len().
//
// Pop is equivalent to `RemoveAt(0)`.
func (h *Heap[A]) Pop() (output A, ok bool) {
	if len(h.Values) == 0 {
		return
	}
	n := len(h.Values) - 1
	h.swap(0, n)
	h.down(0, n)
	output, ok = h.arrayPop()
	return
}

// FixAt re-establishes the heap ordering after the element at index i has changed its value.
// Changing the value of the element at index i and then calling Fix is equivalent to,
// but less expensive than, calling `RemoveAt(i)` followed by a Push of the new value.
//
// The complexity is O(log n) where n = h.Len().
func (h *Heap[A]) FixAt(i int) {
	if !h.down(i, len(h.Values)) {
		h.up(i)
	}
}

// RemoveAt removes and returns the element at index i from the heap.
// The complexity is O(log n) where n = h.Len().
//
// Note that `RemoveAt` takes an index, you should first determine
// the index by scanning the heap values for the appropriate index.
//
// Further note; we don't have a value based function for this because
// not all types implement `comparable` and we don't want to make assumptions
// about the heap type for a subset of the heap methods.
func (h *Heap[A]) RemoveAt(i int) (output A, ok bool) {
	if len(h.Values) == 0 {
		return
	}
	n := len(h.Values) - 1
	if n != i {
		h.swap(i, n)
		if !h.down(i, n) {
			h.up(i)
		}
	}
	return h.arrayPop()
}

//
// internal helpers
//

func (h *Heap[A]) arrayPop() (output A, ok bool) {
	old := h.Values
	n := len(old)
	output = old[n-1]
	ok = true
	h.Values = old[0 : n-1]
	return
}

func (h *Heap[A]) swap(i, j int) {
	h.Values[i], h.Values[j] = h.Values[j], h.Values[i]
}

func (h *Heap[A]) less(i, j int) bool {
	return h.LessFn(h.Values[i], h.Values[j])
}

func (h *Heap[A]) up(j int) {
	for {
		i := (j - 1) / 2 // parent
		if i == j || !h.less(j, i) {
			break
		}
		h.swap(i, j)
		j = i
	}
}

func (h *Heap[A]) down(i0, n int) bool {
	i := i0
	for {
		j1 := 2*i + 1
		if j1 >= n || j1 < 0 { // j1 < 0 after int overflow
			break
		}
		j := j1 // left child
		if j2 := j1 + 1; j2 < n && h.less(j2, j1) {
			j = j2 // = 2*i + 2  // right child
		}
		if !h.less(j, i) {
			break
		}
		h.swap(i, j)
		i = j
	}
	return i > i0
}
