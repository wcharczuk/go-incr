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

func mockObserver() IObserver {
	return &observeIncr[any]{
		n: NewNode(),
	}
}

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

func allHeight(values []*recomputeHeapItem, height int) bool {
	for _, v := range values {
		if v.node.Node().height != height {
			return false
		}
	}
	return true
}

func newList(items ...INode) map[Identifier]*recomputeHeapItem {
	l := make(map[Identifier]*recomputeHeapItem, len(items))
	for _, i := range items {
		l[i.Node().id] = &recomputeHeapItem{node: i, height: i.Node().height}
	}
	return l
}

func newListWithItems(items ...INode) (l map[Identifier]*recomputeHeapItem, outputItems []*recomputeHeapItem) {
	l = make(map[Identifier]*recomputeHeapItem)
	for _, i := range items {
		newItem := &recomputeHeapItem{node: i, height: i.Node().height}
		l[i.Node().id] = newItem
		outputItems = append(outputItems, newItem)
	}
	return
}

func createDynamicMaps(ctx context.Context, label string) Incr[string] {
	mapVar0 := Var(ctx, fmt.Sprintf("%s-0", label))
	mapVar0.Node().SetLabel(fmt.Sprintf("%sv-0", label))
	mapVar1 := Var(ctx, fmt.Sprintf("%s-1", label))
	mapVar1.Node().SetLabel(fmt.Sprintf("%sv-1", label))
	m := Map2(ctx, mapVar0, mapVar1, func(a, b string) string {
		return a + "+" + b
	})
	m.Node().SetLabel(label)
	return m
}

func createDynamicBind(ctx context.Context, label string, a, b Incr[string]) (VarIncr[string], BindIncr[string]) {
	bindVar := Var(ctx, "a")
	bindVar.Node().SetLabel(fmt.Sprintf("bind - %s - var", label))

	bind := Bind(ctx, bindVar, func(ctx context.Context, which string) Incr[string] {
		if which == "a" {
			return Map(ctx, a, func(v string) string {
				return v + "->" + label
			})
		}
		if which == "b" {
			return Map(ctx, b, func(v string) string {
				return v + "->" + label
			})
		}
		return nil
	})
	bind.Node().SetLabel(fmt.Sprintf("bind - %s", label))
	return bindVar, bind
}
