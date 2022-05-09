package incr

import "testing"

func Test_Identifier(t *testing.T) {
	id := NewIdentifier()

	long := make([]byte, 36)
	encodeHex(long, id)
	ItsEqual(t, string(long), id.String())

	short := make([]byte, 8)
	encodeHexShort(short, id)
	ItsEqual(t, string(short), id.Short())
}
