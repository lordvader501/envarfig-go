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

func TestLoadEnv(t *testing.T) {
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
		// once = sync.Once{}               // Reset the once variable to allow re-execution of the test
		cachedConfigs = sync.Map{}       // Reset the cached config to allow re-execution of the test
		mockGodotenv.ExpectedCalls = nil // Reset the expected calls to the mock
	}

	// Set environment variables for testing

	t.Run("Test with valid struct", func(t *testing.T) {
		setup()
		t.Cleanup(resetCache)

		var config Config
		err := LoadEnv(&config)
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
		err := LoadEnv(&configNotTag)
		assert.Error(t, err)
		assert.ErrorIs(t, err, errTagNotFound)
		resetCache()
		setup()
		var configInvalidTagName InvalidTagNaeConfig
		err = LoadEnv(&configInvalidTagName)
		assert.Error(t, err)
		assert.ErrorIs(t, err, errTagNotFound)
		resetCache()
		setup()
		var configEmptyTag EmptyTagConfig
		err = LoadEnv(&configEmptyTag)
		assert.Error(t, err)
		assert.ErrorIs(t, err, errTagNotFound)
	})

	t.Run("Test with nil config", func(t *testing.T) {
		setup()
		t.Cleanup(resetCache)
		var nilConfig *Config
		err := LoadEnv(nilConfig)
		assert.Error(t, err)
		assert.Equal(t, errNilConfig, err)
	})
	t.Run("Test config not struct", func(t *testing.T) {
		setup()
		t.Cleanup(resetCache)
		var invalidConfig *int
		err := LoadEnv(&invalidConfig)
		assert.Error(t, err)
		assert.ErrorIs(t, err, errConfigNotPtrToStruct)
	})
	t.Run("Test with multiple configs of different variables", func(t *testing.T) {
		setup()
		t.Cleanup(resetCache)
		type Config1 struct {
			Host string `env:"HOST"`
			Port int    `env:"PORT"`
		}
		type Config2 struct {
			Host string `env:"HOST"`
			Port int    `env:"PORT"`
		}
		var config1 Config1
		var config2 Config2
		t.Setenv("HOST", "localhost")
		t.Setenv("PORT", "8080")
		err1 := LoadEnv(&config1)
		t.Setenv("PORT", "8081")
		err2 := LoadEnv(&config2)
		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.Equal(t, "localhost", config1.Host)
		assert.Equal(t, 8080, config1.Port)
		assert.Equal(t, "localhost", config2.Host)
		assert.Equal(t, 8081, config2.Port)
	})
	t.Run("Test with multiple configs of same variable", func(t *testing.T) {
		setup()
		t.Cleanup(resetCache)
		type Config struct {
			Host string `env:"HOST"`
			Port int    `env:"PORT"`
		}
		var config1 Config
		var config2 Config
		t.Setenv("HOST", "localhost")
		t.Setenv("PORT", "8080")
		err1 := LoadEnv(&config1)
		t.Setenv("PORT", "8081")
		err2 := LoadEnv(&config2)
		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.Equal(t, "localhost", config1.Host)
		assert.Equal(t, 8080, config1.Port)
		assert.Equal(t, "localhost", config2.Host)
		assert.Equal(t, 8080, config2.Port)
	})

	t.Run("Test with cacheing off", func(t *testing.T) {
		setup()
		t.Cleanup(resetCache)
		type Config struct {
			Host string `env:"HOST"`
			Port int    `env:"PORT"`
		}
		var config1 Config
		var config2 Config
		t.Setenv("HOST", "localhost")
		t.Setenv("PORT", "8080")
		err1 := LoadEnv(&config1, WithCacheConfig(false))
		t.Setenv("PORT", "8081")
		err2 := LoadEnv(&config2, WithCacheConfig(false))
		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.Equal(t, "localhost", config1.Host)
		assert.Equal(t, 8080, config1.Port)
		assert.Equal(t, "localhost", config2.Host)
		assert.Equal(t, 8081, config2.Port)
	})

	t.Run("Test with invalid env variable with incorrect datatype", func(t *testing.T) {
		setup()
		t.Cleanup(resetCache)
		// Set an invalid environment variable
		t.Setenv("PORT", "invalid")
		var newConf Config

		err := LoadEnv(&newConf)
		assert.Error(t, err)
		assert.Equal(t, "failed to convert PORT to int: strconv.ParseInt: parsing \"invalid\": invalid syntax", err.Error())
	})

	t.Run("Test without godotenv", func(t *testing.T) {
		setup()
		t.Cleanup(resetCache)
		// Set environment variables for testing

		var config Config
		err := LoadEnv(
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
		err := LoadEnv(
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
		err := LoadEnv(
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
		err := LoadEnv(
			&config,
			WithEnvFiles("invalid.env"),
		)
		assert.Error(t, err)
		assert.ErrorIs(t, err, errInvalidEnvPathArgs)
		mockGodotenv.AssertExpectations(t)

	})
	// testing for different supported datatypes
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
		err := LoadEnv(&dataTypesConfig)
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
		err := LoadEnv(&dataTypesConfig)
		assert.Error(t, err)
		assert.Equal(t, fmt.Sprintf("error parsing env var BOOLVAL: strconv.ParseBool: parsing \"%s\": invalid syntax", invalidBool), err.Error())
		mockGodotenv.AssertExpectations(t)
	})
	t.Run("Test unint datatypes", func(t *testing.T) {
		setup()
		t.Cleanup(resetCache)
		type DataTypesConfig struct {
			Uintval  uint  `env:"UINTVAL"`
			Uint8val uint8 `env:"UINT8VAL"`
		}
		t.Setenv("UINTVAL", "24")
		t.Setenv("UINT8VAL", "24")
		var dataTypesConfig DataTypesConfig
		err := LoadEnv(&dataTypesConfig)
		assert.NoError(t, err)
		assert.Equal(t, uint(24), dataTypesConfig.Uintval)
		assert.Equal(t, uint8(24), dataTypesConfig.Uint8val)
		mockGodotenv.AssertExpectations(t)
	})
	t.Run("Test unint datatypes for errors", func(t *testing.T) {
		setup()
		t.Cleanup(resetCache)
		type DataTypesConfig struct {
			Uintval  uint  `env:"UINTVAL"`
			Uint8val uint8 `env:"UINT8VAL"`
		}
		invalidUint := "24asdf"
		t.Setenv("UINTVAL", invalidUint)
		t.Setenv("UINT8VAL", invalidUint)
		var dataTypesConfig DataTypesConfig
		err := LoadEnv(&dataTypesConfig)
		assert.Error(t, err)
		assert.Equal(t, fmt.Sprintf("failed to convert UINTVAL to uint: strconv.ParseUint: parsing \"%s\": invalid syntax", invalidUint), err.Error())
		mockGodotenv.AssertExpectations(t)
	})
	t.Run("Test float data types", func(t *testing.T) {
		setup()
		t.Cleanup(resetCache)
		type DataTypesConfig struct {
			Floatval   float32 `env:"FLOATVAL"`
			Float64val float64 `env:"FLOAT64VAL"`
		}
		t.Setenv("FLOATVAL", "24.5")
		t.Setenv("FLOAT64VAL", "24.5")
		var dataTypesConfig DataTypesConfig
		err := LoadEnv(&dataTypesConfig)
		assert.NoError(t, err)
		assert.Equal(t, float32(24.5), dataTypesConfig.Floatval)
		assert.Equal(t, float64(24.5), dataTypesConfig.Float64val)
		mockGodotenv.AssertExpectations(t)
	})
	t.Run("Test float data types for errors", func(t *testing.T) {
		setup()
		t.Cleanup(resetCache)
		type Float32Config struct {
			Floatval float32 `env:"FLOATVAL"`
		}
		type Float64Config struct {
			Float64val float64 `env:"FLOAT64VAL"`
		}
		invalidFloat := "24.5asdf"
		t.Setenv("FLOATVAL", invalidFloat)
		t.Setenv("FLOAT64VAL", invalidFloat)
		var float32val Float32Config
		var float64val Float64Config
		err1 := LoadEnv(&float32val)
		err2 := LoadEnv(&float64val)
		assert.Error(t, err1)
		assert.Error(t, err2)
		assert.Equal(t, fmt.Sprintf("failed to convert FLOATVAL to float: strconv.ParseFloat: parsing \"%s\": invalid syntax", invalidFloat), err1.Error())
		assert.Equal(t, fmt.Sprintf("failed to convert FLOAT64VAL to float: strconv.ParseFloat: parsing \"%s\": invalid syntax", invalidFloat), err2.Error())
		mockGodotenv.AssertExpectations(t)
	})
	t.Run("Test complex data types", func(t *testing.T) {
		setup()
		t.Cleanup(resetCache)
		type ComplexConfig struct {
			Complex64Val  complex64  `env:"COMPLEX64VAL"`
			Complex128Val complex128 `env:"COMPLEX128VAL"`
		}
		t.Setenv("COMPLEX64VAL", "1+2i")
		t.Setenv("COMPLEX128VAL", "1+2i")
		var complexConfig ComplexConfig
		err := LoadEnv(&complexConfig)
		assert.NoError(t, err)
		assert.Equal(t, complex64(1+2i), complexConfig.Complex64Val)
		assert.Equal(t, complex128(1+2i), complexConfig.Complex128Val)
		mockGodotenv.AssertExpectations(t)
	})
	t.Run("Test complex data types for errors", func(t *testing.T) {
		setup()
		t.Cleanup(resetCache)
		type Complex64Config struct {
			Complex64Val  complex64  `env:"COMPLEX64VAL"`
			Complex128Val complex128 `env:"COMPLEX128VAL"`
		}
		type Complex128Config struct {
			Complex128Val complex128 `env:"COMPLEX128VAL"`
		}
		invalidComplex := "1+2i+3"
		t.Setenv("COMPLEX64VAL", invalidComplex)
		t.Setenv("COMPLEX128VAL", invalidComplex)
		var complex64Config Complex64Config
		var complex128Config Complex128Config
		err1 := LoadEnv(&complex64Config)
		err2 := LoadEnv(&complex128Config)
		assert.Error(t, err1)
		assert.Error(t, err2)
		assert.Equal(t, fmt.Sprintf("failed to convert COMPLEX64VAL to complex: strconv.ParseComplex: parsing \"%s\": invalid syntax", invalidComplex), err1.Error())
		assert.Equal(t, fmt.Sprintf("failed to convert COMPLEX128VAL to complex: strconv.ParseComplex: parsing \"%s\": invalid syntax", invalidComplex), err2.Error())
		mockGodotenv.AssertExpectations(t)
	})
	t.Run("Test any data types", func(t *testing.T) {
		setup()
		t.Cleanup(resetCache)
		type AnyConfig struct {
			Anyval any `env:"ANYVAL"`
		}
		t.Setenv("ANYVAL", "any_value")
		var anyConfig AnyConfig
		err := LoadEnv(&anyConfig)
		assert.NoError(t, err)
		assert.Equal(t, "any_value", anyConfig.Anyval)
		mockGodotenv.AssertExpectations(t)
	})
	t.Run("Test array data types", func(t *testing.T) {
		setup()
		t.Cleanup(resetCache)
		type ArrayConfig struct {
			Strval     [2]string    `env:"STRVAL"`
			Intval     [2]int       `env:"INTVAL"`
			Uintval    [2]uint      `env:"UINTVAL"`
			Floatval   [2]float32   `env:"FLOATVAL, delimiter=';'"`
			Boolval    [2]bool      `env:"BOOLVAL"`
			Complexval [2]complex64 `env:"COMPLEXVAL"`
			AnyVal     [2]any       `env:"ANYVAL"`
		}
		t.Setenv("STRVAL", "hello,world")
		t.Setenv("INTVAL", "1,2")
		t.Setenv("UINTVAL", "1,2")
		t.Setenv("FLOATVAL", "1.1;2.2")
		t.Setenv("BOOLVAL", "true,false")
		t.Setenv("COMPLEXVAL", "1+2i,3+4i")
		t.Setenv("ANYVAL", "any_value1,any_value2")
		var arrayConfig ArrayConfig
		err := LoadEnv(&arrayConfig)
		assert.NoError(t, err)
		assert.Equal(t, [2]string{"hello", "world"}, arrayConfig.Strval)
		assert.Equal(t, [2]int{1, 2}, arrayConfig.Intval)
		assert.Equal(t, [2]uint{1, 2}, arrayConfig.Uintval)
		assert.Equal(t, [2]float32{1.1, 2.2}, arrayConfig.Floatval)
		assert.Equal(t, [2]bool{true, false}, arrayConfig.Boolval)
		assert.Equal(t, [2]complex64{1 + 2i, 3 + 4i}, arrayConfig.Complexval)
		assert.Equal(t, [2]any{"any_value1", "any_value2"}, arrayConfig.AnyVal)
		mockGodotenv.AssertExpectations(t)
	})
	t.Run("Test slice data types", func(t *testing.T) {
		setup()
		t.Cleanup(resetCache)
		type ArrayConfig struct {
			Strval     []string    `env:"STRVAL"`
			Intval     []int       `env:"INTVAL"`
			Uintval    []uint      `env:"UINTVAL"`
			Floatval   []float32   `env:"FLOATVAL, delimiter=';'"`
			Boolval    []bool      `env:"BOOLVAL"`
			Complexval []complex64 `env:"COMPLEXVAL"`
			AnyVal     []any       `env:"ANYVAL"`
			KeyBytes   []byte      `env:"KEY_BYTES,default='hello',isstring=true"`
			KeyRunes   []rune      `env:"KEY_RUNES,default='हेलो',isstring"`
		}
		t.Setenv("STRVAL", "hello,world")
		t.Setenv("INTVAL", "1,2")
		t.Setenv("UINTVAL", "1,2")
		t.Setenv("FLOATVAL", "1.1;2.2")
		t.Setenv("BOOLVAL", "true,false")
		t.Setenv("COMPLEXVAL", "1+2i,3+4i")
		t.Setenv("ANYVAL", "any_value1,any_value2")
		var arrayConfig ArrayConfig
		err := LoadEnv(&arrayConfig)
		assert.NoError(t, err)
		assert.Equal(t, []string{"hello", "world"}, arrayConfig.Strval)
		assert.Equal(t, []int{1, 2}, arrayConfig.Intval)
		assert.Equal(t, []uint{1, 2}, arrayConfig.Uintval)
		assert.Equal(t, []float32{1.1, 2.2}, arrayConfig.Floatval)
		assert.Equal(t, []bool{true, false}, arrayConfig.Boolval)
		assert.Equal(t, []complex64{1 + 2i, 3 + 4i}, arrayConfig.Complexval)
		assert.Equal(t, []any{"any_value1", "any_value2"}, arrayConfig.AnyVal)
		assert.Equal(t, "hello", string(arrayConfig.KeyBytes))
		assert.Equal(t, "हेलो", string(arrayConfig.KeyRunes))
		mockGodotenv.AssertExpectations(t)
	})
	t.Run("Test array data types for errors", func(t *testing.T) {
		setup()
		t.Cleanup(resetCache)
		type ArrayInvalidLengthConfig struct {
			Strval [2]string `env:"STRVAL"`
		}
		type ArrayInvalidIntConfig struct {
			Intval [2]int `env:"INTVAL"`
		}
		type ArrayInvalidUintConfig struct {
			Uintval [2]uint `env:"UINTVAL"`
		}
		type ArrayInvalidFloatConfig struct {
			Floatval [2]float32 `env:"FLOATVAL, delimiter=';'"`
		}
		type ArrayInvalidBoolConfig struct {
			Boolval [2]bool `env:"BOOLVAL"`
		}
		type ArrayInvalidComplexConfig struct {
			Complexval [2]complex64 `env:"COMPLEXVAL"`
		}
		type ArrayInvalidNotSupportedConfig struct {
			NotSupportedval [2]struct{} `env:"NOTSUPPORTEDVAL"`
		}
		type ArrayInvalidDelimiterConfig struct {
			Strval [3]string `env:"STRVAL, delimiter"`
		}

		var arrayInvalidLengthConfig ArrayInvalidLengthConfig
		var arrayInvalidIntConfig ArrayInvalidIntConfig
		var arrayInvalidUintConfig ArrayInvalidUintConfig
		var arrayInvalidFloatConfig ArrayInvalidFloatConfig
		var arrayInvalidBoolConfig ArrayInvalidBoolConfig
		var arrayInvalidComplexConfig ArrayInvalidComplexConfig
		var arrayInvalidNotSupportedConfig ArrayInvalidNotSupportedConfig
		var arrayInvalidDelimiterConfig ArrayInvalidDelimiterConfig
		t.Setenv("STRVAL", "hello,world,foo")
		t.Setenv("INTVAL", "1,2a")
		t.Setenv("UINTVAL", "1,2b")
		t.Setenv("FLOATVAL", "1.1;2.2aa")
		t.Setenv("BOOLVAL", "true,falsea")
		t.Setenv("COMPLEXVAL", "1+2i,3+4")
		t.Setenv("NOTSUPPORTEDVAL", "1,2")
		err1 := LoadEnv(&arrayInvalidLengthConfig)
		err2 := LoadEnv(&arrayInvalidIntConfig)
		err3 := LoadEnv(&arrayInvalidUintConfig)
		err4 := LoadEnv(&arrayInvalidFloatConfig)
		err5 := LoadEnv(&arrayInvalidBoolConfig)
		err6 := LoadEnv(&arrayInvalidComplexConfig)
		err7 := LoadEnv(&arrayInvalidNotSupportedConfig)
		err8 := LoadEnv(&arrayInvalidDelimiterConfig)
		assert.Error(t, err1)
		assert.Error(t, err2)
		assert.Error(t, err3)
		assert.Error(t, err4)
		assert.Error(t, err5)
		assert.Error(t, err6)
		assert.Error(t, err7)
		assert.NoError(t, err8)
		assert.Equal(t, "env var STRVAL has 3 values, but array expects 2", err1.Error())
		assert.Equal(t, "failed to convert INTVAL to int: strconv.ParseInt: parsing \"2a\": invalid syntax", err2.Error())
		assert.Equal(t, "failed to convert UINTVAL to uint: strconv.ParseUint: parsing \"2b\": invalid syntax", err3.Error())
		assert.Equal(t, "failed to convert FLOATVAL to float: strconv.ParseFloat: parsing \"2.2aa\": invalid syntax", err4.Error())
		assert.Equal(t, "error parsing env var BOOLVAL: strconv.ParseBool: parsing \"falsea\": invalid syntax", err5.Error())
		assert.Equal(t, "failed to convert COMPLEXVAL to complex: strconv.ParseComplex: parsing \"3+4\": invalid syntax", err6.Error())
		assert.Equal(t, "unsupported slice/array element type: struct", err7.Error())
		mockGodotenv.AssertExpectations(t)
	})
	t.Run("Test slice data types for errors", func(t *testing.T) {
		setup()
		t.Cleanup(resetCache)
		type ArrayInvalidLengthConfig struct {
			Strval []string `env:"STRVAL"`
		}
		type ArrayInvalidIntConfig struct {
			Intval []int `env:"INTVAL"`
		}
		type ArrayInvalidUintConfig struct {
			Uintval []uint `env:"UINTVAL"`
		}
		type ArrayInvalidFloatConfig struct {
			Floatval []float32 `env:"FLOATVAL, delimiter=';'"`
		}
		type ArrayInvalidBoolConfig struct {
			Boolval []bool `env:"BOOLVAL"`
		}
		type ArrayInvalidComplexConfig struct {
			Complexval []complex64 `env:"COMPLEXVAL"`
		}
		type ArrayInvalidNotSupportedConfig struct {
			NotSupportedval []struct{} `env:"NOTSUPPORTEDVAL"`
		}
		type ArrayInvalidDelimiterConfig struct {
			Strval []string `env:"STRVAL, delimiter"`
		}
		type SliceInvalidIsStringFalse struct {
			KeyBytes []byte `env:"KEY_BYTES,default='hello',isstring=false"`
		}

		var arrayInvalidLengthConfig ArrayInvalidLengthConfig
		var arrayInvalidIntConfig ArrayInvalidIntConfig
		var arrayInvalidUintConfig ArrayInvalidUintConfig
		var arrayInvalidFloatConfig ArrayInvalidFloatConfig
		var arrayInvalidBoolConfig ArrayInvalidBoolConfig
		var arrayInvalidComplexConfig ArrayInvalidComplexConfig
		var arrayInvalidNotSupportedConfig ArrayInvalidNotSupportedConfig
		var arrayInvalidDelimiterConfig ArrayInvalidDelimiterConfig
		var sliceInvalidIsStringFalseConfig SliceInvalidIsStringFalse
		t.Setenv("STRVAL", "hello,world,foo")
		t.Setenv("INTVAL", "1,2a")
		t.Setenv("UINTVAL", "1,2b")
		t.Setenv("FLOATVAL", "1.1;2.2aa")
		t.Setenv("BOOLVAL", "true,falsea")
		t.Setenv("COMPLEXVAL", "1+2i,3+4")
		t.Setenv("NOTSUPPORTEDVAL", "1,2")
		err1 := LoadEnv(&arrayInvalidLengthConfig)
		err2 := LoadEnv(&arrayInvalidIntConfig)
		err3 := LoadEnv(&arrayInvalidUintConfig)
		err4 := LoadEnv(&arrayInvalidFloatConfig)
		err5 := LoadEnv(&arrayInvalidBoolConfig)
		err6 := LoadEnv(&arrayInvalidComplexConfig)
		err7 := LoadEnv(&arrayInvalidNotSupportedConfig)
		err8 := LoadEnv(&arrayInvalidDelimiterConfig)
		err9 := LoadEnv(&sliceInvalidIsStringFalseConfig)
		assert.NoError(t, err1)
		assert.Error(t, err2)
		assert.Error(t, err3)
		assert.Error(t, err4)
		assert.Error(t, err5)
		assert.Error(t, err6)
		assert.Error(t, err7)
		assert.NoError(t, err8)
		assert.Error(t, err9)
		assert.Equal(t, "failed to convert INTVAL to int: strconv.ParseInt: parsing \"2a\": invalid syntax", err2.Error())
		assert.Equal(t, "failed to convert UINTVAL to uint: strconv.ParseUint: parsing \"2b\": invalid syntax", err3.Error())
		assert.Equal(t, "failed to convert FLOATVAL to float: strconv.ParseFloat: parsing \"2.2aa\": invalid syntax", err4.Error())
		assert.Equal(t, "error parsing env var BOOLVAL: strconv.ParseBool: parsing \"falsea\": invalid syntax", err5.Error())
		assert.Equal(t, "failed to convert COMPLEXVAL to complex: strconv.ParseComplex: parsing \"3+4\": invalid syntax", err6.Error())
		assert.Equal(t, "unsupported slice/array element type: struct", err7.Error())
		assert.Equal(t, "failed to convert KEY_BYTES to uint: strconv.ParseUint: parsing \"hello\": invalid syntax", err9.Error())
		mockGodotenv.AssertExpectations(t)
	})
	t.Run("Test map data types", func(t *testing.T) {
		setup()
		t.Cleanup(resetCache)
		type MapValConfig struct {
			Strval     map[string]string       `env:"STRVAL"`
			Intval     map[int]int             `env:"INTVAL"`
			Uintval    map[uint]uint           `env:"UINTVAL"`
			Floatval   map[float32]float32     `env:"FLOATVAL, delimiter=';'"`
			Boolval    map[bool]bool           `env:"BOOLVAL"`
			Complexval map[complex64]complex64 `env:"COMPLEXVAL"`
			AnyVal     map[any]any             `env:"ANYVAL"`
		}
		t.Setenv("STRVAL", "{hello:world,foo:bar}")
		t.Setenv("INTVAL", "{1:2,3:4}")
		t.Setenv("UINTVAL", "{1:2,3:4}")
		t.Setenv("FLOATVAL", "{1.1:2.2;3.3:4.4}")
		t.Setenv("BOOLVAL", "{true:false}")
		t.Setenv("COMPLEXVAL", "{(1+2i):(3+4i)}")
		t.Setenv("ANYVAL", "{any_value1:any_value2}")
		var mapValConfig MapValConfig
		err := LoadEnv(&mapValConfig)
		assert.NoError(t, err)
		assert.Equal(t, map[string]string{"hello": "world", "foo": "bar"}, mapValConfig.Strval)
		assert.Equal(t, map[int]int{1: 2, 3: 4}, mapValConfig.Intval)
		assert.Equal(t, map[uint]uint{1: 2, 3: 4}, mapValConfig.Uintval)
		assert.Equal(t, map[float32]float32{1.1: 2.2, 3.3: 4.4}, mapValConfig.Floatval)
		assert.Equal(t, map[bool]bool{true: false}, mapValConfig.Boolval)
		assert.Equal(t, map[complex64]complex64{complex64(1 + 2i): complex64(3 + 4i)}, mapValConfig.Complexval)
		assert.Equal(t, map[any]any{"any_value1": "any_value2"}, mapValConfig.AnyVal)
		mockGodotenv.AssertExpectations(t)
	})
	t.Run("Test map values types for errors", func(t *testing.T) {
		setup()
		t.Cleanup(resetCache)

		type MapValConfigInvalid struct {
			Strval map[string]struct{} `env:"STRVAL"`
		}
		type MapValConfigInvalidInt struct {
			Intval map[int]int `env:"INTVAL"`
		}
		type MapValConfigInvalidUint struct {
			Uintval map[uint]uint `env:"UINTVAL"`
		}
		type MapValConfigInvalidFloat struct {
			Floatval map[float32]float32 `env:"FLOATVAL, delimiter=';'"`
		}
		type MapValConfigInvalidBool struct {
			Boolval map[bool]bool `env:"BOOLVAL"`
		}
		type MapValConfigInvalidComplex struct {
			Complexval map[complex64]complex64 `env:"COMPLEXVAL"`
		}
		type MapValConfigInvalidAny struct {
			AnyVal map[any]struct{} `env:"ANYVAL"`
		}
		type MapInvalidKeyValConfig struct {
			InvalidVal map[struct{}]string `env:"INVALIDVAL"`
		}
		var mapValConfigInvalid MapValConfigInvalid
		var mapValConfigInvalidInt MapValConfigInvalidInt
		var mapValConfigInvalidUint MapValConfigInvalidUint
		var mapValConfigInvalidFloat MapValConfigInvalidFloat
		var mapValConfigInvalidBool MapValConfigInvalidBool
		var mapValConfigInvalidComplex MapValConfigInvalidComplex
		var mapValConfigInvalidAny MapValConfigInvalidAny
		var mapInvalidKeyValConfig MapInvalidKeyValConfig
		t.Setenv("STRVAL", "{hello:world,foo:bar}")
		t.Setenv("INTVAL", "{1:2a,3:4}")
		t.Setenv("UINTVAL", "{1:2a,3:4}")
		t.Setenv("FLOATVAL", "{1.1:2.2a;3.3:4.4}")
		t.Setenv("BOOLVAL", "{true:falsea}")
		t.Setenv("COMPLEXVAL", "{(1+2i):(3+4)}")
		t.Setenv("ANYVAL", "{any_value1:any_value2}")
		t.Setenv("NOTSUPPORTEDVAL", "{1:2,3:4}")
		t.Setenv("INVALIDVAL", "{helloworld,foo:bar}")
		err1 := LoadEnv(&mapValConfigInvalid)
		err2 := LoadEnv(&mapValConfigInvalidInt)
		err3 := LoadEnv(&mapValConfigInvalidUint)
		err4 := LoadEnv(&mapValConfigInvalidFloat)
		err5 := LoadEnv(&mapValConfigInvalidBool)
		err6 := LoadEnv(&mapValConfigInvalidComplex)
		err7 := LoadEnv(&mapValConfigInvalidAny)
		err8 := LoadEnv(&mapInvalidKeyValConfig)
		assert.Error(t, err1)
		assert.Error(t, err2)
		assert.Error(t, err3)
		assert.Error(t, err4)
		assert.Error(t, err5)
		assert.Error(t, err6)
		assert.Error(t, err7)
		assert.Error(t, err8)
		assert.Equal(t, "unsupported map value type: struct", err1.Error())
		assert.Equal(t, "failed to convert map value 2a to int: strconv.ParseInt: parsing \"2a\": invalid syntax", err2.Error())
		assert.Equal(t, "failed to convert map value 2a to uint: strconv.ParseUint: parsing \"2a\": invalid syntax", err3.Error())
		assert.Equal(t, "failed to convert map value 2.2a to float: strconv.ParseFloat: parsing \"2.2a\": invalid syntax", err4.Error())
		assert.Equal(t, "failed to convert map value falsea to bool: strconv.ParseBool: parsing \"falsea\": invalid syntax", err5.Error())
		assert.Equal(t, "failed to convert map value (3+4) to complex: strconv.ParseComplex: parsing \"(3+4)\": invalid syntax", err6.Error())
		assert.Equal(t, "unsupported map value type: struct", err7.Error())
		assert.Equal(t, "invalid map entry for INVALIDVAL: helloworld", err8.Error())
		mockGodotenv.AssertExpectations(t)
	})
	t.Run("Test map key for errors", func(t *testing.T) {
		setup()
		t.Cleanup(resetCache)
		type MapKeyIntConfig struct {
			Intval map[int]string `env:"INTVAL"`
		}
		type MapKeyUintConfig struct {
			Uintval map[uint]string `env:"UINTVAL"`
		}
		type MapKeyFloatConfig struct {
			Floatval map[float32]string `env:"FLOATVAL, delimiter=';'"`
		}
		type MapKeyBoolConfig struct {
			Boolval map[bool]string `env:"BOOLVAL"`
		}
		type MapKeyComplexConfig struct {
			Complexval map[complex64]string `env:"COMPLEXVAL"`
		}
		type MapKeyStructConfig struct {
			Anyval map[struct{}]string `env:"ANYVAL"`
		}

		var mapKeyIntConfig MapKeyIntConfig
		var mapKeyUintConfig MapKeyUintConfig
		var mapKeyFloatConfig MapKeyFloatConfig
		var mapKeyBoolConfig MapKeyBoolConfig
		var mapKeyComplexConfig MapKeyComplexConfig
		var mapKeyStructConfig MapKeyStructConfig
		t.Setenv("INTVAL", "{1a:2,3:4}")
		t.Setenv("UINTVAL", "{1a:2,3:4}")
		t.Setenv("FLOATVAL", "{1.1a:2.2;3.3:4.4}")
		t.Setenv("BOOLVAL", "{truea:false}")
		t.Setenv("COMPLEXVAL", "{(1+2):(3+4i)}")
		t.Setenv("ANYVAL", "{any_value1:any_value2}")
		err1 := LoadEnv(&mapKeyIntConfig)
		err2 := LoadEnv(&mapKeyUintConfig)
		err3 := LoadEnv(&mapKeyFloatConfig)
		err4 := LoadEnv(&mapKeyBoolConfig)
		err5 := LoadEnv(&mapKeyComplexConfig)
		err6 := LoadEnv(&mapKeyStructConfig)
		assert.Error(t, err1)
		assert.Error(t, err2)
		assert.Error(t, err3)
		assert.Error(t, err4)
		assert.Error(t, err5)
		assert.Error(t, err6)
		assert.Equal(t, "failed to convert map key 1a to int: strconv.ParseInt: parsing \"1a\": invalid syntax", err1.Error())
		assert.Equal(t, "failed to convert map key 1a to uint: strconv.ParseUint: parsing \"1a\": invalid syntax", err2.Error())
		assert.Equal(t, "failed to convert map key 1.1a to float: strconv.ParseFloat: parsing \"1.1a\": invalid syntax", err3.Error())
		assert.Equal(t, "failed to convert map key truea to bool: strconv.ParseBool: parsing \"truea\": invalid syntax", err4.Error())
		assert.Equal(t, "failed to convert map key (1+2) to complex: strconv.ParseComplex: parsing \"(1+2)\": invalid syntax", err5.Error())
		assert.Equal(t, "unsupported map key type: struct", err6.Error())
		mockGodotenv.AssertExpectations(t)
	})
	t.Run("Test unsupported data types", func(t *testing.T) {
		setup()
		t.Cleanup(resetCache)
		type UnsupportedConfig struct {
			UnsupportedField struct{} `env:"UNSUPPORTED_FIELD"`
		}
		var config UnsupportedConfig
		err := LoadEnv(&config)
		assert.Error(t, err)
		assert.Equal(t, "unsupported field type: struct", err.Error())
		mockGodotenv.AssertExpectations(t)
	})

	// testing for required env var with different cases
	t.Run("Test with required env var", func(t *testing.T) {
		setup()
		t.Cleanup(resetCache)
		type RequiredConfig struct {
			RequiredField string `env:"REQUIRED_FIELD,required"`
		}
		t.Setenv("REQUIRED_FIELD", "required_value")
		var config RequiredConfig
		err := LoadEnv(&config)
		assert.NoError(t, err)
		assert.Equal(t, "required_value", config.RequiredField)
		mockGodotenv.AssertExpectations(t)
	})
	t.Run("Test with required env var not set", func(t *testing.T) {
		setup()
		t.Cleanup(resetCache)
		type RequiredConfig struct {
			RequiredField string `env:"REQUIRED_FIELD,required"`
		}
		var config RequiredConfig
		err := LoadEnv(&config)
		assert.Error(t, err)
		assert.Equal(t, "required environment variable REQUIRED_FIELD not found", err.Error())
		mockGodotenv.AssertExpectations(t)
	})
	t.Run("Test with required env var set to false", func(t *testing.T) {
		setup()
		t.Cleanup(resetCache)
		type RequiredConfig struct {
			RequiredField string `env:"REQUIRED_FIELD,required=false"`
		}
		t.Setenv("REQUIRED_FIELD", "required_value")
		var config RequiredConfig
		err := LoadEnv(&config)
		assert.NoError(t, err)
		assert.Equal(t, "required_value", config.RequiredField)
		mockGodotenv.AssertExpectations(t)
	})
	t.Run("Test with required env var set to true", func(t *testing.T) {
		setup()
		t.Cleanup(resetCache)
		type RequiredConfig struct {
			RequiredField string `env:"REQUIRED_FIELD,required=true"`
		}
		t.Setenv("REQUIRED_FIELD", "required_value")
		var config RequiredConfig
		err := LoadEnv(&config)
		assert.NoError(t, err)
		assert.Equal(t, "required_value", config.RequiredField)
		mockGodotenv.AssertExpectations(t)
	})
	// testing for default env var with different cases
	t.Run("Test with default env var", func(t *testing.T) {
		setup()
		t.Cleanup(resetCache)
		type DefaultConfig struct {
			DefaultField string `env:"DEFAULT_FIELD,default=default_value"`
		}
		var config DefaultConfig
		err := LoadEnv(&config)
		assert.NoError(t, err)
		assert.Equal(t, "default_value", config.DefaultField)
		mockGodotenv.AssertExpectations(t)
	})
	t.Run("Test with default env var set to empty", func(t *testing.T) {
		setup()
		t.Cleanup(resetCache)
		type DefaultConfig struct {
			DefaultField string `env:"DEFAULT_FIELD,default="`
		}
		t.Setenv("DEFAULT_FIELD", "default_value")
		var config DefaultConfig
		err := LoadEnv(&config)
		assert.NoError(t, err)
		assert.Equal(t, "default_value", config.DefaultField)
		mockGodotenv.AssertExpectations(t)
	})
	t.Run("Test with default env var set to empty string", func(t *testing.T) {
		setup()
		t.Cleanup(resetCache)
		type DefaultConfig struct {
			DefaultField string `env:"DEFAULT_FIELD,default=\"\""`
		}
		var config DefaultConfig
		err := LoadEnv(&config)
		assert.NoError(t, err)
		assert.Equal(t, "", config.DefaultField)
		mockGodotenv.AssertExpectations(t)
	})
	t.Run("Test with default env var set to empty string with quotes", func(t *testing.T) {
		setup()
		t.Cleanup(resetCache)
		type DefaultConfig struct {
			DefaultField string `env:"DEFAULT_FIELD,default=''"`
		}
		var config DefaultConfig
		err := LoadEnv(&config)
		assert.NoError(t, err)
		assert.Equal(t, "", config.DefaultField)
		mockGodotenv.AssertExpectations(t)
	})
	t.Run("Test default with no equal to sign", func(t *testing.T) {
		setup()
		t.Cleanup(resetCache)
		type DefaultConfig struct {
			DefaultField string `env:"DEFAULT_FIELD,default"`
		}
		var config DefaultConfig
		err := LoadEnv(&config)
		assert.NoError(t, err)
		assert.Equal(t, "", config.DefaultField)
		mockGodotenv.AssertExpectations(t)
	})
	t.Run("Test with default env var set to empty string with quotes and spaces", func(t *testing.T) {
		setup()
		t.Cleanup(resetCache)
		type DefaultConfig struct {
			DefaultField string `env:"DEFAULT_FIELD,default='  '"`
		}
		var config DefaultConfig
		err := LoadEnv(&config)
		assert.NoError(t, err)
		assert.Equal(t, "", config.DefaultField)
		mockGodotenv.AssertExpectations(t)
	})
	t.Run("Test with both default and required env var", func(t *testing.T) {
		setup()
		t.Cleanup(resetCache)
		type DefaultRequiredConfig struct {
			DefaultField string `env:"DEFAULT_FIELD,default=default_value,required"`
		}
		var config DefaultRequiredConfig
		err := LoadEnv(&config)
		assert.NoError(t, err)
		assert.Equal(t, "default_value", config.DefaultField)
		mockGodotenv.AssertExpectations(t)
	})
	t.Run("Test with both default and required env var set to empty", func(t *testing.T) {
		setup()
		t.Cleanup(resetCache)
		type DefaultRequiredConfig struct {
			DefaultField string `env:"DEFAULT_FIELD,default=,required"`
		}
		t.Setenv("DEFAULT_FIELD", "default_value")
		var config DefaultRequiredConfig
		err := LoadEnv(&config)
		assert.NoError(t, err)
		assert.Equal(t, "default_value", config.DefaultField)
		mockGodotenv.AssertExpectations(t)
	})
	t.Run("Test with both default and required env var set to empty string", func(t *testing.T) {
		setup()
		t.Cleanup(resetCache)
		type DefaultRequiredConfig struct {
			DefaultField string `env:"DEFAULT_FIELD,default=\"\",required"`
		}
		var config DefaultRequiredConfig
		err := LoadEnv(&config)
		assert.Error(t, err)
		assert.Equal(t, "required environment variable DEFAULT_FIELD not found", err.Error())
		assert.Equal(t, "", config.DefaultField)
		mockGodotenv.AssertExpectations(t)
	})
	t.Run("Test with both default and required env var set to empty string with quotes", func(t *testing.T) {
		setup()
		t.Cleanup(resetCache)
		type DefaultRequiredConfig struct {
			DefaultField string `env:"DEFAULT_FIELD,default='',required"`
		}
		var config DefaultRequiredConfig
		err := LoadEnv(&config)
		assert.Error(t, err)
		assert.Equal(t, "required environment variable DEFAULT_FIELD not found", err.Error())
		assert.Equal(t, "", config.DefaultField)
		mockGodotenv.AssertExpectations(t)
	})
	t.Run("Test with both default and required env var set to empty string with quotes and spaces", func(t *testing.T) {
		setup()
		t.Cleanup(resetCache)
		type DefaultRequiredConfig struct {
			DefaultField string `env:"DEFAULT_FIELD,default='  ',required"`
		}
		var config DefaultRequiredConfig
		err := LoadEnv(&config)
		assert.Error(t, err)
		assert.Equal(t, "required environment variable DEFAULT_FIELD not found", err.Error())
		assert.Equal(t, "", config.DefaultField)
		mockGodotenv.AssertExpectations(t)
	})
}
