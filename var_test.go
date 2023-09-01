package incr

import (
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_Var_Set_unobserved(t *testing.T) {
	v := Var("foo")
	v.Set("not-foo")

	testutil.ItsEqual(t, "not-foo", v.(*varIncr[string]).setValue)
}
