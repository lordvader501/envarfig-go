package envarfig

import (
	"reflect"
	"sync"
)

var cachedConfigs sync.Map // Map to store cached configurations

/*
args:
  - envConfig: a pointer to a struct
  - options: variadic options for configuration (e.g., env file paths, auto-load settings)

returns:
  - error: an error if any
*/
func LoadEnv[T any](envConfig *T, options ...option) error {
	if envConfig == nil {
		return errNilConfig
	}

	// Load the settings
	settings := loadSettings(options...)

	// Get the type of the struct to use as a cache key
	structType := reflect.TypeOf(envConfig).Elem()

	// Check if caching is enabled and the struct is already cached
	if settings.CacheConfig {
		if cachedConfig, ok := cachedConfigs.Load(structType); ok {
			*envConfig = cachedConfig.(T) // Load from cache
			return nil
		}
	}

	var err error
	var once sync.Once

	// Ensure the struct is only loaded once
	once.Do(func() {
		// Load the env file
		err = loadEnvFile(settings.AutoLoadEnv, settings.EnvFiles)
		if err != nil {
			err = errInvalidEnvPathArgs
			return
		}

		// Parse the environment variables into the struct
		err = parseEnvVar(envConfig)
		if err == nil && settings.CacheConfig {
			// Cache the struct configuration
			cachedConfigs.Store(structType, *envConfig)
		}
	})

	if err != nil {
		return err
	}

	return nil
}
