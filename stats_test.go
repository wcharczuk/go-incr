package incr

import (
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_NodeStats(t *testing.T) {
	v0 := Var("a")
	v0.Node().SetLabel("v0")

	av := Var("a-value")
	a0 := Map(av, ident)
	a1 := Map(a0, ident)

	bv := Var("b-value")
	b0 := Map(bv, ident)
	b1 := Map(b0, ident)
	b2 := Map(b1, ident)

	bind := Bind(v0, func(which string) Incr[string] {
		if which == "a" {
			return a1
		}
		if which == "b" {
			return b2
		}
		return nil
	})

	s0 := Return("hello")
	s1 := Map(s0, ident)
	o := Map2(bind, s1, concat)

	_ = New(o)

	oStats := NodeStats(o)
	testutil.ItsEqual(t, 2, oStats.Children())
	testutil.ItsEqual(t, 0, oStats.Parents())
}
