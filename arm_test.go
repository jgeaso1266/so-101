package so_arm

import (
	"math"
	"testing"
)

func TestCalculateJointLimits_MatchesNormalization(t *testing.T) {
	cal := &MotorCalibration{
		ID: 1, DriveMode: 0, HomingOffset: 0,
		RangeMin: 500, RangeMax: 3500,
		NormMode: NormModeDegrees,
	}
	fullCal := SO101FullCalibration{
		ShoulderPan:  cal,
		ShoulderLift: cal,
		ElbowFlex:    cal,
		WristFlex:    cal,
		WristRoll:    cal,
		Gripper:      DefaultSO101FullCalibration.Gripper,
	}

	arm := &so101{
		armServoIDs: []int{1, 2, 3, 4, 5},
		controller: &SafeSoArmController{
			calibration: fullCal,
		},
	}

	limits := arm.calculateJointLimits()

	if len(limits) != 5 {
		t.Fatalf("expected 5 limits, got %d", len(limits))
	}

	// With range 500-3500: Normalize(500)=-180°, Normalize(3500)=+180°
	// Converted to radians: ±π
	expectedMin := -math.Pi
	expectedMax := math.Pi

	for i, lim := range limits {
		if math.Abs(lim[0]-expectedMin) > 0.001 {
			t.Errorf("joint %d min limit = %.4f rad, want %.4f rad", i, lim[0], expectedMin)
		}
		if math.Abs(lim[1]-expectedMax) > 0.001 {
			t.Errorf("joint %d max limit = %.4f rad, want %.4f rad", i, lim[1], expectedMax)
		}
	}
}

func TestCalculateJointLimits_DriveMode(t *testing.T) {
	cal := &MotorCalibration{
		ID: 1, DriveMode: 1, HomingOffset: 0, // drive mode inverts sign
		RangeMin: 500, RangeMax: 3500,
		NormMode: NormModeDegrees,
	}
	fullCal := SO101FullCalibration{
		ShoulderPan: cal, ShoulderLift: cal, ElbowFlex: cal,
		WristFlex: cal, WristRoll: cal,
		Gripper: DefaultSO101FullCalibration.Gripper,
	}

	arm := &so101{
		armServoIDs: []int{1, 2, 3, 4, 5},
		controller:  &SafeSoArmController{calibration: fullCal},
	}

	limits := arm.calculateJointLimits()

	for i, lim := range limits {
		if lim[0] > lim[1] {
			t.Errorf("joint %d: inverted limits [%.4f, %.4f] — min must be <= max", i, lim[0], lim[1])
		}
		// With drive mode, limits are still ±π (just swapped then re-sorted)
		if math.Abs(lim[0]-(-math.Pi)) > 0.001 {
			t.Errorf("joint %d min = %.4f, want -π", i, lim[0])
		}
		if math.Abs(lim[1]-math.Pi) > 0.001 {
			t.Errorf("joint %d max = %.4f, want +π", i, lim[1])
		}
	}
}

func TestCalculateJointLimits_AsymmetricCalibration(t *testing.T) {
	cal := &MotorCalibration{
		ID: 1, DriveMode: 0, HomingOffset: 0,
		RangeMin: 800, RangeMax: 3200,
		NormMode: NormModeDegrees,
	}
	fullCal := SO101FullCalibration{
		ShoulderPan: cal, ShoulderLift: cal, ElbowFlex: cal,
		WristFlex: cal, WristRoll: cal,
		Gripper: DefaultSO101FullCalibration.Gripper,
	}

	arm := &so101{
		armServoIDs: []int{1, 2, 3, 4, 5},
		controller:  &SafeSoArmController{calibration: fullCal},
	}

	limits := arm.calculateJointLimits()

	// Normalize(800)=-180°, Normalize(3200)=+180° → ±π rad
	for i, lim := range limits {
		if math.Abs(lim[0]-(-math.Pi)) > 0.001 {
			t.Errorf("joint %d min = %.4f, want %.4f", i, lim[0], -math.Pi)
		}
		if math.Abs(lim[1]-math.Pi) > 0.001 {
			t.Errorf("joint %d max = %.4f, want %.4f", i, lim[1], math.Pi)
		}
	}
}
