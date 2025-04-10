package envarfig

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

type tagProperties struct {
	EnvName      string
	DefaultValue string
	Required     bool
}

func (tp *tagProperties) setEnvName(envName string) {
	tp.EnvName = envName
}
func (tp *tagProperties) setDefaultValue(defaultValue string) {
	tp.DefaultValue = defaultValue
}
func (tp *tagProperties) setRequired(required bool) {
	tp.Required = required
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
		tagProp := parseTagAndTagValues(tagValues)
		// TODO: tag properties feature will be implemented in future

		//get and set the env var value
		envValue, exist := os.LookupEnv(tagProp.EnvName)
		if !exist {
			// check if the field is required
			if tagProp.Required && tagProp.DefaultValue == "" {
				return fmt.Errorf("required environment variable %s not found", tagProp.EnvName)
			}
			// set the field value to the default value
			envValue = tagProp.DefaultValue
		}
		// set the field value
		fieldValue := value.Field(i)
		if err := setEnvVarValues(fieldValue, tagProp.EnvName, envValue); err != nil {
			return err
		}
	}

	return nil
}

func parseTagAndTagValues(tag string) tagProperties {
	properties := strings.Split(tag, ",")
	for i, v := range properties {
		properties[i] = strings.TrimSpace(v)
	}
	tagProp := tagProperties{}
	envName := properties[0]
	tagProp.setEnvName(envName)
	if len(properties) > 1 {
		for _, prop := range properties[1:] {
			// the required field in prop is of type "required" or "required=true"
			checkAndSetRequired(prop, &tagProp)
			checkAndSetDefaultValue(prop, &tagProp)
		}
	} else {
		tagProp.setDefaultValue("")
		tagProp.setRequired(false)
	}

	return tagProp
}

func setEnvVarValues(fieldValue reflect.Value, envName string, envValue string) error {
	switch fieldValue.Kind() {
	case reflect.String:
		// set the field value to the env var value
		fieldValue.SetString(envValue)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intValue, err := strconv.ParseInt(envValue, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to convert %s to int: %w", envName, err)
		}
		fieldValue.SetInt(intValue)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintValue, err := strconv.ParseUint(envValue, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to convert %s to uint: %w", envName, err)
		}
		fieldValue.SetUint(uintValue)
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

func checkAndSetRequired(property string, tagProp *tagProperties) {
	if !strings.Contains(strings.ToLower(property), "required") {
		return
	}
	// check if the required field is set to true or false
	if strings.Contains(property, "=") {
		property = strings.Split(property, "=")[1]
		property = strings.TrimSpace(property)
		property = strings.ToLower(property)
	}
	if property == "true" {
		tagProp.setRequired(true)
	} else if property == "false" {
		tagProp.setRequired(false)
	} else {
		tagProp.setRequired(true)
	}

}

func checkAndSetDefaultValue(property string, tagProp *tagProperties) {
	if !strings.Contains(strings.ToLower(property), "default") {
		return
	}
	// check if the default field is set to true or false
	if !strings.Contains(property, "=") {
		return
	}
	property = strings.SplitN(property, "=", 2)[1]
	property = strings.TrimSpace(property)
	property = strings.ToLower(property)
	valLen := len(property)

	if valLen >= 2 {
		first, last := property[0], property[valLen-1]
		if (first == last) && (first == '"' || first == '\'') {
			property = strings.TrimSpace(property[1 : valLen-1])
		}
	}
	tagProp.setDefaultValue(property)
}
