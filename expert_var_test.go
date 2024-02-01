package incr

import (
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_ExpertVar(t *testing.T) {
	v := Var(testContext(), "hello")
	ev := ExpertVar(v)
	ev.SetValue("not-hello")
	testutil.ItsEqual(t, "not-hello", v.Value())
}
