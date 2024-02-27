package incr

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
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
		testutil.BlueDye(ctx, t)
		return math.Abs(v1-v0) <= delta, nil
	}
}

type mathTypes interface {
	~int | ~float64
}

func epsilonFn[A, B mathTypes](eps A, oldv, newv B) bool {
	if oldv > newv {
		return oldv-newv <= B(eps)
	}
	return newv-oldv <= B(eps)
}

func concat(a, b string) string {
	return a + b
}

func mapAppend(suffix string) func(string) string {
	return func(v string) string {
		return v + suffix
	}
}

type addable interface {
	~int | ~float64
}

func add[T addable](v0, v1 T) T {
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

func mockObserver(scope Scope) IObserver {
	return WithinScope(scope, &observeIncr[any]{
		n: NewNode("mock_observer"),
	})
}

func newMockBareNodeWithHeight(scope Scope, height int) *mockBareNode {
	mbn := WithinScope(scope, &mockBareNode{
		n: NewNode("bare_node"),
	})
	mbn.n.height = height
	return mbn
}

func newMockBareNode(scope Scope) *mockBareNode {
	o := WithinScope(scope, &mockBareNode{
		n: NewNode("bare_node"),
	})
	o.Node().height = 0
	return o
}

type mockBareNode struct {
	n       *Node
	parents []INode
}

func (mn *mockBareNode) String() string { return mn.n.String() }

func (mn *mockBareNode) Parents() []INode { return mn.parents }

func (mn *mockBareNode) Node() *Node {
	return mn.n
}

func (mn *mockBareNode) Value() any {
	return nil
}

func newHeightIncr(scope Scope, height int) *heightIncr {
	return WithinScope(scope, &heightIncr{
		n: &Node{
			id:     NewIdentifier(),
			height: height,
		},
	})
}

type heightIncr struct {
	Incr[struct{}]
	n *Node
}

func (hi *heightIncr) String() string { return hi.n.String() }

func (hi *heightIncr) Parents() []INode { return nil }

func (hi heightIncr) Node() *Node {
	return hi.n
}

func newList(items ...INode) *recomputeHeapList {
	l := new(recomputeHeapList)
	for _, i := range items {
		i.Node().heightInRecomputeHeap = i.Node().height
		l.push(i)
	}
	return l
}

func createDynamicMaps(scope Scope, label string) Incr[string] {
	mapVar0 := Var(scope, fmt.Sprintf("%s-0", label))
	mapVar0.Node().SetLabel(fmt.Sprintf("%sv-0", label))
	mapVar1 := Var(scope, fmt.Sprintf("%s-1", label))
	mapVar1.Node().SetLabel(fmt.Sprintf("%sv-1", label))
	m := Map2(scope, mapVar0, mapVar1, func(a, b string) string {
		return a + "+" + b
	})
	m.Node().SetLabel(label)
	return m
}

func createDynamicBind(scope Scope, label string, a, b Incr[string]) (VarIncr[string], BindIncr[string]) {
	bindVar := Var(scope, "a")
	bindVar.Node().SetLabel(fmt.Sprintf("bind - %s - var", label))
	bind := Bind(scope, bindVar, func(scope Scope, which string) Incr[string] {
		if which == "a" {
			m := Map(scope, a, func(v string) string {
				return v + "->" + label
			})
			m.Node().SetLabel(fmt.Sprintf("bind - %s - %s - map", label, which))
			return m
		}
		if which == "b" {
			m := Map(scope, b, func(v string) string {
				return v + "->" + label
			})
			m.Node().SetLabel(fmt.Sprintf("bind - %s - %s - map", label, which))
			return m
		}
		return nil
	})
	bind.Node().SetLabel(fmt.Sprintf("bind - %s", label))
	return bindVar, bind
}

func homedir(filename string) string {
	var rootDir string
	if rootDir = os.Getenv("INCR_DEBUG_DOT_ROOT"); rootDir == "" {
		rootDir = os.ExpandEnv("$HOME/Desktop")
	}
	return filepath.Join(rootDir, filename)
}

func dumpDot(g *Graph, path string) error {
	if os.Getenv("INCR_DEBUG_DOT") != "true" {
		return nil
	}

	dotContents := new(bytes.Buffer)
	if err := Dot(dotContents, g); err != nil {
		return err
	}

	dotOutput, err := os.Create(os.ExpandEnv(path))
	if err != nil {
		return err
	}
	defer func() { _ = dotOutput.Close() }()
	dotFullPath, err := exec.LookPath("dot")
	if err != nil {
		return fmt.Errorf("there was an issue finding `dot` in your path; you may need to install the `graphviz` package or similar on your platform: %w", err)
	}

	errOut := new(bytes.Buffer)
	cmd := exec.Command(dotFullPath, "-Tpng")
	cmd.Stdin = dotContents
	cmd.Stdout = dotOutput
	cmd.Stderr = errOut
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%v; %w", errOut.String(), err)
	}
	return nil
}

func hasKey[A INode](nodes []A, id Identifier) bool {
	for _, n := range nodes {
		if n.Node().id == id {
			return true
		}
	}
	return false
}

func cutoffAlways[A, B any](scope Scope, input Incr[A], cutoff func(context.Context, A) (bool, error), fn func(context.Context, A) (B, error)) Incr[B] {
	return WithinScope(scope, &ccutoffAlwaysIncr[A, B]{
		n:      NewNode("cutoff-always"),
		input:  input,
		cutoff: cutoff,
		fn:     fn,
	})
}

type ccutoffAlwaysIncr[A, B any] struct {
	n      *Node
	input  Incr[A]
	cutoff func(context.Context, A) (bool, error)
	fn     func(context.Context, A) (B, error)
	value  B
}

func (n *ccutoffAlwaysIncr[A, B]) Parents() []INode {
	return []INode{n.input}
}

func (n *ccutoffAlwaysIncr[A, B]) Node() *Node { return n.n }

func (n *ccutoffAlwaysIncr[A, B]) Value() B { return n.value }

func (n *ccutoffAlwaysIncr[A, B]) Always() {}

func (n *ccutoffAlwaysIncr[A, B]) Cutoff(ctx context.Context) (bool, error) {
	return n.cutoff(ctx, n.input.Value())
}

func (n *ccutoffAlwaysIncr[A, B]) Stabilize(ctx context.Context) (err error) {
	var value B
	value, err = n.fn(ctx, n.input.Value())
	if err != nil {
		return
	}
	n.value = value
	return
}

func (n *ccutoffAlwaysIncr[A, B]) String() string {
	return n.n.String()
}
