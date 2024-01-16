package incr

import (
	"context"
	"fmt"
	"math"
	"os"
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

// testContext returns a test context.
func testContext() context.Context {
	ctx := context.Background()
	ctx = testutil.WithBlueDye(ctx)
	if os.Getenv("INCR_DEBUG_TRACING") != "" {
		ctx = WithTracing(ctx)
	}
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

func newMockBareNodeWithHeight(height int) *mockBareNode {
	mbn := &mockBareNode{
		n: NewNode(),
	}
	mbn.n.height = height
	return mbn
}

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

func allHeight(values []INode, height int) bool {
	for _, v := range values {
		if v.Node().height != height {
			return false
		}
	}
	return true
}

func newList(items ...INode) *list[Identifier, INode] {
	l := new(list[Identifier, INode])
	for _, i := range items {
		l.Push(i.Node().id, i, i.Node().height)
	}
	return l
}

func newListWithItems(items ...INode) (l *list[Identifier, INode], outputItems []*listItem[Identifier, INode]) {
	l = new(list[Identifier, INode])
	for _, i := range items {
		outputItems = append(outputItems, l.Push(i.Node().id, i, i.Node().height))
	}
	return
}

func createDynamicMaps(label string) Incr[string] {
	mapVar0 := Var(fmt.Sprintf("%s-0", label))
	mapVar1 := Var(fmt.Sprintf("%s-1", label))
	m := Map2(mapVar0, mapVar1, func(a, b string) string {
		return a + "+" + b
	})
	m.Node().SetLabel(label)
	return m
}

func createDynamicBind(label string, a, b Incr[string]) (VarIncr[string], BindIncr[string]) {
	bindVar := Var("a")
	bindVar.Node().SetLabel(fmt.Sprintf("bind - %s - var", label))

	bind := Bind(bindVar, func(which string) Incr[string] {
		if which == "a" {
			return Map(a, func(v string) string {
				return v + "->" + label
			})
		}
		if which == "b" {
			return Map(b, func(v string) string {
				return v + "->" + label
			})
		}
		return nil
	})
	bind.Node().SetLabel(fmt.Sprintf("bind - %s", label))
	return bindVar, bind
}
