package incr

import (
	"encoding/hex"
	"fmt"
	"testing"
)

func Test_Identifier(t *testing.T) {
	id := NewIdentifier()
	ItsEqual(t, hex.EncodeToString(id[:]), id.String())
	ItsEqual(t, hex.EncodeToString(id[12:]), id.Short())
}

func Test_ParseIdentifier(t *testing.T) {
	knownID := NewIdentifier()
	testCases := [...]struct {
		Input       string
		Expected    Identifier
		ExpectedErr error
	}{
		{"", Identifier{}, nil},
		{"zzzz", Identifier{}, fmt.Errorf("encoding/hex: invalid byte: U+007A 'z'")},
		{"deadbeef", Identifier{}, fmt.Errorf("invalid identifier; must be 16 bytes")},
		{knownID.String(), knownID, nil},
	}

	for _, tc := range testCases {
		parsed, err := ParseIdentifier(tc.Input)
		if tc.ExpectedErr != nil {
			ItsNotNil(t, err)
			ItsEqual(t, tc.ExpectedErr.Error(), err.Error())
		} else {
			ItsNil(t, err, tc.Input)
			ItsEqual(t, tc.Expected, parsed, tc.Input)
		}
	}
}
