package incr

// Heap is a generic priority queue.
//
// You must provide a `Less(...) bool` function, but values can be omitted.
type Heap[A any] struct {
	Values []A
	Less   func(A, A) bool
}

// Init establishes the heap invariants required by the other routines in this package.
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

// Peek returns the first (smallest) element in the heap.
func (h *Heap[A]) Peek() (output A, ok bool) {
	if len(h.Values) == 0 {
		return
	}
	output = h.Values[0]
	ok = true
	return
}

// Pop removes and returns the minimum element (according to Less) from the heap.
// The complexity is O(log n) where n = h.Len().
// Pop is equivalent to Remove(h, 0).
func (h *Heap[A]) Pop() (output A, ok bool) {
	if len(h.Values) == 0 {
		return
	}

	// heap pop
	n := len(h.Values) - 1
	h.swap(0, n)
	h.down(0, n)

	// intheap pop
	old := h.Values
	n = len(old)
	output = old[n-1]
	ok = true
	h.Values = old[0 : n-1]
	return
}

// Fix re-establishes the heap ordering after the element at index i has changed its value.
// Changing the value of the element at index i and then calling Fix is equivalent to,
// but less expensive than, calling Remove(h, i) followed by a Push of the new value.
// The complexity is O(log n) where n = h.Len().
func (h *Heap[A]) Fix(i int) {
	if !h.down(i, len(h.Values)) {
		h.up(i)
	}
}

// Remove removes and returns the element at index i from the heap.
// The complexity is O(log n) where n = h.Len().
func (h *Heap[A]) Remove(i int) (output A, ok bool) {
	n := len(h.Values) - 1
	if n != i {
		h.swap(i, n)
		if !h.down(i, n) {
			h.up(i)
		}
	}
	return h.Pop()
}

//
// internal helpers
//

func (h *Heap[A]) swap(i, j int) {
	h.Values[i], h.Values[j] = h.Values[j], h.Values[i]
}

func (h *Heap[A]) up(j int) {
	for {
		i := (j - 1) / 2 // parent
		if i == j || !h.Less(h.Values[j], h.Values[i]) {
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
		if j2 := j1 + 1; j2 < n && h.Less(h.Values[j2], h.Values[j1]) {
			j = j2 // = 2*i + 2  // right child
		}
		if !h.Less(h.Values[j], h.Values[i]) {
			break
		}
		h.swap(i, j)
		i = j
	}
	return i > i0
}
