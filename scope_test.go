package incr

import (
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_graphFromNodeCreatedIn(t *testing.T) {
	g := New()
	n := newMockBareNode(g)
	verify := GraphForNode(n)
	testutil.NotNil(t, verify)
	testutil.Equal(t, verify.id, g.id)
}

func Test_graphFromNodeCreatedIn_unset(t *testing.T) {
	g := GraphForNode(nil)
	testutil.Nil(t, g)
}
