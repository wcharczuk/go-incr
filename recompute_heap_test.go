package incr

import (
	"testing"
)

func Test_recomuteHeap_Add(t *testing.T) {

	rh := newRecomputeHeap(32)

	n50 := newHeightIncr(5)
	n60 := newHeightIncr(6)
	n70 := newHeightIncr(7)

	rh.Add(n50)

	// assertions post add n50
	{
		ItsEqual(t, 1, rh.Len())
		ItsEqual(t, 1, rh.heights[5].len)
		ItsNotNil(t, 1, rh.heights[5].head)
		ItsNotNil(t, 1, rh.heights[5].head.value)
		ItsEqual(t, n50.Node().id, rh.heights[5].head.value.Node().id)
		ItsEqual(t, true, rh.Has(n50))
		ItsEqual(t, false, rh.Has(n60))
		ItsEqual(t, false, rh.Has(n70))
		ItsEqual(t, 5, rh.MinHeight())
		ItsEqual(t, 5, rh.MaxHeight())
	}

	rh.Add(n60)

	// assertions post add n60
	{
		ItsEqual(t, 2, rh.Len())
		ItsEqual(t, 1, rh.heights[5].len)
		ItsNotNil(t, 1, rh.heights[5].head)
		ItsNotNil(t, 1, rh.heights[5].head.value)
		ItsEqual(t, n50.Node().id, rh.heights[5].head.value.Node().id)
		ItsEqual(t, 1, rh.heights[6].len)
		ItsNotNil(t, 1, rh.heights[6].head)
		ItsNotNil(t, 1, rh.heights[6].head.value)
		ItsEqual(t, n60.Node().id, rh.heights[6].head.value.Node().id)
		ItsEqual(t, true, rh.Has(n50))
		ItsEqual(t, true, rh.Has(n50))
		ItsEqual(t, false, rh.Has(n70))
		ItsEqual(t, 5, rh.MinHeight())
		ItsEqual(t, 6, rh.MaxHeight())
	}

	rh.Add(n70)

	// assertions post add n70
	{
		ItsEqual(t, 3, rh.Len())
		ItsEqual(t, 1, rh.heights[5].len)
		ItsNotNil(t, 1, rh.heights[5].head)
		ItsNotNil(t, 1, rh.heights[5].head.value)
		ItsEqual(t, n50.Node().id, rh.heights[5].head.value.Node().id)
		ItsEqual(t, 1, rh.heights[6].len)
		ItsNotNil(t, 1, rh.heights[6].head)
		ItsNotNil(t, 1, rh.heights[6].head.value)
		ItsEqual(t, n60.Node().id, rh.heights[6].head.value.Node().id)
		ItsEqual(t, 1, rh.heights[7].len)
		ItsNotNil(t, 1, rh.heights[7].head)
		ItsNotNil(t, 1, rh.heights[7].head.value)
		ItsEqual(t, n70.Node().id, rh.heights[7].head.value.Node().id)
		ItsEqual(t, true, rh.Has(n50))
		ItsEqual(t, true, rh.Has(n60))
		ItsEqual(t, true, rh.Has(n70))
		ItsEqual(t, 5, rh.MinHeight())
		ItsEqual(t, 7, rh.MaxHeight())
	}

	panics := make(chan any, 1)
	func() {
		defer func() {
			panics <- recover()
		}()
		rh.Add(newHeightIncr(33))
	}()

	ItsNotNil(t, <-panics)
}

