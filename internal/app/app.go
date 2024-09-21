package app

import (
	"go-waf/config"
	delivery_http "go-waf/internal/delivery/http"
	"go-waf/pkg/httpserver"
	"go-waf/pkg/logger"
	"log"
	"os"
	"time"
)

type App struct {
	config *config.Config

	notify chan os.Signal
}

func NewApp(config *config.Config) *App {
	app := &App{
		config: config,

		notify: make(chan os.Signal, 1),
	}

	return app
}

func (a *App) execute() {
	server := httpserver.NewHttpServer(a.config)
	router := delivery_http.NewHttpRouter(a.config)

	server.SetHandler(router.GetHandler())
	server.Start()
}

func (a *App) Start() {
	start := time.Now()
	log.Println("[Info] HttpServer listen and serve at ", a.config.ADDR)
	a.execute()

	logger.Logger(map[string]any{
		"message":       "App stopped",
		"stopped_after": time.Since(start),
		"causer":        <-a.notify,
	}).Info()
}
