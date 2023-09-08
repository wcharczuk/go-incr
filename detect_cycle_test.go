package incr

import (
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_DetectCycle(t *testing.T) {
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
	err = DetectCycle(n13, n12)
	testutil.ItsNil(t, err)

	err = DetectCycle(n1, n12)
	testutil.ItsNotNil(t, err)
}

func Test_DetectCycle_regression(t *testing.T) {
	table := Var("table")
	columnDownload := Map(table, ident)
	lastDownload := Map(columnDownload, ident)
	targetUpload := Map(lastDownload, ident)

	columnUpload := Map(table, ident)
	lastUpload := Map(columnUpload, ident)
	uploadRemaining := MapN(identMany, lastUpload)

	err := DetectCycle(uploadRemaining, targetUpload)
	testutil.ItsNil(t, err)
}
