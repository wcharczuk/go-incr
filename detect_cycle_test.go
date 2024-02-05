package incr

import (
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_DetectCycleIfLinked(t *testing.T) {
	n0 := MapN[any, any](Root(), identMany)
	n01 := MapN[any, any](Root(), identMany)
	n02 := MapN[any, any](Root(), identMany)
	n03 := MapN[any, any](Root(), identMany)
	n1 := MapN[any, any](Root(), identMany)
	n11 := MapN[any, any](Root(), identMany)
	n12 := MapN[any, any](Root(), identMany)
	n13 := MapN[any, any](Root(), identMany)

	var err error
	err = n01.AddInput(n0)
	testutil.ItsNil(t, err)
	err = n02.AddInput(n01)
	testutil.ItsNil(t, err)
	err = n03.AddInput(n02)
	testutil.ItsNil(t, err)
	err = n1.AddInput(n02)
	testutil.ItsNil(t, err)
	err = n11.AddInput(n1)
	testutil.ItsNil(t, err)
	err = n12.AddInput(n11)
	testutil.ItsNil(t, err)

	err = DetectCycleIfLinked(n13, n12)
	testutil.ItsNil(t, err)

	err = DetectCycleIfLinked(n1, n12)
	testutil.ItsNotNil(t, err)
}

func detectCycleNode(label string) MapNIncr[any, any] {
	n := MapN[any, any](Root(), identMany)
	n.Node().SetLabel(label)
	return n
}

func Test_DetectCycleIfLinked_complex(t *testing.T) {
	n0 := detectCycleNode("n0")
	n1 := detectCycleNode("n1")
	n2 := detectCycleNode("n2")

	var err error
	err = n1.AddInput(n0)
	testutil.ItsNil(t, err)
	err = n2.AddInput(n1)
	testutil.ItsNil(t, err)

	n01 := detectCycleNode("n01")
	n02 := detectCycleNode("n02")

	err = n01.AddInput(n2)
	testutil.ItsNil(t, err)
	err = n02.AddInput(n01)
	testutil.ItsNil(t, err)

	n11 := detectCycleNode("n11")
	n12 := detectCycleNode("n12")
	n13 := detectCycleNode("n13")

	err = n11.AddInput(n01)
	testutil.ItsNil(t, err)
	err = n12.AddInput(n11)
	testutil.ItsNil(t, err)
	err = n13.AddInput(n12)
	testutil.ItsNil(t, err)

	n21 := detectCycleNode("n21")
	n22 := detectCycleNode("n22")
	n23 := detectCycleNode("n23")
	n24 := detectCycleNode("n24")

	err = n21.AddInput(n11)
	testutil.ItsNil(t, err)
	err = n22.AddInput(n21)
	testutil.ItsNil(t, err)
	err = n23.AddInput(n22)
	testutil.ItsNil(t, err)
	err = n24.AddInput(n23)
	testutil.ItsNil(t, err)

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

	var err error
	err = n10.AddInput(n0)
	testutil.ItsNil(t, err)
	err = n11.AddInput(n10)
	testutil.ItsNil(t, err)
	err = n12.AddInput(n11)
	testutil.ItsNil(t, err)
	err = n21.AddInput(n11)
	testutil.ItsNil(t, err)
	err = n22.AddInput(n21)
	testutil.ItsNil(t, err)

	err = DetectCycleIfLinked(n10, n22)
	testutil.ItsNotNil(t, err)
}

func Test_DetectCycleIfLinked_2(t *testing.T) {
	/* these are some trivial cases to make sure bases are covered */
	n0 := MapN[any, any](Root(), identMany)
	n1 := MapN[any, any](Root(), identMany)
	n2 := MapN[any, any](Root(), identMany)

	err := DetectCycleIfLinked(n0, n0)
	testutil.ItsNotNil(t, err)
	err = n1.AddInput(n0)
	testutil.ItsNil(t, err)

	err = DetectCycleIfLinked(n2, n1)
	testutil.ItsNil(t, err)

	err = n2.AddInput(n1)
	testutil.ItsNil(t, err)

	err = DetectCycleIfLinked(n0, n2)
	testutil.ItsNotNil(t, err)
}

func Test_DetectCycleIfLinked_regression(t *testing.T) {
	table := Var(Root(), "table")
	columnDownload := Map(Root(), table, ident)
	lastDownload := Map(Root(), columnDownload, ident)
	targetUpload := Map(Root(), lastDownload, ident)

	columnUpload := Map(Root(), table, ident)
	lastUpload := Map(Root(), columnUpload, ident)
	uploadRemaining := MapN(Root(), identMany, lastUpload)

	err := DetectCycleIfLinked(uploadRemaining, targetUpload)
	testutil.ItsNil(t, err, "this should _not_ cause a cycle!")
}

func Test_DetectCycleIfLinked_regression2(t *testing.T) {
	table := Var(Root(), "table")
	columnDownload := MapN(Root(), identMany[string])
	lastDownload := Map(Root(), columnDownload, ident)
	_ = Map(Root(), lastDownload, ident)

	columnUpload := Map(Root(), table, ident)
	lastUpload := Map(Root(), columnUpload, ident)
	_ = MapN(Root(), identMany, lastUpload)

	err := DetectCycleIfLinked(columnDownload, table)
	testutil.ItsNil(t, err, "this should _not_ cause a cycle!")
}
