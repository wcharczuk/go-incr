package incr

import (
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_remove(t *testing.T) {
	g := New()

	n0 := newMockBareNode(g)
	n1 := newMockBareNode(g)
	n2 := newMockBareNode(g)
	nodes := []INode{
		n0, n1, n2,
	}
	nodes = remove(nodes, n1.Node().id)

	testutil.Equal(t, 2, len(nodes))
}
