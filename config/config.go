package config

import (
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	ADDR string `env:"ADDR" env-default:":8080"`

	HOST              string `env:"HOST"`
	HOST_DESTINATION  string `env:"HOST_DESTINATION" env-default:"https://www.google.com"`
	IGNORE_SSL_VERIFY bool   `env:"IGNORE_SSL_VERIFY" env-default:"false"`

	USE_SSL  bool   `env:"USE_SSL" env-default:"false"`
	SSL_CERT string `env:"SSL_CERT"`
	SSL_KEY  string `env:"SSL_KEY"`

	USE_RATELIMIT    bool `env:"USE_RATELIMIT" env-default:"false"`
	RATELIMIT_SECOND int  `env:"RATELIMIT_SECOND" env-default:"1"`
	RATELIMIT_MAX    uint `env:"RATELIMIT_MAX" env-default:"5"`

	USE_CACHE    string `env:"USE_CACHE" env-default:"false"`
	CACHE_TTL    int    `env:"CACHE_TTL" env-default:"1209600"`   // default 2 week
	CACHE_DRIVER string `env:"CACHE_DRIVER" env-default:"memory"` // TODO: Redis and File

	REDIS_ADDR string `env:"REDIS_ADDR" env-default:"localhost:6379"`
	REDIS_SSL  bool   `env:"REDIS_SSL" env-default:"false"`
	REDIS_USER string `env:"REDIS_USER"`
	REDIS_PASS string `env:"REDIS_PASS"`
	REDIS_DB   int    `env:"REDIS_DB" env-default:"0"`

	// debug
	GIN_MODE  string `env:"GIN_MODE" env-default:"debug"`
	LOG_LEVEL string `env:"LOG_LEVEL" env-default:"debug"`
	LOG_FILE  string `env:"LOG_FILE"`
}

func Load() *Config {
	env := ".env"
	conf := Config{}

	report := func(err error) {
		if err != nil {
			log.Println("[Warn] config: ", err)
		}
	}

	err := cleanenv.ReadEnv(&conf)
	report(err)

	if _, err := os.Stat(env); err == nil {
		err := cleanenv.ReadConfig(env, &conf)
		report(err)
	}

	return &conf
}
