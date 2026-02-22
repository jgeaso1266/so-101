package so_arm

import (
	"testing"
)

func TestEncodeSignMagnitude12(t *testing.T) {
	tests := []struct {
		name     string
		value    int
		expected [2]byte
	}{
		{"zero", 0, [2]byte{0x00, 0x00}},
		{"positive small", 100, [2]byte{0x64, 0x00}},
		{"positive 256", 256, [2]byte{0x00, 0x01}},
		{"positive max (2047)", 2047, [2]byte{0xFF, 0x07}},
		{"negative small", -100, [2]byte{0x64, 0x08}},
		{"negative 256", -256, [2]byte{0x00, 0x09}},
		{"negative max (-2047)", -2047, [2]byte{0xFF, 0x0F}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := encodeSignMagnitude12(tt.value)
			if got[0] != tt.expected[0] || got[1] != tt.expected[1] {
				t.Errorf("encodeSignMagnitude12(%d) = [0x%02X, 0x%02X], want [0x%02X, 0x%02X]",
					tt.value, got[0], got[1], tt.expected[0], tt.expected[1])
			}
		})
	}
}

func TestEncodeSignMagnitude12_RoundTrip(t *testing.T) {
	values := []int{0, 1, 100, 255, 256, 1000, 2047, -1, -100, -255, -256, -1000, -2047}
	for _, v := range values {
		encoded := encodeSignMagnitude12(v)
		raw := uint16(encoded[0]) | (uint16(encoded[1]) << 8)

		// Decode using same logic as readInt16Register in config.go
		signBit := 11
		directionBit := (raw >> uint(signBit)) & 1
		magnitudeMask := uint16((1 << uint(signBit)) - 1)
		magnitude := int(raw & magnitudeMask)

		var decoded int
		if directionBit != 0 {
			decoded = -magnitude
		} else {
			decoded = magnitude
		}

		if decoded != v {
			t.Errorf("Round-trip failed for %d: encoded to [0x%02X,0x%02X], decoded to %d",
				v, encoded[0], encoded[1], decoded)
		}
	}
}
