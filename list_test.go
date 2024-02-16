package incr

import (
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

func idWithNode(n INode) (Identifier, INode) {
	return n.Node().id, n
}

func Test_list_push_pop(t *testing.T) {
	g := New()
	q := new(list[Identifier, INode])

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
	}

	// Region: push 0
	{
		q.push(idWithNode(n0))
		testutil.NotNil(t, q.head)
		testutil.Nil(t, q.head.next)
		testutil.NotNil(t, q.tail)
		testutil.Nil(t, q.tail.previous)
		testutil.Equal(t, q.head, q.tail)
		testutil.Equal(t, 1, q.len())
		testutil.Equal(t, false, q.isEmpty())
		testutil.Equal(t, true, q.has(n0.Node().ID()))
		testutil.Equal(t, false, q.has(n1.Node().ID()))
		testutil.Equal(t, false, q.has(n2.Node().ID()))
		testutil.Equal(t, false, q.has(n3.Node().ID()))
		testutil.Equal(t, n0.n.id, q.head.value.Node().id)
		testutil.Equal(t, n0.n.id, q.tail.value.Node().id)
	}

	// Region: push 1
	{
		q.push(idWithNode(n1))
		testutil.NotNil(t, q.head)
		testutil.Nil(t, q.head.previous)
		testutil.NotNil(t, q.head.next)
		testutil.Nil(t, q.head.next.next)
		testutil.NotNil(t, q.tail)
		testutil.NotNil(t, q.tail.previous)
		testutil.Nil(t, q.tail.previous.previous)
		testutil.Nil(t, q.tail.next)
		testutil.NotEqual(t, q.head, q.tail)
		testutil.Equal(t, q.head.next, q.tail)
		testutil.Equal(t, q.tail.previous, q.head)
		testutil.Equal(t, 2, q.len())
		testutil.Equal(t, false, q.isEmpty())
		testutil.Equal(t, n0.Node().ID(), q.head.value.Node().ID())
		testutil.Equal(t, true, q.has(n0.Node().ID()))
		testutil.Equal(t, true, q.has(n1.Node().ID()))
		testutil.Equal(t, false, q.has(n2.Node().ID()))
		testutil.Equal(t, false, q.has(n3.Node().ID()))
		testutil.Equal(t, n1.Node().ID(), q.tail.value.Node().ID())
	}

	// Region: push 2
	{
		q.push(idWithNode(n2))
		testutil.Nil(t, q.head.previous)
		testutil.NotNil(t, q.head)
		testutil.NotNil(t, q.head.next)
		testutil.NotNil(t, q.head.next.next)
		testutil.Nil(t, q.head.next.next.next)
		testutil.Equal(t, q.head.next.next, q.tail)
		testutil.NotNil(t, q.tail)
		testutil.NotNil(t, q.tail.previous)
		testutil.NotNil(t, q.tail.previous.previous)
		testutil.Nil(t, q.tail.previous.previous.previous)
		testutil.Nil(t, q.tail.next)
		testutil.Equal(t, q.tail.previous.previous, q.head)
		testutil.NotEqual(t, q.head, q.tail)
		testutil.Equal(t, 3, q.len())
		testutil.Equal(t, true, q.has(n0.Node().ID()))
		testutil.Equal(t, true, q.has(n1.Node().ID()))
		testutil.Equal(t, true, q.has(n2.Node().ID()))
		testutil.Equal(t, false, q.has(n3.Node().ID()))
		testutil.Equal(t, n0.Node().ID(), q.head.value.Node().ID())
		testutil.Equal(t, n2.Node().ID(), q.tail.value.Node().ID())
	}

	// Region: push 3
	{
		q.push(idWithNode(n3))
		testutil.Nil(t, q.head.previous)
		testutil.NotNil(t, q.head)
		testutil.NotNil(t, q.head.next)
		testutil.NotNil(t, q.head.next.next)
		testutil.NotNil(t, q.head.next.next.next)
		testutil.Nil(t, q.head.next.next.next.next)
		testutil.Equal(t, q.head.next.next.next, q.tail)
		testutil.NotNil(t, q.tail)
		testutil.NotNil(t, q.tail.previous)
		testutil.NotNil(t, q.tail.previous.previous)
		testutil.NotNil(t, q.tail.previous.previous.previous)
		testutil.Nil(t, q.tail.previous.previous.previous.previous)
		testutil.Equal(t, q.tail.previous.previous.previous, q.head)
		testutil.Nil(t, q.tail.next)
		testutil.NotEqual(t, q.head, q.tail)
		testutil.Equal(t, 4, q.len())
		testutil.Equal(t, true, q.has(n0.Node().ID()))
		testutil.Equal(t, true, q.has(n1.Node().ID()))
		testutil.Equal(t, true, q.has(n2.Node().ID()))
		testutil.Equal(t, true, q.has(n3.Node().ID()))
		testutil.Equal(t, n0.Node().ID(), q.head.value.Node().ID())
		testutil.Equal(t, n3.Node().ID(), q.tail.value.Node().ID())
	}

	// Region: pop 0
	{
		id, n, ok = q.pop()
		testutil.Equal(t, true, ok)
		testutil.Equal(t, n0.n.id, id)
		testutil.Equal(t, n0.n.id, n.Node().id)
		testutil.NotNil(t, q.head)
		testutil.NotNil(t, q.head.next)
		testutil.NotNil(t, q.head.next.next)
		testutil.Nil(t, q.head.next.next.next)
		testutil.Equal(t, q.head.next.next, q.tail)
		testutil.NotNil(t, q.tail)
		testutil.NotEqual(t, q.head, q.tail)
		testutil.Equal(t, 3, q.len())
		testutil.Equal(t, false, q.has(n0.Node().ID()))
		testutil.Equal(t, true, q.has(n1.Node().ID()))
		testutil.Equal(t, true, q.has(n2.Node().ID()))
		testutil.Equal(t, true, q.has(n3.Node().ID()))
		testutil.Equal(t, n1.Node().ID(), q.head.value.Node().ID())
		testutil.Equal(t, n3.Node().ID(), q.tail.value.Node().ID())
	}

	// Region: pop 1
	{
		id, n, ok = q.pop()
		testutil.Equal(t, true, ok)
		testutil.Equal(t, n1.n.id, id)
		testutil.Equal(t, n1.n.id, n.Node().id)
		testutil.NotNil(t, q.head)
		testutil.NotNil(t, q.head.next)
		testutil.Nil(t, q.head.next.next)
		testutil.Equal(t, q.head.next, q.tail)
		testutil.NotNil(t, q.tail)
		testutil.NotNil(t, q.tail.previous)
		testutil.Nil(t, q.tail.previous.previous)
		testutil.Equal(t, q.tail.previous, q.head)
		testutil.NotEqual(t, q.head, q.tail)
		testutil.Equal(t, 2, q.len())
		testutil.Equal(t, false, q.has(n0.Node().ID()))
		testutil.Equal(t, false, q.has(n1.Node().ID()))
		testutil.Equal(t, true, q.has(n2.Node().ID()))
		testutil.Equal(t, true, q.has(n3.Node().ID()))
		testutil.Equal(t, n2.n.id, q.head.value.Node().id)
		testutil.Equal(t, n3.n.id, q.tail.value.Node().id)
	}

	// Region: pop 2
	{
		id, n, ok = q.pop()
		testutil.Equal(t, true, ok)
		testutil.Equal(t, n2.n.id, id)
		testutil.Equal(t, n2.n.id, n.Node().id)
		testutil.NotNil(t, q.head)
		testutil.Nil(t, q.head.previous)
		testutil.NotNil(t, q.tail)
		testutil.Equal(t, q.head, q.tail)
		testutil.Equal(t, 1, q.len())
		testutil.Equal(t, n3.n.id, q.head.value.Node().id)
		testutil.Equal(t, n3.n.id, q.tail.value.Node().id)
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
	}

	q.push(idWithNode(n0))
	q.push(idWithNode(n1))
	q.push(idWithNode(n2))
	testutil.Equal(t, 3, q.len())
}

