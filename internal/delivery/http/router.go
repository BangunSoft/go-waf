package delivery_http

import (
	"strings"

	"github.com/jahrulnr/go-waf/config"
	http_clearcache_handler "github.com/jahrulnr/go-waf/internal/delivery/http/clear_cache"
	http_reverseproxy_handler "github.com/jahrulnr/go-waf/internal/delivery/http/reverse_proxy"
	"github.com/jahrulnr/go-waf/internal/interface/service"
	"github.com/jahrulnr/go-waf/internal/middleware/device"
	"github.com/jahrulnr/go-waf/internal/middleware/ratelimit"
	"github.com/jahrulnr/go-waf/internal/middleware/waf"
	service_waf "github.com/jahrulnr/go-waf/internal/service/waf"
	"github.com/jahrulnr/go-waf/pkg/logger"
	"github.com/nanmu42/gzip"

	"github.com/gin-gonic/gin"
)

type Router struct {
	config  *config.Config
	handler *gin.Engine

	wafHandler   service.WAFInterface
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
	// var middlewareList []gin.HandlerFunc

	if h.config.USE_WAF {
		wafService := service_waf.NewWAFService(h.config, h.config.WAF_CONFIG)
		// middlewareList = append(middlewareList, waf.NewWAFMiddleware(wafService))
		h.handler.Use(waf.NewWAFMiddleware(wafService))
	}

	// this will used for clear cache
	h.handler.HandleMethodNotAllowed = h.config.USE_CACHE

	// ratelimiter
	if h.config.USE_RATELIMIT {
		if h.config.CACHE_DRIVER == "redis" {
			h.rateLimiter.Driver("redis")
		} else {
			h.rateLimiter.Driver("memory")
		}
		// middlewareList = append(middlewareList, h.rateLimiter.RateLimit())
		h.handler.Use(h.rateLimiter.RateLimit())
	}

	// gzip compress
	if h.config.ENABLE_GZIP {
		gzipHandler := func(c *gin.Context) {
			logger.Logger(c.Request.Header.Get("Accept-Encoding")).Debug()
			if strings.Contains(c.Request.Header.Get("Accept-Encoding"), "gzip") {
				gzip.NewHandler(gzip.Config{
					CompressionLevel: h.config.GZIP_COMPRESSION_LEVEL,
					MinContentLength: h.config.GZIP_MIN_CONTENT_LENGTH,
					RequestFilter: []gzip.RequestFilter{
						gzip.NewCommonRequestFilter(),
						gzip.DefaultExtensionFilter(),
					},
					ResponseHeaderFilter: []gzip.ResponseHeaderFilter{
						gzip.NewSkipCompressedFilter(),
						gzip.DefaultContentTypeFilter(),
					},
				}).Gin(c)
			}
		}

		// middlewareList = append(middlewareList, gzipHandler)
		h.handler.Use(gzipHandler)
	}

	if h.config.DETECT_DEVICE {
		deviceHandler := device.NewCheckDevice(h.config)
		// middlewareList = append(middlewareList, deviceHandler.SendHeader())
		h.handler.Use(deviceHandler.SendHeader())
	}

	// if len(middlewareList) > 0 {
	// 	h.handler.Use(middlewareList...)
	// }

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
