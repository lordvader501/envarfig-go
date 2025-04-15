//go:build ignore

package main

import (
	"fmt"

	"github.com/lordvader501/envarfig-go"
)

type Config struct {
	PORT map[string]any `env:"PORT, default='{ hello : 123, hi : a234, foo : 234 }', required=false"`
	HOST string         `env:"HOST, default=localhost:4000, required"`
}

var config Config

func main() {
	// os.Setenv("PORT", "8080,8090")
	err := envarfig.LoadEnv(&config, envarfig.WithAutoLoadEnv(false))
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(config)
	fmt.Printf("%v\n", map[any]any{complex64(1 + 2i): complex64(2 + 3i)})
}
