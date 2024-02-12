package incr

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_NewNode(t *testing.T) {
	n := NewNode("test_node")
	testutil.NotNil(t, n.id)
	testutil.Nil(t, n.graph)
	testutil.Equal(t, "test_node", n.kind)
	testutil.Equal(t, 0, len(n.parents))
	testutil.Equal(t, 0, len(n.children))
	testutil.Equal(t, "", n.label)
	testutil.Equal(t, 0, n.height)
	testutil.Equal(t, 0, n.changedAt)
	testutil.Equal(t, 0, n.setAt)
	testutil.Equal(t, 0, n.recomputedAt)
	testutil.Nil(t, n.onUpdateHandlers)
	testutil.Nil(t, n.onErrorHandlers)
	testutil.Nil(t, n.stabilize)
	testutil.Nil(t, n.cutoff)
	testutil.Equal(t, 0, n.numRecomputes)
	testutil.Nil(t, n.metadata)
}

func Test_Node_ID(t *testing.T) {
	n := NewNode("test_node")
	testutil.Equal(t, false, n.ID().IsZero())

	other := NewIdentifier()
	n.id = other
	testutil.Equal(t, other, n.ID())
}

func Test_Node_Label(t *testing.T) {
	n := NewNode("test_node")
	testutil.Equal(t, "", n.Label())
	n.SetLabel("foo")
	testutil.Equal(t, "foo", n.Label())
}

func Test_Node_Metadata(t *testing.T) {
	n := NewNode("test_node")
	testutil.Equal(t, nil, n.Metadata())
	n.SetMetadata("foo")
	testutil.Equal(t, "foo", n.Metadata())
}

func Test_Link(t *testing.T) {
	g := New()

	c := newMockBareNode(g)
	p0 := newMockBareNode(g)
	p1 := newMockBareNode(g)
	p2 := newMockBareNode(g)

	// set up P with (3) inputs
	Link(c, p0, p1, p2)

	// no nodes depend on p, p is not an input to any nodes
	testutil.Equal(t, 3, len(c.n.parents))
	testutil.Equal(t, 0, len(c.n.children))

	testutil.Equal(t, true, hasKey(c.n.parents, p0.n.id))
	testutil.Equal(t, true, hasKey(c.n.parents, p1.n.id))
	testutil.Equal(t, true, hasKey(c.n.parents, p2.n.id))

	testutil.Equal(t, 1, len(p0.n.children))
	testutil.Equal(t, true, hasKey(p0.n.children, c.n.id))
	testutil.Equal(t, 0, len(p0.n.parents))

	testutil.Equal(t, 1, len(p1.n.children))
	testutil.Equal(t, true, hasKey(p1.n.children, c.n.id))
	testutil.Equal(t, 0, len(p1.n.parents))

	testutil.Equal(t, 1, len(p2.n.children))
	testutil.Equal(t, true, hasKey(p2.n.children, c.n.id))
	testutil.Equal(t, 0, len(p2.n.parents))
}

func Test_Node_String(t *testing.T) {
	g := New()

	n := newMockBareNode(g)
	n.n.height = 2

	testutil.Equal(t, "bare_node["+n.n.id.Short()+"]@2", n.Node().String())

	n.Node().SetLabel("test_label")
	testutil.Equal(t, "bare_node["+n.n.id.Short()+"]:test_label@2", n.Node().String())
}

func Test_SetStale(t *testing.T) {
	g := New()
	n := newMockBareNode(g)
	_ = Observe(g, n)
	g.SetStale(n)

	testutil.Equal(t, 0, n.n.changedAt)
	testutil.Equal(t, 0, n.n.recomputedAt)
	testutil.Equal(t, 1, n.n.setAt)

	testutil.Equal(t, true, n.n.graph.recomputeHeap.has(n))

	// find the node in the recompute heap layer
	testutil.Equal(t, 1, len(n.n.graph.recomputeHeap.heights[1]))

	g.SetStale(n)

	testutil.Equal(t, 0, n.n.changedAt)
	testutil.Equal(t, 0, n.n.recomputedAt)
	testutil.Equal(t, 1, n.n.setAt)

	testutil.Equal(t, true, n.n.graph.recomputeHeap.has(n))

	testutil.Equal(t, 1, len(n.n.graph.recomputeHeap.heights[1]))
}

func Test_Node_OnUpdate(t *testing.T) {
	n := NewNode("test_node")

	testutil.Equal(t, 0, len(n.onUpdateHandlers))
	n.OnUpdate(func(_ context.Context) {})
	testutil.Equal(t, 1, len(n.onUpdateHandlers))
}

