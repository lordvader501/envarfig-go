//go:build ignore

package main

import (
	"fmt"
	"os"

	"github.com/lordvader501/envarfig-go"
)

type Config struct {
	PORT int    `env:"PORT"`
	HOST string `env:"HOST, default=localhost, required"`
}

var config Config

func main() {

	os.Setenv("HOST", "localhost1")
	os.Setenv("PORT", "hello")
	err := envarfig.GetEnvVar(&config)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(config, config.HOST)

}
