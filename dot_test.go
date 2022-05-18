package incr

import (
	"bytes"
	"testing"
)

func Test_Dot(t *testing.T) {
	ctx := testContext()
	v0 := Var("foo")
	v1 := Var("bar")
	m0 := Apply2(v0.Read(), v1.Read(), concat)

	buf := new(bytes.Buffer)
	err := Dot(buf, m0)
	ItsNil(t, err)

}