func Test_Node_OnError(t *testing.T) {
	n := NewNode("test_node")

	testutil.Equal(t, 0, len(n.onErrorHandlers))
	n.OnError(func(_ context.Context, _ error) {})
	testutil.Equal(t, 1, len(n.onErrorHandlers))
}

func Test_Node_SetLabel(t *testing.T) {
	n := NewNode("test_node")

	testutil.Equal(t, "", n.label)
	n.SetLabel("test-label")
	testutil.Equal(t, "test-label", n.label)
}

func Test_Node_addChildren(t *testing.T) {
	g := New()

	n := newMockBareNode(g)
	_ = n.Node()

	c0 := newMockBareNode(g)
	_ = c0.Node()

	c1 := newMockBareNode(g)
	_ = c1.Node()

	testutil.Equal(t, 0, len(n.n.parents))
	testutil.Equal(t, 0, len(n.n.children))

	n.Node().addChildren(c0, c1)

	testutil.Equal(t, 0, len(n.n.parents))
	testutil.Equal(t, 2, len(n.n.children))

	testutil.Equal(t, true, hasKey(n.n.children, c0.n.id))
	testutil.Equal(t, true, hasKey(n.n.children, c1.n.id))
}

func Test_Node_removeChild(t *testing.T) {
	g := New()

	n := newMockBareNode(g)
	_ = n.Node()

	c0 := newMockBareNode(g)
	_ = c0.Node()

	c1 := newMockBareNode(g)
	_ = c1.Node()

	c2 := newMockBareNode(g)
	_ = c2.Node()

	n.Node().addChildren(c0, c1, c2)

	testutil.Equal(t, 0, len(n.n.parents))
	testutil.Equal(t, 3, len(n.n.children))

	testutil.Equal(t, true, hasKey(n.n.children, c0.n.id))
	testutil.Equal(t, true, hasKey(n.n.children, c1.n.id))
	testutil.Equal(t, true, hasKey(n.n.children, c2.n.id))

	n.Node().removeChild(c1.n.id)

	testutil.Equal(t, 0, len(n.n.parents))
	testutil.Equal(t, 2, len(n.n.children))
	testutil.Equal(t, true, hasKey(n.n.children, c0.n.id))
	testutil.Equal(t, false, hasKey(n.n.children, c1.n.id))
	testutil.Equal(t, true, hasKey(n.n.children, c2.n.id))
}

func Test_Node_addParents(t *testing.T) {
	g := New()

	n := newMockBareNode(g)
	_ = n.Node()

	c0 := newMockBareNode(g)
	_ = c0.Node()

	c1 := newMockBareNode(g)
	_ = c1.Node()

	testutil.Equal(t, 0, len(n.n.parents))
	testutil.Equal(t, 0, len(n.n.children))

	n.Node().addParents(c0, c1)

	testutil.Equal(t, 2, len(n.n.parents))
	testutil.Equal(t, 0, len(n.n.children))

	testutil.Equal(t, true, hasKey(n.n.parents, c0.n.id))
	testutil.Equal(t, true, hasKey(n.n.parents, c1.n.id))
}

func Test_Node_removeParent(t *testing.T) {
	g := New()

	n := newMockBareNode(g)
	_ = n.Node()

	c0 := newMockBareNode(g)
	_ = c0.Node()

	c1 := newMockBareNode(g)
	_ = c1.Node()

	c2 := newMockBareNode(g)
	_ = c2.Node()

	n.Node().addParents(c0, c1, c2)

	testutil.Equal(t, 3, len(n.n.parents))
	testutil.Equal(t, 0, len(n.n.children))

	testutil.Equal(t, true, hasKey(n.n.parents, c0.n.id))
	testutil.Equal(t, true, hasKey(n.n.parents, c1.n.id))
	testutil.Equal(t, true, hasKey(n.n.parents, c2.n.id))

	n.Node().removeParent(c1.n.id)

	testutil.Equal(t, 2, len(n.n.parents))
	testutil.Equal(t, 0, len(n.n.children))

	testutil.Equal(t, true, hasKey(n.n.parents, c0.n.id))
	testutil.Equal(t, false, hasKey(n.n.parents, c1.n.id))
	testutil.Equal(t, true, hasKey(n.n.parents, c2.n.id))
}

func Test_Node_maybeStabilize(t *testing.T) {
	ctx := testContext()
	n := NewNode("test_node")

	// assert it doesn't panic
	err := n.maybeStabilize(ctx)
	testutil.Nil(t, err)

	var calledStabilize bool
	n.stabilize = func(ictx context.Context) error {
		calledStabilize = true
		testutil.BlueDye(ictx, t)
		return nil
	}

	err = n.maybeStabilize(ctx)
	testutil.Nil(t, err)
	testutil.Equal(t, true, calledStabilize)
}

