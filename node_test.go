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
	testutil.Equal(t, "test_node", n.kind)
	testutil.Equal(t, 0, len(n.parents))
	testutil.Equal(t, 0, len(n.children))
	testutil.Equal(t, "", n.label)
	testutil.Equal(t, HeightUnset, n.height)
	testutil.Equal(t, HeightUnset, n.heightInRecomputeHeap)
	testutil.Equal(t, HeightUnset, n.heightInAdjustHeightsHeap)
	testutil.Equal(t, 0, n.changedAt)
	testutil.Equal(t, 0, n.setAt)
	testutil.Equal(t, 0, n.recomputedAt)
	testutil.Nil(t, n.onUpdateHandlers)
	testutil.Nil(t, n.onErrorHandlers)
	testutil.Nil(t, n.stabilizeFn)
	testutil.Nil(t, n.cutoffFn)
	testutil.Equal(t, 0, n.numRecomputes)
	testutil.Nil(t, n.metadata)
}

func Test_Node_ID(t *testing.T) {
	n := NewNode("test_node")
	testutil.Equal(t, true, n.ID().IsZero(), "identifiers are provided by the scope through WithinScope")

	other := _defaultIdentifierProvider.NewIdentifier()
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
	_ = MustObserve(g, n)
	g.SetStale(n)

	testutil.Equal(t, 0, n.n.changedAt)
	testutil.Equal(t, 0, n.n.recomputedAt)
	testutil.Equal(t, 1, n.n.setAt)

	testutil.Equal(t, true, GraphForNode(n).recomputeHeap.has(n))

	// find the node in the recompute heap layer
	testutil.Equal(t, 1, GraphForNode(n).recomputeHeap.heights[0].len())

	g.SetStale(n)

	testutil.Equal(t, 0, n.n.changedAt)
	testutil.Equal(t, 0, n.n.recomputedAt)
	testutil.Equal(t, 1, n.n.setAt)

	testutil.Equal(t, true, GraphForNode(n).recomputeHeap.has(n))

	testutil.Equal(t, 1, GraphForNode(n).recomputeHeap.heights[0].len())
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

func Test_Node_addObserver(t *testing.T) {
	g := New()

	n := newMockBareNode(g)
	o0 := mockObserver(g)
	o1 := mockObserver(g)

	testutil.Equal(t, 0, len(n.n.observers))

	n.Node().addObservers(o0, o1)

	testutil.Equal(t, 2, len(n.n.observers))

	testutil.Equal(t, true, hasKey(n.n.observers, o0.Node().id))
	testutil.Equal(t, true, hasKey(n.n.observers, o1.Node().id))
}

func Test_Node_removeObservers(t *testing.T) {
	g := New()

	n := newMockBareNode(g)
	o0 := mockObserver(g)
	o1 := mockObserver(g)

	testutil.Equal(t, 0, len(n.n.observers))

	n.Node().addObservers(o0, o1)

	testutil.Equal(t, 2, len(n.n.observers))

	n.Node().removeObserver(o1.Node().id)

	testutil.Equal(t, 1, len(n.n.observers))

	testutil.Equal(t, true, hasKey(n.n.observers, o0.Node().id))
	testutil.Equal(t, false, hasKey(n.n.observers, o1.Node().id))
}

func Test_Node_maybeStabilize(t *testing.T) {
	ctx := testContext()
	n := NewNode("test_node")

	// assert it doesn't panic
	err := n.maybeStabilize(ctx)
	testutil.Nil(t, err)

	var calledStabilize bool
	n.stabilizeFn = func(ictx context.Context) error {
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

	n.stabilizeFn = func(ictx context.Context) error {
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

	n.cutoffFn = func(ictx context.Context) (bool, error) {
		testutil.BlueDye(ictx, t)
		return true, nil
	}

	shouldCutoff, err = n.maybeCutoff(ctx)
	testutil.Nil(t, err)
	testutil.Equal(t, true, shouldCutoff)

	n.cutoffFn = func(ictx context.Context) (bool, error) {
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
	testutil.NotNil(t, yes.cutoffFn)

	no := NewNode("test_node")
	no.detectCutoff(new(mockBareNode))
	testutil.Nil(t, no.cutoffFn)
}

func Test_Node_detectStabilize(t *testing.T) {
	yes := NewNode("test_node")
	yes.detectStabilize(new(mapIncr[string, string]))
	testutil.NotNil(t, yes.stabilizeFn)

	no := NewNode("test_node")
	no.detectStabilize(new(mockBareNode))
	testutil.Nil(t, no.stabilizeFn)
}

func Test_Node_isStale(t *testing.T) {
	n := NewNode("test_node")
	testutil.Equal(t, true, n.isStale())

	n.valid = false
	testutil.Equal(t, false, n.isStale())

	n.valid = true
	n.staleFn = func() bool { return true }
	testutil.Equal(t, true, n.isStale())

	n.valid = true
	n.staleFn = nil
	n.recomputedAt = 1
	testutil.Equal(t, false, n.isStale())
}

func Test_Node_isNecessary(t *testing.T) {
	g := New()
	n := NewNode("test_node")
	testutil.Equal(t, false, n.isNecessary())

	n.observer = true
	testutil.Equal(t, true, n.isNecessary())

	n.observer = false
	n.forceNecessary = true
	testutil.Equal(t, true, n.isNecessary())

	n.observer = false
	n.forceNecessary = false
	n.children = []INode{newMockBareNode(g)}
	testutil.Equal(t, true, n.isNecessary())

	n.observer = false
	n.forceNecessary = false
	n.children = nil
	n.observers = []IObserver{mockObserver(g)}
	testutil.Equal(t, true, n.isNecessary())

	n.observer = false
	n.forceNecessary = false
	n.children = nil
	n.observers = nil
	testutil.Equal(t, false, n.isNecessary())
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
	_ = MustObserve(g, m0)

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
	id := _defaultIdentifierProvider.NewIdentifier()
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
		{Map4[string, int, float64, bool, string](g, Return(g, ""), Return(g, 0), Return(g, 1.0), Return(g, false), nil), "map4"},
		{Map5[string, int, float64, bool, string, string](g, Return(g, ""), Return(g, 0), Return(g, 1.0), Return(g, false), Return(g, ""), nil), "map5"},
		{Map6[string, int, float64, bool, string, int, string](g, Return(g, ""), Return(g, 0), Return(g, 1.0), Return(g, false), Return(g, ""), Return(g, 2), nil), "map6"},
		{MapIf(g, Return(g, ""), Return(g, ""), Return(g, false)), "map_if"},
		{Return(g, ""), "return"},
		{Watch(g, Return(g, "")), "watch"},
		{Freeze(g, Return(g, "")), "freeze"},
		{Var(g, ""), "var"},
		// {MustObserve(g, Return(g, "")), "observer"},
		{Always(g, Return(g, "")), "always"},
	}

	for _, tc := range testCases {
		tc.Node.Node().id = id
		tc.Node.Node().height = 2
		testutil.Equal(t, fmt.Sprintf("%s[%s]@2", tc.Label, id.Short()), fmt.Sprint(tc.Node))
		tc.Node.Node().label = "test-label"
		testutil.Equal(t, fmt.Sprintf("%s[%s]:test-label@2", tc.Label, id.Short()), fmt.Sprint(tc.Node))
	}
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

	areaObs := MustObserve(g, area)
	areaObs.Node().SetLabel("areaObs")
	scaledVolumeObs := MustObserve(g, scaledVolume)
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

func Test_Node_shouldBeInvalidated(t *testing.T) {
	n := NewNode("bogus")
	n.valid = false
	testutil.Equal(t, false, n.shouldBeInvalidated())
}

func Test_Node_shouldBeInvalidated_fn(t *testing.T) {
	n := NewNode("bogus")
	n.valid = true
	n.shouldBeInvalidatedFn = func() bool {
		return true
	}
	testutil.Equal(t, true, n.shouldBeInvalidated())
	n.shouldBeInvalidatedFn = func() bool {
		return false
	}
	testutil.Equal(t, false, n.shouldBeInvalidated())
}

func Test_Node_shouldBeInvalidated_parent(t *testing.T) {
	g := New()
	n := NewNode("bogus")
	n.valid = true
	n.shouldBeInvalidatedFn = nil
	okParent := newMockBareNode(g)
	okParent.Node().valid = true
	notOkParent := newMockBareNode(g)
	notOkParent.Node().valid = false
	n.parents = []INode{
		okParent, notOkParent,
	}
	testutil.Equal(t, true, n.shouldBeInvalidated())
}

func Test_Node_shouldBeInvalidated_fallThrough(t *testing.T) {
	g := New()
	n := NewNode("bogus")
	n.valid = true
	n.shouldBeInvalidatedFn = nil
	okParent := newMockBareNode(g)
	okParent.Node().valid = true
	n.parents = []INode{
		okParent,
	}
	testutil.Equal(t, false, n.shouldBeInvalidated())
}
