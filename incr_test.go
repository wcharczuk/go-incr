package incremental

import "testing"

func Test_Return(t *testing.T) {
	expected := "foo"
	n := Return(expected)

	if value := n.Value(); value != expected {
		t.Errorf("expected %q, actual: %q", expected, value)
		t.FailNow()
	}
}

func Test_Map(t *testing.T) {
	m := Map(Return("foo"), func(v string) string {
		return "not "+v
	})

	expected := "not foo"
	if value := m.Value(); value != expected {
		t.Errorf("expected %q, actual: %q", expected, value)
		t.FailNow()
	}
}

func Test_Map2(t *testing.T) {
	m2 := Map2(Return("foo"), Return("bar"), func(v0, v1 string) string {
		return v0+" "+v1
	})

	expected := "foo bar"
	if value := m2.Value(); value != "foo bar" {
		t.Errorf("expected %q, actual: %q", expected, value)
		t.FailNow()
	}
}

func Test_Bind(t *testing.T) {
	b := Bind(Return("foo"), func(v string) Incr[string] {
		return IncrFunc[string](func() string {
			return v + " bar"
		})
	})

	expected := "foo bar"
	if value := b.Value(); value != expected {
		t.Errorf("expected %q, actual: %q", expected, value)
		t.FailNow()
	}
}
