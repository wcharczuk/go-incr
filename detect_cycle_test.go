package incr

import (
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_DetectCycle(t *testing.T) {
	n00 := MapN[any, any](identMany)
	n01 := MapN[any, any](identMany)
	n02 := MapN[any, any](identMany)
	n03 := MapN[any, any](identMany)
	n10 := MapN[any, any](identMany)
	n11 := MapN[any, any](identMany)
	n12 := MapN[any, any](identMany)
	n13 := MapN[any, any](identMany)

	n01.AddInput(n00)
	n02.AddInput(n01)
	n03.AddInput(n02)
	n10.AddInput(n02)
	n11.AddInput(n10)
	n12.AddInput(n11)
	n13.AddInput(n12)
	n10.AddInput(n12)

	err := DetectCycle(n10)
	testutil.ItsNotNil(t, err)

	err = DetectCycle(n02)
	testutil.ItsNil(t, err)
}

func Test_DetectCycleIfLinked(t *testing.T) {
	n0 := MapN[any, any](identMany)
	n01 := MapN[any, any](identMany)
	n02 := MapN[any, any](identMany)
	n03 := MapN[any, any](identMany)
	n1 := MapN[any, any](identMany)
	n11 := MapN[any, any](identMany)
	n12 := MapN[any, any](identMany)
	n13 := MapN[any, any](identMany)

	n01.AddInput(n0)
	n02.AddInput(n01)
	n03.AddInput(n02)
	n1.AddInput(n02)
	n11.AddInput(n1)
	n12.AddInput(n11)

	var err error
	err = DetectCycleIfLinked(n13, n12)
	testutil.ItsNil(t, err)

	err = DetectCycleIfLinked(n1, n12)
	testutil.ItsNotNil(t, err)
}

func detectCycleNode(label string) MapNIncr[any, any] {
	n := MapN[any, any](identMany)
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

	var err error
	err = DetectCycleIfLinked(n2, n02)
	testutil.ItsNotNil(t, err)

	err = DetectCycleIfLinked(n2, n13)
	testutil.ItsNotNil(t, err)

	err = DetectCycleIfLinked(n2, n24)
	testutil.ItsNotNil(t, err)

	err = DetectCycleIfLinked(n02, n13)
	testutil.ItsNil(t, err, "this should _not_ cause a cycle")

	err = DetectCycleIfLinked(n01, n13)
	testutil.ItsNotNil(t, err)
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
	testutil.ItsNotNil(t, err)
}

func Test_DetectCycleIfLinked_2(t *testing.T) {
	/* these are some trivial cases to make sure bases are covered */

	n0 := MapN[any, any](identMany)
	n1 := MapN[any, any](identMany)
	n2 := MapN[any, any](identMany)

	err := DetectCycleIfLinked(n0, n0)
	testutil.ItsNotNil(t, err)
	n1.AddInput(n0)

	err = DetectCycleIfLinked(n2, n1)
	testutil.ItsNil(t, err)

	n2.AddInput(n1)

	err = DetectCycleIfLinked(n0, n2)
	testutil.ItsNotNil(t, err)
}

func Test_DetectCycleIfLinked_regression(t *testing.T) {
	table := Var("table")
	columnDownload := Map(table, ident)
	lastDownload := Map(columnDownload, ident)
	targetUpload := Map(lastDownload, ident)

	columnUpload := Map(table, ident)
	lastUpload := Map(columnUpload, ident)
	uploadRemaining := MapN(identMany, lastUpload)

	err := DetectCycleIfLinked(uploadRemaining, targetUpload)
	testutil.ItsNil(t, err, "this should _not_ cause a cycle!")
}

func Test_DetectCycleIfLinked_regression2(t *testing.T) {
	table := Var("table")
	columnDownload := MapN(identMany[string])
	lastDownload := Map(columnDownload, ident)
	_ = Map(lastDownload, ident)

	columnUpload := Map(table, ident)
	lastUpload := Map(columnUpload, ident)
	_ = MapN(identMany, lastUpload)

	err := DetectCycleIfLinked(columnDownload, table)
	testutil.ItsNil(t, err, "this should _not_ cause a cycle!")
}
