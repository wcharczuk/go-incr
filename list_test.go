package incr

import (
	"testing"

	. "github.com/wcharczuk/go-incr/testutil"
)

func idWithNode(n INode) (Identifier, INode) {
	return n.Node().id, n
}

func Test_list_Push_Pop(t *testing.T) {
	q := new(list[Identifier, INode])

	n0 := newHeightIncr(0)
	n1 := newHeightIncr(0)
	n2 := newHeightIncr(0)
	n3 := newHeightIncr(0)

	var zeroID Identifier
	id, n, ok := q.Pop()

	// Region: empty list
	{
		ItsEqual(t, false, ok)
		ItsNil(t, q.head)
		ItsNil(t, q.tail)
		ItsEqual(t, zeroID, id)
		ItsNil(t, n)
		ItsEqual(t, 0, q.Len())
		ItsEqual(t, true, q.IsEmpty())
	}

	// Region: push 0
	{
		q.Push(idWithNode(n0))
		ItsNotNil(t, q.head)
		ItsNil(t, q.head.next)
		ItsNotNil(t, q.tail)
		ItsNil(t, q.tail.previous)
		ItsEqual(t, q.head, q.tail)
		ItsEqual(t, 1, q.Len())
		ItsEqual(t, false, q.IsEmpty())
		ItsEqual(t, true, q.Has(n0.Node().ID()))
		ItsEqual(t, false, q.Has(n1.Node().ID()))
		ItsEqual(t, false, q.Has(n2.Node().ID()))
		ItsEqual(t, false, q.Has(n3.Node().ID()))
		ItsEqual(t, n0.n.id, q.head.value.Node().id)
		ItsEqual(t, n0.n.id, q.tail.value.Node().id)
	}

	// Region: push 1
	{
		q.Push(idWithNode(n1))
		ItsNotNil(t, q.head)
		ItsNil(t, q.head.previous)
		ItsNotNil(t, q.head.next)
		ItsNil(t, q.head.next.next)
		ItsNotNil(t, q.tail)
		ItsNotNil(t, q.tail.previous)
		ItsNil(t, q.tail.previous.previous)
		ItsNil(t, q.tail.next)
		ItsNotEqual(t, q.head, q.tail)
		ItsEqual(t, q.head.next, q.tail)
		ItsEqual(t, q.tail.previous, q.head)
		ItsEqual(t, 2, q.Len())
		ItsEqual(t, false, q.IsEmpty())
		ItsEqual(t, n0.Node().ID(), q.head.value.Node().ID())
		ItsEqual(t, true, q.Has(n0.Node().ID()))
		ItsEqual(t, true, q.Has(n1.Node().ID()))
		ItsEqual(t, false, q.Has(n2.Node().ID()))
		ItsEqual(t, false, q.Has(n3.Node().ID()))
		ItsEqual(t, n1.Node().ID(), q.tail.value.Node().ID())
	}

	// Region: push 2
	{
		q.Push(idWithNode(n2))
		ItsNil(t, q.head.previous)
		ItsNotNil(t, q.head)
		ItsNotNil(t, q.head.next)
		ItsNotNil(t, q.head.next.next)
		ItsNil(t, q.head.next.next.next)
		ItsEqual(t, q.head.next.next, q.tail)
		ItsNotNil(t, q.tail)
		ItsNotNil(t, q.tail.previous)
		ItsNotNil(t, q.tail.previous.previous)
		ItsNil(t, q.tail.previous.previous.previous)
		ItsNil(t, q.tail.next)
		ItsEqual(t, q.tail.previous.previous, q.head)
		ItsNotEqual(t, q.head, q.tail)
		ItsEqual(t, 3, q.Len())
		ItsEqual(t, true, q.Has(n0.Node().ID()))
		ItsEqual(t, true, q.Has(n1.Node().ID()))
		ItsEqual(t, true, q.Has(n2.Node().ID()))
		ItsEqual(t, false, q.Has(n3.Node().ID()))
		ItsEqual(t, n0.Node().ID(), q.head.value.Node().ID())
		ItsEqual(t, n2.Node().ID(), q.tail.value.Node().ID())
	}

	// Region: push 3
	{
		q.Push(idWithNode(n3))
		ItsNil(t, q.head.previous)
		ItsNotNil(t, q.head)
		ItsNotNil(t, q.head.next)
		ItsNotNil(t, q.head.next.next)
		ItsNotNil(t, q.head.next.next.next)
		ItsNil(t, q.head.next.next.next.next)
		ItsEqual(t, q.head.next.next.next, q.tail)
		ItsNotNil(t, q.tail)
		ItsNotNil(t, q.tail.previous)
		ItsNotNil(t, q.tail.previous.previous)
		ItsNotNil(t, q.tail.previous.previous.previous)
		ItsNil(t, q.tail.previous.previous.previous.previous)
		ItsEqual(t, q.tail.previous.previous.previous, q.head)
		ItsNil(t, q.tail.next)
		ItsNotEqual(t, q.head, q.tail)
		ItsEqual(t, 4, q.Len())
		ItsEqual(t, true, q.Has(n0.Node().ID()))
		ItsEqual(t, true, q.Has(n1.Node().ID()))
		ItsEqual(t, true, q.Has(n2.Node().ID()))
		ItsEqual(t, true, q.Has(n3.Node().ID()))
		ItsEqual(t, n0.Node().ID(), q.head.value.Node().ID())
		ItsEqual(t, n3.Node().ID(), q.tail.value.Node().ID())
	}

	// Region: pop 0
	{
		id, n, ok = q.Pop()
		ItsEqual(t, true, ok)
		ItsEqual(t, n0.n.id, id)
		ItsEqual(t, n0.n.id, n.Node().id)
		ItsNotNil(t, q.head)
		ItsNotNil(t, q.head.next)
		ItsNotNil(t, q.head.next.next)
		ItsNil(t, q.head.next.next.next)
		ItsEqual(t, q.head.next.next, q.tail)
		ItsNotNil(t, q.tail)
		ItsNotEqual(t, q.head, q.tail)
		ItsEqual(t, 3, q.Len())
		ItsEqual(t, false, q.Has(n0.Node().ID()))
		ItsEqual(t, true, q.Has(n1.Node().ID()))
		ItsEqual(t, true, q.Has(n2.Node().ID()))
		ItsEqual(t, true, q.Has(n3.Node().ID()))
		ItsEqual(t, n1.Node().ID(), q.head.value.Node().ID())
		ItsEqual(t, n3.Node().ID(), q.tail.value.Node().ID())
	}

	// Region: pop 1
	{
		id, n, ok = q.Pop()
		ItsEqual(t, true, ok)
		ItsEqual(t, n1.n.id, id)
		ItsEqual(t, n1.n.id, n.Node().id)
		ItsNotNil(t, q.head)
		ItsNotNil(t, q.head.next)
		ItsNil(t, q.head.next.next)
		ItsEqual(t, q.head.next, q.tail)
		ItsNotNil(t, q.tail)
		ItsNotNil(t, q.tail.previous)
		ItsNil(t, q.tail.previous.previous)
		ItsEqual(t, q.tail.previous, q.head)
		ItsNotEqual(t, q.head, q.tail)
		ItsEqual(t, 2, q.Len())
		ItsEqual(t, false, q.Has(n0.Node().ID()))
		ItsEqual(t, false, q.Has(n1.Node().ID()))
		ItsEqual(t, true, q.Has(n2.Node().ID()))
		ItsEqual(t, true, q.Has(n3.Node().ID()))
		ItsEqual(t, n2.n.id, q.head.value.Node().id)
		ItsEqual(t, n3.n.id, q.tail.value.Node().id)
	}

	// Region: pop 2
	{
		id, n, ok = q.Pop()
		ItsEqual(t, true, ok)
		ItsEqual(t, n2.n.id, id)
		ItsEqual(t, n2.n.id, n.Node().id)
		ItsNotNil(t, q.head)
		ItsNil(t, q.head.previous)
		ItsNotNil(t, q.tail)
		ItsEqual(t, q.head, q.tail)
		ItsEqual(t, 1, q.Len())
		ItsEqual(t, n3.n.id, q.head.value.Node().id)
		ItsEqual(t, n3.n.id, q.tail.value.Node().id)
	}

	// Region: pop 3
	{
		id, n, ok = q.Pop()
		ItsEqual(t, true, ok)
		ItsEqual(t, n3.n.id, id)
		ItsEqual(t, n3.n.id, n.Node().id)
		ItsNil(t, q.head)
		ItsNil(t, q.tail)
		ItsEqual(t, 0, q.Len())
	}

	q.Push(idWithNode(n0))
	q.Push(idWithNode(n1))
	q.Push(idWithNode(n2))
	ItsEqual(t, 3, q.Len())
}