func Test_recomputeHeap_RemoveMin(t *testing.T) {
	rh := newRecomputeHeap(32)

	ItsNil(t, rh.RemoveMin())

	n00 := newHeightIncr(0)
	n01 := newHeightIncr(0)
	n10 := newHeightIncr(1)
	n100 := newHeightIncr(10)

	rh.Add(n00)
	ItsEqual(t, 1, rh.Len())
	ItsEqual(t, 32, len(rh.heights))
	ItsEqual(t, 1, rh.heights[0].len)
	rh.Add(n01)
	ItsEqual(t, 2, rh.Len())
	ItsEqual(t, 32, len(rh.heights))
	ItsEqual(t, 2, rh.heights[0].len)

	ItsEqual(t, 0, rh.minHeight)
	ItsEqual(t, 0, rh.maxHeight)

	rh.Add(n10)
	ItsEqual(t, 3, rh.Len())
	ItsEqual(t, 32, len(rh.heights))
	ItsEqual(t, 2, rh.heights[0].len)
	ItsEqual(t, 1, rh.heights[1].len)
	ItsEqual(t, 0, rh.minHeight)
	ItsEqual(t, 1, rh.maxHeight)

	rh.Add(n100)
	ItsEqual(t, 4, rh.Len())
	ItsEqual(t, 32, len(rh.heights))
	ItsEqual(t, 2, rh.heights[0].len)
	ItsEqual(t, 1, rh.heights[1].len)
	ItsEqual(t, 1, rh.heights[10].len)
	ItsEqual(t, 0, rh.minHeight)
	ItsEqual(t, 10, rh.maxHeight)

	r00 := rh.RemoveMin()
	ItsNotNil(t, r00)
	ItsNotNil(t, r00.Node())
	ItsEqual(t, n00.n.id, r00.Node().id)
	ItsEqual(t, 3, rh.Len())
	ItsEqual(t, false, rh.Has(n00))
	ItsEqual(t, true, rh.Has(n01))
	ItsEqual(t, true, rh.Has(n10))
	ItsEqual(t, true, rh.Has(n100))
	ItsEqual(t, 1, rh.heights[0].len)
	ItsEqual(t, 1, rh.heights[1].len)
	ItsEqual(t, 1, rh.heights[10].len)
	ItsEqual(t, 0, rh.minHeight)
	ItsEqual(t, 10, rh.maxHeight)

	r01 := rh.RemoveMin()
	ItsNotNil(t, r01)
	ItsNotNil(t, r01.Node())
	ItsEqual(t, n01.n.id, r01.Node().id)
	ItsEqual(t, 2, rh.Len())
	ItsEqual(t, false, rh.Has(n00))
	ItsEqual(t, false, rh.Has(n01))
	ItsEqual(t, true, rh.Has(n10))
	ItsEqual(t, true, rh.Has(n100))
	ItsEqual(t, 0, rh.heights[0].len)
	ItsEqual(t, 1, rh.heights[1].len)
	ItsEqual(t, 1, rh.heights[10].len)
	ItsEqual(t, 1, rh.minHeight)
	ItsEqual(t, 10, rh.maxHeight)

	r10 := rh.RemoveMin()
	ItsNotNil(t, r10)
	ItsNotNil(t, r10.Node())
	ItsEqual(t, n10.n.id, r10.Node().id)
	ItsEqual(t, 1, rh.Len())
	ItsEqual(t, false, rh.Has(n00))
	ItsEqual(t, false, rh.Has(n01))
	ItsEqual(t, false, rh.Has(n10))
	ItsEqual(t, true, rh.Has(n100))
	ItsEqual(t, 0, rh.heights[0].len)
	ItsEqual(t, 0, rh.heights[1].len)
	ItsEqual(t, 1, rh.heights[10].len)
	ItsEqual(t, 10, rh.minHeight)
	ItsEqual(t, 10, rh.maxHeight)

	r100 := rh.RemoveMin()
	ItsNotNil(t, r100)
	ItsNotNil(t, r100.Node())
	ItsEqual(t, n100.n.id, r100.Node().id)
	ItsEqual(t, 0, rh.Len())
	ItsEqual(t, false, rh.Has(n00))
	ItsEqual(t, false, rh.Has(n01))
	ItsEqual(t, false, rh.Has(n10))
	ItsEqual(t, false, rh.Has(n100))
	ItsEqual(t, 0, rh.heights[0].len)
	ItsEqual(t, 0, rh.heights[1].len)
	ItsEqual(t, 0, rh.heights[10].len)
	ItsEqual(t, 0, rh.minHeight)
	ItsEqual(t, 10, rh.maxHeight)
}