func Test_Node_maybeStabilize_error(t *testing.T) {
	ctx := testContext()
	n := NewNode("test_node")

	n.stabilize = func(ictx context.Context) error {
		testutil.BlueDye(ictx, t)
		return fmt.Errorf("just a test")
	}

	err := n.maybeStabilize(ctx)
	testutil.NotNil(t, err)
	testutil.Equal(t, "just a test", err.Error())
	testutil.Equal(t, 0, n.numRecomputes)
	testutil.Equal(t, 0, n.recomputedAt)
}

func Test_Node_maybeCutoff(t *testing.T) {
	ctx := testContext()
	n := NewNode("test_node")

	shouldCutoff, err := n.maybeCutoff(ctx)
	testutil.Nil(t, err)
	testutil.Equal(t, false, shouldCutoff)

	n.cutoff = func(ictx context.Context) (bool, error) {
		testutil.BlueDye(ictx, t)
		return true, nil
	}

	shouldCutoff, err = n.maybeCutoff(ctx)
	testutil.Nil(t, err)
	testutil.Equal(t, true, shouldCutoff)

	n.cutoff = func(ictx context.Context) (bool, error) {
		testutil.BlueDye(ictx, t)
		return true, fmt.Errorf("this is just a test")
	}
	shouldCutoff, err = n.maybeCutoff(ctx)
	testutil.NotNil(t, err)
	testutil.Equal(t, true, shouldCutoff)
}

func Test_Node_detectCutoff(t *testing.T) {
	yes := NewNode("test_node")
	yes.detectCutoff(new(cutoffIncr[string]))
	testutil.NotNil(t, yes.cutoff)

	no := NewNode("test_node")
	no.detectCutoff(new(mockBareNode))
	testutil.Nil(t, no.cutoff)
}

func Test_Node_detectStabilize(t *testing.T) {
	yes := NewNode("test_node")
	yes.detectStabilize(new(mapIncr[string, string]))
	testutil.NotNil(t, yes.stabilize)

	no := NewNode("test_node")
	no.detectStabilize(new(mockBareNode))
	testutil.Nil(t, no.stabilize)
}

func Test_Node_shouldRecompute(t *testing.T) {
	g := New()

	n := NewNode("test_node")
	testutil.Equal(t, true, n.ShouldRecompute())

	n.recomputedAt = 1
	testutil.Equal(t, false, n.ShouldRecompute())

	n.stabilize = func(_ context.Context) error { return nil }
	n.setAt = 2
	testutil.Equal(t, true, n.ShouldRecompute())

	n.setAt = 1
	n.changedAt = 2
	testutil.Equal(t, true, n.ShouldRecompute())

	n.changedAt = 1
	c1 := newMockBareNode(g)
	c1.Node().changedAt = 2

	n.addParents(newMockBareNode(g), c1)
	testutil.Equal(t, true, n.ShouldRecompute())

	c1.Node().changedAt = 1
	testutil.Equal(t, false, n.ShouldRecompute())
}

func Test_Node_recompute(t *testing.T) {
	ctx := testContext()

	g := New()
	var calledStabilize bool
	m0 := MapContext(g, Return(g, ""), func(ictx context.Context, _ string) (string, error) {
		calledStabilize = true
		testutil.BlueDye(ictx, t)
		return "hello", nil
	})

	p := newMockBareNode(g)
	m0.Node().addParents(p)
	_ = Observe(g, m0)

	var calledUpdateHandler0, calledUpdateHandler1 bool
	m0.Node().OnUpdate(func(ictx context.Context) {
		testutil.BlueDye(ictx, t)
		calledUpdateHandler0 = true
	})
	m0.Node().OnUpdate(func(ictx context.Context) {
		testutil.BlueDye(ictx, t)
		calledUpdateHandler1 = true
	})

	var calledErrorHandler0, calledErrorHandler1 bool
	m0.Node().OnError(func(ictx context.Context, err error) {
		testutil.BlueDye(ictx, t)
		calledErrorHandler0 = true
	})
	m0.Node().OnError(func(ictx context.Context, err error) {
		testutil.BlueDye(ictx, t)
		calledErrorHandler1 = true
	})

	err := g.recompute(ctx, m0)
	testutil.Nil(t, err)

	// find the node in the recompute heap layer
	testutil.Equal(t, true, g.recomputeHeap.has(p))
	testutil.Equal(t, true, calledStabilize)

	// we don't call these handlers directly
	testutil.Equal(t, false, calledUpdateHandler0)
	testutil.Equal(t, false, calledUpdateHandler1)
	// we don't call these handlers at all b/c no error
	testutil.Equal(t, false, calledErrorHandler0)
	testutil.Equal(t, false, calledErrorHandler1)

	testutil.Equal(t, 1, len(g.handleAfterStabilization))

	for _, handlers := range g.handleAfterStabilization {
		for _, h := range handlers {
			h(ctx)
		}
	}
	testutil.Equal(t, true, calledUpdateHandler0)
	testutil.Equal(t, true, calledUpdateHandler1)
}

