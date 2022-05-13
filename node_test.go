package incr

import (
	"context"
	"fmt"
	"testing"
)

func Test_Node_NewNode(t *testing.T) {
	n := NewNode()
	ItsNotNil(t, n.id)
	ItsNil(t, n.gs)
	ItsNil(t, n.parents)
	ItsNil(t, n.children)
	ItsEqual(t, "", n.label)
	ItsEqual(t, 0, n.height)
	ItsEqual(t, 0, n.changedAt)
	ItsEqual(t, 0, n.setAt)
	ItsEqual(t, 0, n.recomputedAt)
	ItsNil(t, n.onUpdateHandlers)
	ItsNil(t, n.onErrorHandlers)
	ItsNil(t, n.stabilize)
	ItsNil(t, n.cutoff)
	ItsEqual(t, 0, n.numRecomputes)
}

func Test_Link(t *testing.T) {
	p := new(mockBareNode)
	c0 := new(mockBareNode)
	c1 := new(mockBareNode)
	c2 := new(mockBareNode)

	// set up P with (3) inputs
	Link(p, c0, c1, c2)

	// no nodes depend on p, p is not an input to any nodes
	ItsEqual(t, 0, len(p.n.parents))
	ItsEqual(t, 3, len(p.n.children))
	ItsEqual(t, c0.n.id, p.n.children[0].Node().id)
	ItsEqual(t, c1.n.id, p.n.children[1].Node().id)
	ItsEqual(t, c2.n.id, p.n.children[2].Node().id)

	ItsEqual(t, 1, len(c0.n.parents))
	ItsEqual(t, p.n.id, c0.n.parents[0].Node().id)
	ItsEqual(t, 0, len(c0.n.children))

	ItsEqual(t, 1, len(c1.n.parents))
	ItsEqual(t, p.n.id, c1.n.parents[0].Node().id)
	ItsEqual(t, 0, len(c1.n.children))

	ItsEqual(t, 1, len(c2.n.parents))
	ItsEqual(t, p.n.id, c2.n.parents[0].Node().id)
	ItsEqual(t, 0, len(c2.n.children))
}

func Test_FormatNode(t *testing.T) {
	n := new(mockBareNode)
	_ = n.Node()

	ItsEqual(t, "test["+n.n.id.Short()+"]", FormatNode(n.Node(), "test"))

	n.Node().SetLabel("test_label")
	ItsEqual(t, "test["+n.n.id.Short()+"]:test_label", FormatNode(n.Node(), "test"))
}

func Test_SetStale(t *testing.T) {
	n := new(mockBareNode)
	n.n = NewNode()
	n.n.gs = newGraphState()

	SetStale(n)

	ItsEqual(t, 0, n.n.changedAt)
	ItsEqual(t, 0, n.n.recomputedAt)
	ItsEqual(t, 1, n.n.setAt)

	ItsEqual(t, true, n.n.gs.rh.Has(n))

	// find the node in the recompute heap layer
	ItsEqual(t, 1, n.n.gs.rh.heights[0].len)
	ItsEqual(t, n.n.id, n.n.gs.rh.heights[0].head.key)

	SetStale(n)

	ItsEqual(t, 0, n.n.changedAt)
	ItsEqual(t, 0, n.n.recomputedAt)
	ItsEqual(t, 1, n.n.setAt)

	ItsEqual(t, true, n.n.gs.rh.Has(n))

	ItsEqual(t, 1, n.n.gs.rh.heights[0].len)
	ItsEqual(t, n.n.id, n.n.gs.rh.heights[0].head.key)
}

func Test_Node_OnUpdate(t *testing.T) {
	n := NewNode()

	ItsEqual(t, 0, len(n.onUpdateHandlers))
	n.OnUpdate(func(_ context.Context) {})
	ItsEqual(t, 1, len(n.onUpdateHandlers))
}

func Test_Node_OnError(t *testing.T) {
	n := NewNode()

	ItsEqual(t, 0, len(n.onErrorHandlers))
	n.OnError(func(_ context.Context, _ error) {})
	ItsEqual(t, 1, len(n.onErrorHandlers))
}

func Test_Node_SetLabel(t *testing.T) {
	n := NewNode()

	ItsEqual(t, "", n.label)
	n.SetLabel("test-label")
	ItsEqual(t, "test-label", n.label)
}

