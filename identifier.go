package incr

import (
	"crypto/rand"
	"encoding/hex"
)

type identifier [16]byte

func newIdentifier() (output identifier) {
	_, _ = rand.Read(output[:])
	output[6] = (output[6] & 0x0f) | 0x40 // Version 4
	output[8] = (output[8] & 0x3f) | 0x80 // Variant is 10
	return
}

func (id identifier) String() string {
	var buf [36]byte
	encodeHex(buf[:], id)
	return string(buf[:])
}

func (id identifier) Short() string {
	var buf [8]byte
	encodeHexShort(buf[:], id)
	return string(buf[:])
}

func encodeHex(dst []byte, id identifier) {
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

func encodeHexShort(dst []byte, id identifier) {
	hex.Encode(dst[:], id[12:])
}
