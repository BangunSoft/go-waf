package app

import (
	"fmt"
	"go-waf/config"
	delivery_http "go-waf/internal/delivery/http"
	service_cache "go-waf/internal/service/cache"
	"go-waf/pkg/httpserver"
	"go-waf/pkg/logger"
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
	cacheHandler := service_cache.NewCacheService(a.config)
	router := delivery_http.NewHttpRouter(a.config, cacheHandler)

	server.SetHandler(router.GetHandler())
	server.Start()

	err := <-server.Notify()
	if err != nil {
		logger.Logger(err).Error()
		a.notify <- os.Kill
	}
}

func (a *App) Start() {
	start := time.Now()
	logger.Logger(fmt.Sprintf("[Info] HttpServer listen and serve at %s", a.config.ADDR)).Info()
	a.execute()

	logger.Logger(map[string]any{
		"message":       "App stopped",
		"stopped_after": time.Since(start),
		"causer":        <-a.notify,
	}).Info()
}