func Test_list_PushFront(t *testing.T) {
	q := new(list[Identifier, INode])

	n0 := newHeightIncr(0)
	n1 := newHeightIncr(0)
	n2 := newHeightIncr(0)
	n3 := newHeightIncr(0)
	q.PushFront(idWithNode(n0))
	q.PushFront(idWithNode(n1))
	q.PushFront(idWithNode(n2))
	q.PushFront(idWithNode(n3))

	all := q.PopAll()
	ItsEqual(t, n3.Node().ID(), all[0].Node().ID())
	ItsEqual(t, n2.Node().ID(), all[1].Node().ID())
	ItsEqual(t, n1.Node().ID(), all[2].Node().ID())
	ItsEqual(t, n0.Node().ID(), all[3].Node().ID())
}

func Test_list_PopBack(t *testing.T) {
	q := new(list[Identifier, INode])

	var zeroID Identifier
	id, n, ok := q.PopBack()
	ItsEqual(t, false, ok)
	ItsEqual(t, zeroID, id)
	ItsNil(t, n)

	n0 := newHeightIncrLabel(0, "n0")
	n1 := newHeightIncrLabel(0, "n1")
	n2 := newHeightIncrLabel(0, "n2")
	n3 := newHeightIncrLabel(0, "n3")

	q.Push(idWithNode(n0))
	q.Push(idWithNode(n1))
	q.Push(idWithNode(n2))
	q.Push(idWithNode(n3))

	id, n, ok = q.PopBack()
	ItsEqual(t, true, ok)
	ItsEqual(t, n3.Node().id, id)
	ItsEqual(t, n3.Node().id, n.Node().id)

	id, n, ok = q.PopBack()
	ItsEqual(t, true, ok)
	ItsEqual(t, n2.Node().id, id)
	ItsEqual(t, n2.Node().id, n.Node().id)

	id, n, ok = q.PopBack()
	ItsEqual(t, true, ok)
	ItsEqual(t, n1.Node().id, id)
	ItsEqual(t, n1.Node().id, n.Node().id)

	id, n, ok = q.PopBack()
	ItsEqual(t, true, ok)
	ItsEqual(t, n0.Node().id, id)
	ItsEqual(t, n0.Node().id, n.Node().id)

	id, n, ok = q.PopBack()
	ItsEqual(t, false, ok)
	ItsEqual(t, zeroID, id)
	ItsNil(t, n)
}

