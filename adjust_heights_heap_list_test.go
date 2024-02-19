package incr

import (
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_adjustHeightsHeapList_push_pop(t *testing.T) {
	g := New()
	q := new(adjustHeightsHeapList)

	n0 := newHeightIncr(g, 0)
	n1 := newHeightIncr(g, 0)
	n2 := newHeightIncr(g, 0)
	n3 := newHeightIncr(g, 0)

	var zeroID Identifier
	id, n, ok := q.pop()

	// Region: empty list
	{
		testutil.Equal(t, false, ok)
		testutil.Nil(t, q.head)
		testutil.Nil(t, q.tail)
		testutil.Equal(t, zeroID, id)
		testutil.Nil(t, n)
		testutil.Equal(t, 0, q.len())
		testutil.Equal(t, true, q.isEmpty())

		testutil.Equal(t, false, q.has(n0.Node().id))
		testutil.Equal(t, false, q.has(n1.Node().id))
		testutil.Equal(t, false, q.has(n2.Node().id))
		testutil.Equal(t, false, q.has(n3.Node().id))
	}

	// Region: push 0
	{
		q.push(n0)
		testutil.NotNil(t, q.head)
		testutil.Nil(t, q.head.Node().nextInAdjustHeightsHeap)
		testutil.NotNil(t, q.tail)
		testutil.Nil(t, q.tail.Node().previousInAdjustHeightsHeap)
		testutil.Equal(t, q.head, q.tail)
		testutil.Equal(t, 1, q.len())
		testutil.Equal(t, false, q.isEmpty())
		testutil.Equal(t, true, q.has(n0.Node().id))
		testutil.Equal(t, false, q.has(n1.Node().id))
		testutil.Equal(t, false, q.has(n2.Node().id))
		testutil.Equal(t, false, q.has(n3.Node().id))
		testutil.Equal(t, n0.n.id, q.head.Node().id)
		testutil.Equal(t, n0.n.id, q.tail.Node().id)

		testutil.Equal(t, true, q.has(n0.Node().id))
		testutil.Equal(t, false, q.has(n1.Node().id))
		testutil.Equal(t, false, q.has(n2.Node().id))
		testutil.Equal(t, false, q.has(n3.Node().id))
	}

	// Region: push 1
	{
		q.push(n1)
		testutil.NotNil(t, q.head)
		testutil.Nil(t, q.head.Node().previousInAdjustHeightsHeap)
		testutil.NotNil(t, q.head.Node().nextInAdjustHeightsHeap)
		testutil.Nil(t, q.head.Node().nextInAdjustHeightsHeap.Node().nextInAdjustHeightsHeap)
		testutil.NotNil(t, q.tail)
		testutil.NotNil(t, q.tail.Node().previousInAdjustHeightsHeap)
		testutil.Nil(t, q.tail.Node().previousInAdjustHeightsHeap.Node().previousInAdjustHeightsHeap)
		testutil.Nil(t, q.tail.Node().nextInAdjustHeightsHeap)
		testutil.NotEqual(t, q.head, q.tail)
		testutil.Equal(t, q.head.Node().nextInAdjustHeightsHeap, q.tail)
		testutil.Equal(t, q.tail.Node().previousInAdjustHeightsHeap, q.head)
		testutil.Equal(t, 2, q.len())
		testutil.Equal(t, false, q.isEmpty())
		testutil.Equal(t, n0.Node().id, q.head.Node().id)
		testutil.Equal(t, true, q.has(n0.Node().id))
		testutil.Equal(t, true, q.has(n1.Node().id))
		testutil.Equal(t, false, q.has(n2.Node().id))
		testutil.Equal(t, false, q.has(n3.Node().id))
		testutil.Equal(t, n1.Node().id, q.tail.Node().id)

		testutil.Equal(t, true, q.has(n0.Node().id))
		testutil.Equal(t, true, q.has(n1.Node().id))
		testutil.Equal(t, false, q.has(n2.Node().id))
		testutil.Equal(t, false, q.has(n3.Node().id))
	}

	// Region: push 2
	{
		q.push(n2)
		testutil.Nil(t, q.head.Node().previousInAdjustHeightsHeap)
		testutil.NotNil(t, q.head)
		testutil.NotNil(t, q.head.Node().nextInAdjustHeightsHeap)
		testutil.NotNil(t, q.head.Node().nextInAdjustHeightsHeap.Node().nextInAdjustHeightsHeap)
		testutil.Nil(t, q.head.Node().nextInAdjustHeightsHeap.Node().nextInAdjustHeightsHeap.Node().nextInAdjustHeightsHeap)
		testutil.Equal(t, q.head.Node().nextInAdjustHeightsHeap.Node().nextInAdjustHeightsHeap, q.tail)
		testutil.NotNil(t, q.tail)
		testutil.NotNil(t, q.tail.Node().previousInAdjustHeightsHeap)
		testutil.NotNil(t, q.tail.Node().previousInAdjustHeightsHeap.Node().previousInAdjustHeightsHeap)
		testutil.Nil(t, q.tail.Node().previousInAdjustHeightsHeap.Node().previousInAdjustHeightsHeap.Node().previousInAdjustHeightsHeap)
		testutil.Nil(t, q.tail.Node().nextInAdjustHeightsHeap)
		testutil.Equal(t, q.tail.Node().previousInAdjustHeightsHeap.Node().previousInAdjustHeightsHeap, q.head)
		testutil.NotEqual(t, q.head, q.tail)
		testutil.Equal(t, 3, q.len())
		testutil.Equal(t, true, q.has(n0.Node().id))
		testutil.Equal(t, true, q.has(n1.Node().id))
		testutil.Equal(t, true, q.has(n2.Node().id))
		testutil.Equal(t, false, q.has(n3.Node().id))
		testutil.Equal(t, n0.Node().id, q.head.Node().id)
		testutil.Equal(t, n2.Node().id, q.tail.Node().id)

		testutil.Equal(t, true, q.has(n0.Node().id))
		testutil.Equal(t, true, q.has(n1.Node().id))
		testutil.Equal(t, true, q.has(n2.Node().id))
		testutil.Equal(t, false, q.has(n3.Node().id))
	}

	// Region: push 3
	{
		q.push(n3)
		testutil.Nil(t, q.head.Node().previousInAdjustHeightsHeap)
		testutil.NotNil(t, q.head)
		testutil.NotNil(t, q.head.Node().nextInAdjustHeightsHeap)
		testutil.NotNil(t, q.head.Node().nextInAdjustHeightsHeap.Node().nextInAdjustHeightsHeap)
		testutil.NotNil(t, q.head.Node().nextInAdjustHeightsHeap.Node().nextInAdjustHeightsHeap.Node().nextInAdjustHeightsHeap)
		testutil.Nil(t, q.head.Node().nextInAdjustHeightsHeap.Node().nextInAdjustHeightsHeap.Node().nextInAdjustHeightsHeap.Node().nextInAdjustHeightsHeap)
		testutil.Equal(t, q.head.Node().nextInAdjustHeightsHeap.Node().nextInAdjustHeightsHeap.Node().nextInAdjustHeightsHeap, q.tail)
		testutil.NotNil(t, q.tail)
		testutil.NotNil(t, q.tail.Node().previousInAdjustHeightsHeap)
		testutil.NotNil(t, q.tail.Node().previousInAdjustHeightsHeap.Node().previousInAdjustHeightsHeap)
		testutil.NotNil(t, q.tail.Node().previousInAdjustHeightsHeap.Node().previousInAdjustHeightsHeap.Node().previousInAdjustHeightsHeap)
		testutil.Nil(t, q.tail.Node().previousInAdjustHeightsHeap.Node().previousInAdjustHeightsHeap.Node().previousInAdjustHeightsHeap.Node().previousInAdjustHeightsHeap)
		testutil.Equal(t, q.tail.Node().previousInAdjustHeightsHeap.Node().previousInAdjustHeightsHeap.Node().previousInAdjustHeightsHeap, q.head)
		testutil.Nil(t, q.tail.Node().nextInAdjustHeightsHeap)
		testutil.NotEqual(t, q.head, q.tail)
		testutil.Equal(t, 4, q.len())
		testutil.Equal(t, true, q.has(n0.Node().id))
		testutil.Equal(t, true, q.has(n1.Node().id))
		testutil.Equal(t, true, q.has(n2.Node().id))
		testutil.Equal(t, true, q.has(n3.Node().id))
		testutil.Equal(t, n0.Node().id, q.head.Node().id)
		testutil.Equal(t, n3.Node().id, q.tail.Node().id)

		testutil.Equal(t, true, q.has(n0.Node().id))
		testutil.Equal(t, true, q.has(n1.Node().id))
		testutil.Equal(t, true, q.has(n2.Node().id))
		testutil.Equal(t, true, q.has(n3.Node().id))
	}

	// Region: pop 0
	{
		id, n, ok = q.pop()
		testutil.Equal(t, true, ok)
		testutil.Equal(t, n0.n.id, id)
		testutil.Equal(t, n0.n.id, n.Node().id)
		testutil.NotNil(t, q.head)
		testutil.NotNil(t, q.head.Node().nextInAdjustHeightsHeap)
		testutil.NotNil(t, q.head.Node().nextInAdjustHeightsHeap.Node().nextInAdjustHeightsHeap)
		testutil.Nil(t, q.head.Node().nextInAdjustHeightsHeap.Node().nextInAdjustHeightsHeap.Node().nextInAdjustHeightsHeap)
		testutil.Equal(t, q.head.Node().nextInAdjustHeightsHeap.Node().nextInAdjustHeightsHeap, q.tail)
		testutil.NotNil(t, q.tail)
		testutil.NotEqual(t, q.head, q.tail)
		testutil.Equal(t, 3, q.len())
		testutil.Equal(t, false, q.has(n0.Node().id))
		testutil.Equal(t, true, q.has(n1.Node().id))
		testutil.Equal(t, true, q.has(n2.Node().id))
		testutil.Equal(t, true, q.has(n3.Node().id))
		testutil.Equal(t, n1.Node().id, q.head.Node().id)
		testutil.Equal(t, n3.Node().id, q.tail.Node().id)

		testutil.Equal(t, false, q.has(n0.Node().id))
		testutil.Equal(t, true, q.has(n1.Node().id))
		testutil.Equal(t, true, q.has(n2.Node().id))
		testutil.Equal(t, true, q.has(n3.Node().id))

		testutil.Nil(t, n.Node().nextInAdjustHeightsHeap)
		testutil.Nil(t, n.Node().previousInAdjustHeightsHeap)
	}

	// Region: pop 1
	{
		id, n, ok = q.pop()
		testutil.Equal(t, true, ok)
		testutil.Equal(t, n1.n.id, id)
		testutil.Equal(t, n1.n.id, n.Node().id)
		testutil.NotNil(t, q.head)
		testutil.NotNil(t, q.head.Node().nextInAdjustHeightsHeap)
		testutil.Nil(t, q.head.Node().nextInAdjustHeightsHeap.Node().nextInAdjustHeightsHeap)
		testutil.Equal(t, q.head.Node().nextInAdjustHeightsHeap, q.tail)
		testutil.NotNil(t, q.tail)
		testutil.NotNil(t, q.tail.Node().previousInAdjustHeightsHeap)
		testutil.Nil(t, q.tail.Node().previousInAdjustHeightsHeap.Node().previousInAdjustHeightsHeap)
		testutil.Equal(t, q.tail.Node().previousInAdjustHeightsHeap, q.head)
		testutil.NotEqual(t, q.head, q.tail)
		testutil.Equal(t, 2, q.len())
		testutil.Equal(t, false, q.has(n0.Node().id))
		testutil.Equal(t, false, q.has(n1.Node().id))
		testutil.Equal(t, true, q.has(n2.Node().id))
		testutil.Equal(t, true, q.has(n3.Node().id))
		testutil.Equal(t, n2.n.id, q.head.Node().id)
		testutil.Equal(t, n3.n.id, q.tail.Node().id)

		testutil.Equal(t, false, q.has(n0.Node().id))
		testutil.Equal(t, false, q.has(n1.Node().id))
		testutil.Equal(t, true, q.has(n2.Node().id))
		testutil.Equal(t, true, q.has(n3.Node().id))

		testutil.Nil(t, n.Node().nextInAdjustHeightsHeap)
		testutil.Nil(t, n.Node().previousInAdjustHeightsHeap)
	}

	// Region: pop 2
	{
		id, n, ok = q.pop()
		testutil.Equal(t, true, ok)
		testutil.Equal(t, n2.n.id, id)
		testutil.Equal(t, n2.n.id, n.Node().id)
		testutil.NotNil(t, q.head)
		testutil.Nil(t, q.head.Node().previousInAdjustHeightsHeap)
		testutil.NotNil(t, q.tail)
		testutil.Equal(t, q.head, q.tail)
		testutil.Equal(t, 1, q.len())
		testutil.Equal(t, n3.n.id, q.head.Node().id)
		testutil.Equal(t, n3.n.id, q.tail.Node().id)

		testutil.Equal(t, false, q.has(n0.Node().id))
		testutil.Equal(t, false, q.has(n1.Node().id))
		testutil.Equal(t, false, q.has(n2.Node().id))
		testutil.Equal(t, true, q.has(n3.Node().id))

		testutil.Nil(t, n.Node().nextInAdjustHeightsHeap)
		testutil.Nil(t, n.Node().previousInAdjustHeightsHeap)
	}

	// Region: pop 3
	{
		id, n, ok = q.pop()
		testutil.Equal(t, true, ok)
		testutil.Equal(t, n3.n.id, id)
		testutil.Equal(t, n3.n.id, n.Node().id)
		testutil.Nil(t, q.head)
		testutil.Nil(t, q.tail)
		testutil.Equal(t, 0, q.len())

		testutil.Equal(t, false, q.has(n0.Node().id))
		testutil.Equal(t, false, q.has(n1.Node().id))
		testutil.Equal(t, false, q.has(n2.Node().id))
		testutil.Equal(t, false, q.has(n3.Node().id))

		testutil.Nil(t, n.Node().nextInAdjustHeightsHeap)
		testutil.Nil(t, n.Node().previousInAdjustHeightsHeap)
	}

	// Region: pop empty
	{
		id, n, ok = q.pop()
		testutil.Equal(t, false, ok)
		testutil.Nil(t, n)
		testutil.Equal(t, zeroID, id)

		testutil.Equal(t, false, q.has(n0.Node().id))
		testutil.Equal(t, false, q.has(n1.Node().id))
		testutil.Equal(t, false, q.has(n2.Node().id))
		testutil.Equal(t, false, q.has(n3.Node().id))
	}

	q.push(n0)
	q.push(n1)
	q.push(n2)
	testutil.Equal(t, 3, q.len())
}

func Test_adjustHeightsHeapList_remove_0(t *testing.T) {
	g := New()
	q := new(adjustHeightsHeapList)

	n0 := newHeightIncr(g, 0)
	n1 := newHeightIncr(g, 1)
	n2 := newHeightIncr(g, 2)
	n3 := newHeightIncr(g, 3)
	n4 := newHeightIncr(g, 4)

	q.push(n0)
	q.push(n1)
	q.push(n2)
	q.push(n3)
	q.push(n4)

	testutil.Equal(t, q.head.Node().id, n0.n.id)
	testutil.NotNil(t, q.head.Node().nextInAdjustHeightsHeap)
	testutil.NotNil(t, q.tail.Node().previousInAdjustHeightsHeap)
	testutil.Equal(t, q.tail.Node().id, n4.n.id)

	ok := q.remove(n0.Node().id)

	testutil.Equal(t, true, ok)
	testutil.Equal(t, 4, q.len())
	testutil.NotNil(t, q.head)
	testutil.Equal(t, q.head.Node().id, n1.n.id)
	testutil.NotNil(t, q.tail)
	testutil.Equal(t, q.tail.Node().id, n4.n.id)

	testutil.Nil(t, n0.Node().nextInAdjustHeightsHeap)
	testutil.Nil(t, n0.Node().previousInAdjustHeightsHeap)

	testutil.Equal(t, n1.Node().id, q.head.Node().id)
	testutil.Equal(t, n2.Node().id, q.head.Node().nextInAdjustHeightsHeap.Node().id)
	testutil.Equal(t, n3.Node().id, q.head.Node().nextInAdjustHeightsHeap.Node().nextInAdjustHeightsHeap.Node().id)
	testutil.Equal(t, n4.Node().id, q.head.Node().nextInAdjustHeightsHeap.Node().nextInAdjustHeightsHeap.Node().nextInAdjustHeightsHeap.Node().id)

	testutil.Equal(t, n4.Node().id, q.tail.Node().id)
	testutil.Equal(t, n3.Node().id, q.tail.Node().previousInAdjustHeightsHeap.Node().id)
	testutil.Equal(t, n2.Node().id, q.tail.Node().previousInAdjustHeightsHeap.Node().previousInAdjustHeightsHeap.Node().id)
	testutil.Equal(t, n1.Node().id, q.tail.Node().previousInAdjustHeightsHeap.Node().previousInAdjustHeightsHeap.Node().previousInAdjustHeightsHeap.Node().id)
}

func Test_adjustHeightsHeapList_remove_1(t *testing.T) {
	g := New()
	q := new(adjustHeightsHeapList)

	n0 := newHeightIncr(g, 0)
	n1 := newHeightIncr(g, 1)
	n2 := newHeightIncr(g, 2)
	n3 := newHeightIncr(g, 3)
	n4 := newHeightIncr(g, 4)

	q.push(n0)
	q.push(n1)
	q.push(n2)
	q.push(n3)
	q.push(n4)

	testutil.Equal(t, q.head.Node().id, n0.n.id)
	testutil.NotNil(t, q.head.Node().nextInAdjustHeightsHeap)
	testutil.NotNil(t, q.tail.Node().previousInAdjustHeightsHeap)
	testutil.Equal(t, q.tail.Node().id, n4.n.id)

	ok := q.remove(n1.Node().id)

	testutil.Equal(t, true, ok)
	testutil.Equal(t, 4, q.len())
	testutil.NotNil(t, q.head)
	testutil.Equal(t, q.head.Node().id, n0.n.id)
	testutil.NotNil(t, q.tail)
	testutil.Equal(t, q.tail.Node().id, n4.n.id)

	testutil.Nil(t, n1.Node().nextInAdjustHeightsHeap)
	testutil.Nil(t, n1.Node().previousInAdjustHeightsHeap)

	testutil.Equal(t, n0.Node().id, q.head.Node().id)
	testutil.Equal(t, n2.Node().id, q.head.Node().nextInAdjustHeightsHeap.Node().id)
	testutil.Equal(t, n3.Node().id, q.head.Node().nextInAdjustHeightsHeap.Node().nextInAdjustHeightsHeap.Node().id)
	testutil.Equal(t, n4.Node().id, q.head.Node().nextInAdjustHeightsHeap.Node().nextInAdjustHeightsHeap.Node().nextInAdjustHeightsHeap.Node().id)

	testutil.Equal(t, n4.Node().id, q.tail.Node().id)
	testutil.Equal(t, n3.Node().id, q.tail.Node().previousInAdjustHeightsHeap.Node().id)
	testutil.Equal(t, n2.Node().id, q.tail.Node().previousInAdjustHeightsHeap.Node().previousInAdjustHeightsHeap.Node().id)
	testutil.Equal(t, n0.Node().id, q.tail.Node().previousInAdjustHeightsHeap.Node().previousInAdjustHeightsHeap.Node().previousInAdjustHeightsHeap.Node().id)
}

func Test_adjustHeightsHeapList_remove_2(t *testing.T) {
	g := New()
	q := new(adjustHeightsHeapList)

	n0 := newHeightIncr(g, 0)
	n1 := newHeightIncr(g, 1)
	n2 := newHeightIncr(g, 2)
	n3 := newHeightIncr(g, 3)
	n4 := newHeightIncr(g, 4)

	q.push(n0)
	q.push(n1)
	q.push(n2)
	q.push(n3)
	q.push(n4)

	testutil.Equal(t, q.head.Node().id, n0.n.id)
	testutil.NotNil(t, q.head.Node().nextInAdjustHeightsHeap)
	testutil.NotNil(t, q.tail.Node().previousInAdjustHeightsHeap)
	testutil.Equal(t, q.tail.Node().id, n4.n.id)

	ok := q.remove(n2.Node().id)

	testutil.Equal(t, true, ok)
	testutil.Equal(t, 4, q.len())
	testutil.NotNil(t, q.head)
	testutil.Equal(t, q.head.Node().id, n0.n.id)
	testutil.NotNil(t, q.tail)
	testutil.Equal(t, q.tail.Node().id, n4.n.id)

	testutil.Nil(t, n2.Node().nextInAdjustHeightsHeap)
	testutil.Nil(t, n2.Node().previousInAdjustHeightsHeap)

	testutil.Equal(t, n0.Node().id, q.head.Node().id)
	testutil.Equal(t, n1.Node().id, q.head.Node().nextInAdjustHeightsHeap.Node().id)
	testutil.Equal(t, n3.Node().id, q.head.Node().nextInAdjustHeightsHeap.Node().nextInAdjustHeightsHeap.Node().id)
	testutil.Equal(t, n4.Node().id, q.head.Node().nextInAdjustHeightsHeap.Node().nextInAdjustHeightsHeap.Node().nextInAdjustHeightsHeap.Node().id)

	testutil.Equal(t, n4.Node().id, q.tail.Node().id)
	testutil.Equal(t, n3.Node().id, q.tail.Node().previousInAdjustHeightsHeap.Node().id)
	testutil.Equal(t, n1.Node().id, q.tail.Node().previousInAdjustHeightsHeap.Node().previousInAdjustHeightsHeap.Node().id)
	testutil.Equal(t, n0.Node().id, q.tail.Node().previousInAdjustHeightsHeap.Node().previousInAdjustHeightsHeap.Node().previousInAdjustHeightsHeap.Node().id)
}

func Test_adjustHeightsHeapList_remove_3(t *testing.T) {
	g := New()
	q := new(adjustHeightsHeapList)

	n0 := newHeightIncr(g, 0)
	n1 := newHeightIncr(g, 1)
	n2 := newHeightIncr(g, 2)
	n3 := newHeightIncr(g, 3)
	n4 := newHeightIncr(g, 4)

	q.push(n0)
	q.push(n1)
	q.push(n2)
	q.push(n3)
	q.push(n4)

	testutil.Equal(t, q.head.Node().id, n0.n.id)
	testutil.NotNil(t, q.head.Node().nextInAdjustHeightsHeap)
	testutil.NotNil(t, q.tail.Node().previousInAdjustHeightsHeap)
	testutil.Equal(t, q.tail.Node().id, n4.n.id)

	ok := q.remove(n3.Node().id)

	testutil.Equal(t, true, ok)
	testutil.Equal(t, 4, q.len())
	testutil.NotNil(t, q.head)
	testutil.Equal(t, q.head.Node().id, n0.n.id)
	testutil.NotNil(t, q.tail)
	testutil.Equal(t, q.tail.Node().id, n4.n.id)

	testutil.Nil(t, n3.Node().nextInAdjustHeightsHeap)
	testutil.Nil(t, n3.Node().previousInAdjustHeightsHeap)

	testutil.Equal(t, n0.Node().id, q.head.Node().id)
	testutil.Equal(t, n1.Node().id, q.head.Node().nextInAdjustHeightsHeap.Node().id)
	testutil.Equal(t, n2.Node().id, q.head.Node().nextInAdjustHeightsHeap.Node().nextInAdjustHeightsHeap.Node().id)
	testutil.Equal(t, n4.Node().id, q.head.Node().nextInAdjustHeightsHeap.Node().nextInAdjustHeightsHeap.Node().nextInAdjustHeightsHeap.Node().id)

	testutil.Equal(t, n4.Node().id, q.tail.Node().id)
	testutil.Equal(t, n2.Node().id, q.tail.Node().previousInAdjustHeightsHeap.Node().id)
	testutil.Equal(t, n1.Node().id, q.tail.Node().previousInAdjustHeightsHeap.Node().previousInAdjustHeightsHeap.Node().id)
	testutil.Equal(t, n0.Node().id, q.tail.Node().previousInAdjustHeightsHeap.Node().previousInAdjustHeightsHeap.Node().previousInAdjustHeightsHeap.Node().id)
}

func Test_adjustHeightsHeapList_remove_4(t *testing.T) {
	g := New()
	q := new(adjustHeightsHeapList)

	n0 := newHeightIncr(g, 0)
	n1 := newHeightIncr(g, 1)
	n2 := newHeightIncr(g, 2)
	n3 := newHeightIncr(g, 3)
	n4 := newHeightIncr(g, 4)

	q.push(n0)
	q.push(n1)
	q.push(n2)
	q.push(n3)
	q.push(n4)

	testutil.Equal(t, q.head.Node().id, n0.n.id)
	testutil.NotNil(t, q.head.Node().nextInAdjustHeightsHeap)
	testutil.NotNil(t, q.tail.Node().previousInAdjustHeightsHeap)
	testutil.Equal(t, q.tail.Node().id, n4.n.id)

	ok := q.remove(n4.Node().id)

	testutil.Equal(t, true, ok)
	testutil.Equal(t, 4, q.len())
	testutil.NotNil(t, q.head)
	testutil.Equal(t, q.head.Node().id, n0.n.id)
	testutil.NotNil(t, q.tail)
	testutil.Equal(t, q.tail.Node().id, n3.n.id)

	testutil.Nil(t, n4.Node().nextInAdjustHeightsHeap)
	testutil.Nil(t, n4.Node().previousInAdjustHeightsHeap)

	testutil.Equal(t, n0.Node().id, q.head.Node().id)
	testutil.Equal(t, n1.Node().id, q.head.Node().nextInAdjustHeightsHeap.Node().id)
	testutil.Equal(t, n2.Node().id, q.head.Node().nextInAdjustHeightsHeap.Node().nextInAdjustHeightsHeap.Node().id)
	testutil.Equal(t, n3.Node().id, q.head.Node().nextInAdjustHeightsHeap.Node().nextInAdjustHeightsHeap.Node().nextInAdjustHeightsHeap.Node().id)

	testutil.Equal(t, n3.Node().id, q.tail.Node().id)
	testutil.Equal(t, n2.Node().id, q.tail.Node().previousInAdjustHeightsHeap.Node().id)
	testutil.Equal(t, n1.Node().id, q.tail.Node().previousInAdjustHeightsHeap.Node().previousInAdjustHeightsHeap.Node().id)
	testutil.Equal(t, n0.Node().id, q.tail.Node().previousInAdjustHeightsHeap.Node().previousInAdjustHeightsHeap.Node().previousInAdjustHeightsHeap.Node().id)
}

func Test_adjustHeightsHeapList_remove_empty(t *testing.T) {
	q := new(adjustHeightsHeapList)

	ok := q.remove(NewIdentifier())
	testutil.Equal(t, false, ok)
}

func Test_adjustHeightsHeapList_remove_notFound(t *testing.T) {
	g := New()
	q := new(adjustHeightsHeapList)

	n0 := newHeightIncr(g, 0)
	n1 := newHeightIncr(g, 1)
	n2 := newHeightIncr(g, 2)
	n3 := newHeightIncr(g, 3)

	q.push(n0)
	q.push(n1)
	q.push(n2)
	q.push(n3)

	testutil.Equal(t, q.head.Node().id, n0.n.id)
	testutil.NotNil(t, q.head.Node().nextInAdjustHeightsHeap)
	testutil.NotNil(t, q.tail.Node().previousInAdjustHeightsHeap)
	testutil.Equal(t, q.tail.Node().id, n3.n.id)

	ok := q.remove(NewIdentifier())
	testutil.Equal(t, false, ok)
}

func Test_adjustHeightsHeapList_remove_head(t *testing.T) {
	g := New()
	q := new(adjustHeightsHeapList)

	n0 := newHeightIncr(g, 0)
	n1 := newHeightIncr(g, 1)
	n2 := newHeightIncr(g, 2)
	n3 := newHeightIncr(g, 3)

	q.push(n0)
	q.push(n1)
	q.push(n2)
	q.push(n3)

	testutil.Equal(t, q.head.Node().id, n0.n.id)
	testutil.NotNil(t, q.head.Node().nextInAdjustHeightsHeap)
	testutil.NotNil(t, q.tail.Node().previousInAdjustHeightsHeap)
	testutil.Equal(t, q.tail.Node().id, n3.n.id)

	ok := q.remove(n0.Node().id)
	testutil.Equal(t, ok, true)
	testutil.NotNil(t, q.head)
	testutil.Equal(t, q.head.Node().id, n1.n.id)

	testutil.Nil(t, n0.Node().nextInAdjustHeightsHeap)
	testutil.Nil(t, n0.Node().previousInAdjustHeightsHeap)
}

func Test_adjustHeightsHeapList_remove_tail(t *testing.T) {
	g := New()
	q := new(adjustHeightsHeapList)

	n0 := newHeightIncr(g, 0)
	n1 := newHeightIncr(g, 1)
	n2 := newHeightIncr(g, 2)
	n3 := newHeightIncr(g, 3)

	q.push(n0)
	q.push(n1)
	q.push(n2)
	q.push(n3)

	testutil.Equal(t, q.head.Node().id, n0.n.id)
	testutil.NotNil(t, q.head.Node().nextInAdjustHeightsHeap)
	testutil.NotNil(t, q.tail.Node().previousInAdjustHeightsHeap)
	testutil.Equal(t, q.tail.Node().id, n3.n.id)

	ok := q.remove(n3.Node().id)
	testutil.Equal(t, ok, true)
	testutil.NotNil(t, q.tail)
	testutil.Equal(t, q.tail.Node().id, n2.n.id)

	testutil.Nil(t, n3.Node().nextInAdjustHeightsHeap)
	testutil.Nil(t, n3.Node().previousInAdjustHeightsHeap)
}

func Test_adjustHeightsHeapList_removeHeadItem_full(t *testing.T) {
	g := New()
	q := new(adjustHeightsHeapList)

	n0 := newHeightIncr(g, 0)
	n1 := newHeightIncr(g, 1)
	n2 := newHeightIncr(g, 2)
	n3 := newHeightIncr(g, 3)
	n4 := newHeightIncr(g, 4)

	q.push(n0)
	q.push(n1)
	q.push(n2)
	q.push(n3)
	q.push(n4)

	testutil.NotNil(t, q.head)
	testutil.NotNil(t, q.head.Node().nextInAdjustHeightsHeap)
	testutil.Equal(t, n1.Node().id, q.head.Node().nextInAdjustHeightsHeap.Node().id)
	testutil.Equal(t, n0.Node().id, q.head.Node().nextInAdjustHeightsHeap.Node().previousInAdjustHeightsHeap.Node().id)

	testutil.NotNil(t, q.tail)
	testutil.Equal(t, n4.Node().id, q.tail.Node().id)
	testutil.NotNil(t, q.tail.Node().previousInAdjustHeightsHeap)
	testutil.Equal(t, n3.Node().id, q.tail.Node().previousInAdjustHeightsHeap.Node().id)

	q.removeHeadItem()

	testutil.NotNil(t, q.head)
	testutil.NotNil(t, q.head.Node().nextInAdjustHeightsHeap)
	testutil.Equal(t, n2.Node().id, q.head.Node().nextInAdjustHeightsHeap.Node().id)
	testutil.Equal(t, n1.Node().id, q.head.Node().nextInAdjustHeightsHeap.Node().previousInAdjustHeightsHeap.Node().id)

	testutil.Nil(t, n0.Node().nextInAdjustHeightsHeap)
	testutil.Nil(t, n0.Node().previousInAdjustHeightsHeap)
}

func Test_adjustHeightsHeapList_removeHeadItem_one(t *testing.T) {
	g := New()
	q := new(adjustHeightsHeapList)

	n0 := newHeightIncr(g, 0)

	q.push(n0)

	testutil.NotNil(t, q.head)
	testutil.Nil(t, q.head.Node().nextInAdjustHeightsHeap)
	testutil.NotNil(t, q.tail)
	testutil.Nil(t, q.tail.Node().previousInAdjustHeightsHeap)

	q.removeHeadItem()

	testutil.Nil(t, q.head)
	testutil.Nil(t, q.tail)
}