func Test_list_PushFront(t *testing.T) {
	g := New()
	q := new(list[Identifier, INode])

	n0 := newHeightIncr(g, 0)
	n1 := newHeightIncr(g, 0)
	n2 := newHeightIncr(g, 0)
	n3 := newHeightIncr(g, 0)
	q.pushFront(idWithNode(n0))
	q.pushFront(idWithNode(n1))
	q.pushFront(idWithNode(n2))
	q.pushFront(idWithNode(n3))

	all := q.popAll()
	testutil.Equal(t, n3.Node().ID(), all[0].Node().ID())
	testutil.Equal(t, n2.Node().ID(), all[1].Node().ID())
	testutil.Equal(t, n1.Node().ID(), all[2].Node().ID())
	testutil.Equal(t, n0.Node().ID(), all[3].Node().ID())
}

func Test_list_PopBack(t *testing.T) {
	g := New()
	q := new(list[Identifier, INode])

	var zeroID Identifier
	id, n, ok := q.popBack()
	testutil.Equal(t, false, ok)
	testutil.Equal(t, zeroID, id)
	testutil.Nil(t, n)

	n0 := newHeightIncrLabel(g, 0, "n0")
	n1 := newHeightIncrLabel(g, 0, "n1")
	n2 := newHeightIncrLabel(g, 0, "n2")
	n3 := newHeightIncrLabel(g, 0, "n3")

	q.push(idWithNode(n0))
	q.push(idWithNode(n1))
	q.push(idWithNode(n2))
	q.push(idWithNode(n3))

	id, n, ok = q.popBack()
	testutil.Equal(t, true, ok)
	testutil.Equal(t, n3.Node().id, id)
	testutil.Equal(t, n3.Node().id, n.Node().id)

	id, n, ok = q.popBack()
	testutil.Equal(t, true, ok)
	testutil.Equal(t, n2.Node().id, id)
	testutil.Equal(t, n2.Node().id, n.Node().id)

	id, n, ok = q.popBack()
	testutil.Equal(t, true, ok)
	testutil.Equal(t, n1.Node().id, id)
	testutil.Equal(t, n1.Node().id, n.Node().id)

	id, n, ok = q.popBack()
	testutil.Equal(t, true, ok)
	testutil.Equal(t, n0.Node().id, id)
	testutil.Equal(t, n0.Node().id, n.Node().id)

	id, n, ok = q.popBack()
	testutil.Equal(t, false, ok)
	testutil.Equal(t, zeroID, id)
	testutil.Nil(t, n)
}

