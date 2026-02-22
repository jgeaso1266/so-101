// discovery_test.go
package so_arm

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
)

func TestFilterCandidatePorts(t *testing.T) {
	tests := []struct {
		name     string
		ports    []string
		expected []string
	}{
		{
			name:     "Linux USB ports",
			ports:    []string{"/dev/ttyUSB0", "/dev/ttyS0", "/dev/ttyACM0", "/dev/null"},
			expected: []string{"/dev/ttyUSB0", "/dev/ttyACM0"},
		},
		{
			name:     "macOS USB ports",
			ports:    []string{"/dev/tty.usbmodem123", "/dev/tty.Bluetooth", "/dev/tty.usbserial-AB", "/dev/cu.usbmodem5A7A0594021", "/dev/cu.usbserial-123"},
			expected: []string{"/dev/tty.usbmodem123", "/dev/tty.usbserial-AB", "/dev/cu.usbmodem5A7A0594021", "/dev/cu.usbserial-123"},
		},
		{
			name:     "Windows COM ports",
			ports:    []string{"COM3", "COM10", "LPT1", "PRN"},
			expected: []string{"COM3", "COM10"},
		},
		{
			name:     "Empty list",
			ports:    []string{},
			expected: []string{},
		},
		{
			name:     "No matching ports",
			ports:    []string{"/dev/null", "/dev/zero"},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterCandidatePorts(tt.ports)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractPortSuffix(t *testing.T) {
	tests := []struct {
		name     string
		portPath string
		expected string
	}{
		{
			name:     "Linux ttyUSB",
			portPath: "/dev/ttyUSB0",
			expected: "ttyUSB0",
		},
		{
			name:     "Linux ttyACM",
			portPath: "/dev/ttyACM1",
			expected: "ttyACM1",
		},
		{
			name:     "macOS usbmodem",
			portPath: "/dev/tty.usbmodem14201",
			expected: "usbmodem14201",
		},
		{
			name:     "macOS usbserial",
			portPath: "/dev/tty.usbserial-AB123",
			expected: "usbserial-AB123",
		},
		{
			name:     "Windows COM port",
			portPath: "COM3",
			expected: "COM3",
		},
		{
			name:     "Windows double digit COM",
			portPath: "COM10",
			expected: "COM10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractPortSuffix(tt.portPath)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFindCalibrationFile(t *testing.T) {
	// Create temp directory for test files
	tempDir := t.TempDir()

	tests := []struct {
		name       string
		setup      func(string) // Setup function to create test files
		portSuffix string
		expected   string
	}{
		{
			name: "Port-specific file exists",
			setup: func(dir string) {
				os.WriteFile(filepath.Join(dir, "ttyUSB0_calibration.json"), []byte("{}"), 0644)
				os.WriteFile(filepath.Join(dir, "so101_calibration.json"), []byte("{}"), 0644)
			},
			portSuffix: "ttyUSB0",
			expected:   "ttyUSB0_calibration.json",
		},
		{
			name: "Only default file exists",
			setup: func(dir string) {
				os.WriteFile(filepath.Join(dir, "so101_calibration.json"), []byte("{}"), 0644)
			},
			portSuffix: "ttyUSB0",
			expected:   "so101_calibration.json",
		},
		{
			name: "No calibration files",
			setup: func(dir string) {
				// No files created
			},
			portSuffix: "ttyUSB0",
			expected:   "",
		},
	}

	logger := logging.NewTestLogger(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDir := filepath.Join(tempDir, tt.name)
			os.MkdirAll(testDir, 0755)
			tt.setup(testDir)

			result := findCalibrationFile(testDir, tt.portSuffix, logger)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEnumerateSerialPorts(t *testing.T) {
	// This is a system-dependent test - just verify it doesn't panic and returns a slice
	ports := enumerateSerialPorts()
	assert.NotNil(t, ports)
	// Ports list can be empty on systems without serial devices
	t.Logf("Found %d serial ports", len(ports))
}

func TestGenerateConfigs_CalibrationSensorIncludesFile(t *testing.T) {
	logger := logging.NewTestLogger(t)
	dis := &so101Discovery{logger: logger}

	// With calibration file: sensor config must include calibration_file attribute
	configs := dis.generateConfigs("/dev/ttyUSB0", "ttyUSB0", true, true, "so101_calibration.json")

	var sensorConfig *resource.Config
	for i := range configs {
		if configs[i].Model == SO101CalibrationSensorModel {
			sensorConfig = &configs[i]
			break
		}
	}
	if sensorConfig == nil {
		t.Fatal("no calibration sensor config generated")
	}
	calFile, exists := sensorConfig.Attributes["calibration_file"]
	if !exists {
		t.Error("calibration sensor config missing calibration_file attribute")
	}
	if calFile != "so101_calibration.json" {
		t.Errorf("calibration_file = %v, want so101_calibration.json", calFile)
	}

	// Without calibration file: calibration_file key must be absent
	configs2 := dis.generateConfigs("/dev/ttyUSB0", "ttyUSB0", true, false, "")
	for i := range configs2 {
		if configs2[i].Model == SO101CalibrationSensorModel {
			if _, exists := configs2[i].Attributes["calibration_file"]; exists {
				t.Error("calibration_file should be absent when no calibration file provided")
			}
			return
		}
	}
	t.Fatal("no calibration sensor config generated for second case")
}
