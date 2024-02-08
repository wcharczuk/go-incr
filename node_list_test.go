package incr

import (
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_remove(t *testing.T) {
	n0 := newMockBareNode()
	n1 := newMockBareNode()
	n2 := newMockBareNode()
	nodes := []INode{
		n0, n1, n2,
	}
	nodes = remove(nodes, n1.Node().id)

	testutil.ItsEqual(t, 2, len(nodes))
}