func Test_list_Get(t *testing.T) {
	g := New()
	q := new(list[Identifier, INode])

	n0 := newHeightIncr(g, 0)
	n1 := newHeightIncr(g, 0)
	n2 := newHeightIncr(g, 0)
	n3 := newHeightIncr(g, 0)

	q.push(idWithNode(n0))
	q.push(idWithNode(n1))
	q.push(idWithNode(n2))
	q.push(idWithNode(n3))

	testutil.NotNil(t, q.head.next)
	testutil.NotNil(t, q.head.next.next)
	testutil.NotNil(t, q.head.next.next.next)

	found, ok := q.get(n3.Node().ID())
	testutil.Equal(t, true, ok)
	testutil.NotNil(t, found)
	testutil.Equal(t, found.Node().ID(), n3.Node().ID())

	found, ok = q.get(n2.Node().ID())
	testutil.Equal(t, true, ok)
	testutil.Equal(t, found.Node().ID(), n2.Node().ID())

	found, ok = q.get(n1.Node().ID())
	testutil.Equal(t, true, ok)
	testutil.Equal(t, found.Node().ID(), n1.Node().ID())

	found, ok = q.get(n0.Node().ID())
	testutil.Equal(t, true, ok)
	testutil.Equal(t, found.Node().ID(), n0.Node().ID())
}

func Test_list_Remove(t *testing.T) {
	g := New()
	q := new(list[Identifier, INode])

	n0 := newHeightIncr(g, 0)
	n1 := newHeightIncr(g, 1)
	n2 := newHeightIncr(g, 2)
	n3 := newHeightIncr(g, 3)
	n4 := newHeightIncr(g, 4)

	q.push(idWithNode(n0))
	q.push(idWithNode(n1))
	q.push(idWithNode(n2))
	q.push(idWithNode(n3))
	q.push(idWithNode(n4))

	testutil.Equal(t, q.head.key, n0.n.id)
	testutil.NotNil(t, q.head.next)
	testutil.NotNil(t, q.tail.previous)
	testutil.Equal(t, q.tail.key, n4.n.id)

	ok := q.remove(n2.Node().id)

	testutil.Equal(t, true, ok)
	testutil.Equal(t, 4, q.len())
	testutil.NotNil(t, q.tail)
	testutil.Equal(t, q.tail.key, n4.n.id)

	testutil.Equal(t, n0.Node().ID(), q.head.key)
	testutil.Equal(t, n1.Node().ID(), q.head.next.key)
	testutil.Equal(t, n3.Node().ID(), q.head.next.next.key)
	testutil.Equal(t, n4.Node().ID(), q.head.next.next.next.key)

	testutil.Equal(t, n4.Node().ID(), q.tail.key)
	testutil.Equal(t, n3.Node().ID(), q.tail.previous.key)
	testutil.Equal(t, n1.Node().ID(), q.tail.previous.previous.key)
	testutil.Equal(t, n0.Node().ID(), q.tail.previous.previous.previous.key)
}

func Test_List_Has(t *testing.T) {
	testutil.Equal(t, false, (&list[string, string]{}).has("foo"))
	testutil.Equal(t, true, (&list[string, string]{
		items: map[string]*listItem[string, string]{"foo": {key: "foo", value: "bar"}},
	}).has("foo"))
	testutil.Equal(t, false, (&list[string, string]{
		items: map[string]*listItem[string, string]{"foo": {key: "foo", value: "bar"}},
	}).has("not-foo"))
}

