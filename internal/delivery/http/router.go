package delivery_http

import (
	"go-waf/config"
	http_reverseproxy_handler "go-waf/internal/delivery/http/reverse_proxy"
	"go-waf/internal/interface/service"
	"go-waf/internal/middleware/ratelimit"

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

	// set handler
	h.handler.Any("/*path", func(ctx *gin.Context) {
		if ctx.Param("path") != "/ping" {
			proxyHandler.ReverseProxy(ctx, h.cacheHandler)
		} else {
			ctx.String(200, "PONG")
		}
	})
}

func (h *Router) GetHandler() *gin.Engine {
	h.setRouter()

	// return handler
	return h.handler
}
