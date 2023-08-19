package incr

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_Identifier(t *testing.T) {
	id := NewIdentifier()
	testutil.ItsEqual(t, hex.EncodeToString(id[:]), id.String())
	testutil.ItsEqual(t, hex.EncodeToString(id[12:]), id.Short())
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
			testutil.ItsNotNil(t, err)
			testutil.ItsEqual(t, tc.ExpectedErr.Error(), err.Error())
		} else {
			testutil.ItsNil(t, err, tc.Input)
			testutil.ItsEqual(t, tc.Expected, parsed, tc.Input)
		}
	}
}
