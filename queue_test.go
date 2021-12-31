/*

Copyright (c) 2021 - Present. Blend Labs, Inc. All rights reserved
Use of this source code is governed by a MIT license that can be found in the LICENSE file.

*/

package incr

import (
	"testing"
)

func okValue[A any](v A, _ bool) A {
	return v
}

func Test_Queue(t *testing.T) {
	buffer := new(Queue[int])

	buffer.Push(1)
	itsEqual(t, 1, buffer.Len())
	itsEqual(t, 1, okValue(buffer.Peek()))
	itsEqual(t, 1, okValue(buffer.PeekBack()))

	buffer.Push(2)
	itsEqual(t, 2, buffer.Len())
	itsEqual(t, 1, okValue(buffer.Peek()))
	itsEqual(t, 2, okValue(buffer.PeekBack()))

	buffer.Push(3)
	itsEqual(t, 3, buffer.Len())
	itsEqual(t, 1, okValue(buffer.Peek()))
	itsEqual(t, 3, okValue(buffer.PeekBack()))

	buffer.Push(4)
	itsEqual(t, 4, buffer.Len())
	itsEqual(t, 1, okValue(buffer.Peek()))
	itsEqual(t, 4, okValue(buffer.PeekBack()))

	buffer.Push(5)
	itsEqual(t, 5, buffer.Len())
	itsEqual(t, 1, okValue(buffer.Peek()))
	itsEqual(t, 5, okValue(buffer.PeekBack()))

	buffer.Push(6)
	itsEqual(t, 6, buffer.Len())
	itsEqual(t, 1, okValue(buffer.Peek()))
	itsEqual(t, 6, okValue(buffer.PeekBack()))

	buffer.Push(7)
	itsEqual(t, 7, buffer.Len())
	itsEqual(t, 1, okValue(buffer.Peek()))
	itsEqual(t, 7, okValue(buffer.PeekBack()))

	buffer.Push(8)
	itsEqual(t, 8, buffer.Len())
	itsEqual(t, 1, okValue(buffer.Peek()))
	itsEqual(t, 8, okValue(buffer.PeekBack()))

	value, ok := buffer.Pop()
	itsEqual(t, true, ok)
	itsEqual(t, 1, value)
	itsEqual(t, 7, buffer.Len())
	itsEqual(t, 2, okValue(buffer.Peek()))
	itsEqual(t, 8, okValue(buffer.PeekBack()))

	value, ok = buffer.Pop()
	itsEqual(t, true, ok)
	itsEqual(t, 2, value)
	itsEqual(t, 6, buffer.Len())
	itsEqual(t, 3, okValue(buffer.Peek()))
	itsEqual(t, 8, okValue(buffer.PeekBack()))

	value, ok = buffer.Pop()
	itsEqual(t, true, ok)
	itsEqual(t, 3, value)
	itsEqual(t, 5, buffer.Len())
	itsEqual(t, 4, okValue(buffer.Peek()))
	itsEqual(t, 8, okValue(buffer.PeekBack()))

	value, ok = buffer.Pop()
	itsEqual(t, true, ok)
	itsEqual(t, 4, value)
	itsEqual(t, 4, buffer.Len())
	itsEqual(t, 5, okValue(buffer.Peek()))
	itsEqual(t, 8, okValue(buffer.PeekBack()))

	value, ok = buffer.Pop()
	itsEqual(t, true, ok)
	itsEqual(t, 5, value)
	itsEqual(t, 3, buffer.Len())
	itsEqual(t, 6, okValue(buffer.Peek()))
	itsEqual(t, 8, okValue(buffer.PeekBack()))

	value, ok = buffer.Pop()
	itsEqual(t, true, ok)
	itsEqual(t, 6, value)
	itsEqual(t, 2, buffer.Len())
	itsEqual(t, 7, okValue(buffer.Peek()))
	itsEqual(t, 8, okValue(buffer.PeekBack()))

	value, ok = buffer.Pop()
	itsEqual(t, true, ok)
	itsEqual(t, 7, value)
	itsEqual(t, 1, buffer.Len())
	itsEqual(t, 8, okValue(buffer.Peek()))
	itsEqual(t, 8, okValue(buffer.PeekBack()))

	value, ok = buffer.Pop()
	itsEqual(t, true, ok)
	itsEqual(t, 8, value)
	itsEqual(t, 0, buffer.Len())
	itsEqual(t, 0, okValue(buffer.Peek()))
	itsEqual(t, 0, okValue(buffer.PeekBack()))
}

func Test_QueueClear(t *testing.T) {
	buffer := new(Queue[int])
	buffer.Push(1)
	buffer.Push(1)
	buffer.Push(1)
	buffer.Push(1)
	buffer.Push(1)
	buffer.Push(1)
	buffer.Push(1)
	buffer.Push(1)

	itsEqual(t, 8, buffer.Len())

	buffer.Clear()

	itsEqual(t, 0, buffer.Len())
	itsEqual(t, 0, okValue(buffer.Peek()))
	itsEqual(t, 0, okValue(buffer.PeekBack()))
}

func Test_QueueEach(t *testing.T) {
	buffer := new(Queue[int])

	for x := 1; x < 17; x++ {
		buffer.Push(x)
	}

	called := 0
	buffer.Each(func(v int) error {
		if v == (called + 1) {
			called++
		}
		return nil
	})

	itsEqual(t, 16, called)
}

func Test_QueueReverseEach(t *testing.T) {
	buffer := new(Queue[int])

	for x := 1; x < 17; x++ {
		buffer.Push(x)
	}

	called := 17
	buffer.ReverseEach(func(v int) error {
		if v == (called - 1) {
			called--
		}
		return nil
	})

	itsEqual(t, 1, called)
}
