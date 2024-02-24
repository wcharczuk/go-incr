package incr

import (
	"bytes"
	"strings"
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_Dot(t *testing.T) {
	g := New()

	v0 := Var(g, "foo")
	v1 := Var(g, "bar")

	m2 := Map2(g, v0, v1, concat)
	m3 := Map2(g, m2, Return(g, "const"), concat)

	s := Sentinel(g, func() bool { return true }, m2)

	o := MustObserve(g, m3)

	buffer := new(bytes.Buffer)

	err := Dot(buffer, g)
	testutil.NoError(t, err)

	testutil.NotEqual(t, "", buffer.String())

	testutil.Equal(t, true, strings.Contains(buffer.String(), o.Node().id.Short()))
	testutil.Equal(t, true, strings.Contains(buffer.String(), s.Node().id.Short()))
	testutil.Equal(t, true, strings.Contains(buffer.String(), m2.Node().id.Short()))
	testutil.Equal(t, true, strings.Contains(buffer.String(), m3.Node().id.Short()))
	testutil.Equal(t, true, strings.Contains(buffer.String(), v0.Node().id.Short()))
	testutil.Equal(t, true, strings.Contains(buffer.String(), v1.Node().id.Short()))
}
