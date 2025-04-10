//go:build unit

package envarfig

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockEnv struct {
	mock.Mock
}

func (m *MockEnv) Load(filenames ...string) error {
	if len(filenames) == 0 {
		return m.Called().Error(0)
	}
	args := m.Called(filenames)
	return args.Error(0)
}

func TestLoadGoDotEnv(t *testing.T) {
	originalEnvLoader := envLoader // Store the original envLoader
	defer func() {
		envLoader = originalEnvLoader // Restore the original envLoader after the test
	}()
	mockGodotenv := new(MockEnv)

	cleanup := func() {
		mockGodotenv.ExpectedCalls = nil // Reset the expected calls to the mock
	}

	envLoader = func(filenames ...string) error {
		return mockGodotenv.Load(filenames...)
	}

	tests := []struct {
		name        string
		autoLoad    bool
		filePath    []string
		expectError bool
		err         error
	}{
		{"AutoLoad with default env file", true, nil, false, nil},
		{"AutoLoad with custom env file", true, []string{"path/to/envfile"}, false, nil},
		{"No AutoLoad with default env file", false, nil, false, nil},
		{"No AutoLoad with custom env file", false, []string{"path/to/envfile"}, true, errAutoLoadFalseFilePath},
		{"Invalid file path", true, []string{"invalid/path/to/envfile"}, true, errInvalidEnvPathArgs},
		{"Empty file path", true, []string{""}, true, errInvalidEnvPathArgs},
		{"Invalid file path with no AutoLoad", false, []string{"invalid/path/to/envfile"}, true, errAutoLoadFalseFilePath},
		{"Empty file path with no AutoLoad", false, []string{""}, true, errAutoLoadFalseFilePath},
		{"Invalid file path with AutoLoad", true, []string{"invalid/path/to/envfile"}, true, errInvalidEnvPathArgs},
		{"Empty file path with AutoLoad", true, []string{""}, true, errInvalidEnvPathArgs},
		{"multiple file paths", true, []string{"path/to/envfile1", "path/to/envfile2"}, false, nil},
		{"multiple file paths with no AutoLoad", false, []string{"path/to/envfile1", "path/to/envfile2"}, true, errAutoLoadFalseFilePath},
		{"multiple file paths with invalid path", true, []string{"path/to/envfile1", "invalid/path/to/envfile"}, true, errInvalidEnvPathArgs},
		{"multiple file paths with empty path", true, []string{"path/to/envfile1", ""}, true, errInvalidEnvPathArgs},
		{"multiple file paths with empty path and no AutoLoad", false, []string{"path/to/envfile1", ""}, true, errAutoLoadFalseFilePath},
		{"multiple file paths with invalid path and no AutoLoad", false, []string{"path/to/envfile1", "invalid/path/to/envfile"}, true, errAutoLoadFalseFilePath},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Cleanup(cleanup)
			if tt.filePath == nil {
				mockGodotenv.On("Load").Return(tt.err)
			} else {
				mockGodotenv.On("Load", tt.filePath).Return(tt.err)
			}
			err := loadEnvFile(tt.autoLoad, tt.filePath)
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.err, err)
			} else {
				assert.NoError(t, err)
			}
			mockGodotenv.AssertExpectations(t)
		})
	}

}
