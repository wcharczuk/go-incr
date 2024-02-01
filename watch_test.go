package incr

import (
	"fmt"
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_Watch(t *testing.T) {
	ctx := testContext()
	r0 := Return(ctx, "hello")
	w0 := Watch(ctx, r0)
	w0.Node().SetLabel("w0")

	testutil.ItMatches(t, "watch\\[.*\\]:w0", w0.(fmt.Stringer).String())
}
