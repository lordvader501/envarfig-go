package envarfig

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

// constants
const (
	// DefaultTagName is the default tag name for the env tag
	defaultTagName = "env"
)

type tagProperties struct {
	EnvName      string
	DefaultValue string
	Delimiter    string
	Required     bool
	isString     bool
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
func (tp *tagProperties) setDelimiter(s string) {
	tp.Delimiter = s
}
func (tp *tagProperties) setIsString() {
	tp.isString = true
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
		if err := setEnvVarValues(fieldValue, tagProp, envValue); err != nil {
			return err
		}
	}

	return nil
}

func parseTagAndTagValues(tag string) tagProperties {
	properties := splitTagRespectingQuotes(tag)
	tagProp := tagProperties{}
	envName := properties[0]
	tagProp.setEnvName(envName)
	// setting defaults
	tagProp.setDefaultValue("")
	tagProp.setRequired(false)
	tagProp.setDelimiter(",")
	if len(properties) > 1 {
		for _, prop := range properties[1:] {
			// the required field in prop is of type "required" or "required=true"
			checkAndSetTagPropRequired(prop, &tagProp)
			checkAndSetTagPropDefaultValue(prop, &tagProp)
			checkAndSetTagPropDelimiterForSliceOrArray(prop, &tagProp)
			cehckAndSetIsStringForByteOrRuneArray(prop, &tagProp)
		}
	}

	return tagProp
}

func setEnvVarValues(fieldValue reflect.Value, tagProp tagProperties, envValue string) error {
	switch fieldValue.Kind() {
	case reflect.String:
		// set the field value to the env var value
		fieldValue.SetString(envValue)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intValue, err := strconv.ParseInt(envValue, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to convert %s to int: %w", tagProp.EnvName, err)
		}
		fieldValue.SetInt(intValue)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintValue, err := strconv.ParseUint(envValue, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to convert %s to uint: %w", tagProp.EnvName, err)
		}
		fieldValue.SetUint(uintValue)
	case reflect.Float32, reflect.Float64:
		floatValue, err := strconv.ParseFloat(envValue, 64)
		if err != nil {
			return fmt.Errorf("failed to convert %s to float: %w", tagProp.EnvName, err)
		}
		fieldValue.SetFloat(floatValue)
	case reflect.Complex64, reflect.Complex128:
		envValue = strings.ReplaceAll(envValue, " ", "")
		complexValue, err := strconv.ParseComplex(envValue, 128)
		if err != nil {
			return fmt.Errorf("failed to convert %s to complex: %w", tagProp.EnvName, err)
		}
		fieldValue.SetComplex(complexValue)
	case reflect.Slice, reflect.Array:
		if err := setEnvVarSliceOrArrayValues(fieldValue, tagProp.EnvName, envValue, tagProp); err != nil {
			return err
		}
	case reflect.Map:
		if err := setEnvVarMapValues(fieldValue, tagProp.EnvName, envValue, tagProp); err != nil {
			return err
		}
	case reflect.Bool:
		// set the field value to the env var value
		boolValue, err := strconv.ParseBool(envValue)
		if err != nil {
			return fmt.Errorf("error parsing env var %s: %w", tagProp.EnvName, err)
		}
		fieldValue.SetBool(boolValue)
	case reflect.Interface:
		// set the field value to the env var value
		fieldValue.Set(reflect.ValueOf(envValue))
	default:
		return fmt.Errorf("unsupported field type: %s", fieldValue.Kind())
	}
	return nil
}

