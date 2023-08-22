package incr

import (
	"fmt"
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_Watch(t *testing.T) {
	r0 := Return("hello")
	w0 := Watch(r0)
	w0.Node().SetLabel("w0")

	testutil.ItMatches(t, "watch\\[.*\\]:w0", w0.(fmt.Stringer).String())
}
