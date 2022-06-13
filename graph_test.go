package incr

import "testing"

func Test_New(t *testing.T) {
	r0 := Return("hello")
	r1 := Return("world!")
	m0 := Map2(r0, r1, func(v0, v1 string) string { return v0 + v1 })
	g := New(m0)

	ItsEqual(t, 3, len(g.observed))
	ItsEqual(t, true, g.isObserving(r0))
	ItsEqual(t, true, g.isObserving(r1))
	ItsEqual(t, true, g.isObserving(m0))

	m1 := Map2(r0, r1, func(v0, v1 string) string { return v0 + v1 })
	ItsEqual(t, false, g.isObserving(m1))
}
