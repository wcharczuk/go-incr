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
		testutil.ItsBlueDye(ctx, t)
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
		n: NewNode("mock_observer"),
	}
}

func newMockBareNodeWithHeight(height int) *mockBareNode {
	mbn := &mockBareNode{
		n: NewNode("bare_node"),
	}
	mbn.n.height = height
	return mbn
}

func newMockBareNode() *mockBareNode {
	return &mockBareNode{
		n: NewNode("bare_node"),
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

func newList(items ...INode) map[Identifier]INode {
	l := make(map[Identifier]INode, len(items))
	for _, i := range items {
		i.Node().heightInRecomputeHeap = i.Node().height
		l[i.Node().id] = i
	}
	return l
}

func createDynamicMaps(scope *BindScope, label string) Incr[string] {
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

func createDynamicBind(scope *BindScope, label string, a, b Incr[string]) (VarIncr[string], BindIncr[string]) {
	bindVar := Var(scope, "a")
	bindVar.Node().SetLabel(fmt.Sprintf("bind - %s - var", label))
	bind := Bind(scope, bindVar, func(scope *BindScope, which string) Incr[string] {
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
