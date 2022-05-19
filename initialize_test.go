package incr

import (
	"testing"
)

func Test_Initialize(t *testing.T) {
	ctx := testContext()

	v0 := Var("foo")
	v1 := Var("moo")
	v2 := Var("bar")
	v3 := Var("baz")

	m0 := Map2(v0.Read(), v1.Read(), func(a, b string) string {
		return a + " " + b
	})
	m1 := Map2(v2.Read(), v3.Read(), func(a, b string) string {
		return a + "/" + b
	})
	r0 := Return("hello")
	m2 := Map2(m0, r0, func(a, b string) string {
		return a + "+" + b
	})
	m3 := Map2(m1, m2, func(a, b string) string {
		return a + "+" + b
	})

	Initialize(ctx, m3)

	ItsEqual(t, 1, v0.Node().height)
	ItsEqual(t, 1, v1.Node().height)
	ItsEqual(t, 1, v2.Node().height)
	ItsEqual(t, 1, v3.Node().height)
	ItsEqual(t, 1, r0.Node().height)

	ItsEqual(t, 2, m0.Node().height)
	ItsEqual(t, 2, m1.Node().height)

	ItsEqual(t, 3, m2.Node().height)
	ItsEqual(t, 4, m3.Node().height)
}
