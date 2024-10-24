package main

import (
	"io"
	"os"
	"strings"

	"github.com/jahrulnr/go-waf/config"
	"github.com/jahrulnr/go-waf/internal/app"
	"github.com/jahrulnr/go-waf/pkg/logger"

	"github.com/gin-gonic/gin"
)

func main() {
	// load config from os or from .env file
	config := config.Load()

	// initialize logger based on config
	logger.SetLevel(config.LOG_LEVEL)
	if config.LOG_FILE != "" {
		logger.SetOutput(loggerOutput(config.LOG_FILE))
	}

	// set gin mode based on config
	gin.SetMode(strings.ToLower(config.GIN_MODE))

	// register app functional with config
	app := app.NewApp(config)

	// start app
	app.Start()
}

func loggerOutput(path string) io.Writer {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		logger.Logger(map[string]any{
			"message": "Fail to create log file",
			"causer":  err,
		}).Fatal()
	}

	return file
}
