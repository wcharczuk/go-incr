package incr

import (
	"testing"

	. "github.com/wcharczuk/go-incr/testutil"
)

func Test_set(t *testing.T) {
	s := make(set[string])
	Equal(t, false, s.has("foo"))
	Equal(t, false, s.has("bar"))

	s.add("foo")
	Equal(t, true, s.has("foo"))
	Equal(t, false, s.has("bar"))
}
