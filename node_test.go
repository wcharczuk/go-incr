package incr

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_NewNode(t *testing.T) {
	n := NewNode()
	testutil.ItsNotNil(t, n.id)
	testutil.ItsNil(t, n.graph)
	testutil.ItsEqual(t, 0, len(n.parents))
	testutil.ItsEqual(t, 0, len(n.children))
	testutil.ItsEqual(t, "", n.label)
	testutil.ItsEqual(t, 0, n.height)
	testutil.ItsEqual(t, 0, n.changedAt)
	testutil.ItsEqual(t, 0, n.setAt)
	testutil.ItsEqual(t, 0, n.recomputedAt)
	testutil.ItsNil(t, n.onUpdateHandlers)
	testutil.ItsNil(t, n.onErrorHandlers)
	testutil.ItsNil(t, n.stabilize)
	testutil.ItsNil(t, n.cutoff)
	testutil.ItsEqual(t, 0, n.numRecomputes)
	testutil.ItsNil(t, n.metadata)
}

func Test_Node_ID(t *testing.T) {
	n := NewNode()
	testutil.ItsEqual(t, false, n.ID().IsZero())

	other := NewIdentifier()
	n.id = other
	testutil.ItsEqual(t, other, n.ID())
}

func Test_Node_Label(t *testing.T) {
	n := NewNode()
	testutil.ItsEqual(t, "", n.Label())
	n.SetLabel("foo")
	testutil.ItsEqual(t, "foo", n.Label())
}

func Test_Node_Metadata(t *testing.T) {
	n := NewNode()
	testutil.ItsEqual(t, nil, n.Metadata())
	n.SetMetadata("foo")
	testutil.ItsEqual(t, "foo", n.Metadata())
}

func Test_Link(t *testing.T) {
	c := newMockBareNode()
	p0 := newMockBareNode()
	p1 := newMockBareNode()
	p2 := newMockBareNode()

	// set up P with (3) inputs
	Link(c, p0, p1, p2)

	// no nodes depend on p, p is not an input to any nodes
	testutil.ItsEqual(t, 3, len(c.n.parents))
	testutil.ItsEqual(t, 0, len(c.n.children))

	testutil.ItsEqual(t, true, hasKey(c.n.parents, p0.n.id))
	testutil.ItsEqual(t, true, hasKey(c.n.parents, p1.n.id))
	testutil.ItsEqual(t, true, hasKey(c.n.parents, p2.n.id))

	testutil.ItsEqual(t, 1, len(p0.n.children))
	testutil.ItsEqual(t, true, hasKey(p0.n.children, c.n.id))
	testutil.ItsEqual(t, 0, len(p0.n.parents))

	testutil.ItsEqual(t, 1, len(p1.n.children))
	testutil.ItsEqual(t, true, hasKey(p1.n.children, c.n.id))
	testutil.ItsEqual(t, 0, len(p1.n.parents))

	testutil.ItsEqual(t, 1, len(p2.n.children))
	testutil.ItsEqual(t, true, hasKey(p2.n.children, c.n.id))
	testutil.ItsEqual(t, 0, len(p2.n.parents))
}

func Test_Node_String(t *testing.T) {
	n := newMockBareNode()
	n.n.height = 2

	testutil.ItsEqual(t, "test["+n.n.id.Short()+"]@2", n.Node().String("test"))

	n.Node().SetLabel("test_label")
	testutil.ItsEqual(t, "test["+n.n.id.Short()+"]:test_label@2", n.Node().String("test"))
}

func Test_SetStale(t *testing.T) {
	g := New()
	n := newMockBareNode()
	_ = Observe(Root(), g, n)
	g.SetStale(n)

	testutil.ItsEqual(t, 0, n.n.changedAt)
	testutil.ItsEqual(t, 0, n.n.recomputedAt)
	testutil.ItsEqual(t, 1, n.n.setAt)

	testutil.ItsEqual(t, true, n.n.graph.recomputeHeap.has(n))

	// find the node in the recompute heap layer
	testutil.ItsEqual(t, 1, len(n.n.graph.recomputeHeap.heights[1]))

	g.SetStale(n)

	testutil.ItsEqual(t, 0, n.n.changedAt)
	testutil.ItsEqual(t, 0, n.n.recomputedAt)
	testutil.ItsEqual(t, 1, n.n.setAt)

	testutil.ItsEqual(t, true, n.n.graph.recomputeHeap.has(n))

	testutil.ItsEqual(t, 1, len(n.n.graph.recomputeHeap.heights[1]))
}

