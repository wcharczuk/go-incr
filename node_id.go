package incr

import (
	"crypto/rand"
	"encoding/hex"
)

// NodeID is effectively a uuidv4 but we don't advertise as such.
type NodeID [16]byte

// NewNodeID returns a new node id, effectively a uuidv4.
func NewNodeID() (output NodeID) {
	_, _ = rand.Read(output[:])
	output[6] = (output[6] & 0x0f) | 0x40 // Version 4
	output[8] = (output[8] & 0x3f) | 0x80 // Variant is 10
	return
}

// String returns the string form of the
// nodeID as xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx.
func (id NodeID) String() string {
	var buf [36]byte
	encodeHex(buf[:], id)
	return string(buf[:])
}

func encodeHex(dst []byte, id NodeID) {
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
