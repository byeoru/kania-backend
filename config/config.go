package config

import (
	"os"

	"github.com/naoina/toml"
)

type Config struct {
	Server struct {
		Port string
	}
	Database struct {
		DatabaseDriver string
		DatabaseSource string
	}
}

func NewConfig(filePath string) *Config {
	c := new(Config)

	if file, err := os.Open(filePath); err != nil {
		panic(err)
	} else if err = toml.NewDecoder(file).Decode(c); err != nil {
		panic(err)
	} else {
		return c
	}
}
