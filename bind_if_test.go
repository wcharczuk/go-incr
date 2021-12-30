package incr

import (
	"testing"
)

func Test_BindIf(t *testing.T) {

	a := Return("a")
	b := Return("b")
	c := Var(true)

	bi := BindIf(a, b, c)

	if len(bi.getNode().parents) != 3 {
		t.Errorf("expected (3) parents")
	}
	if len(a.getNode().children) != 1 {
		t.Errorf("expected (1) children for 'a'")
	}
	if len(b.getNode().children) != 1 {
		t.Errorf("expected (1) children for 'b'")
	}
	if len(c.getNode().children) != 1 {
		t.Errorf("expected (1) children for 'c'")
	}

	// c.Set(true)
	value := bi.Value()
	if value != "a" {
		t.Errorf("expected value to be 'a', actual: %q", value)
	}

	c.Set(false)
	value = bi.Value()
	if value != "b" {
		t.Errorf("expected value to be 'b', actual: %q", value)
	}
}
