package incr

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

var (
	_ json.Marshaler   = (*Identifier)(nil)
	_ json.Unmarshaler = (*Identifier)(nil)
)

func Test_Identifier(t *testing.T) {
	id := NewIdentifier()
	testutil.Equal(t, hex.EncodeToString(id[:]), id.String())
	testutil.Equal(t, hex.EncodeToString(id[12:]), id.Short())
}

func Test_SetIdentifierProvider(t *testing.T) {
	t.Cleanup(func() {
		SetIdentifierProvider(cryptoRandIdentifierProvider)
	})

	var counter uint64
	SetIdentifierProvider(func() (output Identifier) {
		newCounter := atomic.AddUint64(&counter, 1)
		output[15] = byte(newCounter)
		output[14] = byte(newCounter >> 8)
		output[13] = byte(newCounter >> 16)
		output[12] = byte(newCounter >> 24)
		output[11] = byte(newCounter >> 32)
		output[10] = byte(newCounter >> 40)
		output[9] = byte(newCounter >> 48)
		output[8] = byte(newCounter >> 56)
		return
	})
	testutil.Equal(t, "00000000000000000000000000000001", NewIdentifier().String())
	testutil.Equal(t, "00000000000000000000000000000002", NewIdentifier().String())
	testutil.Equal(t, "00000000000000000000000000000003", NewIdentifier().String())
}

func Test_Identifier_IsZero(t *testing.T) {
	id := NewIdentifier()
	testutil.Equal(t, false, id.IsZero())
	testutil.Equal(t, true, zero.IsZero())
	var test Identifier
	testutil.Equal(t, true, test.IsZero())
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
			testutil.NotNil(t, err)
			testutil.Equal(t, tc.ExpectedErr.Error(), err.Error())
		} else {
			testutil.Nil(t, err, tc.Input)
			testutil.Equal(t, tc.Expected, parsed, tc.Input)
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
	testutil.Nil(t, err)
	testutil.NotEqual(t, 0, len(data))

	var verify jsonTest
	err = json.Unmarshal(data, &verify)
	testutil.Nil(t, err)
	testutil.Equal(t, testValue.ID, verify.ID)
}

func Test_Identifier_jsonError(t *testing.T) {
	data := []byte(`{"id":"----------"}`)
	var verify jsonTest
	err := json.Unmarshal(data, &verify)
	testutil.NotNil(t, err)
	var zero Identifier
	testutil.Equal(t, zero, verify.ID)
}
