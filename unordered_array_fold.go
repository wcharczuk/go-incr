package incr

import (
	"context"
	"fmt"
)

// UnorderedArrayFold folds many inputs into a single value, updating the result
// from just the input that changed rather than re-reading all of them.
//
// The initial pass folds every input with fold, starting from initial. Afterwards,
// when one input changes, update is called with the accumulator and that input's
// old and new values, and its result becomes the new accumulator. A change
// therefore costs O(1) rather than O(inputs), which is the difference between an
// aggregate over thousands of inputs being practical and not.
//
// update must undo the old value's contribution and apply the new one, so the
// result matches what folding from scratch would give. For a sum that is
//
//	func(acc, old, new int) int { return acc - old + new }
//
// This requires the operation to have an inverse. Aggregates without one -- min,
// max, concatenation -- cannot be maintained this way; use [ReduceBalanced], which
// costs O(log n) per change and needs only associativity.
//
// The order in which inputs are folded is not specified, hence "unordered": with a
// correct update the accumulator is independent of it.
func UnorderedArrayFold[A, B any](
	scope Scope,
	initial B,
	fold func(B, A) B,
	update func(acc B, oldValue, newValue A) B,
	inputs ...Incr[A],
) Incr[B] {
	f := &unorderedArrayFoldIncr[A, B]{
		n:       NewNode(KindUnorderedArrayFold),
		initial: initial,
		fold:    fold,
		update:  update,
		inputs:  inputs,
		last:    make([]A, len(inputs)),
		slots:   make(map[Identifier][]int, len(inputs)),
	}
	f.value = initial
	for index, input := range inputs {
		id := input.Node().id
		f.slots[id] = append(f.slots[id], index)
	}
	return WithinScope(scope, f)
}

var (
	_ Incr[int]     = (*unorderedArrayFoldIncr[int, int])(nil)
	_ INode         = (*unorderedArrayFoldIncr[int, int])(nil)
	_ IStabilize    = (*unorderedArrayFoldIncr[int, int])(nil)
	_ IParents      = (*unorderedArrayFoldIncr[int, int])(nil)
	_ IChildChanged = (*unorderedArrayFoldIncr[int, int])(nil)
	_ fmt.Stringer  = (*unorderedArrayFoldIncr[int, int])(nil)
)

type unorderedArrayFoldIncr[A, B any] struct {
	n       *Node
	initial B
	fold    func(B, A) B
	update  func(B, A, A) B
	inputs  []Incr[A]
	value   B
	// last holds the value each input contributed to the accumulator, so that a
	// change can be applied as a delta.
	last []A
	// slots maps an input's identifier to the positions it occupies, since a fold
	// may take the same input more than once and each occurrence contributes
	// separately.
	slots map[Identifier][]int
	// folded records whether the initial full fold has happened; before it has,
	// there is no accumulator for update to adjust.
	folded bool
	// pending accumulates the deltas reported since the last recompute. It is
	// applied in Stabilize rather than on notification so that the node's value
	// only moves when the graph recomputes it.
	pending []unorderedArrayFoldChange[A]
}

type unorderedArrayFoldChange[A any] struct {
	slot     int
	oldValue A
	newValue A
}

func (f *unorderedArrayFoldIncr[A, B]) Node() *Node { return f.n }

func (f *unorderedArrayFoldIncr[A, B]) Parents() []INode {
	out := make([]INode, len(f.inputs))
	for index, input := range f.inputs {
		out[index] = input
	}
	return out
}

func (f *unorderedArrayFoldIncr[A, B]) Value() B { return f.value }

// ChildChanged records the delta for an input that just took a new value.
//
// The graph calls this while the input is recomputing, before this node is
// recomputed, so the deltas are queued rather than applied here.
func (f *unorderedArrayFoldIncr[A, B]) ChildChanged(child INode) {
	if !f.folded {
		// the initial fold reads every input directly, so deltas before it would be
		// applied to an accumulator that does not exist yet.
		return
	}
	slots, ok := f.slots[child.Node().id]
	if !ok {
		return
	}
	for _, slot := range slots {
		newValue := f.inputs[slot].Value()
		f.pending = append(f.pending, unorderedArrayFoldChange[A]{
			slot:     slot,
			oldValue: f.last[slot],
			newValue: newValue,
		})
		f.last[slot] = newValue
	}
}

func (f *unorderedArrayFoldIncr[A, B]) Stabilize(_ context.Context) error {
	if !f.folded {
		f.value = f.initial
		for index, input := range f.inputs {
			value := input.Value()
			f.last[index] = value
			f.value = f.fold(f.value, value)
		}
		f.folded = true
		f.pending = f.pending[:0]
		return nil
	}
	for _, change := range f.pending {
		f.value = f.update(f.value, change.oldValue, change.newValue)
	}
	f.pending = f.pending[:0]
	return nil
}

func (f *unorderedArrayFoldIncr[A, B]) String() string { return f.n.String() }