func Test_recomputeHeap_RemoveMinHeight(t *testing.T) {
	rh := newRecomputeHeap(10)

	n00 := newHeightIncr(0)
	n01 := newHeightIncr(0)
	n02 := newHeightIncr(0)

	n10 := newHeightIncr(1)
	n11 := newHeightIncr(1)
	n12 := newHeightIncr(1)
	n13 := newHeightIncr(1)

	n50 := newHeightIncr(5)
	n51 := newHeightIncr(5)
	n52 := newHeightIncr(5)
	n53 := newHeightIncr(5)
	n54 := newHeightIncr(5)

	rh.Add(n00)
	rh.Add(n01)
	rh.Add(n02)
	rh.Add(n10)
	rh.Add(n11)
	rh.Add(n12)
	rh.Add(n13)
	rh.Add(n50)
	rh.Add(n51)
	rh.Add(n52)
	rh.Add(n53)
	rh.Add(n54)

	ItsEqual(t, 12, rh.Len())
	ItsEqual(t, 0, rh.MinHeight())
	ItsEqual(t, 5, rh.MaxHeight())

	output := rh.RemoveMinHeight()
	ItsEqual(t, 9, rh.Len())
	ItsEqual(t, 3, len(output))
	ItsNil(t, rh.heights[0].head)
	ItsNil(t, rh.heights[0].tail)
	ItsEqual(t, 0, rh.heights[0].len)
	ItsNotNil(t, rh.heights[1].head)
	ItsNotNil(t, rh.heights[1].tail)
	ItsEqual(t, 4, rh.heights[1].len)
	ItsNotNil(t, rh.heights[5].head)
	ItsNotNil(t, rh.heights[5].tail)
	ItsEqual(t, 5, rh.heights[5].len)

	ItsEqual(t, 1, rh.MinHeight())
	ItsEqual(t, 5, rh.MaxHeight())

	output = rh.RemoveMinHeight()
	ItsEqual(t, 5, rh.Len())
	ItsEqual(t, 4, len(output))
	ItsNil(t, rh.heights[0].head)
	ItsNil(t, rh.heights[0].tail)
	ItsEqual(t, 0, rh.heights[0].len)
	ItsNil(t, rh.heights[1].head)
	ItsNil(t, rh.heights[1].tail)
	ItsEqual(t, 0, rh.heights[1].len)
	ItsNotNil(t, rh.heights[5].head)
	ItsNotNil(t, rh.heights[5].tail)
	ItsEqual(t, 5, rh.heights[5].len)

	output = rh.RemoveMinHeight()
	ItsEqual(t, 0, rh.Len())
	ItsEqual(t, 5, len(output))
	ItsNil(t, rh.heights[0].head)
	ItsNil(t, rh.heights[0].tail)
	ItsEqual(t, 0, rh.heights[0].len)
	ItsNil(t, rh.heights[1].head)
	ItsNil(t, rh.heights[1].tail)
	ItsEqual(t, 0, rh.heights[1].len)
	ItsNil(t, rh.heights[5].head)
	ItsNil(t, rh.heights[5].tail)
	ItsEqual(t, 0, rh.heights[5].len)

	rh.Add(n50)
	rh.Add(n51)
	rh.Add(n52)
	rh.Add(n53)
	rh.Add(n54)

	ItsEqual(t, 5, rh.Len())
	ItsEqual(t, 5, rh.MinHeight())
	ItsEqual(t, 5, rh.MaxHeight())

	output = rh.RemoveMinHeight()
	ItsEqual(t, 0, rh.Len())
	ItsEqual(t, 5, len(output))
	ItsNil(t, rh.heights[0].head)
	ItsNil(t, rh.heights[0].tail)
	ItsEqual(t, 0, rh.heights[0].len)
	ItsNil(t, rh.heights[1].head)
	ItsNil(t, rh.heights[1].tail)
	ItsEqual(t, 0, rh.heights[1].len)
	ItsNil(t, rh.heights[5].head)
	ItsNil(t, rh.heights[5].tail)
	ItsEqual(t, 0, rh.heights[5].len)
}

