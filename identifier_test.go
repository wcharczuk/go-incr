package incr

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

var (
	_ json.Marshaler   = (*Identifier)(nil)
	_ json.Unmarshaler = (*Identifier)(nil)
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

type jsonTest struct {
	ID Identifier `json:"id"`
}

func Test_Identifier_json(t *testing.T) {
	testValue := jsonTest{
		ID: NewIdentifier(),
	}
	data, err := json.Marshal(testValue)
	testutil.ItsNil(t, err)
	testutil.ItsNotEqual(t, 0, len(data))

	var verify jsonTest
	err = json.Unmarshal(data, &verify)
	testutil.ItsNil(t, err)
	testutil.ItsEqual(t, testValue.ID, verify.ID)
}

func Test_Identifier_jsonError(t *testing.T) {
	data := []byte(`{"id":"----------"}`)
	var verify jsonTest
	err := json.Unmarshal(data, &verify)
	testutil.ItsNotNil(t, err)
	var zero Identifier
	testutil.ItsEqual(t, zero, verify.ID)
}
