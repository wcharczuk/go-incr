package incr

import (
	"context"
	"fmt"
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_NewNode(t *testing.T) {
	n := NewNode()
	testutil.ItsNotNil(t, n.id)
	testutil.ItsNil(t, n.graph)
	testutil.ItsNil(t, n.parents)
	testutil.ItsNil(t, n.children)
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
	p := newMockBareNode()
	c0 := newMockBareNode()
	c1 := newMockBareNode()
	c2 := newMockBareNode()

	// set up P with (3) inputs
	Link(p, c0, c1, c2)

	// no nodes depend on p, p is not an input to any nodes
	testutil.ItsEqual(t, 0, len(p.n.parents))
	testutil.ItsEqual(t, 3, len(p.n.children))
	testutil.ItsEqual(t, c0.n.id, p.n.children[0].Node().id)
	testutil.ItsEqual(t, c1.n.id, p.n.children[1].Node().id)
	testutil.ItsEqual(t, c2.n.id, p.n.children[2].Node().id)

	testutil.ItsEqual(t, 1, len(c0.n.parents))
	testutil.ItsEqual(t, p.n.id, c0.n.parents[0].Node().id)
	testutil.ItsEqual(t, 0, len(c0.n.children))

	testutil.ItsEqual(t, 1, len(c1.n.parents))
	testutil.ItsEqual(t, p.n.id, c1.n.parents[0].Node().id)
	testutil.ItsEqual(t, 0, len(c1.n.children))

	testutil.ItsEqual(t, 1, len(c2.n.parents))
	testutil.ItsEqual(t, p.n.id, c2.n.parents[0].Node().id)
	testutil.ItsEqual(t, 0, len(c2.n.children))
}

func Test_Node_String(t *testing.T) {
	n := newMockBareNode()

	testutil.ItsEqual(t, "test["+n.n.id.Short()+"]", n.Node().String("test"))

	n.Node().SetLabel("test_label")
	testutil.ItsEqual(t, "test["+n.n.id.Short()+"]:test_label", n.Node().String("test"))
}

func Test_SetStale(t *testing.T) {
	g := New()
	n := newMockBareNode()
	g.Observe(n)
	g.SetStale(n)

	testutil.ItsEqual(t, 0, n.n.changedAt)
	testutil.ItsEqual(t, 0, n.n.recomputedAt)
	testutil.ItsEqual(t, 1, n.n.setAt)

	testutil.ItsEqual(t, true, n.n.graph.recomputeHeap.Has(n))

	// find the node in the recompute heap layer
	testutil.ItsEqual(t, 1, n.n.graph.recomputeHeap.heights[1].Len())
	testutil.ItsEqual(t, n.n.id, n.n.graph.recomputeHeap.heights[1].head.key)

	g.SetStale(n)

	testutil.ItsEqual(t, 0, n.n.changedAt)
	testutil.ItsEqual(t, 0, n.n.recomputedAt)
	testutil.ItsEqual(t, 1, n.n.setAt)

	testutil.ItsEqual(t, true, n.n.graph.recomputeHeap.Has(n))

	testutil.ItsEqual(t, 1, n.n.graph.recomputeHeap.heights[1].Len())
	testutil.ItsEqual(t, n.n.id, n.n.graph.recomputeHeap.heights[1].head.key)
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
	testutil.ItsEqual(t, c0.n.id, n.n.children[0].Node().id)
	testutil.ItsEqual(t, c1.n.id, n.n.children[1].Node().id)
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
	testutil.ItsEqual(t, c0.n.id, n.n.children[0].Node().id)
	testutil.ItsEqual(t, c1.n.id, n.n.children[1].Node().id)
	testutil.ItsEqual(t, c2.n.id, n.n.children[2].Node().id)

	n.Node().removeChild(c1.n.id)

	testutil.ItsEqual(t, 0, len(n.n.parents))
	testutil.ItsEqual(t, 2, len(n.n.children))
	testutil.ItsEqual(t, c0.n.id, n.n.children[0].Node().id)
	testutil.ItsEqual(t, c2.n.id, n.n.children[1].Node().id)
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
	testutil.ItsEqual(t, c0.n.id, n.n.parents[0].Node().id)
	testutil.ItsEqual(t, c1.n.id, n.n.parents[1].Node().id)
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
	testutil.ItsEqual(t, c0.n.id, n.n.parents[0].Node().id)
	testutil.ItsEqual(t, c1.n.id, n.n.parents[1].Node().id)
	testutil.ItsEqual(t, c2.n.id, n.n.parents[2].Node().id)

	n.Node().removeParent(c1.n.id)

	testutil.ItsEqual(t, 2, len(n.n.parents))
	testutil.ItsEqual(t, 0, len(n.n.children))
	testutil.ItsEqual(t, c0.n.id, n.n.parents[0].Node().id)
	testutil.ItsEqual(t, c2.n.id, n.n.parents[1].Node().id)
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

	testutil.ItsEqual(t, false, n.maybeCutoff(ctx))

	n.cutoff = func(ictx context.Context) bool {
		testutil.ItsBlueDye(ictx, t)
		return true
	}

	testutil.ItsEqual(t, true, n.maybeCutoff(ctx))
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
	testutil.ItsEqual(t, true, n.shouldRecompute())

	n.recomputedAt = 1
	testutil.ItsEqual(t, false, n.shouldRecompute())

	n.stabilize = func(_ context.Context) error { return nil }
	n.setAt = 2
	testutil.ItsEqual(t, true, n.shouldRecompute())

	n.setAt = 1
	n.changedAt = 2
	testutil.ItsEqual(t, true, n.shouldRecompute())

	n.changedAt = 1
	c1 := newMockBareNode()
	c1.Node().changedAt = 2
	n.children = append(n.children, newMockBareNode(), c1)
	testutil.ItsEqual(t, true, n.shouldRecompute())

	c1.Node().changedAt = 1
	testutil.ItsEqual(t, false, n.shouldRecompute())
}

func Test_Node_computePseudoHeight(t *testing.T) {
	c010 := newMockBareNode()
	c10 := newMockBareNode()
	c00 := newMockBareNode()
	c01 := newMockBareNode()
	c0 := newMockBareNode()
	c1 := newMockBareNode()
	c2 := newMockBareNode()
	p := newMockBareNode()

	Link(c01, c010)
	Link(c0, c00, c01)
	Link(c1, c10)
	Link(p, c0, c1, c2)

	testutil.ItsEqual(t, 4, p.n.computePseudoHeight())
	testutil.ItsEqual(t, 3, c0.n.computePseudoHeight())
	testutil.ItsEqual(t, 2, c1.n.computePseudoHeight())
	testutil.ItsEqual(t, 1, c2.n.computePseudoHeight())
}

func Test_Node_recompute(t *testing.T) {
	ctx := testContext()

	g := New()
	var calledStabilize bool
	m0 := MapContext(Return(""), func(ictx context.Context, _ string) (string, error) {
		calledStabilize = true
		testutil.ItsBlueDye(ictx, t)
		return "hello", nil
	})

	p := newMockBareNode()
	m0.Node().addParents(p)
	g.Observe(m0)

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
	testutil.ItsEqual(t, true, g.recomputeHeap.Has(p))
	testutil.ItsEqual(t, true, calledStabilize)

	// we don't call these handlers directly
	testutil.ItsEqual(t, false, calledUpdateHandler0)
	testutil.ItsEqual(t, false, calledUpdateHandler1)
	// we don't call these handlers at all b/c no error
	testutil.ItsEqual(t, false, calledErrorHandler0)
	testutil.ItsEqual(t, false, calledErrorHandler1)

	testutil.ItsEqual(t, 1, g.handleAfterStabilization.Len())
	testutil.ItsEqual(t, 2, len(g.handleAfterStabilization.head.value))

	var handlers []func(context.Context)
	for g.handleAfterStabilization.Len() > 0 {
		_, handlers, _ = g.handleAfterStabilization.Pop()
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
	m0 := MapContext(Return(""), func(ictx context.Context, _ string) (string, error) {
		calledStabilize = true
		testutil.ItsBlueDye(ictx, t)
		return "", fmt.Errorf("test error")
	})

	p := newMockBareNode()
	m0.Node().addParents(p)
	g.Observe(m0)

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
	testutil.ItsEqual(t, false, g.recomputeHeap.Has(p))
	testutil.ItsEqual(t, true, calledStabilize)
	testutil.ItsEqual(t, false, calledUpdateHandler0)
	testutil.ItsEqual(t, false, calledUpdateHandler1)
	testutil.ItsEqual(t, true, calledErrorHandler0)
	testutil.ItsEqual(t, true, calledErrorHandler1)
}

func Test_nodeFormatters(t *testing.T) {
	id := NewIdentifier()

	testCases := [...]struct {
		Node  INode
		Label string
	}{
		{Bind[string, string](Return(""), nil), "bind"},
		{BindIf[string](Return(false), nil), "bind_if"},
		{Cutoff(Return(""), nil), "cutoff"},
		{Func[string](nil), "func"},
		{MapN[string, bool](nil), "map_n"},
		{Map[string, bool](Return(""), nil), "map"},
		{Map2[string, int, bool](Return(""), Return(0), nil), "map2"},
		{Map3[string, int, float64, bool](Return(""), Return(0), Return(1.0), nil), "map3"},
		{MapIf(Return(""), Return(""), Return(false)), "map_if"},
		{Return(""), "return"},
		{Watch(Return("")), "watch"},
		{Freeze(Return("")), "freeze"},
		{Var(""), "var"},
		{FoldLeft(Return([]string{}), "", nil), "fold_left"},
		{FoldRight(Return([]string{}), "", nil), "fold_right"},
		{FoldMap(Return(map[string]int{}), "", nil), "fold_map"},
	}

	for _, tc := range testCases {
		tc.Node.Node().id = id
		testutil.ItsEqual(t, fmt.Sprintf("%s[%s]", tc.Label, id.Short()), fmt.Sprint(tc.Node))
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

	testutil.ItsEqual(t, 1, n.Height())
	testutil.ItsEqual(t, 2, len(n.Children()))
	testutil.ItsEqual(t, 3, len(n.Parents()))
}

type emptyNode struct {
	n *Node
}

func (en emptyNode) Node() *Node {
	return en.n
}

func Test_Node_recomputeParentHeightsOnBindChange(t *testing.T) {
	n0 := emptyNode{NewNode()}
	n1 := emptyNode{NewNode()}
	n2 := emptyNode{NewNode()}
	n3 := emptyNode{NewNode()}

	Link(n1, n0)
	Link(n2, n1)
	Link(n3, n2)

	n1.Node().recomputeParentHeightsOnBindChange()

	testutil.ItsEqual(t, 0, n0.n.height)
	testutil.ItsEqual(t, 2, n1.n.height)
	testutil.ItsEqual(t, 3, n2.n.height)
	testutil.ItsEqual(t, 4, n3.n.height)
}

func Test_Node_shouldRecompute_unit(t *testing.T) {
	var noop = func(_ context.Context) error { return nil }
	testutil.ItsEqual(t, true, (&Node{recomputedAt: 0}).shouldRecompute())
	testutil.ItsEqual(t, false, (&Node{recomputedAt: 1}).shouldRecompute())
	testutil.ItsEqual(t, true, (&Node{recomputedAt: 1, stabilize: noop, setAt: 2}).shouldRecompute())
	testutil.ItsEqual(t, true, (&Node{recomputedAt: 2, stabilize: noop, setAt: 2, boundAt: 3}).shouldRecompute())
	testutil.ItsEqual(t, true, (&Node{recomputedAt: 2, stabilize: noop, setAt: 2, boundAt: 2, changedAt: 3}).shouldRecompute())
	testutil.ItsEqual(t, true, (&Node{recomputedAt: 2, stabilize: noop, setAt: 2, boundAt: 2, changedAt: 2, children: []INode{
		emptyNode{&Node{changedAt: 3}},
	}}).shouldRecompute())
	testutil.ItsEqual(t, false, (&Node{recomputedAt: 2, stabilize: noop, setAt: 2, boundAt: 2, changedAt: 2, children: []INode{
		emptyNode{&Node{changedAt: 2}},
	}}).shouldRecompute())
}
