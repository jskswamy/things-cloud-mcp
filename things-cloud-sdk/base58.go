package thingscloud

import (
	"crypto/sha1"
	"fmt"
	"math/big"
	"strings"

	"github.com/google/uuid"
)

// base58Alphabet is the Bitcoin Base58 alphabet used by Things identifiers.
const base58Alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

// EncodeUUID encodes a UUID as canonical Base58, preserving leading zero bytes.
func EncodeUUID(u uuid.UUID) string {
	zeros := 0
	for zeros < len(u) && u[zeros] == 0 {
		zeros++
	}

	n := new(big.Int).SetBytes(u[:])
	base := big.NewInt(58)
	mod := new(big.Int)
	var encoded []byte
	for n.Sign() > 0 {
		n.DivMod(n, base, mod)
		encoded = append(encoded, base58Alphabet[mod.Int64()])
	}
	for i := 0; i < zeros; i++ {
		encoded = append(encoded, '1')
	}
	for i, j := 0, len(encoded)-1; i < j; i, j = i+1, j-1 {
		encoded[i], encoded[j] = encoded[j], encoded[i]
	}
	return string(encoded)
}

// EncodeLegacyIdentifier derives the current Base58 identifier for an object
// created before Things Cloud migrated away from UUID-style identifiers. The
// input is hashed byte-for-byte because case normalization changes the result.
// Recurring instances use the historical two-stage <uuid>-YYYYMMDD derivation.
func EncodeLegacyIdentifier(legacyID string) string {
	if legacyID == "" {
		return ""
	}
	input := []byte(legacyID)
	if prefix, suffix, ok := splitLegacyRecurrenceIdentifier(legacyID); ok {
		prefixSum := sha1.Sum([]byte(prefix))
		input = make([]byte, 0, 16+len(suffix))
		input = append(input, prefixSum[:16]...)
		input = append(input, suffix...)
	}
	sum := sha1.Sum(input)
	var u uuid.UUID
	copy(u[:], sum[:len(u)])
	return EncodeUUID(u)
}

func splitLegacyRecurrenceIdentifier(id string) (string, []byte, bool) {
	const uuidLength = 36
	const dateSuffixLength = len("-YYYYMMDD")
	if len(id) != uuidLength+dateSuffixLength {
		return "", nil, false
	}
	prefix := id[:uuidLength]
	if _, err := uuid.Parse(prefix); err != nil || id[uuidLength] != '-' {
		return "", nil, false
	}
	for i := uuidLength + 1; i < len(id); i++ {
		if id[i] < '0' || id[i] > '9' {
			return "", nil, false
		}
	}
	return prefix, []byte(id[uuidLength:]), true
}

// DecodeUUID decodes a canonical Base58 Things identifier into a UUID.
func DecodeUUID(s string) (uuid.UUID, error) {
	var u uuid.UUID
	if s == "" {
		return u, fmt.Errorf("thingscloud: empty identifier")
	}

	zeros := 0
	for zeros < len(s) && s[zeros] == '1' {
		zeros++
	}

	n := new(big.Int)
	base := big.NewInt(58)
	for i := zeros; i < len(s); i++ {
		idx := strings.IndexByte(base58Alphabet, s[i])
		if idx < 0 {
			return u, fmt.Errorf("thingscloud: invalid Base58 character %q in identifier %q", s[i], s)
		}
		n.Mul(n, base)
		n.Add(n, big.NewInt(int64(idx)))
	}

	value := n.Bytes()
	if zeros+len(value) != len(u) {
		return u, fmt.Errorf("thingscloud: identifier %q decodes to %d bytes, want 16", s, zeros+len(value))
	}
	copy(u[len(u)-len(value):], value)
	return u, nil
}

// ValidateUUID checks that s is a canonical 16-byte Base58 identifier.
func ValidateUUID(s string) error {
	_, err := DecodeUUID(s)
	return err
}

// NewUUID returns a new canonical Things identifier.
func NewUUID() string {
	return EncodeUUID(uuid.New())
}
