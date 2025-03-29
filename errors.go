package envarfig

import "errors"

// errors
var (
	// Error if config is nil
	errNilConfig = errors.New("env config is nil")
	// Error if config is not a pointer to a struct
	errConfigNotPtrToStruct = errors.New("config must be a pointer to a struct")
	// Error if tag is not found
	errTagNotFound = errors.New("tag not found")
	// Error if env file is invalid type
	errInvalidEnvPathArgs = errors.New("invalid env path args")
)