func setEnvVarSliceOrArrayValues(fieldValue reflect.Value, envName string, envValue string, tagProp tagProperties) error {
	envValSliceOrArray := strings.Split(envValue, tagProp.Delimiter)
	isString := tagProp.isString

	// Determine the type: slice or array
	kind := fieldValue.Kind()
	elemType := fieldValue.Type().Elem()

	// Create new slice or get a new array instance
	var newValue reflect.Value
	switch kind {
	case reflect.Slice:
		newValue = reflect.MakeSlice(fieldValue.Type(), len(envValSliceOrArray), len(envValSliceOrArray))
	case reflect.Array:
		if len(envValSliceOrArray) != fieldValue.Len() {
			return fmt.Errorf("env var %s has %d values, but array expects %d", envName, len(envValSliceOrArray), fieldValue.Len())
		}
		newValue = fieldValue
	}

	// Set elements
	for i, v := range envValSliceOrArray {
		strVal := strings.TrimSpace(v)

		switch elemType.Kind() {
		case reflect.String:
			newValue.Index(i).SetString(strVal)

		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if isString && elemType.Kind() == reflect.Int32 {
				fieldValue.Set(reflect.ValueOf([]rune(envValue)))
				return nil
			}
			intValue, err := strconv.ParseInt(strVal, 10, elemType.Bits())
			if err != nil {
				return fmt.Errorf("failed to convert %s to int: %w", envName, err)
			}
			newValue.Index(i).SetInt(intValue)

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if isString && elemType.Kind() == reflect.Uint8 {
				fieldValue.SetBytes([]byte(envValue))
				return nil
			}
			uintValue, err := strconv.ParseUint(strVal, 10, elemType.Bits())
			if err != nil {
				return fmt.Errorf("failed to convert %s to uint: %w", envName, err)
			}
			newValue.Index(i).SetUint(uintValue)

		case reflect.Float32, reflect.Float64:
			floatValue, err := strconv.ParseFloat(strVal, elemType.Bits())
			if err != nil {
				return fmt.Errorf("failed to convert %s to float: %w", envName, err)
			}
			newValue.Index(i).SetFloat(floatValue)

		case reflect.Complex64, reflect.Complex128:
			complexValue, err := strconv.ParseComplex(strVal, elemType.Bits())
			if err != nil {
				return fmt.Errorf("failed to convert %s to complex: %w", envName, err)
			}
			newValue.Index(i).SetComplex(complexValue)

		case reflect.Bool:
			boolValue, err := strconv.ParseBool(strVal)
			if err != nil {
				return fmt.Errorf("error parsing env var %s: %w", envName, err)
			}
			newValue.Index(i).SetBool(boolValue)

		case reflect.Interface:
			newValue.Index(i).Set(reflect.ValueOf(strVal))
		default:
			return fmt.Errorf("unsupported slice/array element type: %s", elemType.Kind())
		}
	}

	// Set the final value
	fieldValue.Set(newValue)
	return nil
}

func setEnvVarMapValues(fieldValue reflect.Value, envName string, envValue string, tagProp tagProperties) error {
	// set the field value to the env var value
	mapValues := strings.Split(envValue, tagProp.Delimiter)
	lenMapValues := len(mapValues)
	//replace starting braces and ending braces
	mapValues[0] = strings.ReplaceAll(mapValues[0], "{", "")
	mapValues[lenMapValues-1] = strings.ReplaceAll(mapValues[lenMapValues-1], "}", "")
	newMap := reflect.MakeMapWithSize(fieldValue.Type(), lenMapValues)

	for _, pair := range mapValues {
		keyValue := strings.SplitN(pair, ":", 2)
		if len(keyValue) != 2 {
			return fmt.Errorf("invalid map entry for %s: %s", envName, pair)
		}

		key := strings.TrimSpace(keyValue[0])
		value := strings.TrimSpace(keyValue[1])

		mapKey := reflect.New(fieldValue.Type().Key()).Elem()
		mapValue := reflect.New(fieldValue.Type().Elem()).Elem()

		// Set key
		switch mapKey.Kind() {
		case reflect.String:
			mapKey.SetString(key)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			intKey, err := strconv.ParseInt(key, 10, mapKey.Type().Bits())
			if err != nil {
				return fmt.Errorf("failed to convert map key %s to int: %w", key, err)
			}
			mapKey.SetInt(intKey)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			uintKey, err := strconv.ParseUint(key, 10, mapKey.Type().Bits())
			if err != nil {
				return fmt.Errorf("failed to convert map key %s to uint: %w", key, err)
			}
			mapKey.SetUint(uintKey)
		case reflect.Float32, reflect.Float64:
			floatKey, err := strconv.ParseFloat(key, mapKey.Type().Bits())
			if err != nil {
				return fmt.Errorf("failed to convert map key %s to float: %w", key, err)
			}
			mapKey.SetFloat(floatKey)
		case reflect.Complex64, reflect.Complex128:
			complexKey, err := strconv.ParseComplex(key, mapKey.Type().Bits())
			if err != nil {
				return fmt.Errorf("failed to convert map key %s to complex: %w", key, err)
			}
			mapKey.SetComplex(complexKey)
		case reflect.Bool:
			boolKey, err := strconv.ParseBool(key)
			if err != nil {
				return fmt.Errorf("failed to convert map key %s to bool: %w", key, err)
			}
			mapKey.SetBool(boolKey)
		case reflect.Interface:
			mapKey.Set(reflect.ValueOf(key))
		default:
			return fmt.Errorf("unsupported map key type: %s", mapKey.Kind())
		}

		// Set value
		switch mapValue.Kind() {
		case reflect.String:
			mapValue.SetString(value)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			intValue, err := strconv.ParseInt(value, 10, mapValue.Type().Bits())
			if err != nil {
				return fmt.Errorf("failed to convert map value %s to int: %w", value, err)
			}
			mapValue.SetInt(intValue)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			uintValue, err := strconv.ParseUint(value, 10, mapValue.Type().Bits())
			if err != nil {
				return fmt.Errorf("failed to convert map value %s to uint: %w", value, err)
			}
			mapValue.SetUint(uintValue)
		case reflect.Float32, reflect.Float64:
			floatValue, err := strconv.ParseFloat(value, mapValue.Type().Bits())
			if err != nil {
				return fmt.Errorf("failed to convert map value %s to float: %w", value, err)
			}
			mapValue.SetFloat(floatValue)
		case reflect.Bool:
			boolValue, err := strconv.ParseBool(value)
			if err != nil {
				return fmt.Errorf("failed to convert map value %s to bool: %w", value, err)
			}
			mapValue.SetBool(boolValue)
		case reflect.Complex64, reflect.Complex128:
			complexValue, err := strconv.ParseComplex(value, mapValue.Type().Bits())
			if err != nil {
				return fmt.Errorf("failed to convert map value %s to complex: %w", value, err)
			}
			mapValue.SetComplex(complexValue)
		case reflect.Interface:
			mapValue.Set(reflect.ValueOf(value))
		default:
			return fmt.Errorf("unsupported map value type: %s", mapValue.Kind())
		}

		newMap.SetMapIndex(mapKey, mapValue)
	}

	fieldValue.Set(newMap)
	return nil
}

