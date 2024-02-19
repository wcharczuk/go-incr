package incr

import (
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

func nodePtrID(n *INode) Identifier {
	if n == nil {
		return zero
	}
	return (*n).Node().id
}

func rhnext(n *INode) *INode {
	if n == nil {
		return nil
	}
	return (*n).Node().nextInRecomputeHeap
}

func rhprev(n *INode) *INode {
	if n == nil {
		return nil
	}
	return (*n).Node().previousInRecomputeHeap
}

func Test_recomputeHeapList_push_pop(t *testing.T) {
	g := New()
	q := new(recomputeHeapList)

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

		testutil.Equal(t, false, q.has(n0.Node().id))
		testutil.Equal(t, false, q.has(n1.Node().id))
		testutil.Equal(t, false, q.has(n2.Node().id))
		testutil.Equal(t, false, q.has(n3.Node().id))
	}

	// Region: push 0
	{
		q.push(n0)
		testutil.NotNil(t, q.head)
		testutil.Nil(t, rhnext(q.head))
		testutil.NotNil(t, q.tail)
		testutil.Nil(t, rhprev(q.tail))
		testutil.Equal(t, q.head, q.tail)
		testutil.Equal(t, 1, q.len())
		testutil.Equal(t, true, q.has(n0.Node().id))
		testutil.Equal(t, false, q.has(n1.Node().id))
		testutil.Equal(t, false, q.has(n2.Node().id))
		testutil.Equal(t, false, q.has(n3.Node().id))
		testutil.Equal(t, n0.n.id, nodePtrID(q.head))
		testutil.Equal(t, n0.n.id, nodePtrID(q.tail))

		testutil.Equal(t, true, q.has(n0.Node().id))
		testutil.Equal(t, false, q.has(n1.Node().id))
		testutil.Equal(t, false, q.has(n2.Node().id))
		testutil.Equal(t, false, q.has(n3.Node().id))
	}

	// Region: push 1
	{
		q.push(n1)
		testutil.NotNil(t, q.head)
		testutil.Nil(t, rhprev(q.head))
		testutil.NotNil(t, rhnext(q.head))
		testutil.Nil(t, rhnext(rhnext(q.head)))
		testutil.NotNil(t, q.tail)
		testutil.NotNil(t, rhprev(q.tail))
		testutil.Nil(t, rhprev(rhprev(q.tail)))
		testutil.Nil(t, rhnext(q.tail))
		testutil.NotEqual(t, q.head, q.tail)
		testutil.Equal(t, rhnext(q.head), q.tail)
		testutil.Equal(t, rhprev(q.tail), q.head)
		testutil.Equal(t, 2, q.len())
		testutil.Equal(t, n0.Node().id, nodePtrID(q.head))
		testutil.Equal(t, true, q.has(n0.Node().id))
		testutil.Equal(t, true, q.has(n1.Node().id))
		testutil.Equal(t, false, q.has(n2.Node().id))
		testutil.Equal(t, false, q.has(n3.Node().id))
		testutil.Equal(t, n1.Node().id, nodePtrID(q.tail))

		testutil.Equal(t, true, q.has(n0.Node().id))
		testutil.Equal(t, true, q.has(n1.Node().id))
		testutil.Equal(t, false, q.has(n2.Node().id))
		testutil.Equal(t, false, q.has(n3.Node().id))
	}

	// Region: push 2
	{
		q.push(n2)
		testutil.NotNil(t, q.head)
		testutil.Nil(t, rhprev(q.head))
		testutil.NotNil(t, rhnext(q.head))
		testutil.NotNil(t, rhnext(rhnext(q.head)))
		testutil.Nil(t, rhnext(rhnext(rhnext(q.head))))
		testutil.Equal(t, rhnext(rhnext(q.head)), q.tail)
		testutil.NotNil(t, q.tail)
		testutil.NotNil(t, rhprev(q.tail))
		testutil.NotNil(t, rhprev(rhprev(q.tail)))
		testutil.Nil(t, rhprev(rhprev(rhprev(q.tail))))
		testutil.Nil(t, rhnext(q.tail))
		testutil.Equal(t, rhprev(rhprev(q.tail)), q.head)
		testutil.NotEqual(t, q.head, q.tail)
		testutil.Equal(t, 3, q.len())
		testutil.Equal(t, true, q.has(n0.Node().id))
		testutil.Equal(t, true, q.has(n1.Node().id))
		testutil.Equal(t, true, q.has(n2.Node().id))
		testutil.Equal(t, false, q.has(n3.Node().id))
		testutil.Equal(t, n0.Node().id, nodePtrID(q.head))
		testutil.Equal(t, n2.Node().id, nodePtrID(q.tail))

		testutil.Equal(t, true, q.has(n0.Node().id))
		testutil.Equal(t, true, q.has(n1.Node().id))
		testutil.Equal(t, true, q.has(n2.Node().id))
		testutil.Equal(t, false, q.has(n3.Node().id))
	}

	// Region: push 3
	{
		q.push(n3)
		testutil.NotNil(t, q.head)
		testutil.Nil(t, rhprev(q.head))
		testutil.NotNil(t, rhnext(q.head))
		testutil.NotNil(t, rhnext(rhnext(q.head)))
		testutil.NotNil(t, rhnext(rhnext(rhnext(q.head))))
		testutil.Nil(t, rhnext(rhnext(rhnext(rhnext(q.head)))))
		testutil.Equal(t, rhnext(rhnext(rhnext(q.head))), q.tail)
		testutil.NotNil(t, q.tail)
		testutil.NotNil(t, rhprev(q.tail))
		testutil.NotNil(t, rhprev(rhprev(q.tail)))
		testutil.NotNil(t, rhprev(rhprev(rhprev(q.tail))))
		testutil.Nil(t, rhprev(rhprev(rhprev(rhprev(q.tail)))))
		testutil.Nil(t, rhnext(q.tail))
		testutil.Equal(t, rhprev(rhprev(rhprev(q.tail))), q.head)
		testutil.NotEqual(t, q.head, q.tail)
		testutil.Equal(t, 4, q.len())
		testutil.Equal(t, true, q.has(n0.Node().id))
		testutil.Equal(t, true, q.has(n1.Node().id))
		testutil.Equal(t, true, q.has(n2.Node().id))
		testutil.Equal(t, true, q.has(n3.Node().id))
		testutil.Equal(t, n0.Node().id, nodePtrID(q.head))
		testutil.Equal(t, n3.Node().id, nodePtrID(q.tail))
	}

	// Region: pop 0
	{
		id, n, ok = q.pop()
		testutil.Equal(t, true, ok)
		testutil.Equal(t, n0.n.id, id)
		testutil.Equal(t, n0.n.id, n.Node().id)
		testutil.NotNil(t, q.head)
		testutil.NotNil(t, rhnext(q.head))
		testutil.NotNil(t, rhnext(rhnext(q.head)))
		testutil.Nil(t, rhnext(rhnext(rhnext(q.head))))
		testutil.Equal(t, rhnext(rhnext(q.head)), q.tail)
		testutil.NotNil(t, q.tail)
		testutil.NotEqual(t, q.head, q.tail)
		testutil.Equal(t, 3, q.len())
		testutil.Equal(t, false, q.has(n0.Node().id))
		testutil.Equal(t, true, q.has(n1.Node().id))
		testutil.Equal(t, true, q.has(n2.Node().id))
		testutil.Equal(t, true, q.has(n3.Node().id))
		testutil.Equal(t, n1.Node().id, nodePtrID(q.head))
		testutil.Equal(t, n3.Node().id, nodePtrID(q.tail))

		testutil.Nil(t, n.Node().nextInRecomputeHeap)
		testutil.Nil(t, n.Node().previousInRecomputeHeap)
	}

	// Region: pop 1
	{
		id, n, ok = q.pop()
		testutil.Equal(t, true, ok)
		testutil.Equal(t, n1.n.id, id)
		testutil.Equal(t, n1.n.id, n.Node().id)
		testutil.NotNil(t, q.head)
		testutil.NotNil(t, rhnext(q.head))
		testutil.Nil(t, rhnext(rhnext(q.head)))
		testutil.Equal(t, rhnext(q.head), q.tail)
		testutil.NotNil(t, q.tail)
		testutil.NotNil(t, rhprev(q.tail))
		testutil.Nil(t, rhprev(rhprev(q.tail)))
		testutil.Equal(t, rhprev(q.tail), q.head)
		testutil.NotEqual(t, q.head, q.tail)
		testutil.Equal(t, 2, q.len())
		testutil.Equal(t, false, q.has(n0.Node().id))
		testutil.Equal(t, false, q.has(n1.Node().id))
		testutil.Equal(t, true, q.has(n2.Node().id))
		testutil.Equal(t, true, q.has(n3.Node().id))
		testutil.Equal(t, n2.n.id, nodePtrID(q.head))
		testutil.Equal(t, n3.n.id, nodePtrID(q.tail))

		testutil.Nil(t, n.Node().nextInRecomputeHeap)
		testutil.Nil(t, n.Node().previousInRecomputeHeap)
	}

	// Region: pop 2
	{
		id, n, ok = q.pop()
		testutil.Equal(t, true, ok)
		testutil.Equal(t, n2.n.id, id)
		testutil.Equal(t, n2.n.id, n.Node().id)
		testutil.NotNil(t, q.head)
		testutil.Nil(t, rhnext(q.head))
		testutil.NotNil(t, q.tail)
		testutil.Equal(t, q.head, q.tail)
		testutil.Equal(t, 1, q.len())
		testutil.Equal(t, n3.n.id, nodePtrID(q.head))
		testutil.Equal(t, n3.n.id, nodePtrID(q.tail))

		testutil.Equal(t, false, q.has(n0.Node().id))
		testutil.Equal(t, false, q.has(n1.Node().id))
		testutil.Equal(t, false, q.has(n2.Node().id))
		testutil.Equal(t, true, q.has(n3.Node().id))

		testutil.Nil(t, n.Node().nextInRecomputeHeap)
		testutil.Nil(t, n.Node().previousInRecomputeHeap)
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

		testutil.Nil(t, n.Node().nextInRecomputeHeap)
		testutil.Nil(t, n.Node().previousInRecomputeHeap)
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

func Test_recomputeHeapList_remove_0(t *testing.T) {
	g := New()
	q := new(recomputeHeapList)

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

	testutil.NotNil(t, rhnext(q.head))
	testutil.Equal(t, nodePtrID(q.head), n0.n.id)
	testutil.NotNil(t, rhprev(q.tail))
	testutil.Equal(t, nodePtrID(q.tail), n4.n.id)

	ok := q.remove(n0.Node().id)

	testutil.Equal(t, true, ok)
	testutil.Equal(t, 4, q.len())
	testutil.NotNil(t, q.head)
	testutil.Equal(t, nodePtrID(q.head), n1.n.id)
	testutil.NotNil(t, q.tail)
	testutil.Equal(t, nodePtrID(q.tail), n4.n.id)

	testutil.Nil(t, n0.Node().nextInRecomputeHeap)
	testutil.Nil(t, n0.Node().previousInRecomputeHeap)

	testutil.Equal(t, n1.Node().id, nodePtrID(q.head))
	testutil.Equal(t, n2.Node().id, nodePtrID(rhnext(q.head)))
	testutil.Equal(t, n3.Node().id, nodePtrID(rhnext(rhnext(q.head))))
	testutil.Equal(t, n4.Node().id, nodePtrID(rhnext(rhnext(rhnext(q.head)))))

	testutil.Equal(t, n4.Node().id, nodePtrID(q.tail))
	testutil.Equal(t, n3.Node().id, nodePtrID(rhprev(q.tail)))
	testutil.Equal(t, n2.Node().id, nodePtrID(rhprev(rhprev(q.tail))))
	testutil.Equal(t, n1.Node().id, nodePtrID(rhprev(rhprev(rhprev(q.tail)))))
}

func Test_recomputeHeapList_remove_1(t *testing.T) {
	g := New()
	q := new(recomputeHeapList)

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

	testutil.NotNil(t, rhnext(q.head))
	testutil.Equal(t, nodePtrID(q.head), n0.n.id)
	testutil.NotNil(t, rhprev(q.tail))
	testutil.Equal(t, nodePtrID(q.tail), n4.n.id)

	ok := q.remove(n1.Node().id)

	testutil.Equal(t, true, ok)
	testutil.Equal(t, 4, q.len())
	testutil.NotNil(t, q.head)
	testutil.Equal(t, nodePtrID(q.head), n0.n.id)
	testutil.NotNil(t, q.tail)
	testutil.Equal(t, nodePtrID(q.tail), n4.n.id)

	testutil.Nil(t, n1.Node().nextInRecomputeHeap)
	testutil.Nil(t, n1.Node().previousInRecomputeHeap)

	testutil.Equal(t, n0.Node().id, nodePtrID(q.head))
	testutil.Equal(t, n2.Node().id, nodePtrID(rhnext(q.head)))
	testutil.Equal(t, n3.Node().id, nodePtrID(rhnext(rhnext(q.head))))
	testutil.Equal(t, n4.Node().id, nodePtrID(rhnext(rhnext(rhnext(q.head)))))

	testutil.Equal(t, n4.Node().id, nodePtrID(q.tail))
	testutil.Equal(t, n3.Node().id, nodePtrID(rhprev(q.tail)))
	testutil.Equal(t, n2.Node().id, nodePtrID(rhprev(rhprev(q.tail))))
	testutil.Equal(t, n0.Node().id, nodePtrID(rhprev(rhprev(rhprev(q.tail)))))
}

func Test_recomputeHeapList_remove_2(t *testing.T) {
	g := New()
	q := new(recomputeHeapList)

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

	testutil.NotNil(t, rhnext(q.head))
	testutil.Equal(t, nodePtrID(q.head), n0.n.id)
	testutil.NotNil(t, rhprev(q.tail))
	testutil.Equal(t, nodePtrID(q.tail), n4.n.id)

	ok := q.remove(n2.Node().id)

	testutil.Equal(t, true, ok)
	testutil.Equal(t, 4, q.len())
	testutil.NotNil(t, q.head)
	testutil.Equal(t, nodePtrID(q.head), n0.n.id)
	testutil.NotNil(t, q.tail)
	testutil.Equal(t, nodePtrID(q.tail), n4.n.id)

	testutil.Nil(t, n2.Node().nextInRecomputeHeap)
	testutil.Nil(t, n2.Node().previousInRecomputeHeap)

	testutil.Equal(t, n0.Node().id, nodePtrID(q.head))
	testutil.Equal(t, n1.Node().id, nodePtrID(rhnext(q.head)))
	testutil.Equal(t, n3.Node().id, nodePtrID(rhnext(rhnext(q.head))))
	testutil.Equal(t, n4.Node().id, nodePtrID(rhnext(rhnext(rhnext(q.head)))))

	testutil.Equal(t, n4.Node().id, nodePtrID(q.tail))
	testutil.Equal(t, n3.Node().id, nodePtrID(rhprev(q.tail)))
	testutil.Equal(t, n1.Node().id, nodePtrID(rhprev(rhprev(q.tail))))
	testutil.Equal(t, n0.Node().id, nodePtrID(rhprev(rhprev(rhprev(q.tail)))))
}

func Test_recomputeHeapList_remove_3(t *testing.T) {
	g := New()
	q := new(recomputeHeapList)

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

	testutil.NotNil(t, rhnext(q.head))
	testutil.Equal(t, nodePtrID(q.head), n0.n.id)
	testutil.NotNil(t, rhprev(q.tail))
	testutil.Equal(t, nodePtrID(q.tail), n4.n.id)

	ok := q.remove(n3.Node().id)

	testutil.Equal(t, true, ok)
	testutil.Equal(t, 4, q.len())
	testutil.NotNil(t, q.head)
	testutil.Equal(t, nodePtrID(q.head), n0.n.id)
	testutil.NotNil(t, q.tail)
	testutil.Equal(t, nodePtrID(q.tail), n4.n.id)

	testutil.Nil(t, n3.Node().nextInRecomputeHeap)
	testutil.Nil(t, n3.Node().previousInRecomputeHeap)

	testutil.Equal(t, n0.Node().id, nodePtrID(q.head))
	testutil.Equal(t, n1.Node().id, nodePtrID(rhnext(q.head)))
	testutil.Equal(t, n2.Node().id, nodePtrID(rhnext(rhnext(q.head))))
	testutil.Equal(t, n4.Node().id, nodePtrID(rhnext(rhnext(rhnext(q.head)))))

	testutil.Equal(t, n4.Node().id, nodePtrID(q.tail))
	testutil.Equal(t, n2.Node().id, nodePtrID(rhprev(q.tail)))
	testutil.Equal(t, n1.Node().id, nodePtrID(rhprev(rhprev(q.tail))))
	testutil.Equal(t, n0.Node().id, nodePtrID(rhprev(rhprev(rhprev(q.tail)))))
}

func Test_recomputeHeapList_remove_4(t *testing.T) {
	g := New()
	q := new(recomputeHeapList)

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

	testutil.NotNil(t, rhnext(q.head))
	testutil.Equal(t, nodePtrID(q.head), n0.n.id)
	testutil.NotNil(t, rhprev(q.tail))
	testutil.Equal(t, nodePtrID(q.tail), n4.n.id)

	ok := q.remove(n4.Node().id)

	testutil.Equal(t, true, ok)
	testutil.Equal(t, 4, q.len())
	testutil.NotNil(t, q.head)
	testutil.Equal(t, nodePtrID(q.head), n0.n.id)
	testutil.NotNil(t, q.tail)
	testutil.Equal(t, nodePtrID(q.tail), n3.n.id)

	testutil.Nil(t, n4.Node().nextInRecomputeHeap)
	testutil.Nil(t, n4.Node().previousInRecomputeHeap)

	testutil.Equal(t, n0.Node().id, nodePtrID(q.head))
	testutil.Equal(t, n1.Node().id, nodePtrID(rhnext(q.head)))
	testutil.Equal(t, n2.Node().id, nodePtrID(rhnext(rhnext(q.head))))
	testutil.Equal(t, n3.Node().id, nodePtrID(rhnext(rhnext(rhnext(q.head)))))

	testutil.Equal(t, n3.Node().id, nodePtrID(q.tail))
	testutil.Equal(t, n2.Node().id, nodePtrID(rhprev(q.tail)))
	testutil.Equal(t, n1.Node().id, nodePtrID(rhprev(rhprev(q.tail))))
	testutil.Equal(t, n0.Node().id, nodePtrID(rhprev(rhprev(rhprev(q.tail)))))
}

func Test_recomputeHeapList_remove_empty(t *testing.T) {
	q := new(recomputeHeapList)

	ok := q.remove(NewIdentifier())
	testutil.Equal(t, false, ok)
}

func Test_recomputeHeapList_remove_notFound(t *testing.T) {
	g := New()
	q := new(recomputeHeapList)

	n0 := newHeightIncr(g, 0)
	n1 := newHeightIncr(g, 1)
	n2 := newHeightIncr(g, 2)
	n3 := newHeightIncr(g, 3)

	q.push(n0)
	q.push(n1)
	q.push(n2)
	q.push(n3)

	testutil.NotNil(t, rhnext(q.head))
	testutil.Equal(t, nodePtrID(q.head), n0.n.id)
	testutil.NotNil(t, rhprev(q.tail))
	testutil.Equal(t, nodePtrID(q.tail), n3.n.id)

	ok := q.remove(NewIdentifier())
	testutil.Equal(t, false, ok)
}

func Test_recomputeHeapList_remove_head(t *testing.T) {
	g := New()
	q := new(recomputeHeapList)

	n0 := newHeightIncr(g, 0)
	n1 := newHeightIncr(g, 1)
	n2 := newHeightIncr(g, 2)
	n3 := newHeightIncr(g, 3)

	q.push(n0)
	q.push(n1)
	q.push(n2)
	q.push(n3)

	testutil.NotNil(t, rhnext(q.head))
	testutil.Equal(t, nodePtrID(q.head), n0.n.id)
	testutil.NotNil(t, rhprev(q.tail))
	testutil.Equal(t, nodePtrID(q.tail), n3.n.id)

	ok := q.remove(n0.Node().id)
	testutil.Equal(t, ok, true)
	testutil.NotNil(t, q.head)
	testutil.Equal(t, nodePtrID(q.head), n1.n.id)

	testutil.Nil(t, n0.Node().nextInRecomputeHeap)
	testutil.Nil(t, n0.Node().previousInRecomputeHeap)
}

func Test_recomputeHeapList_remove_tail(t *testing.T) {
	g := New()
	q := new(recomputeHeapList)

	n0 := newHeightIncr(g, 0)
	n1 := newHeightIncr(g, 1)
	n2 := newHeightIncr(g, 2)
	n3 := newHeightIncr(g, 3)

	q.push(n0)
	q.push(n1)
	q.push(n2)
	q.push(n3)

	testutil.NotNil(t, rhnext(q.head))
	testutil.Equal(t, nodePtrID(q.head), n0.n.id)
	testutil.NotNil(t, rhprev(q.tail))
	testutil.Equal(t, nodePtrID(q.tail), n3.n.id)

	ok := q.remove(n3.Node().id)
	testutil.Equal(t, ok, true)
	testutil.NotNil(t, q.tail)
	testutil.Equal(t, nodePtrID(q.tail), n2.n.id)

	testutil.Nil(t, n3.Node().nextInRecomputeHeap)
	testutil.Nil(t, n3.Node().previousInRecomputeHeap)
}

func Test_recomputeHeapList_consume(t *testing.T) {
	g := New()
	q := new(recomputeHeapList)

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

	var seenIDs []Identifier
	var seen []INode

	q.consume(func(k Identifier, v INode) {
		seenIDs = append(seenIDs, k)
		seen = append(seen, v)
	})

	testutil.Equal(t, 0, q.len())
	testutil.Nil(t, q.head)
	testutil.Nil(t, q.tail)

	testutil.Equal(t, 5, len(seenIDs))
	testutil.Equal(t, 5, len(seen))

	for _, n := range seen {
		testutil.Nil(t, n.Node().nextInRecomputeHeap)
		testutil.Nil(t, n.Node().previousInRecomputeHeap)
	}
}

func Test_recomputeHeapList_removeHeadItem_full(t *testing.T) {
	g := New()
	q := new(recomputeHeapList)

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
	testutil.NotNil(t, rhnext(q.head))
	testutil.Equal(t, n0.Node().id, nodePtrID(q.head))
	testutil.Equal(t, n1.Node().id, nodePtrID(rhnext(q.head)))
	testutil.Equal(t, n2.Node().id, nodePtrID(rhnext(rhnext(q.head))))
	testutil.Equal(t, n3.Node().id, nodePtrID(rhnext(rhnext(rhnext(q.head)))))
	testutil.Equal(t, n4.Node().id, nodePtrID(rhnext(rhnext(rhnext(rhnext(q.head))))))

	testutil.NotNil(t, q.tail)
	testutil.NotNil(t, rhprev(q.tail))
	testutil.Equal(t, n4.Node().id, nodePtrID(q.tail))
	testutil.Equal(t, n3.Node().id, nodePtrID(rhprev(q.tail)))
	testutil.Equal(t, n2.Node().id, nodePtrID(rhprev(rhprev(q.tail))))
	testutil.Equal(t, n1.Node().id, nodePtrID(rhprev(rhprev(rhprev(q.tail)))))
	testutil.Equal(t, n0.Node().id, nodePtrID(rhprev(rhprev(rhprev(rhprev(q.tail))))))

	q.removeHeadItem()

	testutil.NotNil(t, q.head)
	testutil.NotNil(t, rhnext(q.head))
	testutil.Equal(t, n1.Node().id, nodePtrID(q.head))
	testutil.Equal(t, n2.Node().id, nodePtrID(rhnext(q.head)))
	testutil.Equal(t, n3.Node().id, nodePtrID(rhnext(rhnext(q.head))))
	testutil.Equal(t, n4.Node().id, nodePtrID(rhnext(rhnext(rhnext(q.head)))))

	testutil.NotNil(t, q.tail)
	testutil.NotNil(t, rhprev(q.tail))
	testutil.Equal(t, n4.Node().id, nodePtrID(q.tail))
	testutil.Equal(t, n3.Node().id, nodePtrID(rhprev(q.tail)))
	testutil.Equal(t, n2.Node().id, nodePtrID(rhprev(rhprev(q.tail))))
	testutil.Equal(t, n1.Node().id, nodePtrID(rhprev(rhprev(rhprev(q.tail)))))

	testutil.Nil(t, n0.Node().nextInRecomputeHeap)
	testutil.Nil(t, n0.Node().previousInRecomputeHeap)
}

func Test_recomputeHeapList_removeHeadItem_one(t *testing.T) {
	g := New()
	q := new(recomputeHeapList)

	n0 := newHeightIncr(g, 0)

	q.push(n0)

	testutil.NotNil(t, q.head)
	testutil.Nil(t, rhnext(q.head))
	testutil.Equal(t, n0.Node().id, nodePtrID(q.head))
	testutil.NotNil(t, q.tail)
	testutil.Nil(t, rhprev(q.tail))
	testutil.Equal(t, n0.Node().id, nodePtrID(q.tail))

	q.removeHeadItem()

	testutil.Nil(t, q.head)
	testutil.Nil(t, q.tail)

	testutil.Nil(t, n0.Node().nextInRecomputeHeap)
	testutil.Nil(t, n0.Node().previousInRecomputeHeap)
}
