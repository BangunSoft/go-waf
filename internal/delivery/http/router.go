package delivery_http

import (
	"go-waf/config"
	http_reverseproxy_handler "go-waf/internal/delivery/http/reverse_proxy"
	"go-waf/internal/middleware/ratelimit"

	"github.com/gin-gonic/gin"
)

type Router struct {
	config  *config.Config
	handler *gin.Engine

	ratelimiter *ratelimit.RateLimit
}

func NewHttpRouter(config *config.Config) *Router {
	return &Router{
		config: config,

		handler:     gin.Default(),
		ratelimiter: ratelimit.NewRateLimit(config),
	}
}

func (h *Router) setRouter() {
	// ratelimiter
	if h.config.USE_RATELIMIT {
		// TODO: add redis as driver option
		h.ratelimiter.Driver("memory")
		h.handler.Use(h.ratelimiter.RateLimit())
	}

	// initial handler
	proxyHandler := http_reverseproxy_handler.NewHttpHandler(h.config, h.handler)

	// set handler
	h.handler.Any("/*path", func(ctx *gin.Context) {
		if ctx.Param("path") != "/ping" {
			proxyHandler.ReverseProxy(ctx)
			return
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
