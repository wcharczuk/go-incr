package incr

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
	"sync"
	"sync/atomic"
)

// Identifier is a unique id.
//
// Create a new identifier with [NewIdentifier].
type Identifier [16]byte

// IdentifierProvider is a type that can provide identifiers.
type IdentifierProvider interface {
	NewIdentifier() Identifier
}

// NewCryptoRandIdentifierProvider a new crypto rand identifier provider.
//
// To use this constructor in practice, pass the [rand.Reader] from crypto/rand as the argument.
func NewCryptoRandIdentifierProvider(randomSource io.Reader) IdentifierProvider {
	return &cryptoRandIdentifierProvider{
		identifierRandPoolPos: cryptoRandPoolSize, // force an initial rand read
		randomSource:          randomSource,
	}
}

// NewSequentialIdentifierProvier returns a new sequential identifier provider.
//
// To use this constructor in practice you can pass a 0 or another known start offset.
func NewSequentialIdentifierProvier(sequenceStart uint64) IdentifierProvider {
	return &sequentialIdentifierProvider{
		seq: sequenceStart,
	}
}

// ParseIdentifier is the reverse of [Identifier.String] and returns
// a parsed identifier from its string representation.
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

// NewIdentifier returns a new random identifier.
//
// This is a convenience method exposed for backwards compatibility reasons.
func NewIdentifier() Identifier {
	return _defaultIdentifierProvider.NewIdentifier()
}

var (
	_defaultIdentifierProvider = NewCryptoRandIdentifierProvider(rand.Reader)
	_zero                      Identifier
)

// IsZero returns if the identifier is unset.
func (id Identifier) IsZero() bool {
	return id == _zero
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

type cryptoRandIdentifierProvider struct {
	identifierRandPoolMu  sync.Mutex
	identifierRandPoolPos int
	identifierRandPool    [cryptoRandPoolSize]byte // protected with poolMu
	randomSource          io.Reader                // random function
}

func (cr *cryptoRandIdentifierProvider) NewIdentifier() (output Identifier) {
	cr.identifierRandPoolMu.Lock()
	if cr.identifierRandPoolPos == cryptoRandPoolSize {
		_, _ = io.ReadFull(cr.randomSource, cr.identifierRandPool[:])
		cr.identifierRandPoolPos = 0
	}
	copy(output[:], cr.identifierRandPool[cr.identifierRandPoolPos:(cr.identifierRandPoolPos+16)])
	cr.identifierRandPoolPos += 16
	cr.identifierRandPoolMu.Unlock()
	output[6] = (output[6] & 0x0f) | 0x40 // Version 4
	output[8] = (output[8] & 0x3f) | 0x80 // Variant is 10
	return
}

const cryptoRandPoolSize = 16 * 16

type sequentialIdentifierProvider struct {
	seq uint64
}

func (s *sequentialIdentifierProvider) NewIdentifier() (output Identifier) {
	newCounter := atomic.AddUint64(&s.seq, 1)
	output[15] = byte(newCounter)
	output[14] = byte(newCounter >> 8)
	output[13] = byte(newCounter >> 16)
	output[12] = byte(newCounter >> 24)
	output[11] = byte(newCounter >> 32)
	output[10] = byte(newCounter >> 40)
	output[9] = byte(newCounter >> 48)
	output[8] = byte(newCounter >> 56)
	return
}
