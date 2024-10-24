package delivery_http

import (
	"strings"

	"github.com/jahrulnr/go-waf/config"
	http_clearcache_handler "github.com/jahrulnr/go-waf/internal/delivery/http/clear_cache"
	http_reverseproxy_handler "github.com/jahrulnr/go-waf/internal/delivery/http/reverse_proxy"
	"github.com/jahrulnr/go-waf/internal/interface/service"
	"github.com/jahrulnr/go-waf/internal/middleware/ratelimit"
	"github.com/jahrulnr/go-waf/pkg/logger"

	"github.com/gin-gonic/gin"
)

type Router struct {
	config  *config.Config
	handler *gin.Engine

	rateLimiter  *ratelimit.RateLimit
	cacheHandler service.CacheInterface
}

func NewHttpRouter(config *config.Config, cacheHandler service.CacheInterface) *Router {
	return &Router{
		config: config,

		handler:      gin.Default(),
		rateLimiter:  ratelimit.NewRateLimit(config),
		cacheHandler: cacheHandler,
	}
}

func (h *Router) setRouter() {
	// this will used for clear cache
	h.handler.HandleMethodNotAllowed = h.config.USE_CACHE

	// ratelimiter
	if h.config.USE_RATELIMIT {
		// TODO: add redis as driver option
		if h.config.CACHE_DRIVER == "redis" {
			h.rateLimiter.Driver("redis")
		} else {
			h.rateLimiter.Driver("memory")
		}
		h.handler.Use(h.rateLimiter.RateLimit())
	}

	// initial handler
	proxyHandler := http_reverseproxy_handler.NewHttpHandler(h.config, h.handler, h.cacheHandler)
	clearCacheHandler := http_clearcache_handler.NewHttpHandler(h.config, h.handler, h.cacheHandler)

	// set handler
	h.handler.Any("/*path", func(ctx *gin.Context) {
		if ctx.Param("path") == "/ping" {
			ctx.String(200, "PONG")
		} else if h.config.USE_CACHE &&
			strings.EqualFold(ctx.Request.Method, h.config.CACHE_REMOVE_METHOD) {
			logger.Logger("[info] clear cache: ", ctx.Param("path")).Info()
			clearCacheHandler.Clear(ctx)
		} else {
			proxyHandler.ReverseProxy(ctx)
		}
	})

	h.handler.NoMethod(func(ctx *gin.Context) {
		if h.config.USE_CACHE &&
			strings.EqualFold(ctx.Request.Method, h.config.CACHE_REMOVE_METHOD) {
			logger.Logger("[info] clear cache: ", ctx.Param("path")).Info()
			clearCacheHandler.Clear(ctx)
		} else {
			ctx.String(404, "404 page not found")
		}
	})
}

func (h *Router) GetHandler() *gin.Engine {
	h.setRouter()

	// return handler
	return h.handler
}