func Test_Node_addChildren(t *testing.T) {
	n := new(mockBareNode)
	_ = n.Node()

	c0 := new(mockBareNode)
	_ = c0.Node()

	c1 := new(mockBareNode)
	_ = c1.Node()

	ItsEqual(t, 0, len(n.n.parents))
	ItsEqual(t, 0, len(n.n.children))

	n.Node().addChildren(c0, c1)

	ItsEqual(t, 0, len(n.n.parents))
	ItsEqual(t, 2, len(n.n.children))
	ItsEqual(t, c0.n.id, n.n.children[0].Node().id)
	ItsEqual(t, c1.n.id, n.n.children[1].Node().id)
}

func Test_Node_removeChild(t *testing.T) {
	n := new(mockBareNode)
	_ = n.Node()

	c0 := new(mockBareNode)
	_ = c0.Node()

	c1 := new(mockBareNode)
	_ = c1.Node()

	c2 := new(mockBareNode)
	_ = c2.Node()

	n.Node().addChildren(c0, c1, c2)

	ItsEqual(t, 0, len(n.n.parents))
	ItsEqual(t, 3, len(n.n.children))
	ItsEqual(t, c0.n.id, n.n.children[0].Node().id)
	ItsEqual(t, c1.n.id, n.n.children[1].Node().id)
	ItsEqual(t, c2.n.id, n.n.children[2].Node().id)

	n.Node().removeChild(c1.n.id)

	ItsEqual(t, 0, len(n.n.parents))
	ItsEqual(t, 2, len(n.n.children))
	ItsEqual(t, c0.n.id, n.n.children[0].Node().id)
	ItsEqual(t, c2.n.id, n.n.children[1].Node().id)
}

func Test_Node_addParents(t *testing.T) {
	n := new(mockBareNode)
	_ = n.Node()

	c0 := new(mockBareNode)
	_ = c0.Node()

	c1 := new(mockBareNode)
	_ = c1.Node()

	ItsEqual(t, 0, len(n.n.parents))
	ItsEqual(t, 0, len(n.n.children))

	n.Node().addParents(c0, c1)

	ItsEqual(t, 2, len(n.n.parents))
	ItsEqual(t, 0, len(n.n.children))
	ItsEqual(t, c0.n.id, n.n.parents[0].Node().id)
	ItsEqual(t, c1.n.id, n.n.parents[1].Node().id)
}

func Test_Node_removeParent(t *testing.T) {
	n := new(mockBareNode)
	_ = n.Node()

	c0 := new(mockBareNode)
	_ = c0.Node()

	c1 := new(mockBareNode)
	_ = c1.Node()

	c2 := new(mockBareNode)
	_ = c2.Node()

	n.Node().addParents(c0, c1, c2)

	ItsEqual(t, 3, len(n.n.parents))
	ItsEqual(t, 0, len(n.n.children))
	ItsEqual(t, c0.n.id, n.n.parents[0].Node().id)
	ItsEqual(t, c1.n.id, n.n.parents[1].Node().id)
	ItsEqual(t, c2.n.id, n.n.parents[2].Node().id)

	n.Node().removeParent(c1.n.id)

	ItsEqual(t, 2, len(n.n.parents))
	ItsEqual(t, 0, len(n.n.children))
	ItsEqual(t, c0.n.id, n.n.parents[0].Node().id)
	ItsEqual(t, c2.n.id, n.n.parents[1].Node().id)
}

func Test_Node_maybeStabilize(t *testing.T) {
	ctx := testContext()
	n := new(mockBareNode)
	n.Node().gs = newGraphState()
	n.Node().gs.stabilizationNum = 5

	var calledStabilize bool
	n.n.stabilize = func(ictx context.Context) error {
		calledStabilize = true
		itsBlueDye(ictx, t)
		return nil
	}

	ItsEqual(t, 0, n.n.numRecomputes)
	ItsEqual(t, 0, n.n.recomputedAt)

	err := n.n.maybeStabilize(ctx)
	ItsNil(t, err)
	ItsEqual(t, true, calledStabilize)
	ItsEqual(t, 1, n.n.numRecomputes)
	ItsEqual(t, 5, n.n.recomputedAt)
}

