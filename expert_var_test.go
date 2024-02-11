package incr

import (
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_ExpertVar(t *testing.T) {
	g := New()
	v := Var(g, "hello")
	ev := ExpertVar(v)
	ev.SetInternalValue("not-hello")
	testutil.Equal(t, "not-hello", v.Value())
}