func Test_Node_OnUpdate(t *testing.T) {
	n := NewNode()

	testutil.ItsEqual(t, 0, len(n.onUpdateHandlers))
	n.OnUpdate(func(_ context.Context) {})
	testutil.ItsEqual(t, 1, len(n.onUpdateHandlers))
}

func Test_Node_OnError(t *testing.T) {
	n := NewNode()

	testutil.ItsEqual(t, 0, len(n.onErrorHandlers))
	n.OnError(func(_ context.Context, _ error) {})
	testutil.ItsEqual(t, 1, len(n.onErrorHandlers))
}

func Test_Node_SetLabel(t *testing.T) {
	n := NewNode()

	testutil.ItsEqual(t, "", n.label)
	n.SetLabel("test-label")
	testutil.ItsEqual(t, "test-label", n.label)
}

func Test_Node_addChildren(t *testing.T) {
	n := newMockBareNode()
	_ = n.Node()

	c0 := newMockBareNode()
	_ = c0.Node()

	c1 := newMockBareNode()
	_ = c1.Node()

	testutil.ItsEqual(t, 0, len(n.n.parents))
	testutil.ItsEqual(t, 0, len(n.n.children))

	n.Node().addChildren(c0, c1)

	testutil.ItsEqual(t, 0, len(n.n.parents))
	testutil.ItsEqual(t, 2, len(n.n.children))

	testutil.ItsEqual(t, true, hasKey(n.n.children, c0.n.id))
	testutil.ItsEqual(t, true, hasKey(n.n.children, c1.n.id))
}

func Test_Node_removeChild(t *testing.T) {
	n := newMockBareNode()
	_ = n.Node()

	c0 := newMockBareNode()
	_ = c0.Node()

	c1 := newMockBareNode()
	_ = c1.Node()

	c2 := newMockBareNode()
	_ = c2.Node()

	n.Node().addChildren(c0, c1, c2)

	testutil.ItsEqual(t, 0, len(n.n.parents))
	testutil.ItsEqual(t, 3, len(n.n.children))

	testutil.ItsEqual(t, true, hasKey(n.n.children, c0.n.id))
	testutil.ItsEqual(t, true, hasKey(n.n.children, c1.n.id))
	testutil.ItsEqual(t, true, hasKey(n.n.children, c2.n.id))

	n.Node().removeChild(c1.n.id)

	testutil.ItsEqual(t, 0, len(n.n.parents))
	testutil.ItsEqual(t, 2, len(n.n.children))
	testutil.ItsEqual(t, true, hasKey(n.n.children, c0.n.id))
	testutil.ItsEqual(t, false, hasKey(n.n.children, c1.n.id))
	testutil.ItsEqual(t, true, hasKey(n.n.children, c2.n.id))
}

func Test_Node_addParents(t *testing.T) {
	n := newMockBareNode()
	_ = n.Node()

	c0 := newMockBareNode()
	_ = c0.Node()

	c1 := newMockBareNode()
	_ = c1.Node()

	testutil.ItsEqual(t, 0, len(n.n.parents))
	testutil.ItsEqual(t, 0, len(n.n.children))

	n.Node().addParents(c0, c1)

	testutil.ItsEqual(t, 2, len(n.n.parents))
	testutil.ItsEqual(t, 0, len(n.n.children))

	testutil.ItsEqual(t, true, hasKey(n.n.parents, c0.n.id))
	testutil.ItsEqual(t, true, hasKey(n.n.parents, c1.n.id))
}

