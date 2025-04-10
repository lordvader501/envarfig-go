//go:build ignore

package main

import (
	"fmt"
	"os"

	"github.com/lordvader501/envarfig-go"
)

type Config struct {
	PORT uint16 `env:"PORT"`
	HOST string `env:"HOST, default=localhost:4000, required"`
}

var config Config

func main() {
	os.Setenv("PORT", "65535")
	err := envarfig.LoadEnv(&config, envarfig.WithAutoLoadEnv(false))
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(config, config.HOST)
}
