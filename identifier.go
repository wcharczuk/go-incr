package incr

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// Identifier is a unique id.
type Identifier [16]byte

// NewIdentifier returns a new identifier.
//
// In practice, the underlying data looks like a uuidv4
// but that shouldn't be relied upon.
func NewIdentifier() (output Identifier) {
	_, _ = rand.Read(output[:])
	output[6] = (output[6] & 0x0f) | 0x40 // Version 4
	output[8] = (output[8] & 0x3f) | 0x80 // Variant is 10
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

// String returns the full hex representation of the id.
func (id Identifier) String() string {
	var buf [32]byte
	hex.Encode(buf[:], id[:])
	return string(buf[:])
}

// Short returns the short hex representation of the id.
func (id Identifier) Short() string {
	var buf [8]byte
	hex.Encode(buf[:], id[12:])
	return string(buf[:])
}
