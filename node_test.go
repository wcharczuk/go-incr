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

	testutil.ItsEqual(t, true, c.n.HasParent(p0.n.id))
	testutil.ItsEqual(t, true, c.n.HasParent(p1.n.id))
	testutil.ItsEqual(t, true, c.n.HasParent(p2.n.id))

	testutil.ItsEqual(t, 1, len(p0.n.children))
	testutil.ItsEqual(t, true, p0.n.HasChild(c.n.id))
	testutil.ItsEqual(t, 0, len(p0.n.parents))

	testutil.ItsEqual(t, 1, len(p1.n.children))
	testutil.ItsEqual(t, true, p1.n.HasChild(c.n.id))
	testutil.ItsEqual(t, 0, len(p1.n.parents))

	testutil.ItsEqual(t, 1, len(p2.n.children))
	testutil.ItsEqual(t, true, p2.n.HasChild(c.n.id))
	testutil.ItsEqual(t, 0, len(p2.n.parents))
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
	_ = Observe(g, n)
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

	testutil.ItsEqual(t, true, n.n.HasChild(c0.n.id))
	testutil.ItsEqual(t, true, n.n.HasChild(c1.n.id))
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

	testutil.ItsEqual(t, true, n.n.HasChild(c0.n.id))
	testutil.ItsEqual(t, true, n.n.HasChild(c1.n.id))
	testutil.ItsEqual(t, true, n.n.HasChild(c2.n.id))

	n.Node().removeChild(c1.n.id)

	testutil.ItsEqual(t, 0, len(n.n.parents))
	testutil.ItsEqual(t, 2, len(n.n.children))
	testutil.ItsEqual(t, true, n.n.HasChild(c0.n.id))
	testutil.ItsEqual(t, false, n.n.HasChild(c1.n.id))
	testutil.ItsEqual(t, true, n.n.HasChild(c2.n.id))
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

	testutil.ItsEqual(t, true, n.n.HasParent(c0.n.id))
	testutil.ItsEqual(t, true, n.n.HasParent(c1.n.id))
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

	testutil.ItsEqual(t, true, n.n.HasParent(c0.n.id))
	testutil.ItsEqual(t, true, n.n.HasParent(c1.n.id))
	testutil.ItsEqual(t, true, n.n.HasParent(c2.n.id))

	n.Node().removeParent(c1.n.id)

	testutil.ItsEqual(t, 2, len(n.n.parents))
	testutil.ItsEqual(t, 0, len(n.n.children))

	testutil.ItsEqual(t, true, n.n.HasParent(c0.n.id))
	testutil.ItsEqual(t, false, n.n.HasParent(c1.n.id))
	testutil.ItsEqual(t, true, n.n.HasParent(c2.n.id))
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
	n.parents.Add(newMockBareNode(), c1)
	testutil.ItsEqual(t, true, n.ShouldRecompute())

	c1.Node().changedAt = 1
	testutil.ItsEqual(t, false, n.ShouldRecompute())
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
	_ = Observe(g, m0)

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
	_ = Observe(g, m0)

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
	g := New()

	testCases := [...]struct {
		Node  INode
		Label string
	}{
		{Bind[string, string](Return(""), nil), "bind"},
		{BindIf[string](Return(false), nil), "bind_if"},
		{Cutoff(Return(""), nil), "cutoff"},
		{Cutoff2(Return(""), Return(""), nil), "cutoff2"},
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
		{Observe[string](g, Return("")), "observer"},
		{Always[string](Return("")), "always"},
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
		children: newNodeLookup(
			newMockBareNode(),
			newMockBareNode(),
		),
		parents: newNodeLookup(
			newMockBareNode(),
			newMockBareNode(),
			newMockBareNode(),
		),
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

func Test_Node_recomputeHeights(t *testing.T) {
	n0 := emptyNode{NewNode()}
	n1 := emptyNode{NewNode()}
	n2 := emptyNode{NewNode()}
	n3 := emptyNode{NewNode()}

	Link(n1, n0)
	Link(n2, n1)
	Link(n3, n2)

	n1.Node().recomputeHeights()

	testutil.ItsEqual(t, 0, n0.n.height)
	testutil.ItsEqual(t, 2, n1.n.height)
	testutil.ItsEqual(t, 3, n2.n.height)
	testutil.ItsEqual(t, 4, n3.n.height)
}

func Test_Node_ShouldRecompute_unit(t *testing.T) {
	var noop = func(_ context.Context) error { return nil }
	testutil.ItsEqual(t, true, (&Node{recomputedAt: 0}).ShouldRecompute())
	testutil.ItsEqual(t, false, (&Node{recomputedAt: 1}).ShouldRecompute())
	testutil.ItsEqual(t, true, (&Node{recomputedAt: 1, stabilize: noop, setAt: 2}).ShouldRecompute())
	testutil.ItsEqual(t, true, (&Node{recomputedAt: 2, stabilize: noop, setAt: 2, boundAt: 3}).ShouldRecompute())
	testutil.ItsEqual(t, true, (&Node{recomputedAt: 2, stabilize: noop, setAt: 2, boundAt: 2, changedAt: 3}).ShouldRecompute())
	testutil.ItsEqual(t, true, (&Node{recomputedAt: 2, stabilize: noop, setAt: 2, boundAt: 2, changedAt: 2, parents: newNodeLookup(emptyNode{&Node{changedAt: 3}})}).ShouldRecompute())
	testutil.ItsEqual(t, false, (&Node{recomputedAt: 2, stabilize: noop, setAt: 2, boundAt: 2, changedAt: 2, parents: newNodeLookup(emptyNode{&Node{changedAt: 2}})}).ShouldRecompute())
}

func Test_Node_HasChild(t *testing.T) {
	c0 := newMockBareNode()
	c1 := newMockBareNode()
	c2 := newMockBareNode()
	n := &Node{
		children: newNodeLookup(c0, c1),
	}

	testutil.ItsEqual(t, true, n.HasChild(c0.Node().ID()))
	testutil.ItsEqual(t, true, n.HasChild(c1.Node().ID()))
	testutil.ItsEqual(t, false, n.HasChild(c2.Node().ID()))
}

func Test_Node_HasParent(t *testing.T) {
	p0 := newMockBareNode()
	p1 := newMockBareNode()
	p2 := newMockBareNode()
	n := &Node{
		parents: newNodeLookup(p0, p1),
	}

	testutil.ItsEqual(t, true, n.HasParent(p0.Node().ID()))
	testutil.ItsEqual(t, true, n.HasParent(p1.Node().ID()))
	testutil.ItsEqual(t, false, n.HasParent(p2.Node().ID()))
}

func Test_Node_IsRoot(t *testing.T) {
	p0 := newMockBareNode()
	p1 := newMockBareNode()
	n := &Node{
		parents: newNodeLookup(p0, p1),
	}

	testutil.ItsEqual(t, false, n.IsRoot())
	n1 := &Node{}
	testutil.ItsEqual(t, true, n1.IsRoot())
}

func Test_Node_IsLeaf(t *testing.T) {
	c0 := newMockBareNode()
	c1 := newMockBareNode()
	n := &Node{
		children: newNodeLookup(c0, c1),
	}

	testutil.ItsEqual(t, false, n.IsLeaf())
	n1 := &Node{}
	testutil.ItsEqual(t, true, n1.IsLeaf())
}

func Test_Node_Observers(t *testing.T) {
	one := &observeIncr[any]{
		n: NewNode(),
	}
	two := &observeIncr[any]{
		n: NewNode(),
	}
	n := &Node{
		observers: map[Identifier]IObserver{
			one.n.id: one,
			two.n.id: two,
		},
	}

	testutil.ItsEqual(t, 2, len(n.Observers()))
}

func Test_Node_HasObserver(t *testing.T) {
	one := &observeIncr[any]{
		n: NewNode(),
	}
	two := &observeIncr[any]{
		n: NewNode(),
	}
	three := &observeIncr[any]{
		n: NewNode(),
	}
	n := &Node{
		observers: map[Identifier]IObserver{
			one.n.id: one,
			two.n.id: two,
		},
	}

	testutil.ItsEqual(t, true, n.HasObserver(one.Node().ID()))
	testutil.ItsEqual(t, true, n.HasObserver(two.Node().ID()))
	testutil.ItsEqual(t, false, n.HasObserver(three.Node().ID()))
}
