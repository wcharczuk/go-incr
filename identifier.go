package incr

import (
	"crypto/rand"
	"encoding/hex"
)

// Identifier is a unique id.
type Identifier [16]byte

// NewIdentifier returns a new identifier.
//
// In practice, the underlying data looks like a uuidv4
// but it is not advisable to rely on this.
func NewIdentifier() (output Identifier) {
	_, _ = rand.Read(output[:])
	output[6] = (output[6] & 0x0f) | 0x40 // Version 4
	output[8] = (output[8] & 0x3f) | 0x80 // Variant is 10
	return
}

// String returns the full hex representation of the id.
func (id Identifier) String() string {
	var buf [36]byte
	encodeHex(buf[:], id)
	return string(buf[:])
}

// Short returns the short hex representation of the id.
func (id Identifier) Short() string {
	var buf [8]byte
	encodeHexShort(buf[:], id)
	return string(buf[:])
}

func encodeHex(dst []byte, id Identifier) {
	hex.Encode(dst, id[:4])
	dst[8] = '-'
	hex.Encode(dst[9:13], id[4:6])
	dst[13] = '-'
	hex.Encode(dst[14:18], id[6:8])
	dst[18] = '-'
	hex.Encode(dst[19:23], id[8:10])
	dst[23] = '-'
	hex.Encode(dst[24:], id[10:])
}

func encodeHexShort(dst []byte, id Identifier) {
	hex.Encode(dst[:], id[12:])
}
