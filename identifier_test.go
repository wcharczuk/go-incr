package incr

import (
	"crypto/rand"
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

func Test_NewCryptoRandIdentifierProvider_NewIdentifier(t *testing.T) {
	provider := NewCryptoRandIdentifierProvider(rand.Reader)
	id00 := provider.NewIdentifier()
	testutil.Equal(t, hex.EncodeToString(id00[:]), id00.String())
	testutil.Equal(t, hex.EncodeToString(id00[12:]), id00.Short())

	id01 := provider.NewIdentifier()
	testutil.Equal(t, hex.EncodeToString(id01[:]), id01.String())
	testutil.Equal(t, hex.EncodeToString(id01[12:]), id01.Short())

	testutil.NotEqual(t, id00, id01, "crypto rand identifiers should be unique")

	// generate another 16 identifiers (to stress the buffer rotation)
	for x := 0; x < 16; x++ {
		id0n := provider.NewIdentifier()
		testutil.Equal(t, hex.EncodeToString(id0n[:]), id0n.String())
		testutil.Equal(t, hex.EncodeToString(id0n[12:]), id0n.Short())

		testutil.NotEqual(t, id0n, id00, "crypto rand identifiers should be unique")
		testutil.NotEqual(t, id0n, id01, "crypto rand identifiers should be unique")
	}
}

func Test_NewSequentialIdentifierProvider_NewIdentifier(t *testing.T) {
	provider := NewSequentialIdentifierProvier(1)
	id00 := provider.NewIdentifier()
	testutil.Equal(t, hex.EncodeToString(id00[:]), id00.String())
	testutil.Equal(t, hex.EncodeToString(id00[12:]), id00.Short())

	id01 := provider.NewIdentifier()
	testutil.Equal(t, hex.EncodeToString(id01[:]), id01.String())
	testutil.Equal(t, hex.EncodeToString(id01[12:]), id01.Short())

	testutil.Equal(t, false, id00.IsZero())
	testutil.Equal(t, false, id01.IsZero())

	testutil.NotEqual(t, id00, id01, "crypto rand identifiers should be unique")

	// generate another 16 identifiers (for no reason other than to be consistent)
	for x := 0; x < 16; x++ {
		id0n := provider.NewIdentifier()
		testutil.Equal(t, hex.EncodeToString(id0n[:]), id0n.String())
		testutil.Equal(t, hex.EncodeToString(id0n[12:]), id0n.Short())

		testutil.NotEqual(t, id0n, id00, "crypto rand identifiers should be unique")
		testutil.NotEqual(t, id0n, id01, "crypto rand identifiers should be unique")
	}
}

func Test_Identifier_IsZero(t *testing.T) {
	provider := NewCryptoRandIdentifierProvider(rand.Reader)
	id := provider.NewIdentifier()
	testutil.Equal(t, false, id.IsZero())
	testutil.Equal(t, true, _zero.IsZero())
	var test Identifier
	testutil.Equal(t, true, test.IsZero())
}

func Test_ParseIdentifier(t *testing.T) {
	provider := NewCryptoRandIdentifierProvider(rand.Reader)
	knownID := provider.NewIdentifier()
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
	provider := NewCryptoRandIdentifierProvider(rand.Reader)
	testValue := jsonTest{
		ID: provider.NewIdentifier(),
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