func Test_list_Get(t *testing.T) {
	q := new(list[Identifier, INode])

	n0 := newHeightIncr(0)
	n1 := newHeightIncr(0)
	n2 := newHeightIncr(0)
	n3 := newHeightIncr(0)

	q.Push(idWithNode(n0))
	q.Push(idWithNode(n1))
	q.Push(idWithNode(n2))
	q.Push(idWithNode(n3))

	ItsNotNil(t, q.head.next)
	ItsNotNil(t, q.head.next.next)
	ItsNotNil(t, q.head.next.next.next)

	found, ok := q.Get(n3.Node().ID())
	ItsEqual(t, true, ok)
	ItsNotNil(t, found)
	ItsEqual(t, found.Node().ID(), n3.Node().ID())

	found, ok = q.Get(n2.Node().ID())
	ItsEqual(t, true, ok)
	ItsEqual(t, found.Node().ID(), n2.Node().ID())

	found, ok = q.Get(n1.Node().ID())
	ItsEqual(t, true, ok)
	ItsEqual(t, found.Node().ID(), n1.Node().ID())

	found, ok = q.Get(n0.Node().ID())
	ItsEqual(t, true, ok)
	ItsEqual(t, found.Node().ID(), n0.Node().ID())
}

func Test_list_Remove(t *testing.T) {
	q := new(list[Identifier, INode])

	n0 := newHeightIncr(0)
	n1 := newHeightIncr(1)
	n2 := newHeightIncr(2)
	n3 := newHeightIncr(3)
	n4 := newHeightIncr(4)

	q.Push(idWithNode(n0))
	q.Push(idWithNode(n1))
	q.Push(idWithNode(n2))
	q.Push(idWithNode(n3))
	q.Push(idWithNode(n4))

	ItsEqual(t, q.head.key, n0.n.id)
	ItsNotNil(t, q.head.next)
	ItsNotNil(t, q.tail.previous)
	ItsEqual(t, q.tail.key, n4.n.id)

	ok := q.Remove(n2.Node().id)

	ItsEqual(t, true, ok)
	ItsEqual(t, 4, q.Len())
	ItsNotNil(t, q.tail)
	ItsEqual(t, q.tail.key, n4.n.id)

	ItsEqual(t, n0.Node().ID(), q.head.key)
	ItsEqual(t, n1.Node().ID(), q.head.next.key)
	ItsEqual(t, n3.Node().ID(), q.head.next.next.key)
	ItsEqual(t, n4.Node().ID(), q.head.next.next.next.key)

	ItsEqual(t, n4.Node().ID(), q.tail.key)
	ItsEqual(t, n3.Node().ID(), q.tail.previous.key)
	ItsEqual(t, n1.Node().ID(), q.tail.previous.previous.key)
	ItsEqual(t, n0.Node().ID(), q.tail.previous.previous.previous.key)
}

