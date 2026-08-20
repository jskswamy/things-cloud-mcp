package thingscloud

import (
	"crypto/sha1"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestEncodeLegacyIdentifier(t *testing.T) {
	t.Parallel()
	tests := []struct {
		id   string
		want string
	}{
		{"AAAAAAAA-BBBB-CCCC-DDDD-EEEEEEEEEEEE", "LXmxn9gakySzcEjKj1DtgD"},
		{"AAAAAAAA-BBBB-CCCC-DDDD-EEEEEEEEEEEE-20240131", "H8Xu72gj7fooPuYoBMZ5TK"},
		{"00000006-1111-2222-3333-000000000006", "fxsSvCT97pJn3XZ4wp5t4"},
		{"000000B4-1111-2222-3333-0000000000B4", "14q4keicwiREVK8EKAuowZ"},
		{"", ""},
	}
	for _, test := range tests {
		if got := EncodeLegacyIdentifier(test.id); got != test.want {
			t.Errorf("EncodeLegacyIdentifier(%q) = %q, want %q", test.id, got, test.want)
		}
	}
}

func TestEncodeLegacyIdentifierDoesNotNormalizeCase(t *testing.T) {
	t.Parallel()
	const legacyID = "AAAAAAAA-BBBB-CCCC-DDDD-EEEEEEEEEEEE"
	if EncodeLegacyIdentifier(strings.ToLower(legacyID)) == EncodeLegacyIdentifier(legacyID) {
		t.Fatal("legacy identifier case was normalized")
	}
}

func TestEncodeLegacyIdentifierMalformedRecurrenceUsesOrdinaryHash(t *testing.T) {
	t.Parallel()
	for _, id := range []string{
		"AAAAAAAA-BBBB-CCCC-DDDD-EEEEEEEEEEEE-2024013X",
		"AAAAAAAA-BBBB-CCCC-DDDD-EEEEEEEEEEEE-2024013",
		"GGGGGGGG-BBBB-CCCC-DDDD-EEEEEEEEEEEE-20240131",
	} {
		sum := sha1.Sum([]byte(id))
		var u uuid.UUID
		copy(u[:], sum[:len(u)])
		if got, want := EncodeLegacyIdentifier(id), EncodeUUID(u); got != want {
			t.Errorf("EncodeLegacyIdentifier(%q) = %q, want %q", id, got, want)
		}
	}
}

func TestEncodeDecodeUUIDRoundTrip(t *testing.T) {
	t.Parallel()
	cases := []uuid.UUID{
		{},
		{0x00, 0x7f, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee},
		{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
	}
	for i := 0; i < 64; i++ {
		cases = append(cases, uuid.New())
	}
	for _, u := range cases {
		encoded := EncodeUUID(u)
		decoded, err := DecodeUUID(encoded)
		if err != nil {
			t.Fatalf("DecodeUUID(%q): %v", encoded, err)
		}
		if decoded != u {
			t.Errorf("round trip: %v -> %q -> %v", u, encoded, decoded)
		}
	}
}

func TestValidateUUID(t *testing.T) {
	t.Parallel()
	for _, valid := range []string{"VJ1edXTP9q3PmFDUuy8EQh", "1111111111111111"} {
		if err := ValidateUUID(valid); err != nil {
			t.Errorf("ValidateUUID(%q): %v", valid, err)
		}
	}
	for _, invalid := range []string{"", "6f9b2c1e-8a4d-4e5f-9c3b-2a1d0e9f8b7c", "VJ0edXTP9q3PmFDUuy8EQh", "zzzzzzzzzzzzzzzzzzzzzz"} {
		if err := ValidateUUID(invalid); err == nil {
			t.Errorf("ValidateUUID(%q) unexpectedly succeeded", invalid)
		}
	}
}

func TestNewUUIDAlwaysCanonical(t *testing.T) {
	t.Parallel()
	for i := 0; i < 1000; i++ {
		if id := NewUUID(); ValidateUUID(id) != nil {
			t.Fatalf("NewUUID() = %q is not canonical", id)
		}
	}
}