func Test_Node_stabilize_error(t *testing.T) {
	ctx := testContext()
	g := New()

	var calledStabilize bool
	m0 := MapContext(g, Return(g, ""), func(ictx context.Context, _ string) (string, error) {
		calledStabilize = true
		testutil.BlueDye(ictx, t)
		return "", fmt.Errorf("test error")
	})

	p := newMockBareNode(g)
	m0.Node().addParents(p)
	_ = Observe(g, m0)

	var calledUpdateHandler0, calledUpdateHandler1 bool
	m0.Node().OnUpdate(func(ictx context.Context) {
		testutil.BlueDye(ictx, t)
		calledUpdateHandler0 = true
	})
	m0.Node().OnUpdate(func(ictx context.Context) {
		testutil.BlueDye(ictx, t)
		calledUpdateHandler1 = true
	})

	var calledErrorHandler0, calledErrorHandler1 bool
	m0.Node().OnError(func(ictx context.Context, err error) {
		testutil.BlueDye(ictx, t)
		calledErrorHandler0 = true
		testutil.NotNil(t, err)
		testutil.Equal(t, "test error", err.Error())
	})
	m0.Node().OnError(func(ictx context.Context, err error) {
		testutil.BlueDye(ictx, t)
		calledErrorHandler1 = true
		testutil.NotNil(t, err)
		testutil.Equal(t, "test error", err.Error())
	})

	err := g.Stabilize(ctx)
	testutil.NotNil(t, err)
	testutil.NotNil(t, "test error", err.Error())

	// find the node in the recompute heap layer
	testutil.Equal(t, false, g.recomputeHeap.has(p))
	testutil.Equal(t, true, calledStabilize)
	testutil.Equal(t, false, calledUpdateHandler0)
	testutil.Equal(t, false, calledUpdateHandler1)
	testutil.Equal(t, true, calledErrorHandler0)
	testutil.Equal(t, true, calledErrorHandler1)
}

func Test_nodeFormatters(t *testing.T) {
	id := NewIdentifier()
	g := New()

	testCases := [...]struct {
		Node  INode
		Label string
	}{
		{Bind[string, string](g, Return(g, ""), nil), "bind"},
		{Cutoff(g, Return(g, ""), nil), "cutoff"},
		{Cutoff2(g, Return(g, ""), Return(g, ""), nil), "cutoff2"},
		{Func[string](g, nil), "func"},
		{MapN[string, bool](g, nil), "map_n"},
		{Map[string, bool](g, Return(g, ""), nil), "map"},
		{Map2[string, int, bool](g, Return(g, ""), Return(g, 0), nil), "map2"},
		{Map3[string, int, float64, bool](g, Return(g, ""), Return(g, 0), Return(g, 1.0), nil), "map3"},
		{MapIf(g, Return(g, ""), Return(g, ""), Return(g, false)), "map_if"},
		{Return(g, ""), "return"},
		{Watch(g, Return(g, "")), "watch"},
		{Freeze(g, Return(g, "")), "freeze"},
		{Var(g, ""), "var"},
		{FoldLeft(g, Return(g, []string{}), "", nil), "fold_left"},
		{FoldRight(g, Return(g, []string{}), "", nil), "fold_right"},
		{FoldMap(g, Return(g, map[string]int{}), "", nil), "fold_map"},
		{Observe(g, Return(g, "")), "observer"},
		{Always(g, Return(g, "")), "always"},
	}

	for _, tc := range testCases {
		tc.Node.Node().id = id
		tc.Node.Node().height = 2
		testutil.Equal(t, fmt.Sprintf("%s[%s]@2", tc.Label, id.Short()), fmt.Sprint(tc.Node))
	}
}

func Test_Node_Properties_readonly(t *testing.T) {
	g := New()

	n := &Node{
		height:    1,
		setAt:     2,
		changedAt: 3,
		children: []INode{
			newMockBareNode(g),
			newMockBareNode(g),
		},
		parents: []INode{
			newMockBareNode(g),
			newMockBareNode(g),
			newMockBareNode(g),
		},
	}

	testutil.Equal(t, 2, len(n.Children()))
	testutil.Equal(t, 3, len(n.Parents()))
}

