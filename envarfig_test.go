//go:build integration

package envarfig

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockGodotenv struct {
	mock.Mock
}

func (m *MockGodotenv) Load(filenames ...string) error {
	if len(filenames) == 0 {
		return m.Called().Error(0)
	}
	args := m.Called(filenames)
	return args.Error(0)
}

func TestGetEnvVar(t *testing.T) {
	// Test with a valid struct and env variables
	originalEnvLoader := envLoader // Store the original envLoader
	defer func() {
		envLoader = originalEnvLoader // Restore the original envLoader after the test
	}()
	type Config struct {
		Host string `env:"HOST"`
		Port int    `env:"PORT"`
	}

	mockGodotenv := new(MockGodotenv)
	setup := func() {
		t.Setenv("HOST", "localhost")
		t.Setenv("PORT", "8080")
		mockGodotenv.On("Load", mock.AnythingOfType("[]string")).Return(nil)
		mockGodotenv.On("Load", mock.Anything).Return(nil)
		mockGodotenv.On("Load").Return(nil)
	}
	envLoader = func(filenames ...string) error {
		return mockGodotenv.Load(filenames...)
	}

	resetCache := func() {
		once = sync.Once{} // Reset the once variable to allow re-execution of the test
		cachedConfig = nil // Reset the cached config to allow re-execution of the test
		t.Setenv("HOST", "")
		t.Setenv("PORT", "")
		mockGodotenv.ExpectedCalls = nil // Reset the expected calls to the mock
	}

	// Set environment variables for testing

	t.Run("Test with valid struct", func(t *testing.T) {
		setup()
		t.Cleanup(resetCache)

		var config Config
		err := GetEnvVar(&config)
		assert.NoError(t, err)
		assert.Equal(t, "localhost", config.Host)
		assert.Equal(t, 8080, config.Port)
	})

	t.Run("Test with no struct tags or invalid tag or empty tag or empty tagname", func(t *testing.T) {
		setup()
		t.Cleanup(resetCache)
		type NoTagConfig struct {
			Host string
		}
		type InvalidTagNaeConfig struct {
			Host string `env1:""`
		}
		type EmptyTagConfig struct {
			Host string `env:""`
		}
		t.Setenv("HOST", "localhost")
		var configNotTag NoTagConfig
		err := GetEnvVar(&configNotTag)
		assert.Error(t, err)
		assert.ErrorIs(t, err, errTagNotFound)
		resetCache()
		setup()
		var configInvalidTagName InvalidTagNaeConfig
		err = GetEnvVar(&configInvalidTagName)
		assert.Error(t, err)
		assert.ErrorIs(t, err, errTagNotFound)
		resetCache()
		setup()
		var configEmptyTag EmptyTagConfig
		err = GetEnvVar(&configEmptyTag)
		assert.Error(t, err)
		assert.ErrorIs(t, err, errTagNotFound)
	})

	t.Run("Test with nil config", func(t *testing.T) {
		setup()
		t.Cleanup(resetCache)
		var nilConfig *Config
		err := GetEnvVar(nilConfig)
		assert.Error(t, err)
		assert.Equal(t, errNilConfig, err)
	})
	t.Run("Test config not struct", func(t *testing.T) {
		setup()
		t.Cleanup(resetCache)
		var invalidConfig *int
		err := GetEnvVar(&invalidConfig)
		assert.Error(t, err)
		assert.ErrorIs(t, err, errConfigNotPtrToStruct)
	})

	t.Run("Test with invalid env variable with incorrect datatype", func(t *testing.T) {
		setup()
		t.Cleanup(resetCache)
		// Set an invalid environment variable
		t.Setenv("PORT", "invalid")
		var newConf Config

		err := GetEnvVar(&newConf)
		assert.Error(t, err)
		assert.Equal(t, "failed to convert PORT to int: strconv.Atoi: parsing \"invalid\": invalid syntax", err.Error())
	})

	t.Run("Test without godotenv", func(t *testing.T) {
		setup()
		t.Cleanup(resetCache)
		// Set environment variables for testing

		var config Config
		err := GetEnvVar(
			&config,
			WithAutoLoadEnv(false),
		)
		assert.NoError(t, err)
		assert.Equal(t, "localhost", config.Host)
		assert.Equal(t, 8080, config.Port)
	})
	t.Run("Test with single env file", func(t *testing.T) {
		setup()
		t.Cleanup(resetCache)

		var config Config
		err := GetEnvVar(
			&config,
			WithEnvFiles("example.env"),
		)
		assert.NoError(t, err)
		assert.Equal(t, "localhost", config.Host)
		assert.Equal(t, 8080, config.Port)
		mockGodotenv.AssertExpectations(t)

	})
	t.Run("Test with multiple env file", func(t *testing.T) {
		setup()
		t.Cleanup(resetCache)

		var config Config
		err := GetEnvVar(
			&config,
			WithEnvFiles("example.env", "example2.env"),
		)
		assert.NoError(t, err)
		assert.Equal(t, "localhost", config.Host)
		assert.Equal(t, 8080, config.Port)
		mockGodotenv.AssertExpectations(t)
	})
	t.Run("Test invalid env file load", func(t *testing.T) {
		setup()
		t.Cleanup(resetCache)
		mockGodotenv.ExpectedCalls = nil
		mockGodotenv.On("Load", mock.AnythingOfType("[]string")).Return(errInvalidEnvPathArgs)
		var config Config
		err := GetEnvVar(
			&config,
			WithEnvFiles("invalid.env"),
		)
		assert.Error(t, err)
		assert.ErrorIs(t, err, errInvalidEnvPathArgs)
		mockGodotenv.AssertExpectations(t)

	})
	t.Run("Test different supported datatypes", func(t *testing.T) {
		setup()
		t.Cleanup(resetCache)
		type DataTypesConfig struct {
			Intval    int    `env:"INTVAL"`
			Boolval   bool   `env:"BOOLVAL"`
			Stringval string `env:"STRINGVAL"`
		}
		t.Setenv("INTVAL", "24")
		t.Setenv("BOOLVAL", "TRUE")
		t.Setenv("STRINGVAL", "HELLO")
		var dataTypesConfig DataTypesConfig
		err := GetEnvVar(&dataTypesConfig)
		assert.NoError(t, err)
		assert.Equal(t, 24, dataTypesConfig.Intval)
		assert.Equal(t, true, dataTypesConfig.Boolval)
		assert.Equal(t, "HELLO", dataTypesConfig.Stringval)
		mockGodotenv.AssertExpectations(t)
	})
	t.Run("Test different supported datatypes for errors", func(t *testing.T) {
		setup()
		t.Cleanup(resetCache)
		type DataTypesConfig struct {
			Boolval bool `env:"BOOLVAL"`
		}
		invalidBool := "TRUEasdf"
		t.Setenv("BOOLVAL", invalidBool)
		var dataTypesConfig DataTypesConfig
		err := GetEnvVar(&dataTypesConfig)
		assert.Error(t, err)
		assert.Equal(t, fmt.Sprintf("error parsing env var BOOLVAL: strconv.ParseBool: parsing \"%s\": invalid syntax", invalidBool), err.Error())
		mockGodotenv.AssertExpectations(t)
	})
}
