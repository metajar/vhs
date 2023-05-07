package git

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"vhs/devices"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSaveDeviceConfiguration(t *testing.T) {
	testCases := []struct {
		name       string
		device     devices.Device
		expectFile bool
	}{
		{
			name: "Save valid device configuration",
			device: devices.NewDevice(
				"test-device",
				[]byte("This is a test device configuration."),
			),
			expectFile: true,
		},
		{
			name: "Save empty device configuration",
			device: devices.NewDevice(
				"empty-device",
				[]byte(""),
			),
			expectFile: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create temporary directory for the test
			tempDir, err := ioutil.TempDir("", "vhs-test")
			require.NoError(t, err)
			defer os.RemoveAll(tempDir)

			g := NewGit(tempDir, "main")
			err = g.SaveDeviceConfiguration(tc.device)
			assert.NoError(t, err)

			deviceFilePath := filepath.Join(tempDir, tc.device.GetDeviceType(), tc.device.Name)
			fileExists := fileExists(deviceFilePath)

			if tc.expectFile {
				assert.True(t, fileExists, "Expected device configuration file to be saved")
				content, err := ioutil.ReadFile(deviceFilePath)
				assert.NoError(t, err, "Error reading device configuration file")
				assert.True(t, strings.Contains(string(content), string(tc.device.Payload)), "Expected device configuration content to be found in the saved file")
			} else {
				assert.False(t, fileExists, "Expected device configuration file not to be saved")
			}
		})
	}
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
