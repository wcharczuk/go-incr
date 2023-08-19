package incr

import (
	"testing"

	. "github.com/wcharczuk/go-incr/testutil"
)

func Test_set(t *testing.T) {
	s := make(set[string])
	ItsEqual(t, false, s.has("foo"))
	ItsEqual(t, false, s.has("bar"))

	s.add("foo")
	ItsEqual(t, true, s.has("foo"))
	ItsEqual(t, false, s.has("bar"))
}