func checkAndSetTagPropRequired(property string, tagProp *tagProperties) {
	if !strings.Contains(strings.ToLower(property), "required") {
		return
	}
	// check if the required field is set to true or false
	if strings.Contains(property, "=") {
		property = strings.Split(property, "=")[1]
		property = strings.TrimSpace(property)
		property = strings.ToLower(property)
	}
	if strings.Contains(property, "true") {
		tagProp.setRequired(true)
	} else if strings.Contains(property, "false") {
		tagProp.setRequired(false)
	} else {
		tagProp.setRequired(true)
	}

}

func checkAndSetTagPropDefaultValue(property string, tagProp *tagProperties) {
	if !strings.Contains(strings.ToLower(property), "default") {
		return
	}
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

func checkAndSetTagPropDelimiterForSliceOrArray(property string, tagProp *tagProperties) {
	if !strings.Contains(strings.ToLower(property), "delimiter") {
		return
	}
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
			tagProp.setDelimiter(property)
		}
	}
}

func cehckAndSetIsStringForByteOrRuneArray(property string, tagProp *tagProperties) {
	if !strings.Contains(strings.ToLower(property), "isstring") {
		return
	}
	// check if the required field is set to true or false
	if strings.Contains(property, "=") {
		property = strings.Split(property, "=")[1]
		property = strings.TrimSpace(property)
		property = strings.ToLower(property)
	}
	if strings.Contains(property, "true") {
		tagProp.setIsString()
	} else if strings.Contains(property, "false") {
		return
	} else {
		tagProp.setIsString()
	}
}

func splitTagRespectingQuotes(tag string) []string {
	var parts []string
	var part strings.Builder
	inQuotes := false
	quoteChar := byte(0)

	for i := 0; i < len(tag); i++ {
		c := tag[i]
		if (c == '\'' || c == '"') && (i == 0 || tag[i-1] != '\\') {
			if inQuotes {
				if c == quoteChar {
					inQuotes = false
				}
			} else {
				inQuotes = true
				quoteChar = c
			}
		}

		if c == ',' && !inQuotes {
			parts = append(parts, strings.TrimSpace(part.String()))
			part.Reset()
		} else {
			part.WriteByte(c)
		}
	}
	if part.Len() > 0 {
		parts = append(parts, strings.TrimSpace(part.String()))
	}
	return parts
}
