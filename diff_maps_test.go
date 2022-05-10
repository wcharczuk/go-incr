package incr

import "testing"

func Test_diffMapsByKeysAdded(t *testing.T) {
	m0 := map[string]string{
		"b": "b",
		"c": "c",
	}
	m1 := map[string]string{
		"a": "a",
		"b": "b",
		"c": "c",
		"d": "d",
	}

	ma, orig := diffMapByKeysAdded(m0, m1)
	ItsEqual(t, map[string]string{
		"a": "a",
		"d": "d",
	}, ma)
	ItsEqual(t, m1, orig)

	ma, orig = diffMapByKeysAdded(nil, m1)
	ItsEqual(t, map[string]string{
		"a": "a",
		"b": "b",
		"c": "c",
		"d": "d",
	}, ma)
	ItsEqual(t, map[string]string{
		"a": "a",
		"b": "b",
		"c": "c",
		"d": "d",
	}, orig)
}

func Test_diffMapsByKeysRemoved(t *testing.T) {
	m0 := map[string]string{
		"a": "a",
		"b": "b",
		"c": "c",
		"d": "d",
	}
	m1 := map[string]string{
		"b": "b",
		"c": "c",
	}

	mr, orig := diffMapByKeysRemoved(m0, m1)
	ItsEqual(t, map[string]string{
		"a": "a",
		"d": "d",
	}, mr)
	ItsEqual(t, orig, m1)

	mr, orig = diffMapByKeysRemoved(m0, nil)
	ItsEqual(t, map[string]string{
		"a": "a",
		"b": "b",
		"c": "c",
		"d": "d",
	}, mr)
	ItsEqual(t, 0, len(orig))
}
