package incr

import (
	"context"
	"math"
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

// testContext returns a test context.
func testContext() context.Context {
	ctx := context.Background()
	ctx = testutil.WithBlueDye(ctx)
	ctx = WithTracing(ctx)
	return ctx
}

func epsilon(delta float64) func(float64, float64) bool {
	return func(v0, v1 float64) bool {
		return math.Abs(v1-v0) <= delta
	}
}

func epsilonContext(t *testing.T, delta float64) func(context.Context, float64, float64) (bool, error) {
	t.Helper()
	return func(ctx context.Context, v0, v1 float64) (bool, error) {
		t.Helper()
		testutil.ItsBlueDye(ctx, t)
		return math.Abs(v1-v0) <= delta, nil
	}
}

// addConst returs a map fn that adds a constant value
// to a given input
func addConst(v float64) func(float64) float64 {
	return func(v0 float64) float64 {
		return v0 + v
	}
}

// add is a map2 fn that adds two values and returns the result
func add[T Ordered](v0, v1 T) T {
	return v0 + v1
}

func ident[T any](v T) T {
	return v
}

func identMany[T any](v ...T) (out T) {
	if len(v) > 0 {
		out = v[0]
	}
	return
}

var _ Incr[any] = (*mockBareNode)(nil)

func newMockBareNode() *mockBareNode {
	return &mockBareNode{
		n: NewNode(),
	}
}

type mockBareNode struct {
	n *Node
}

func (mn *mockBareNode) Node() *Node {
	return mn.n
}

func (mn *mockBareNode) Value() any {
	return nil
}

func newHeightIncr(height int) *heightIncr {
	return &heightIncr{
		n: &Node{
			id:     NewIdentifier(),
			height: height,
		},
	}
}

func newHeightIncrLabel(height int, label string) *heightIncr {
	return &heightIncr{
		n: &Node{
			id:     NewIdentifier(),
			height: height,
			label:  label,
		},
	}
}

type heightIncr struct {
	Incr[struct{}]
	n *Node
}

func (hi heightIncr) Node() *Node {
	return hi.n
}
