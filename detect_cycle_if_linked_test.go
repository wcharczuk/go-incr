package incr

import (
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_DetectCycleIfLinked(t *testing.T) {
	g := New()
	n0 := MapN[any](g, identMany)
	n01 := MapN[any](g, identMany)
	n02 := MapN[any](g, identMany)
	n03 := MapN[any](g, identMany)
	n1 := MapN[any](g, identMany)
	n11 := MapN[any](g, identMany)
	n12 := MapN[any](g, identMany)
	n13 := MapN[any](g, identMany)

	var err error
	err = n01.AddInput(n0)
	testutil.NoError(t, err)
	err = n02.AddInput(n01)
	testutil.NoError(t, err)
	err = n03.AddInput(n02)
	testutil.NoError(t, err)
	err = n1.AddInput(n02)
	testutil.NoError(t, err)
	err = n11.AddInput(n1)
	testutil.NoError(t, err)
	err = n12.AddInput(n11)
	testutil.NoError(t, err)

	err = DetectCycleIfLinked(n13, n12)
	testutil.NoError(t, err)

	err = DetectCycleIfLinked(n1, n12)
	testutil.Error(t, err)
}

func Test_DetectCycleIfLinked_nils(t *testing.T) {
	g := New()
	n0 := newMockBareNode(g)
	n1 := newMockBareNode(g)
	n1.n.parentsFn = func() []INode {
		return []INode{nil}
	}

	var err error
	err = DetectCycleIfLinked(nil, n0)
	testutil.NoError(t, err)

	err = DetectCycleIfLinked(n0, nil)
	testutil.NoError(t, err)

	err = DetectCycleIfLinked(n0, n1)
	testutil.NoError(t, err)
}

func detectCycleNode(label string) MapNIncr[any, any] {
	g := New()
	n := MapN[any](g, identMany)
	n.Node().SetLabel(label)
	return n
}

func Test_DetectCycleIfLinked_complex(t *testing.T) {
	n0 := detectCycleNode("n0")
	n1 := detectCycleNode("n1")
	n2 := detectCycleNode("n2")

	var err error
	err = n1.AddInput(n0)
	testutil.NoError(t, err)
	err = n2.AddInput(n1)
	testutil.NoError(t, err)

	n01 := detectCycleNode("n01")
	n02 := detectCycleNode("n02")

	err = n01.AddInput(n2)
	testutil.NoError(t, err)
	err = n02.AddInput(n01)
	testutil.NoError(t, err)

	n11 := detectCycleNode("n11")
	n12 := detectCycleNode("n12")
	n13 := detectCycleNode("n13")

	err = n11.AddInput(n01)
	testutil.NoError(t, err)
	err = n12.AddInput(n11)
	testutil.NoError(t, err)
	err = n13.AddInput(n12)
	testutil.NoError(t, err)

	n21 := detectCycleNode("n21")
	n22 := detectCycleNode("n22")
	n23 := detectCycleNode("n23")
	n24 := detectCycleNode("n24")

	err = n21.AddInput(n11)
	testutil.NoError(t, err)
	err = n22.AddInput(n21)
	testutil.NoError(t, err)
	err = n23.AddInput(n22)
	testutil.NoError(t, err)
	err = n24.AddInput(n23)
	testutil.NoError(t, err)

	err = DetectCycleIfLinked(n2, n02)
	testutil.NotNil(t, err)

	err = DetectCycleIfLinked(n2, n13)
	testutil.NotNil(t, err)

	err = DetectCycleIfLinked(n2, n24)
	testutil.NotNil(t, err)

	err = DetectCycleIfLinked(n02, n13)
	testutil.Nil(t, err, "this should _not_ cause a cycle")

	err = DetectCycleIfLinked(n01, n13)
	testutil.NotNil(t, err)
}

func Test_DetectCycleIfLinked_complex2(t *testing.T) {
	n0 := detectCycleNode("n0")
	n10 := detectCycleNode("n10")
	n11 := detectCycleNode("n11")
	n12 := detectCycleNode("n12")
	n21 := detectCycleNode("n21")
	n22 := detectCycleNode("n22")

	var err error
	err = n10.AddInput(n0)
	testutil.NoError(t, err)
	err = n11.AddInput(n10)
	testutil.NoError(t, err)
	err = n12.AddInput(n11)
	testutil.NoError(t, err)
	err = n21.AddInput(n11)
	testutil.NoError(t, err)
	err = n22.AddInput(n21)
	testutil.NoError(t, err)

	err = DetectCycleIfLinked(n10, n22)
	testutil.Error(t, err)
}

func Test_DetectCycleIfLinked_2(t *testing.T) {
	/* these are some trivial cases to make sure bases are covered */
	g := New()
	n0 := MapN[any](g, identMany)
	n1 := MapN[any](g, identMany)
	n2 := MapN[any](g, identMany)

	err := DetectCycleIfLinked(n0, n0)
	testutil.Error(t, err)
	err = n1.AddInput(n0)
	testutil.NoError(t, err)

	err = DetectCycleIfLinked(n2, n1)
	testutil.NoError(t, err)

	err = n2.AddInput(n1)
	testutil.NoError(t, err)

	err = DetectCycleIfLinked(n0, n2)
	testutil.NotNil(t, err)
}

func Test_DetectCycleIfLinked_regression(t *testing.T) {
	g := New()
	table := Var(g, "table")
	columnDownload := Map(g, table, ident)
	lastDownload := Map(g, columnDownload, ident)
	targetUpload := Map(g, lastDownload, ident)

	columnUpload := Map(g, table, ident)
	lastUpload := Map(g, columnUpload, ident)
	uploadRemaining := MapN(g, identMany, lastUpload)

	err := DetectCycleIfLinked(uploadRemaining, targetUpload)
	testutil.Nil(t, err, "this should _not_ cause a cycle!")
}

func Test_DetectCycleIfLinked_regression2(t *testing.T) {
	g := New()
	table := Var(g, "table")
	columnDownload := MapN(g, identMany[string])
	lastDownload := Map(g, columnDownload, ident)
	_ = Map(g, lastDownload, ident)

	columnUpload := Map(g, table, ident)
	lastUpload := Map(g, columnUpload, ident)
	_ = MapN(g, identMany, lastUpload)

	err := DetectCycleIfLinked(columnDownload, table)
	testutil.Nil(t, err, "this should _not_ cause a cycle!")
}