type emptyNode struct {
	n *Node
}

func (en emptyNode) Node() *Node {
	return en.n
}

func Test_Node_ShouldRecompute_unit(t *testing.T) {
	var noop = func(_ context.Context) error { return nil }
	testutil.Equal(t, true, (&Node{recomputedAt: 0}).ShouldRecompute())
	testutil.Equal(t, false, (&Node{recomputedAt: 1}).ShouldRecompute())
	testutil.Equal(t, true, (&Node{recomputedAt: 1, stabilize: noop, setAt: 2}).ShouldRecompute())
	testutil.Equal(t, true, (&Node{recomputedAt: 2, stabilize: noop, setAt: 2, boundAt: 3}).ShouldRecompute())
	testutil.Equal(t, true, (&Node{recomputedAt: 2, stabilize: noop, setAt: 2, boundAt: 2, changedAt: 3}).ShouldRecompute())
	testutil.Equal(t, true, (&Node{recomputedAt: 2, stabilize: noop, setAt: 2, boundAt: 2, changedAt: 2, parents: []INode{emptyNode{&Node{changedAt: 3}}}}).ShouldRecompute())
	testutil.Equal(t, false, (&Node{recomputedAt: 2, stabilize: noop, setAt: 2, boundAt: 2, changedAt: 2, parents: []INode{emptyNode{&Node{changedAt: 2}}}}).ShouldRecompute())
}

func Test_Node_Observers(t *testing.T) {
	one := &observeIncr[any]{
		n: NewNode("test_node"),
	}
	two := &observeIncr[any]{
		n: NewNode("test_node"),
	}
	n := &Node{
		observers: []IObserver{one, two},
	}
	testutil.Equal(t, 2, len(n.observers))
}

func Test_nodeSorter(t *testing.T) {
	g := New()

	a := newMockBareNode(g)
	a.Node().height = 1
	a.Node().id, _ = ParseIdentifier(strings.Repeat("0", 32))

	b := newMockBareNode(g)
	b.Node().height = 1
	b.Node().id, _ = ParseIdentifier(strings.Repeat("1", 32))

	c := newMockBareNode(g)
	c.Node().height = 1
	c.Node().id, _ = ParseIdentifier(strings.Repeat("2", 32))

	d := newMockBareNode(g)
	d.Node().height = 2
	d.Node().id, _ = ParseIdentifier(strings.Repeat("3", 32))

	testutil.Equal(t, true, "00000000000000000000000000000000" < "11111111111111111111111111111111")
	testutil.Equal(t, 0, nodeSorter(a, a))
	testutil.Equal(t, 1, nodeSorter(a, b))
	testutil.Equal(t, -1, nodeSorter(c, b))
	testutil.Equal(t, -1, nodeSorter(d, c))
	testutil.Equal(t, 1, nodeSorter(a, d))
}

func Test_Node_onUpdate_regression(t *testing.T) {
	ctx := testContext()
	g := New()

	width := Var(g, 3)
	length := Var(g, 4)

	area := Map2(g, width, length, func(w int, l int) int {
		return w * l
	})
	area.Node().SetLabel("area")

	height := Var(g, 2)
	height.Node().SetLabel("height")
	volume := Map2(g, area, height, func(a int, h int) int {
		return a * h
	})
	volume.Node().SetLabel("volume")
	scaledVolume := Map(g, volume, func(v int) int {
		return v * 2
	})
	scaledVolume.Node().SetLabel("scaledVolume")

	areaObs := Observe(g, area)
	areaObs.Node().SetLabel("areaObs")
	scaledVolumeObs := Observe(g, scaledVolume)
	scaledVolumeObs.Node().SetLabel("scaledVolumeObs")

	err := g.Stabilize(ctx)
	testutil.Nil(t, err)

	_ = dumpDot(g, homedir("node_onupdate_regression_00.png"))

	testutil.Equal(t, 12, areaObs.Value())
	testutil.Equal(t, 48, scaledVolumeObs.Value())

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
	testutil.Nil(t, err)

	_ = dumpDot(g, homedir("node_onupdate_regression_01.png"))

	testutil.Equal(t, true, areaTriggerCount > 0)
	testutil.Equal(t, true, scaledVolumeTriggerCount > 0) // should be triggered because area changed so volume change therefore scaled_volume changed too

	testutil.Equal(t, 15, areaObs.Value())
	testutil.Equal(t, 60, scaledVolumeObs.Value())
}