func Test_recomputeHeap_Remove(t *testing.T) {
	rh := newRecomputeHeap(10)
	n10 := newHeightIncr(1)
	n11 := newHeightIncr(1)
	n20 := newHeightIncr(2)
	n21 := newHeightIncr(2)
	n22 := newHeightIncr(2)
	n30 := newHeightIncr(3)

	// this should just return
	rh.Remove(n10)

	rh.Add(n10)
	rh.Add(n11)
	rh.Add(n20)
	rh.Add(n21)
	rh.Add(n22)
	rh.Add(n30)

	ItsEqual(t, true, rh.Has(n10))
	ItsEqual(t, true, rh.Has(n11))
	ItsEqual(t, true, rh.Has(n20))
	ItsEqual(t, true, rh.Has(n21))
	ItsEqual(t, true, rh.Has(n22))
	ItsEqual(t, true, rh.Has(n30))

	rh.Remove(n21)

	ItsEqual(t, 5, rh.Len())
	ItsEqual(t, true, rh.Has(n10))
	ItsEqual(t, true, rh.Has(n11))
	ItsEqual(t, true, rh.Has(n20))
	ItsEqual(t, false, rh.Has(n21))
	ItsEqual(t, true, rh.Has(n22))
	ItsEqual(t, true, rh.Has(n30))

	ItsEqual(t, 2, rh.heights[2].len)
	ItsEqual(t, n20.Node().id, rh.heights[2].head.value.Node().id)
	ItsEqual(t, n22.Node().id, rh.heights[2].tail.value.Node().id)

	rh.Remove(n10)
	rh.Remove(n11)

	ItsEqual(t, 3, rh.Len())
	ItsEqual(t, 0, rh.heights[1].len)
	ItsNil(t, rh.heights[1].head)
	ItsNil(t, rh.heights[1].tail)
	ItsEqual(t, 2, rh.MinHeight())
	ItsEqual(t, 3, rh.MaxHeight())
}

func Test_recomputeHeapList(t *testing.T) {
	q := new(recomputeHeapList)

	n0 := newHeightIncr(0)
	n1 := newHeightIncr(0)
	n2 := newHeightIncr(0)
	n3 := newHeightIncr(0)

	var zeroID Identifier
	id, n := q.pop()

	ItsNil(t, nil, q.head)
	ItsNil(t, nil, q.tail)
	ItsEqual(t, zeroID, id)
	ItsNil(t, n)
	ItsEqual(t, q.head, q.tail)
	ItsEqual(t, 0, q.len)

	q.push(n0)
	ItsEqual(t, true, q.head != nil)
	ItsEqual(t, nil, q.head.previous)
	ItsEqual(t, true, q.tail != nil)
	ItsEqual(t, nil, q.tail.previous)
	ItsEqual(t, q.head, q.tail)
	ItsEqual(t, 1, q.len)
	ItsEqual(t, n0.n.id, q.head.value.Node().id)
	ItsEqual(t, n0.n.id, q.tail.value.Node().id)

	q.push(n1)
	ItsEqual(t, true, q.head != nil)
	ItsEqual(t, true, q.head.previous != nil)
	ItsEqual(t, nil, q.head.previous.previous)
	ItsEqual(t, q.head.previous, q.tail)
	ItsEqual(t, true, q.tail != nil)
	ItsEqual(t, true, q.tail.next != nil)
	ItsEqual(t, true, q.head != q.tail)
	ItsEqual(t, 2, q.len)
	ItsEqual(t, n0.n.id, q.head.value.Node().id)
	ItsEqual(t, n1.n.id, q.tail.value.Node().id)

	q.push(n2)
	ItsEqual(t, true, q.head != nil)
	ItsEqual(t, true, q.head.previous != nil)
	ItsEqual(t, true, q.head.previous.previous != nil)
	ItsEqual(t, nil, q.head.previous.previous.previous)
	ItsEqual(t, q.head.previous.previous, q.tail)
	ItsEqual(t, true, q.tail != nil)
	ItsEqual(t, true, q.head != q.tail)
	ItsEqual(t, 3, q.len)
	ItsEqual(t, n0.n.id, q.head.value.Node().id)
	ItsEqual(t, n2.n.id, q.tail.value.Node().id)

	q.push(n3)
	ItsEqual(t, true, q.head != nil)
	ItsEqual(t, true, q.head.previous != nil)
	ItsEqual(t, true, q.head.previous.previous != nil)
	ItsEqual(t, true, q.head.previous.previous.previous != nil)
	ItsEqual(t, false, q.head.previous.previous.previous.previous != nil)
	ItsEqual(t, q.head.previous.previous.previous, q.tail)
	ItsEqual(t, true, q.tail != nil)
	ItsEqual(t, true, q.head != q.tail)
	ItsEqual(t, 4, q.len)
	ItsEqual(t, n0.n.id, q.head.value.Node().id)
	ItsEqual(t, n3.n.id, q.tail.value.Node().id)

	id, n = q.pop()
	ItsEqual(t, n0.n.id, id)
	ItsEqual(t, n0.n.id, n.Node().id)
	ItsEqual(t, true, q.head != nil)
	ItsEqual(t, true, q.head.previous != nil)
	ItsEqual(t, true, q.head.previous.previous != nil)
	ItsEqual(t, nil, q.head.previous.previous.previous)
	ItsEqual(t, q.head.previous.previous, q.tail)
	ItsEqual(t, true, q.tail != nil)
	ItsEqual(t, true, q.head != q.tail)
	ItsEqual(t, q.len, 3)
	ItsEqual(t, n1.n.id, q.head.value.Node().id)
	ItsEqual(t, n3.n.id, q.tail.value.Node().id)

	id, n = q.pop()
	ItsEqual(t, n1.n.id, id)
	ItsEqual(t, n1.n.id, n.Node().id)
	ItsEqual(t, true, q.head != nil)
	ItsEqual(t, true, q.head.previous != nil)
	ItsEqual(t, nil, q.head.previous.previous)
	ItsEqual(t, q.head.previous, q.tail)
	ItsEqual(t, true, q.tail != nil)
	ItsEqual(t, true, q.head != q.tail)
	ItsEqual(t, 2, q.len)
	ItsEqual(t, n2.n.id, q.head.value.Node().id)
	ItsEqual(t, n3.n.id, q.tail.value.Node().id)

	id, n = q.pop()
	ItsEqual(t, n2.n.id, id)
	ItsEqual(t, n2.n.id, n.Node().id)
	ItsEqual(t, true, q.head != nil)
	ItsEqual(t, nil, q.head.previous)
	ItsEqual(t, true, q.tail != nil)
	ItsEqual(t, q.head, q.tail)
	ItsEqual(t, 1, q.len)
	ItsEqual(t, n3.n.id, q.head.value.Node().id)
	ItsEqual(t, n3.n.id, q.tail.value.Node().id)

	id, n = q.pop()
	ItsEqual(t, n3.n.id, id)
	ItsEqual(t, n3.n.id, n.Node().id)
	ItsEqual(t, nil, q.head)
	ItsEqual(t, nil, q.tail)
	ItsEqual(t, 0, q.len)

	q.push(n0)
	q.push(n1)
	q.push(n2)
	ItsEqual(t, 3, q.len)
}

