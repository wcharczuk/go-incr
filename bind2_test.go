package incr

import (
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_Bind2_basic(t *testing.T) {
	ctx := testContext()

	bindVarA := Var(Root(), "a")
	bindVarA.Node().SetLabel("bindVarA")
	bindVarB := Var(Root(), "a")
	bindVarB.Node().SetLabel("bindVarB")

	av := Var(Root(), "a-value")
	av.Node().SetLabel("av")
	a0 := Map(Root(), av, ident)
	a0.Node().SetLabel("a0")
	a1 := Map(Root(), a0, ident)
	a1.Node().SetLabel("a1")

	bv := Var(Root(), "b-value")
	bv.Node().SetLabel("bv")
	b0 := Map(Root(), bv, ident)
	b0.Node().SetLabel("b0")
	b1 := Map(Root(), b0, ident)
	b1.Node().SetLabel("b1")
	b2 := Map(Root(), b1, ident)
	b2.Node().SetLabel("b2")

	bind2 := Bind2(Root(), bindVarA, bindVarB, func(_ *BindScope, whichA, whichB string) Incr[string] {
		if whichA == "a" && whichB == "a" {
			return a1
		}
		if whichA == "b" && whichB == "b" {
			return b2
		}
		return nil
	})
	bind2.Node().SetLabel("bind2")

	testutil.ItMatches(t, "bind2\\[(.*)\\]:bind", bind2.String())

	s0 := Return(Root(), "hello")
	s0.Node().SetLabel("s0")
	s1 := Map(Root(), s0, ident)
	s1.Node().SetLabel("s1")

	o := Map2(Root(), bind2, s1, concat)
	o.Node().SetLabel("o")

	// we shouldn't have bind internals set up on construction
	testutil.ItsNil(t, ExpertBind(bind2).BindChange())
	testutil.ItsNil(t, ExpertBind(bind2).Bound())

	g := New()
	_ = Observe(Root(), g, o)

	// we shouldn't have bind internals set up after observation either
	testutil.ItsNil(t, ExpertBind(bind2).BindChange())
	testutil.ItsNil(t, ExpertBind(bind2).Bound())

	var err error
	testutil.ItsEqual(t, 1, bindVarA.Node().height)
	testutil.ItsEqual(t, 1, s0.Node().height)
	testutil.ItsEqual(t, 2, s1.Node().height)

	testutil.ItsEqual(t, 2, bind2.Node().height)
	testutil.ItsEqual(t, 3, o.Node().height)

	testutil.ItsEqual(t, true, g.IsObserving(bindVarA))
	testutil.ItsEqual(t, true, g.IsObserving(s0))
	testutil.ItsEqual(t, true, g.IsObserving(s1))
	testutil.ItsEqual(t, true, g.IsObserving(bind2))
	testutil.ItsEqual(t, true, g.IsObserving(o))

	testutil.ItsEqual(t, false, g.IsObserving(av))
	testutil.ItsEqual(t, false, g.IsObserving(a0))
	testutil.ItsEqual(t, false, g.IsObserving(a1))

	testutil.ItsEqual(t, false, g.IsObserving(bv))
	testutil.ItsEqual(t, false, g.IsObserving(b0))
	testutil.ItsEqual(t, false, g.IsObserving(b1))
	testutil.ItsEqual(t, false, g.IsObserving(b2))

	err = g.Stabilize(ctx)
	testutil.ItsNil(t, err)

	err = dumpDot(g, homedir("bind_basic_00.png"))
	testutil.ItsNil(t, err)

	// we _should_ have bind internals set up after stabilization
	testutil.ItsNotNil(t, ExpertBind(bind2).BindChange())
	testutil.ItsNotNil(t, ExpertBind(bind2).Bound())

	bindChange := ExpertBind(bind2).BindChange()
	testutil.ItsEqual(t, true, bindChange.Node().HasParent(bindVarA.Node().ID()))
	testutil.ItsEqual(t, true, bindChange.Node().HasParent(bindVarB.Node().ID()))
	testutil.ItsEqual(t, true, bindChange.Node().HasChild(a1.Node().ID()))
	testutil.ItsEqual(t, false, bindChange.Node().HasChild(b1.Node().ID()))

	testutil.ItsEqual(t, 1, bind2.Node().boundAt)
	testutil.ItsEqual(t, 1, bind2.Node().changedAt)
	testutil.ItsEqual(t, 1, a1.Node().changedAt)

	testutil.ItsEqual(t, 1, bindVarA.Node().height)
	testutil.ItsEqual(t, 1, bindVarB.Node().height)
	testutil.ItsEqual(t, 1, s0.Node().height)
	testutil.ItsEqual(t, 2, s1.Node().height)

	testutil.ItsEqual(t, 1, av.Node().height)
	testutil.ItsEqual(t, 2, a0.Node().height)
	testutil.ItsEqual(t, 3, a1.Node().height)

	testutil.ItsEqual(t, 2, bind2.Node().height)
	testutil.ItsEqual(t, 4, o.Node().height)

	testutil.ItsEqual(t, true, g.IsObserving(bindVarA))
	testutil.ItsEqual(t, true, g.IsObserving(s0))
	testutil.ItsEqual(t, true, g.IsObserving(s1))
	testutil.ItsEqual(t, true, g.IsObserving(bind2))
	testutil.ItsEqual(t, true, g.IsObserving(o))

	testutil.ItsEqual(t, true, g.IsObserving(av))
	testutil.ItsEqual(t, true, g.IsObserving(a0))
	testutil.ItsEqual(t, true, g.IsObserving(a1))

	testutil.ItsEqual(t, false, g.IsObserving(bv))
	testutil.ItsEqual(t, false, g.IsObserving(b0))
	testutil.ItsEqual(t, false, g.IsObserving(b1))
	testutil.ItsEqual(t, false, g.IsObserving(b2))

	testutil.ItsEqual(t, "a-value", av.Value())
	testutil.ItsEqual(t, "a-value", bind2.Value())
	testutil.ItsEqual(t, "a-valuehello", o.Value())

	bindVarA.Set("b")
	bindVarB.Set("b")
	err = g.Stabilize(ctx)
	testutil.ItsNil(t, err)

	err = dumpDot(g, homedir("bind_basic_01.png"))
	testutil.ItsNil(t, err)

	bindChange = ExpertBind(bind2).BindChange()
	testutil.ItsEqual(t, true, bindChange.Node().HasParent(bindVarA.Node().ID()))
	testutil.ItsEqual(t, true, bindChange.Node().HasParent(bindVarB.Node().ID()))
	testutil.ItsEqual(t, false, bindChange.Node().HasChild(a1.Node().ID()))
	testutil.ItsEqual(t, true, bindChange.Node().HasChild(b2.Node().ID()))

	testutil.ItsEqual(t, 1, bindVarA.Node().height)
	testutil.ItsEqual(t, 1, bindVarB.Node().height)
	testutil.ItsEqual(t, 1, s0.Node().height)
	testutil.ItsEqual(t, 2, s1.Node().height)

	testutil.ItsEqual(t, 1, av.Node().height)
	testutil.ItsEqual(t, 2, a0.Node().height)
	testutil.ItsEqual(t, 3, a1.Node().height)

	testutil.ItsEqual(t, 1, bv.Node().height)
	testutil.ItsEqual(t, 2, b0.Node().height)
	testutil.ItsEqual(t, 3, b1.Node().height)
	testutil.ItsEqual(t, 4, b2.Node().height)

	testutil.ItsEqual(t, 2, bind2.Node().height)
	testutil.ItsEqual(t, 5, o.Node().height)

	testutil.ItsEqual(t, true, g.IsObserving(bindVarA))
	testutil.ItsEqual(t, true, g.IsObserving(bindVarB))
	testutil.ItsEqual(t, true, g.IsObserving(s0))
	testutil.ItsEqual(t, true, g.IsObserving(s1))
	testutil.ItsEqual(t, true, g.IsObserving(bind2))
	testutil.ItsEqual(t, true, g.IsObserving(o))

	testutil.ItsEqual(t, false, g.IsObserving(av), "if we switch to b, we should unobserve the 'a' tree")
	testutil.ItsEqual(t, false, g.IsObserving(a0))
	testutil.ItsEqual(t, false, g.IsObserving(a1))

	testutil.ItsEqual(t, true, g.IsObserving(bv))
	testutil.ItsEqual(t, true, g.IsObserving(b0))
	testutil.ItsEqual(t, true, g.IsObserving(b1))
	testutil.ItsEqual(t, true, g.IsObserving(b2))

	testutil.ItsEqual(t, "a-value", av.Value())
	testutil.ItsEqual(t, "b-value", bv.Value())
	testutil.ItsEqual(t, "b-valuehello", o.Value())

	bindVarA.Set("neither")
	bindVarB.Set("neither")
	err = g.Stabilize(ctx)
	testutil.ItsNil(t, err)

	err = dumpDot(g, homedir("bind_basic_02.png"))
	testutil.ItsNil(t, err)

	bindChange = ExpertBind(bind2).BindChange()
	testutil.ItsNil(t, bindChange)

	testutil.ItsEqual(t, 1, bindVarA.Node().height)
	testutil.ItsEqual(t, 1, bindVarB.Node().height)
	testutil.ItsEqual(t, 1, s0.Node().height)
	testutil.ItsEqual(t, 2, s1.Node().height)

	testutil.ItsEqual(t, 1, av.Node().height)
	testutil.ItsEqual(t, 2, a0.Node().height)
	testutil.ItsEqual(t, 3, a1.Node().height)

	testutil.ItsEqual(t, 1, bv.Node().height)
	testutil.ItsEqual(t, 2, b0.Node().height)
	testutil.ItsEqual(t, 3, b1.Node().height)
	testutil.ItsEqual(t, 4, b2.Node().height)

	testutil.ItsEqual(t, 2, bind2.Node().height)
	testutil.ItsEqual(t, 5, o.Node().height)

	testutil.ItsEqual(t, true, g.IsObserving(bindVarA))
	testutil.ItsEqual(t, true, g.IsObserving(s0))
	testutil.ItsEqual(t, true, g.IsObserving(s1))
	testutil.ItsEqual(t, true, g.IsObserving(bind2))
	testutil.ItsEqual(t, true, g.IsObserving(o))

	testutil.ItsEqual(t, false, g.IsObserving(av))
	testutil.ItsEqual(t, false, g.IsObserving(a0))
	testutil.ItsEqual(t, false, g.IsObserving(a1))

	testutil.ItsEqual(t, false, g.IsObserving(bv))
	testutil.ItsEqual(t, false, g.IsObserving(b0))
	testutil.ItsEqual(t, false, g.IsObserving(b1))
	testutil.ItsEqual(t, false, g.IsObserving(b2))

	testutil.ItsEqual(t, "a-value", av.Value())
	testutil.ItsEqual(t, "b-value", bv.Value())
	testutil.ItsEqual(t, "hello", o.Value())
}