func Test_Node_maybeStabilize_error(t *testing.T) {
	ctx := testContext()
	n := NewNode()
	n.gs = newGraphState()
	n.gs.stabilizationNum = 5

	n.stabilize = func(ictx context.Context) error {
		itsBlueDye(ictx, t)
		return fmt.Errorf("just a test")
	}

	ItsEqual(t, 0, n.numRecomputes)
	ItsEqual(t, 0, n.recomputedAt)

	err := n.maybeStabilize(ctx)
	ItsNotNil(t, err)
	ItsEqual(t, "just a test", err.Error())
	ItsEqual(t, 0, n.numRecomputes)
	ItsEqual(t, 0, n.recomputedAt)
}

func Test_Node_maybeCutoff(t *testing.T) {
	ctx := testContext()
	n := NewNode()

	ItsEqual(t, false, n.maybeCutoff(ctx))

	n.cutoff = func(ictx context.Context) bool {
		itsBlueDye(ictx, t)
		return true
	}

	ItsEqual(t, true, n.maybeCutoff(ctx))
}

func Test_Node_detectCutoff(t *testing.T) {
	yes := NewNode()
	yes.detectCutoff(new(cutoffIncr[string]))
	ItsNotNil(t, yes.cutoff)

	no := NewNode()
	no.detectCutoff(new(mockBareNode))
	ItsNil(t, no.cutoff)
}

func Test_Node_detectStabilize(t *testing.T) {
	yes := NewNode()
	yes.detectStabilize(new(applyIncr[string, string]))
	ItsNotNil(t, yes.stabilize)

	no := NewNode()
	no.detectStabilize(new(mockBareNode))
	ItsNil(t, no.stabilize)
}

func Test_Node_shouldRecompute(t *testing.T) {
	n := NewNode()
	ItsEqual(t, true, n.shouldRecompute())

	n.recomputedAt = 1
	ItsEqual(t, false, n.shouldRecompute())

	n.stabilize = func(_ context.Context) error { return nil }
	n.setAt = 2
	ItsEqual(t, true, n.shouldRecompute())

	n.setAt = 1
	n.changedAt = 2
	ItsEqual(t, true, n.shouldRecompute())

	n.changedAt = 1
	c1 := new(mockBareNode)
	c1.Node().changedAt = 2
	n.children = append(n.children, new(mockBareNode), c1)
	ItsEqual(t, true, n.shouldRecompute())

	c1.Node().changedAt = 1
	ItsEqual(t, false, n.shouldRecompute())
}

func Test_Node_calculateHeight(t *testing.T) {
	c010 := new(mockBareNode)
	c10 := new(mockBareNode)
	c00 := new(mockBareNode)
	c01 := new(mockBareNode)
	c0 := new(mockBareNode)
	c1 := new(mockBareNode)
	c2 := new(mockBareNode)
	p := new(mockBareNode)

	Link(c01, c010)
	Link(c0, c00, c01)
	Link(c1, c10)
	Link(p, c0, c1, c2)

	ItsEqual(t, 4, p.n.calculateHeight())
	ItsEqual(t, 3, c0.n.calculateHeight())
	ItsEqual(t, 2, c1.n.calculateHeight())
	ItsEqual(t, 1, c2.n.calculateHeight())
}

func Test_Node_recompute(t *testing.T) {
	ctx := testContext()

	n := NewNode()
	n.gs = newGraphState()

	ItsNotNil(t, n.gs.rh)

	var calledStabilize bool
	n.stabilize = func(ictx context.Context) error {
		calledStabilize = true
		itsBlueDye(ictx, t)
		return nil
	}
	p := new(mockBareNode)
	_ = p.Node()
	n.parents = append(n.parents, p)

	var calledUpdateHandler0, calledUpdateHandler1 bool
	n.onUpdateHandlers = append(n.onUpdateHandlers, func(ictx context.Context) {
		itsBlueDye(ictx, t)
		calledUpdateHandler0 = true
	})
	n.onUpdateHandlers = append(n.onUpdateHandlers, func(ictx context.Context) {
		itsBlueDye(ictx, t)
		calledUpdateHandler1 = true
	})

	var calledErrorHandler0, calledErrorHandler1 bool
	n.onErrorHandlers = append(n.onErrorHandlers, func(ictx context.Context, err error) {
		itsBlueDye(ictx, t)
		calledErrorHandler0 = true
	})
	n.onErrorHandlers = append(n.onErrorHandlers, func(ictx context.Context, err error) {
		itsBlueDye(ictx, t)
		calledErrorHandler1 = true
	})

	err := n.recompute(ctx, recomputeOptions{
		recomputeIfParentMinHeight: false,
	})
	ItsNil(t, err)

	// find the node in the recompute heap layer
	ItsEqual(t, true, n.gs.rh.Has(p))
	ItsEqual(t, true, calledStabilize)
	ItsEqual(t, true, calledUpdateHandler0)
	ItsEqual(t, true, calledUpdateHandler1)
	ItsEqual(t, false, calledErrorHandler0)
	ItsEqual(t, false, calledErrorHandler1)
}