func Test_Node_removeParent(t *testing.T) {
	n := newMockBareNode()
	_ = n.Node()

	c0 := newMockBareNode()
	_ = c0.Node()

	c1 := newMockBareNode()
	_ = c1.Node()

	c2 := newMockBareNode()
	_ = c2.Node()

	n.Node().addParents(c0, c1, c2)

	testutil.ItsEqual(t, 3, len(n.n.parents))
	testutil.ItsEqual(t, 0, len(n.n.children))

	testutil.ItsEqual(t, true, hasKey(n.n.parents, c0.n.id))
	testutil.ItsEqual(t, true, hasKey(n.n.parents, c1.n.id))
	testutil.ItsEqual(t, true, hasKey(n.n.parents, c2.n.id))

	n.Node().removeParent(c1.n.id)

	testutil.ItsEqual(t, 2, len(n.n.parents))
	testutil.ItsEqual(t, 0, len(n.n.children))

	testutil.ItsEqual(t, true, hasKey(n.n.parents, c0.n.id))
	testutil.ItsEqual(t, false, hasKey(n.n.parents, c1.n.id))
	testutil.ItsEqual(t, true, hasKey(n.n.parents, c2.n.id))
}

func Test_Node_maybeStabilize(t *testing.T) {
	ctx := testContext()
	n := NewNode()

	// assert it doesn't panic
	err := n.maybeStabilize(ctx)
	testutil.ItsNil(t, err)

	var calledStabilize bool
	n.stabilize = func(ictx context.Context) error {
		calledStabilize = true
		testutil.ItsBlueDye(ictx, t)
		return nil
	}

	err = n.maybeStabilize(ctx)
	testutil.ItsNil(t, err)
	testutil.ItsEqual(t, true, calledStabilize)
}

func Test_Node_maybeStabilize_error(t *testing.T) {
	ctx := testContext()
	n := NewNode()

	n.stabilize = func(ictx context.Context) error {
		testutil.ItsBlueDye(ictx, t)
		return fmt.Errorf("just a test")
	}

	err := n.maybeStabilize(ctx)
	testutil.ItsNotNil(t, err)
	testutil.ItsEqual(t, "just a test", err.Error())
	testutil.ItsEqual(t, 0, n.numRecomputes)
	testutil.ItsEqual(t, 0, n.recomputedAt)
}

func Test_Node_maybeCutoff(t *testing.T) {
	ctx := testContext()
	n := NewNode()

	shouldCutoff, err := n.maybeCutoff(ctx)
	testutil.ItsNil(t, err)
	testutil.ItsEqual(t, false, shouldCutoff)

	n.cutoff = func(ictx context.Context) (bool, error) {
		testutil.ItsBlueDye(ictx, t)
		return true, nil
	}

	shouldCutoff, err = n.maybeCutoff(ctx)
	testutil.ItsNil(t, err)
	testutil.ItsEqual(t, true, shouldCutoff)

	n.cutoff = func(ictx context.Context) (bool, error) {
		testutil.ItsBlueDye(ictx, t)
		return true, fmt.Errorf("this is just a test")
	}
	shouldCutoff, err = n.maybeCutoff(ctx)
	testutil.ItsNotNil(t, err)
	testutil.ItsEqual(t, true, shouldCutoff)
}

func Test_Node_detectCutoff(t *testing.T) {
	yes := NewNode()
	yes.detectCutoff(new(cutoffIncr[string]))
	testutil.ItsNotNil(t, yes.cutoff)

	no := NewNode()
	no.detectCutoff(new(mockBareNode))
	testutil.ItsNil(t, no.cutoff)
}

func Test_Node_detectStabilize(t *testing.T) {
	yes := NewNode()
	yes.detectStabilize(new(mapIncr[string, string]))
	testutil.ItsNotNil(t, yes.stabilize)

	no := NewNode()
	no.detectStabilize(new(mockBareNode))
	testutil.ItsNil(t, no.stabilize)
}

