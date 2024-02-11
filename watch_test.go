package incr

import (
	"fmt"
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_Watch(t *testing.T) {
	g := New()
	r0 := Return(g, "hello")
	w0 := Watch(g, r0)
	w0.Node().SetLabel("w0")

	testutil.Matches(t, "watch\\[.*\\]:w0", w0.(fmt.Stringer).String())
}
