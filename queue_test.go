package incr

import (
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_queue_filter(t *testing.T) {
	q := new(queue[Identifier])

	id00 := NewIdentifier()
	id01 := NewIdentifier()
	id02 := NewIdentifier()

	q.push(id00)
	q.push(id01)
	q.push(id02)

	q.filter(func(id Identifier) bool {
		return id != id01
	})

	testutil.ItsEqual(t, 2, q.len())
	testutil.ItsEqual(t, 4, q.cap())
	testutil.ItsEqual(t, id00, q.array[0])
	testutil.ItsEqual(t, id02, q.array[1])
	testutil.ItsEqual(t, true, q.array[2].IsZero())
	testutil.ItsEqual(t, true, q.array[3].IsZero())

	q.push(NewIdentifier())
	q.push(NewIdentifier())

	testutil.ItsEqual(t, 4, q.len())
	testutil.ItsEqual(t, 4, q.cap())
	testutil.ItsEqual(t, id00, q.array[0])
	testutil.ItsEqual(t, id02, q.array[1])
	testutil.ItsEqual(t, false, q.array[2].IsZero())
	testutil.ItsEqual(t, false, q.array[3].IsZero())

	q.push(NewIdentifier())
	q.push(NewIdentifier())

	testutil.ItsEqual(t, 6, q.len())
	testutil.ItsEqual(t, 8, q.cap())
	testutil.ItsEqual(t, id00, q.array[0])
	testutil.ItsEqual(t, id02, q.array[1])
	testutil.ItsEqual(t, false, q.array[2].IsZero())
	testutil.ItsEqual(t, false, q.array[3].IsZero())
	testutil.ItsEqual(t, false, q.array[4].IsZero())
	testutil.ItsEqual(t, false, q.array[5].IsZero())
	testutil.ItsEqual(t, true, q.array[6].IsZero())
	testutil.ItsEqual(t, true, q.array[7].IsZero())
}
