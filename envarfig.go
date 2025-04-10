package envarfig

import (
	"sync"
)

// constants
const (
	// DefaultTagName is the default tag name for the env tag
	defaultTagName = "env"
)

var once sync.Once
var cachedConfig any // Stores the parsed config

/*
args:
  - config: a pointer to a struct
  - useEnvFile: a boolean value to determine if the env file should be used(uses godotenv)
  - rest prams:
  - args[0]- envfile path: the name of the env var or list of paths

returns:
  - error: an error if any
*/
func LoadEnv[T any](envConfig *T, options ...option) error {
	if envConfig == nil {
		return errNilConfig
	}

	var err error

	once.Do(func() {
		// load the settings
		settings := loadSettings(options...)
		// load the env file
		err = loadEnvFile(settings.AutoLoadEnv, settings.EnvFiles)
		if err != nil {
			err = errInvalidEnvPathArgs
			return
		}
		// parse the env var
		err = parseEnvVar(envConfig)
		if err == nil {
			cachedConfig = *envConfig
		}
	})

	if err != nil {
		return err
	}

	*envConfig = cachedConfig.(T)
	return nil
}
