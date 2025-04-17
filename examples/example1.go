//go:build ignore

package main

import (
	"fmt"

	"github.com/lordvader501/envarfig-go"
)

type Config struct {
	MapValues   map[string]any `env:"PORT, default='{ hello : 123, hi : a234, foo : 234 }', required=false"`
	ArrayValues [3]string      `env:"HOST, default='apple;banana;orange', delimiter=';', required"`
	//default delimeter is `,` for array/slices values
	SliceValues []string `env:"HOST, default='apple,banana,orange', required"`
}

var config Config

func main() {
	if err := envarfig.LoadEnv(&config, envarfig.WithAutoLoadEnv(false)); err != nil {
		fmt.Println(err)
	}
	fmt.Println(config)
}
