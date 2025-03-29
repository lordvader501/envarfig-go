# envarfig-go

`envarfig-go` is a lightweight Go library for managing environment variables with ease. It provides a simple way to load and parse environment variables into Go structs, supporting features like default values, type conversion, and more.

## Features

- Load environment variables into Go structs.
- Support for default values and required fields(to be added in future).
- Type-safe parsing for common data types (e.g., `int`, `string`, `bool`).
- Optional `.env` file loading using `godotenv`.
- Customizable settings for environment variable loading.

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
  err := envarfig.GetEnvVar(&config)
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
err := envarfig.GetEnvVar(&config, envarfig.WithEnvFiles(".env"))
```

### Custom Settings

You can disable automatic `.env` file loading:

```go
err := envarfig.GetEnvVar(&config, envarfig.WithAutoLoadEnv(false))
```

## API

### `GetEnvVar`

```go
func GetEnvVar[T any](envConfig *T, options ...option) error
```

- **`envConfig`**: A pointer to a struct where environment variables will be loaded.
- **`options`**: Optional settings for environment variable loading.

### Tag Syntax

- **`env`**: Specifies the environment variable name.
- **`default`**(to be added in future): Specifies a default value if the environment variable is not set.
- **`required`**(to be added in future): Marks the environment variable as required.

Example:

```go
type Config struct {
  Host string `env:"HOST,default=localhost,required"`
}
```

## Testing

Run the tests using:

```bash
go test ./... -v
```

## Contributing

Contributions are welcome! Feel free to open issues or submit pull requests.

## License

This project is licensed under the [MIT License](./LICENSE).

## Acknowledgments

- [godotenv](https://github.com/joho/godotenv) for `.env` file support.
- [Testify](https://github.com/stretchr/testify) for testing utilities.