func Test_List_Get(t *testing.T) {
	got, ok := (&list[string, string]{}).get("foo")
	testutil.Equal(t, false, ok)
	testutil.Equal(t, "", got)

	got, ok = (&list[string, string]{
		items: map[string]*listItem[string, string]{"foo": {key: "foo", value: "bar"}},
	}).get("foo")
	testutil.Equal(t, true, ok)
	testutil.Equal(t, "bar", got)

	got, ok = (&list[string, string]{
		items: map[string]*listItem[string, string]{"foo": {key: "foo", value: "bar"}},
	}).get("not-foo")
	testutil.Equal(t, false, ok)
	testutil.Equal(t, "", got)
}

func Test_list_Remove_notFound(t *testing.T) {
	g := New()
	q := new(list[Identifier, INode])

	n0 := newHeightIncr(g, 0)
	n1 := newHeightIncr(g, 1)
	n2 := newHeightIncr(g, 2)
	n3 := newHeightIncr(g, 3)

	q.push(idWithNode(n0))
	q.push(idWithNode(n1))
	q.push(idWithNode(n2))
	q.push(idWithNode(n3))

	testutil.Equal(t, q.head.key, n0.n.id)
	testutil.NotNil(t, q.head.next)
	testutil.NotNil(t, q.tail.previous)
	testutil.Equal(t, q.tail.key, n3.n.id)

	ok := q.remove(NewIdentifier())
	testutil.Equal(t, false, ok)
}

func Test_list_Remove_head(t *testing.T) {
	g := New()
	q := new(list[Identifier, INode])

	n0 := newHeightIncr(g, 0)
	n1 := newHeightIncr(g, 1)
	n2 := newHeightIncr(g, 2)
	n3 := newHeightIncr(g, 3)

	q.push(idWithNode(n0))
	q.push(idWithNode(n1))
	q.push(idWithNode(n2))
	q.push(idWithNode(n3))

	testutil.Equal(t, q.head.key, n0.n.id)
	testutil.NotNil(t, q.head.next)
	testutil.NotNil(t, q.tail.previous)
	testutil.Equal(t, q.tail.key, n3.n.id)

	ok := q.remove(n0.Node().id)
	testutil.Equal(t, ok, true)
	testutil.NotNil(t, q.head)
	testutil.Equal(t, q.head.key, n1.n.id)
}

func Test_list_Remove_tail(t *testing.T) {
	g := New()
	q := new(list[Identifier, INode])

	n0 := newHeightIncr(g, 0)
	n1 := newHeightIncr(g, 1)
	n2 := newHeightIncr(g, 2)
	n3 := newHeightIncr(g, 3)

	q.push(idWithNode(n0))
	q.push(idWithNode(n1))
	q.push(idWithNode(n2))
	q.push(idWithNode(n3))

	testutil.Equal(t, q.head.key, n0.n.id)
	testutil.NotNil(t, q.head.next)
	testutil.NotNil(t, q.tail.previous)
	testutil.Equal(t, q.tail.key, n3.n.id)

	ok := q.remove(n3.Node().id)
	testutil.Equal(t, ok, true)
	testutil.NotNil(t, q.tail)
	testutil.Equal(t, q.tail.key, n2.n.id)
}

func Test_list_PopAll(t *testing.T) {
	g := New()
	n0 := newHeightIncr(g, 0)
	n1 := newHeightIncr(g, 0)
	n2 := newHeightIncr(g, 0)

	rhl := new(list[Identifier, INode])
	rhl.push(idWithNode(n0))
	rhl.push(idWithNode(n1))
	rhl.push(idWithNode(n2))

	testutil.NotNil(t, rhl.head)
	testutil.NotNil(t, rhl.tail)
	testutil.Equal(t, 3, rhl.len())

	output := rhl.popAll()
	testutil.Equal(t, 3, len(output))
	testutil.Equal(t, 0, rhl.len())
	testutil.Nil(t, rhl.head)
	testutil.Nil(t, rhl.tail)
}

func Test_list_each(t *testing.T) {
	g := New()
	n0 := newHeightIncr(g, 0)
	n1 := newHeightIncr(g, 0)
	n2 := newHeightIncr(g, 0)

	rhl := new(list[Identifier, INode])
	rhl.push(idWithNode(n0))
	rhl.push(idWithNode(n1))
	rhl.push(idWithNode(n2))

	var seen []INode
	var seenIDs []Identifier

	rhl.each(func(k Identifier, v INode) {
		seenIDs = append(seenIDs, k)
		seen = append(seen, v)
	})

	testutil.Equal(t, 3, len(seen))
	testutil.Equal(t, 3, len(seenIDs))

	testutil.Equal(t, n0.Node().id, seen[0].Node().id)
	testutil.Equal(t, n1.Node().id, seen[1].Node().id)
	testutil.Equal(t, n2.Node().id, seen[2].Node().id)

	testutil.Equal(t, n0.Node().id, seenIDs[0])
	testutil.Equal(t, n1.Node().id, seenIDs[1])
	testutil.Equal(t, n2.Node().id, seenIDs[2])
}