func Test_list_Remove_notFound(t *testing.T) {
	q := new(list[Identifier, INode])

	n0 := newHeightIncr(0)
	n1 := newHeightIncr(1)
	n2 := newHeightIncr(2)
	n3 := newHeightIncr(3)

	q.Push(idWithNode(n0))
	q.Push(idWithNode(n1))
	q.Push(idWithNode(n2))
	q.Push(idWithNode(n3))

	ItsEqual(t, q.head.key, n0.n.id)
	ItsNotNil(t, q.head.next)
	ItsNotNil(t, q.tail.previous)
	ItsEqual(t, q.tail.key, n3.n.id)

	ok := q.Remove(NewIdentifier())
	ItsEqual(t, false, ok)
}

func Test_list_Remove_head(t *testing.T) {
	q := new(list[Identifier, INode])

	n0 := newHeightIncr(0)
	n1 := newHeightIncr(1)
	n2 := newHeightIncr(2)
	n3 := newHeightIncr(3)

	q.Push(idWithNode(n0))
	q.Push(idWithNode(n1))
	q.Push(idWithNode(n2))
	q.Push(idWithNode(n3))

	ItsEqual(t, q.head.key, n0.n.id)
	ItsNotNil(t, q.head.next)
	ItsNotNil(t, q.tail.previous)
	ItsEqual(t, q.tail.key, n3.n.id)

	ok := q.Remove(n0.Node().id)
	ItsEqual(t, ok, true)
	ItsNotNil(t, q.head)
	ItsEqual(t, q.head.key, n1.n.id)
}

func Test_list_Remove_tail(t *testing.T) {
	q := new(list[Identifier, INode])

	n0 := newHeightIncr(0)
	n1 := newHeightIncr(1)
	n2 := newHeightIncr(2)
	n3 := newHeightIncr(3)

	q.Push(idWithNode(n0))
	q.Push(idWithNode(n1))
	q.Push(idWithNode(n2))
	q.Push(idWithNode(n3))

	ItsEqual(t, q.head.key, n0.n.id)
	ItsNotNil(t, q.head.next)
	ItsNotNil(t, q.tail.previous)
	ItsEqual(t, q.tail.key, n3.n.id)

	ok := q.Remove(n3.Node().id)
	ItsEqual(t, ok, true)
	ItsNotNil(t, q.tail)
	ItsEqual(t, q.tail.key, n2.n.id)
}

func Test_list_PopAll(t *testing.T) {
	n0 := newHeightIncr(0)
	n1 := newHeightIncr(0)
	n2 := newHeightIncr(0)

	rhl := new(list[Identifier, INode])
	rhl.Push(idWithNode(n0))
	rhl.Push(idWithNode(n1))
	rhl.Push(idWithNode(n2))

	ItsNotNil(t, rhl.head)
	ItsNotNil(t, rhl.tail)
	ItsEqual(t, 3, rhl.Len())

	output := rhl.PopAll()
	ItsEqual(t, 3, len(output))
	ItsEqual(t, 0, rhl.Len())
	ItsNil(t, rhl.head)
	ItsNil(t, rhl.tail)
}
