package so_arm

import (
	"math"
	"testing"
)

func TestNormModeDegrees_Normalize(t *testing.T) {
	cal := &MotorCalibration{
		ID: 1, DriveMode: 0, HomingOffset: 0,
		RangeMin: 500, RangeMax: 3500,
		NormMode: NormModeDegrees,
	}

	tests := []struct {
		name     string
		raw      int
		expected float64
	}{
		{"center maps to 0 degrees", 2000, 0.0},
		{"max maps to +180 degrees", 3500, 180.0},
		{"min maps to -180 degrees", 500, -180.0},
		{"midpoint maps to +90 degrees", 2750, 90.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := cal.Normalize(tt.raw)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if math.Abs(got-tt.expected) > 0.01 {
				t.Errorf("Normalize(%d) = %.4f, want %.4f", tt.raw, got, tt.expected)
			}
		})
	}
}

func TestNormModeDegrees_Denormalize(t *testing.T) {
	cal := &MotorCalibration{
		ID: 1, DriveMode: 0, HomingOffset: 0,
		RangeMin: 500, RangeMax: 3500,
		NormMode: NormModeDegrees,
	}

	tests := []struct {
		name     string
		degrees  float64
		expected int
	}{
		{"0 degrees maps to center (2000)", 0.0, 2000},
		{"+180 degrees maps to max (3500)", 180.0, 3500},
		{"-180 degrees maps to min (500)", -180.0, 500},
		{"+90 degrees maps to 2750", 90.0, 2750},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := cal.Denormalize(tt.degrees)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if absInt(got-tt.expected) > 1 {
				t.Errorf("Denormalize(%.1f) = %d, want %d", tt.degrees, got, tt.expected)
			}
		})
	}
}

func TestNormModeDegrees_RoundTrip(t *testing.T) {
	cal := &MotorCalibration{
		ID: 1, DriveMode: 0, HomingOffset: 0,
		RangeMin: 500, RangeMax: 3500,
		NormMode: NormModeDegrees,
	}

	for raw := 500; raw <= 3500; raw += 100 {
		normalized, err := cal.Normalize(raw)
		if err != nil {
			t.Fatalf("Normalize(%d): %v", raw, err)
		}
		recovered, err := cal.Denormalize(normalized)
		if err != nil {
			t.Fatalf("Denormalize(%.4f): %v", normalized, err)
		}
		if absInt(recovered-raw) > 1 {
			t.Errorf("Round-trip failed for raw=%d: got %d (via %.4f degrees)", raw, recovered, normalized)
		}
	}
}

func TestNormModeDegrees_DriveMode(t *testing.T) {
	cal := &MotorCalibration{
		ID: 1, DriveMode: 1, HomingOffset: 0,
		RangeMin: 500, RangeMax: 3500,
		NormMode: NormModeDegrees,
	}

	center, err := cal.Normalize(2000)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(center) > 0.01 {
		t.Errorf("Center should still be 0 degrees with drive mode, got %.4f", center)
	}

	atMax, _ := cal.Normalize(3500)
	if math.Abs(atMax-(-180.0)) > 0.01 {
		t.Errorf("Max raw should be -180 degrees with drive mode, got %.4f", atMax)
	}
}

func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
