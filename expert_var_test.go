package incr

import (
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_ExpertVar(t *testing.T) {
	v := Var(Root(), "hello")
	ev := ExpertVar(v)
	ev.SetInternalValue("not-hello")
	testutil.Equal(t, "not-hello", v.Value())
}
