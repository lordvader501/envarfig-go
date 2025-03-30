package envarfig

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
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
func GetEnvVar[T any](envConfig *T, options ...option) error {
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

/*
Parse the env var from the config struct
*/
func parseEnvVar[T any](config *T) error {
	// get the value of the config
	value := reflect.ValueOf(config)

	// check if config is a pointer to a struct
	if value.Kind() != reflect.Ptr || value.Elem().Kind() != reflect.Struct {
		return errConfigNotPtrToStruct
	}

	// get the type of the config
	value = value.Elem()
	typ := value.Type()

	// loop through the fields of the struct
	for i := range typ.NumField() {
		field := typ.Field(i)
		tagValues := field.Tag.Get(defaultTagName) // get the tag value

		// check if the tag is empty
		if tagValues == "" {
			return errTagNotFound
		}

		// get the field value
		envName, _ := parseTag(tagValues)
		// TODO: tag properties feature will be implemented in future

		//get and set the env var value
		envValue, _ := os.LookupEnv(envName)
		// set the field value
		fieldValue := value.Field(i)
		if err := setEnvVarValues(fieldValue, envName, envValue); err != nil {
			return err
		}
	}

	return nil
}
func setEnvVarValues(fieldValue reflect.Value, envName string, envValue string) error {
	switch fieldValue.Kind() {
	case reflect.String:
		// set the field value to the env var value
		fieldValue.SetString(envValue)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intValue, err := strconv.Atoi(envValue)
		if err != nil {
			return fmt.Errorf("failed to convert %s to int: %w", envName, err)
		}
		fieldValue.SetInt(int64(intValue))
	case reflect.Bool:
		// set the field value to the env var value
		boolValue, err := strconv.ParseBool(envValue)
		if err != nil {
			return fmt.Errorf("error parsing env var %s: %w", envName, err)
		}
		fieldValue.SetBool(boolValue)
	}
	return nil
}
func parseTag(tag string) (string, []string) {
	properties := strings.Split(tag, ",")
	for i, v := range properties {
		properties[i] = strings.TrimSpace(v)
	}
	return properties[0], properties[1:]
}