func Test_Node_recompute_error(t *testing.T) {
	ctx := testContext()

	n := NewNode()
	n.gs = newGraphState()

	ItsNotNil(t, n.gs.rh)

	var calledStabilize bool
	n.stabilize = func(ictx context.Context) error {
		calledStabilize = true
		itsBlueDye(ictx, t)
		return fmt.Errorf("test error")
	}
	p := new(mockBareNode)
	_ = p.Node()
	n.parents = append(n.parents, p)

	var calledUpdateHandler0, calledUpdateHandler1 bool
	n.onUpdateHandlers = append(n.onUpdateHandlers, func(ictx context.Context) {
		itsBlueDye(ictx, t)
		calledUpdateHandler0 = true
	})
	n.onUpdateHandlers = append(n.onUpdateHandlers, func(ictx context.Context) {
		itsBlueDye(ictx, t)
		calledUpdateHandler1 = true
	})

	var calledErrorHandler0, calledErrorHandler1 bool
	n.onErrorHandlers = append(n.onErrorHandlers, func(ictx context.Context, err error) {
		itsBlueDye(ictx, t)
		calledErrorHandler0 = true
		ItsNotNil(t, err)
		ItsEqual(t, "test error", err.Error())
	})
	n.onErrorHandlers = append(n.onErrorHandlers, func(ictx context.Context, err error) {
		itsBlueDye(ictx, t)
		calledErrorHandler1 = true
		ItsNotNil(t, err)
		ItsEqual(t, "test error", err.Error())
	})

	err := n.recompute(ctx, recomputeOptions{
		recomputeIfParentMinHeight: false,
	})
	ItsNotNil(t, err)
	ItsNotNil(t, "test error", err.Error())

	// find the node in the recompute heap layer
	ItsEqual(t, false, n.gs.rh.Has(p))
	ItsEqual(t, true, calledStabilize)
	ItsEqual(t, false, calledUpdateHandler0)
	ItsEqual(t, false, calledUpdateHandler1)
	ItsEqual(t, true, calledErrorHandler0)
	ItsEqual(t, true, calledErrorHandler1)
}

func Test_nodeFormatters(t *testing.T) {
	id := NewIdentifier()

	testCases := [...]struct {
		Node  INode
		Label string
	}{
		{Bind[string, string](Return(""), nil), "bind"},
		{Bind2[string, int, bool](Return(""), Return(0), nil), "bind2"},
		{Bind3[string, int, float64, bool](Return(""), Return(0), Return(1.0), nil), "bind3"},
		{BindIf[string](Return(false), nil), "bind_if"},
		{Cutoff(Return(""), nil), "cutoff"},
		{Func[string](nil), "func"},
		{Apply[string, bool](Return(""), nil), "apply"},
		{Apply2[string, int, bool](Return(""), Return(0), nil), "apply2"},
		{Apply3[string, int, float64, bool](Return(""), Return(0), Return(1.0), nil), "apply3"},
		{ApplyIf(Return(""), Return(""), Return(false)), "apply_if"},
		{Return(""), "return"},
		{Watch(Return("")), "watch"},
		{Var(""), "var"},
	}

	for _, tc := range testCases {
		tc.Node.Node().id = id
		ItsEqual(t, fmt.Sprintf("%s[%s]", tc.Label, id.Short()), fmt.Sprint(tc.Node))
	}
}