func Test_Node_shouldRecompute(t *testing.T) {
	n := NewNode()
	testutil.ItsEqual(t, true, n.ShouldRecompute())

	n.recomputedAt = 1
	testutil.ItsEqual(t, false, n.ShouldRecompute())

	n.stabilize = func(_ context.Context) error { return nil }
	n.setAt = 2
	testutil.ItsEqual(t, true, n.ShouldRecompute())

	n.setAt = 1
	n.changedAt = 2
	testutil.ItsEqual(t, true, n.ShouldRecompute())

	n.changedAt = 1
	c1 := newMockBareNode()
	c1.Node().changedAt = 2

	n.addParents(newMockBareNode(), c1)
	testutil.ItsEqual(t, true, n.ShouldRecompute())

	c1.Node().changedAt = 1
	testutil.ItsEqual(t, false, n.ShouldRecompute())
}

func Test_Node_recompute(t *testing.T) {
	ctx := testContext()

	g := New()
	var calledStabilize bool
	m0 := MapContext(Root(), Return(Root(), ""), func(ictx context.Context, _ string) (string, error) {
		calledStabilize = true
		testutil.ItsBlueDye(ictx, t)
		return "hello", nil
	})

	p := newMockBareNode()
	m0.Node().addParents(p)
	_ = Observe(Root(), g, m0)

	var calledUpdateHandler0, calledUpdateHandler1 bool
	m0.Node().OnUpdate(func(ictx context.Context) {
		testutil.ItsBlueDye(ictx, t)
		calledUpdateHandler0 = true
	})
	m0.Node().OnUpdate(func(ictx context.Context) {
		testutil.ItsBlueDye(ictx, t)
		calledUpdateHandler1 = true
	})

	var calledErrorHandler0, calledErrorHandler1 bool
	m0.Node().OnError(func(ictx context.Context, err error) {
		testutil.ItsBlueDye(ictx, t)
		calledErrorHandler0 = true
	})
	m0.Node().OnError(func(ictx context.Context, err error) {
		testutil.ItsBlueDye(ictx, t)
		calledErrorHandler1 = true
	})

	err := g.recompute(ctx, m0)
	testutil.ItsNil(t, err)

	// find the node in the recompute heap layer
	testutil.ItsEqual(t, true, g.recomputeHeap.has(p))
	testutil.ItsEqual(t, true, calledStabilize)

	// we don't call these handlers directly
	testutil.ItsEqual(t, false, calledUpdateHandler0)
	testutil.ItsEqual(t, false, calledUpdateHandler1)
	// we don't call these handlers at all b/c no error
	testutil.ItsEqual(t, false, calledErrorHandler0)
	testutil.ItsEqual(t, false, calledErrorHandler1)

	testutil.ItsEqual(t, 1, len(g.handleAfterStabilization))

	for _, handlers := range g.handleAfterStabilization {
		for _, h := range handlers {
			h(ctx)
		}
	}
	testutil.ItsEqual(t, true, calledUpdateHandler0)
	testutil.ItsEqual(t, true, calledUpdateHandler1)
}

func Test_Node_stabilize_error(t *testing.T) {
	ctx := testContext()

	g := New()
	var calledStabilize bool
	m0 := MapContext(Root(), Return(Root(), ""), func(ictx context.Context, _ string) (string, error) {
		calledStabilize = true
		testutil.ItsBlueDye(ictx, t)
		return "", fmt.Errorf("test error")
	})

	p := newMockBareNode()
	m0.Node().addParents(p)
	_ = Observe(Root(), g, m0)

	var calledUpdateHandler0, calledUpdateHandler1 bool
	m0.Node().OnUpdate(func(ictx context.Context) {
		testutil.ItsBlueDye(ictx, t)
		calledUpdateHandler0 = true
	})
	m0.Node().OnUpdate(func(ictx context.Context) {
		testutil.ItsBlueDye(ictx, t)
		calledUpdateHandler1 = true
	})

	var calledErrorHandler0, calledErrorHandler1 bool
	m0.Node().OnError(func(ictx context.Context, err error) {
		testutil.ItsBlueDye(ictx, t)
		calledErrorHandler0 = true
		testutil.ItsNotNil(t, err)
		testutil.ItsEqual(t, "test error", err.Error())
	})
	m0.Node().OnError(func(ictx context.Context, err error) {
		testutil.ItsBlueDye(ictx, t)
		calledErrorHandler1 = true
		testutil.ItsNotNil(t, err)
		testutil.ItsEqual(t, "test error", err.Error())
	})

	err := g.Stabilize(ctx)
	testutil.ItsNotNil(t, err)
	testutil.ItsNotNil(t, "test error", err.Error())

	// find the node in the recompute heap layer
	testutil.ItsEqual(t, false, g.recomputeHeap.has(p))
	testutil.ItsEqual(t, true, calledStabilize)
	testutil.ItsEqual(t, false, calledUpdateHandler0)
	testutil.ItsEqual(t, false, calledUpdateHandler1)
	testutil.ItsEqual(t, true, calledErrorHandler0)
	testutil.ItsEqual(t, true, calledErrorHandler1)
}

