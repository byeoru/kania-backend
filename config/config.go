package config

import (
	"log"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
)

var (
	configInit     sync.Once
	configInstance *Config
)

type Config struct {
	Server struct {
		Port string
	}
	Database struct {
		DatabaseDriver string
		DatabaseSource string
	}
	Token struct {
		TokenSymmertricKey  string
		AccessTokenDuration time.Duration
	}
	Cookie struct {
		CookieDuration   int
		AccessCookieName string
	}
}

func LoadConfig(filePath string) {
	configInit.Do(func() {
		configInstance = new(Config)
		if _, err := toml.DecodeFile(filePath, &configInstance); err != nil {
			log.Panic(err)
		}
	})
}

func GetInstance() *Config {
	return configInstance
}
