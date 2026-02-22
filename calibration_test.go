package so_arm

import (
	"math"
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

func TestGetCurrentPositionsNilCalibrationFallback(t *testing.T) {
	// When calibration is nil, getCurrentPositions uses the fallback formula:
	//   rawPos = int((positions[i]/(2*math.Pi)+0.5)*4095)
	// This test verifies the round-trip math directly.

	fallback := func(pos float64) int {
		return int((pos/(2*math.Pi)+0.5)*4095)
	}

	// At 0 radians, the result should be the midpoint: 2047
	if got := fallback(0); got != 2047 {
		t.Errorf("fallback(0) = %d, want 2047", got)
	}

	// At π radians, the result should be the maximum: 4095
	if got := fallback(math.Pi); got != 4095 {
		t.Errorf("fallback(π) = %d, want 4095", got)
	}
}

func TestGetCurrentPositionsRawConversion(t *testing.T) {
	cal := &MotorCalibration{
		ID: 1, DriveMode: 0, HomingOffset: 0,
		RangeMin: 500, RangeMax: 3500,
		NormMode: NormModeDegrees,
	}

	// Simulate: raw=3500 -> Normalize -> 180° -> DegToRad -> π rad
	normalized, _ := cal.Normalize(3500)       // should be 180.0
	radiansVal := normalized * math.Pi / 180.0 // π

	// The correct inverse: radians -> degrees -> Denormalize -> raw
	degreesVal := radiansVal * 180.0 / math.Pi // 180.0
	raw, err := cal.Denormalize(degreesVal)
	if err != nil {
		t.Fatal(err)
	}
	if absInt(raw-3500) > 1 {
		t.Errorf("round-trip: got raw=%d, want 3500", raw)
	}
}