func Test_nodeFormatters(t *testing.T) {
	id := NewIdentifier()
	g := New()

	testCases := [...]struct {
		Node  INode
		Label string
	}{
		{Bind[string, string](Root(), Return(Root(), ""), nil), "bind"},
		{Cutoff(Root(), Return(Root(), ""), nil), "cutoff"},
		{Cutoff2(Root(), Return(Root(), ""), Return(Root(), ""), nil), "cutoff2"},
		{Func[string](Root(), nil), "func"},
		{MapN[string, bool](Root(), nil), "map_n"},
		{Map[string, bool](Root(), Return(Root(), ""), nil), "map"},
		{Map2[string, int, bool](Root(), Return(Root(), ""), Return(Root(), 0), nil), "map2"},
		{Map3[string, int, float64, bool](Root(), Return(Root(), ""), Return(Root(), 0), Return(Root(), 1.0), nil), "map3"},
		{MapIf(Root(), Return(Root(), ""), Return(Root(), ""), Return(Root(), false)), "map_if"},
		{Return(Root(), ""), "return"},
		{Watch(Root(), Return(Root(), "")), "watch"},
		{Freeze(Root(), Return(Root(), "")), "freeze"},
		{Var(Root(), ""), "var"},
		{FoldLeft(Root(), Return(Root(), []string{}), "", nil), "fold_left"},
		{FoldRight(Root(), Return(Root(), []string{}), "", nil), "fold_right"},
		{FoldMap(Root(), Return(Root(), map[string]int{}), "", nil), "fold_map"},
		{Observe(Root(), g, Return(Root(), "")), "observer"},
		{Always(Root(), Return(Root(), "")), "always"},
	}

	for _, tc := range testCases {
		tc.Node.Node().id = id
		tc.Node.Node().height = 2
		testutil.ItsEqual(t, fmt.Sprintf("%s[%s]@2", tc.Label, id.Short()), fmt.Sprint(tc.Node))
	}
}

func Test_Node_Properties_readonly(t *testing.T) {
	n := &Node{
		height:    1,
		setAt:     2,
		changedAt: 3,
		children: []INode{
			newMockBareNode(),
			newMockBareNode(),
		},
		parents: []INode{
			newMockBareNode(),
			newMockBareNode(),
			newMockBareNode(),
		},
	}

	testutil.ItsEqual(t, 2, len(n.Children()))
	testutil.ItsEqual(t, 3, len(n.Parents()))
}

type emptyNode struct {
	n *Node
}

func (en emptyNode) Node() *Node {
	return en.n
}

func Test_Node_ShouldRecompute_unit(t *testing.T) {
	var noop = func(_ context.Context) error { return nil }
	testutil.ItsEqual(t, true, (&Node{recomputedAt: 0}).ShouldRecompute())
	testutil.ItsEqual(t, false, (&Node{recomputedAt: 1}).ShouldRecompute())
	testutil.ItsEqual(t, true, (&Node{recomputedAt: 1, stabilize: noop, setAt: 2}).ShouldRecompute())
	testutil.ItsEqual(t, true, (&Node{recomputedAt: 2, stabilize: noop, setAt: 2, boundAt: 3}).ShouldRecompute())
	testutil.ItsEqual(t, true, (&Node{recomputedAt: 2, stabilize: noop, setAt: 2, boundAt: 2, changedAt: 3}).ShouldRecompute())
	testutil.ItsEqual(t, true, (&Node{recomputedAt: 2, stabilize: noop, setAt: 2, boundAt: 2, changedAt: 2, parents: []INode{emptyNode{&Node{changedAt: 3}}}}).ShouldRecompute())
	testutil.ItsEqual(t, false, (&Node{recomputedAt: 2, stabilize: noop, setAt: 2, boundAt: 2, changedAt: 2, parents: []INode{emptyNode{&Node{changedAt: 2}}}}).ShouldRecompute())
}

