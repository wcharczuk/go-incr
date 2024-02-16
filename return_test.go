package incr

import (
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_Return(t *testing.T) {
	g := New()
	r := Return(g, "hello")

	testutil.Equal(t, "hello", r.Value())

	testutil.Equal(t, false, r.(*returnIncr[string]).ShouldBeInvalidated())
}
