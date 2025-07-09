# envarfig-go

`envarfig-go` is a lightweight Go library for managing environment variables with ease. It provides a simple way to load and parse environment variables into Go structs, supporting features like default values, type conversion, and more.

[![Go Reference](https://pkg.go.dev/badge/github.com/lordvader501/envarfig-go.svg)](https://pkg.go.dev/github.com/lordvader501/envarfig-go)

## Features

- Load environment variables into Go structs.
- Support for default values and required fields.
- Type-safe parsing for common data types (e.g., `int`, `string`, `bool`, `uint`).
- Optional `.env` file loading using `godotenv`.
- Customizable settings for environment variable loading.
- Error handling for invalid or missing environment variables.

## Installation

To install the package, use:

```bash
go get github.com/lordvader501/envarfig-go
```

## Usage

### Basic Example

```go
package main

import (
    "fmt"
    "github.com/lordvader501/envarfig-go"
)

type Config struct {
    Host string `env:"HOST"`
    Port int    `env:"PORT"`
}

func main() {
    var config Config
    err := envarfig.LoadEnv(&config)
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    fmt.Printf("Host: %s, Port: %d\n", config.Host, config.Port)
}
```

### Using `.env` Files

You can load environment variables from `.env` files:

```go
err := envarfig.LoadEnv(&config, envarfig.WithEnvFiles(".env"))
```

### Custom Settings

You can disable automatic `.env` file loading:

```go
err := envarfig.LoadEnv(&config, envarfig.WithAutoLoadEnv(false))
```

You can also enable or disable caching of configurations:

```go
err := envarfig.LoadEnv(&config, envarfig.WithCacheConfig(false))
```

By default, caching is enabled. Disabling caching ensures that the configuration is reloaded every time `LoadEnv` is called.

### Advanced Example with Default and Required Fields

```go
type Config struct {
    Host string `env:"HOST,default='localhost',required"`
    Port int    `env:"PORT,default='8080'"`
}

var config Config

err := envarfig.LoadEnv(&config)
if err != nil {
    fmt.Println("Error:", err)
}
fmt.Printf("Host: %s, Port: %d\n", config.Host, config.Port)
```

### Supported Data Types

`envarfig-go` supports a wide range of data types for environment variable parsing. Below are examples for each supported type.

#### String

```go
type Config struct {
    Host string `env:"HOST"`
}
```

#### Integer

```go
// all int type(int8, int16....)
type Config struct {
    Port int `env:"PORT"`
}
```

#### Float

```go
// float32 and float64
type Config struct {
    Threshold float64 `env:"THRESHOLD"`
}
```

#### Boolean

```go
type Config struct {
    Debug bool `env:"DEBUG"`
}
```

#### Complex Numbers

```go
// complex64 and complex128
type Config struct {
    ComplexVal complex128 `env:"COMPLEX_VAL, default='1+3i'"`
}
```

#### Arrays

You can use arrays with a fixed size. Use the `delimiter` tag to specify a custom delimiter.

```go
//default delimiter is ','
type Config struct {
    Ports [2]int `env:"PORTS,delimiter=';'"`
}
```

Environment Variable Example:

```
PORTS=8080;9090
```

#### Slices

Slices are supported for dynamic-length lists. Use the `delimiter` tag to specify a custom delimiter.

```go
type Config struct {
    Ports []int `env:"PORTS,delimiter=';'"`
}
```

Environment Variable Example:

```
PORTS=8080;9090;10010
```

#### Maps

Maps are supported with key-value pairs. Use the `delimiter` tag to specify a custom delimiter.

```go
type Config struct {
    Settings map[string]string `env:"SETTINGS,delimiter=';'"`
}
```

Environment Variable Example:

```
SETTINGS=key1:value1;key2:value2
```

#### Any (Interface{})

The `any` type can be used to store any value as a string.

```go
type Config struct {
    Value any `env:"VALUE"`
}
```

#### Example with Multiple Types

```go
type Config struct {
    Host       string            `env:"HOST"`
    Port       int               `env:"PORT"`
    Threshold  float64           `env:"THRESHOLD"`
    Debug      bool              `env:"DEBUG"`
    ComplexVal complex128        `env:"COMPLEX_VAL"`
    Ports      []int             `env:"PORTS,delimiter=';'"`
    Settings   map[string]string `env:"SETTINGS,delimiter=';'"`
    Value      any               `env:"VALUE"`
}
```

Environment Variable Example:

```
HOST=localhost
PORT=8080
THRESHOLD=0.75
DEBUG=true
COMPLEX_VAL=1+2i
PORTS=8080;9090;10010
SETTINGS=key1:value1;key2:value2
VALUE=dynamic_value
```

### Handling Unsupported Field Types

`envarfig-go` does not support certain field types, such as `struct` or other custom types, for environment variable parsing. If you attempt to use unsupported types, the library will return an error indicating the unsupported type.

#### Example

```go
type Config struct {
    UnsupportedField struct{} `env:"UNSUPPORTED_FIELD"`
}

var config Config
err := envarfig.LoadEnv(&config)
if err != nil {
    fmt.Println("Error:", err)
    // Output: Error: unsupported field type: struct
}
```

This ensures that you are aware of unsupported types during development and can handle them appropriately.

## API

### `LoadEnv`

```go
func LoadEnv[T any](envConfig *T, options ...option) error
```

- **`envConfig`**: A pointer to a struct where environment variables will be loaded.
- **`options`**: Optional settings for environment variable loading.

### Tag Syntax

- **`env`**: Specifies the environment variable name.
- **`default`**: Specifies a default value if the environment variable is not set.
- **`required`**: Marks the environment variable as required.
- **`delimiter`**: Delimiter for arrays value seperatior(default = ',')

Example:

```go
type Config struct {
    Host string `env:"HOST,default='localhost',required"`
}
```

## Testing

Run the tests using:

```bash
go test ./... -v
```

### Running Unit Tests

```bash
go test -tags=unit ./... -v
```

### Running Integration Tests

```bash
go test -tags=integration ./... -v
```

## Running all with cover Profile

```bash
go test ./... -race -tags="unit integration" -v -coverprofile="coverage.out"
go tool cover -html "coverage.out"
```

## Contributing

Contributions are welcome! Feel free to open issues or submit pull requests.

## License

This project is licensed under the [MIT License](./LICENSE).

## Acknowledgments

- [godotenv](https://github.com/joho/godotenv) for `.env` file support.
- [Testify](https://github.com/stretchr/testify) for testing utilities.
