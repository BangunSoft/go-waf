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

	USE_WAF            bool   `env:"USE_WAF" env-default:"true"`
	WAF_CONFIG         string `env:"WAF_CONFIG" env-default:"config/keywords.yml"`
	WAF_PROTECT_HEADER bool   `env:"WAF_PROTECT_HEADER" env-default:"true"`
	WAF_PROTECT_BODY   bool   `env:"WAF_PROTECT_BODY" env-default:"false"`

	USE_CACHE             bool   `env:"USE_CACHE" env-default:"false"`
	CACHE_TTL             int    `env:"CACHE_TTL" env-default:"1209600"` // default 2 week
	CACHE_DRIVER          string `env:"CACHE_DRIVER" env-default:"memory"`
	CACHE_REMOVE_METHOD   string `env:"CACHE_REMOVE_METHOD" env-default:"ban"` // example: curl -X BAN http://localhost:8080/blogs/?is_prefix=true
	CACHE_REMOVE_ALLOW_IP string `env:"CACHE_REMOVE_ALLOW_IP" env-default:"127.0.0.0/24"`

	DETECT_DEVICE         bool `env:"DETECT_DEVICE" env-default:"true"`
	SPLIT_CACHE_BY_DEVICE bool `env:"SPLIT_CACHE_BY_DEVICE" env-default:"true"`

	REDIS_ADDR string `env:"REDIS_ADDR" env-default:"localhost:6379"`
	REDIS_SSL  bool   `env:"REDIS_SSL" env-default:"false"`
	REDIS_USER string `env:"REDIS_USER"`
	REDIS_PASS string `env:"REDIS_PASS"`
	REDIS_DB   int    `env:"REDIS_DB" env-default:"0"`

	ENABLE_GZIP             bool  `env:"ENABLE_GZIP" env-default:"false"`
	GZIP_COMPRESSION_LEVEL  int   `env:"GZIP_COMPRESSION_LEVEL" env-default:"6"`
	GZIP_MIN_CONTENT_LENGTH int64 `env:"GZIP_MIN_CONTENT_LENGTH" env-default:"1024"`

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
