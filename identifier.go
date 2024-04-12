package incr

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
	"sync"
)

// Identifier is a unique id.
type Identifier [16]byte

// NewIdentifier returns a new identifier.
//
// Currently the underlying data looks like a
// uuidv4 but that shouldn't be relied upon.
func NewIdentifier() (output Identifier) {
	output = identifierProvider()
	return
}

// MustParseIdentifier is the reverse of `.String()` that will
// panic if an error is returned by `ParseIdentifier`.
func MustParseIdentifier(raw string) (output Identifier) {
	var err error
	output, err = ParseIdentifier(raw)
	if err != nil {
		panic(err)
	}
	return
}

// ParseIdentifier is the reverse of `.String()`.
func ParseIdentifier(raw string) (output Identifier, err error) {
	if raw == "" {
		return
	}
	var parsed []byte
	parsed, err = hex.DecodeString(raw)
	if err != nil {
		return
	}
	if len(parsed) != 16 {
		err = fmt.Errorf("invalid identifier; must be 16 bytes")
		return
	}
	copy(output[:], parsed)
	return
}

// SetIdentifierProvider sets the identifier provider
// to a custom provider separate from the default.
func SetIdentifierProvider(ip func() Identifier) {
	identifierProvider = ip
}

func cryptoRandIdentifierProvider() (output Identifier) {
	identifierRandPoolMu.Lock()
	if identifierRandPoolPos == randPoolSize {
		_, _ = io.ReadFull(randomSource, identifierRandPool[:])
		identifierRandPoolPos = 0
	}
	copy(output[:], identifierRandPool[identifierRandPoolPos:(identifierRandPoolPos+16)])
	identifierRandPoolPos += 16
	identifierRandPoolMu.Unlock()
	output[6] = (output[6] & 0x0f) | 0x40 // Version 4
	output[8] = (output[8] & 0x3f) | 0x80 // Variant is 10
	return
}

const randPoolSize = 16 * 16

var (
	identifierProvider    = cryptoRandIdentifierProvider
	identifierRandPoolMu  sync.Mutex
	identifierRandPoolPos = randPoolSize     // protected with poolMu
	identifierRandPool    [randPoolSize]byte // protected with poolMu
	randomSource          = rand.Reader      // random function
)

var zero Identifier

// IsZero returns if the identifier is unset.
func (id Identifier) IsZero() bool {
	return id == zero
}

// MarshalJSON implements json.Marshaler.
func (id Identifier) MarshalJSON() ([]byte, error) {
	return []byte("\"" + id.String() + "\""), nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (id *Identifier) UnmarshalJSON(data []byte) error {
	dataCleaned := strings.TrimPrefix(string(data), "\"")
	dataCleaned = strings.TrimSuffix(dataCleaned, "\"")
	parsed, err := ParseIdentifier(dataCleaned)
	if err != nil {
		return err
	}
	*id = parsed
	return nil
}

// String returns the full hex representation of the id.
func (id Identifier) String() string {
	var buf [32]byte
	hex.Encode(buf[:], id[:])
	return string(buf[:])
}

// Short returns the short hex representation of the id.
//
// In practice this is the last ~8 bytes of the identifier.
func (id Identifier) Short() string {
	var buf [8]byte
	hex.Encode(buf[:], id[12:])
	return string(buf[:])
}
