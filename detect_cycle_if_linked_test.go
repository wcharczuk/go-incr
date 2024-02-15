package incr

import (
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_DetectCycleIfLinked(t *testing.T) {
	g := New()
	n0 := MapN[any, any](g, identMany)
	n01 := MapN[any, any](g, identMany)
	n02 := MapN[any, any](g, identMany)
	n03 := MapN[any, any](g, identMany)
	n1 := MapN[any, any](g, identMany)
	n11 := MapN[any, any](g, identMany)
	n12 := MapN[any, any](g, identMany)
	n13 := MapN[any, any](g, identMany)

	n01.AddInput(n0)
	n02.AddInput(n01)
	n03.AddInput(n02)
	n1.AddInput(n02)
	n11.AddInput(n1)
	n12.AddInput(n11)

	err := DetectCycleIfLinked(n13, n12)
	testutil.Nil(t, err)

	err = DetectCycleIfLinked(n1, n12)
	testutil.NotNil(t, err)
}

func detectCycleNode(label string) MapNIncr[any, any] {
	g := New()
	n := MapN[any, any](g, identMany)
	n.Node().SetLabel(label)
	return n
}

func Test_DetectCycleIfLinked_complex(t *testing.T) {
	n0 := detectCycleNode("n0")
	n1 := detectCycleNode("n1")
	n2 := detectCycleNode("n2")

	n1.AddInput(n0)
	n2.AddInput(n1)

	n01 := detectCycleNode("n01")
	n02 := detectCycleNode("n02")

	n01.AddInput(n2)
	n02.AddInput(n01)

	n11 := detectCycleNode("n11")
	n12 := detectCycleNode("n12")
	n13 := detectCycleNode("n13")

	n11.AddInput(n01)
	n12.AddInput(n11)
	n13.AddInput(n12)

	n21 := detectCycleNode("n21")
	n22 := detectCycleNode("n22")
	n23 := detectCycleNode("n23")
	n24 := detectCycleNode("n24")

	n21.AddInput(n11)
	n22.AddInput(n21)
	n23.AddInput(n22)
	n24.AddInput(n23)

	err := DetectCycleIfLinked(n2, n02)
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

	n10.AddInput(n0)
	n11.AddInput(n10)
	n12.AddInput(n11)
	n21.AddInput(n11)
	n22.AddInput(n21)

	err := DetectCycleIfLinked(n10, n22)
	testutil.NotNil(t, err)
}

func Test_DetectCycleIfLinked_2(t *testing.T) {
	/* these are some trivial cases to make sure bases are covered */
	g := New()
	n0 := MapN[any, any](g, identMany)
	n1 := MapN[any, any](g, identMany)
	n2 := MapN[any, any](g, identMany)

	err := DetectCycleIfLinked(n0, n0)
	testutil.NotNil(t, err)
	n1.AddInput(n0)

	err = DetectCycleIfLinked(n2, n1)
	testutil.Nil(t, err)

	n2.AddInput(n1)

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