func Test_Node_Observers(t *testing.T) {
	one := &observeIncr[any]{
		n: NewNode(),
	}
	two := &observeIncr[any]{
		n: NewNode(),
	}
	n := &Node{
		observers: []IObserver{one, two},
	}

	testutil.ItsEqual(t, 2, len(n.Observers()))
}

func Test_nodeSorter(t *testing.T) {
	a := newMockBareNode()
	a.Node().height = 1
	a.Node().id, _ = ParseIdentifier(strings.Repeat("0", 32))

	b := newMockBareNode()
	b.Node().height = 1
	b.Node().id, _ = ParseIdentifier(strings.Repeat("1", 32))

	c := newMockBareNode()
	c.Node().height = 1
	c.Node().id, _ = ParseIdentifier(strings.Repeat("2", 32))

	d := newMockBareNode()
	d.Node().height = 2
	d.Node().id, _ = ParseIdentifier(strings.Repeat("3", 32))

	testutil.ItsEqual(t, true, "00000000000000000000000000000000" < "11111111111111111111111111111111")
	testutil.ItsEqual(t, 0, nodeSorter(a, a))
	testutil.ItsEqual(t, 1, nodeSorter(a, b))
	testutil.ItsEqual(t, -1, nodeSorter(c, b))
	testutil.ItsEqual(t, -1, nodeSorter(d, c))
	testutil.ItsEqual(t, 1, nodeSorter(a, d))
}

func Test_Node_onUpdate_regression(t *testing.T) {
	ctx := testContext()

	width := Var(Root(), 3)
	length := Var(Root(), 4)

	area := Map2(Root(), width, length, func(w int, l int) int {
		return w * l
	})
	area.Node().SetLabel("area")

	height := Var(Root(), 2)
	height.Node().SetLabel("height")
	volume := Map2(Root(), area, height, func(a int, h int) int {
		return a * h
	})
	volume.Node().SetLabel("volume")
	scaledVolume := Map(Root(), volume, func(v int) int {
		return v * 2
	})
	scaledVolume.Node().SetLabel("scaledVolume")

	g := New()
	areaObs := Observe(Root(), g, area)
	areaObs.Node().SetLabel("areaObs")
	scaledVolumeObs := Observe(Root(), g, scaledVolume)
	scaledVolumeObs.Node().SetLabel("scaledVolumeObs")

	err := g.Stabilize(ctx)
	testutil.ItsNil(t, err)

	_ = dumpDot(g, homedir("node_onupdate_regression_00.png"))

	testutil.ItsEqual(t, 12, areaObs.Value())
	testutil.ItsEqual(t, 48, scaledVolumeObs.Value())

	areaTriggerCount := 0
	area.Node().OnUpdate(func(ctx context.Context) {
		areaTriggerCount++
	})

	scaledVolumeTriggerCount := 0
	scaledVolumeObs.Node().OnUpdate(func(ctx context.Context) {
		scaledVolumeTriggerCount++
	})

	length.Set(5)

	err = g.Stabilize(ctx)
	testutil.ItsNil(t, err)

	_ = dumpDot(g, homedir("node_onupdate_regression_01.png"))

	testutil.ItsEqual(t, true, areaTriggerCount > 0)
	testutil.ItsEqual(t, true, scaledVolumeTriggerCount > 0) // should be triggered because area changed so volume change therefore scaled_volume changed too

	testutil.ItsEqual(t, 15, areaObs.Value())
	testutil.ItsEqual(t, 60, scaledVolumeObs.Value())
}
