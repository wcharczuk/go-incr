package incr

import "testing"

func idWithNode(n INode) (Identifier, INode) {
	return n.Node().id, n
}

func Test_list_Push_Pop(t *testing.T) {
	q := new(List[listItem[Identifier, INode]])

	n0 := newHeightIncr(0)
	n1 := newHeightIncr(0)
	n2 := newHeightIncr(0)
	n3 := newHeightIncr(0)

	var zeroID Identifier
	node, ok := q.Pop()

	// Region: empty list
	{
		ItsEqual(t, false, ok)
		ItsNil(t, q.head)
		ItsNil(t, q.tail)
		ItsEqual(t, zeroID, node.Key)
		ItsEqual(t, 0, q.size)
	}

	// Region: push 0
	{
		q.Push(idWithNode(n0))
		ItsNotNil(t, q.head)
		ItsNil(t, q.head.next)
		ItsNotNil(t, q.tail)
		ItsNil(t, q.tail.previous)
		ItsEqual(t, q.head, q.tail)
		ItsEqual(t, 1, q.len)
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
		ItsEqual(t, 2, q.len)
		ItsEqual(t, n0.n.id, q.head.value.Node().id)
		ItsEqual(t, n1.n.id, q.tail.value.Node().id)
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
		ItsEqual(t, 3, q.len)
		ItsEqual(t, n0.n.id, q.head.value.Node().id)
		ItsEqual(t, n2.n.id, q.tail.value.Node().id)
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
		ItsEqual(t, 4, q.len)
		ItsEqual(t, n0.n.id, q.head.value.Node().id)
		ItsEqual(t, n3.n.id, q.tail.value.Node().id)
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
		ItsEqual(t, q.len, 3)
		ItsEqual(t, n1.n.id, q.head.value.Node().id)
		ItsEqual(t, n3.n.id, q.tail.value.Node().id)
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
		ItsEqual(t, 2, q.len)
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
		ItsEqual(t, 1, q.len)
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
		ItsEqual(t, 0, q.len)
	}

	q.Push(idWithNode(n0))
	q.Push(idWithNode(n1))
	q.Push(idWithNode(n2))
	ItsEqual(t, 3, q.len)
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
	ItsEqual(t, n3.Node().id, all[0].Node().id)
	ItsEqual(t, n2.Node().id, all[1].Node().id)
	ItsEqual(t, n1.Node().id, all[2].Node().id)
	ItsEqual(t, n0.Node().id, all[3].Node().id)
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

func Test_list_Find(t *testing.T) {
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

	found := q.Find(n3.Node().id)
	ItsNotNil(t, found)
	ItsEqual(t, found.key, n3.Node().id)

	found = q.Find(n2.Node().id)
	ItsNotNil(t, found)
	ItsEqual(t, found.key, n2.Node().id)

	found = q.Find(n1.Node().id)
	ItsNotNil(t, found)
	ItsEqual(t, found.key, n1.Node().id)

	found = q.Find(n0.Node().id)
	ItsNotNil(t, found)
	ItsEqual(t, found.key, n0.Node().id)
}

func Test_list_Remove_tail(t *testing.T) {
	q := new(list[Identifier, INode])

	n0 := newHeightIncr(0)
	n1 := newHeightIncr(0)
	n2 := newHeightIncr(0)
	n3 := newHeightIncr(0)

	q.Push(idWithNode(n0))
	q.Push(idWithNode(n1))
	q.Push(idWithNode(n2))
	q.Push(idWithNode(n3))

	q.Remove(q.Find(n3.Node().id))
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
	ItsEqual(t, 3, rhl.len)

	output := rhl.PopAll()
	ItsEqual(t, 3, len(output))
	ItsEqual(t, 0, rhl.len)
	ItsNil(t, rhl.head)
	ItsNil(t, rhl.tail)
}