func Test_recomputeHeapList_find(t *testing.T) {
	q := new(recomputeHeapList)

	n0 := newHeightIncr(0)
	n1 := newHeightIncr(0)
	n2 := newHeightIncr(0)
	n3 := newHeightIncr(0)
	q.push(n0)
	q.push(n1)
	q.push(n2)
	q.push(n3)

	ItsNotNil(t, q.head.next)
	ItsNotNil(t, q.head.next.next)
	ItsNotNil(t, q.head.next.next.next)

	found := q.find(n3)
	ItsNotNil(t, found)
	ItsEqual(t, found.key, n3.Node().id)

	found = q.find(n2)
	ItsNotNil(t, found)
	ItsEqual(t, found.key, n2.Node().id)

	found = q.find(n1)
	ItsNotNil(t, found)
	ItsEqual(t, found.key, n1.Node().id)

	found = q.find(n0)
	ItsNotNil(t, found)
	ItsEqual(t, found.key, n0.Node().id)
}

func Test_recomputeHeapList_remove_tail(t *testing.T) {
	q := new(recomputeHeapList)

	n0 := newHeightIncr(0)
	n1 := newHeightIncr(0)
	n2 := newHeightIncr(0)
	n3 := newHeightIncr(0)
	q.push(n0)
	q.push(n1)
	q.push(n2)
	q.push(n3)

	q.remove(q.find(n3))
}

func Test_recomputeHeapList_popAll(t *testing.T) {
	n0 := newHeightIncr(0)
	n1 := newHeightIncr(0)
	n2 := newHeightIncr(0)

	rhl := new(recomputeHeapList)
	rhl.push(n0)
	rhl.push(n1)
	rhl.push(n2)

	ItsNotNil(t, rhl.head)
	ItsNotNil(t, rhl.tail)
	ItsEqual(t, 3, rhl.len)

	output := rhl.popAll()
	ItsEqual(t, 3, len(output))
	ItsEqual(t, 0, rhl.len)
	ItsNil(t, rhl.head)
	ItsNil(t, rhl.tail)
}
