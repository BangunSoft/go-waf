package http_reverseproxy_handler

import (
	"github.com/jahrulnr/go-waf/config"
	"github.com/jahrulnr/go-waf/internal/interface/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	config *config.Config

	cacheDriver service.CacheInterface
}

type CacheHandler struct {
	CacheURL     string              `json:"url"`
	CacheHeaders map[string][]string `json:"headers"`
	CacheData    []byte              `json:"data"`
}

func NewHttpHandler(config *config.Config, handler *gin.Engine, cacheDriver service.CacheInterface) *Handler {
	return &Handler{
		config:      config,
		cacheDriver: cacheDriver,
	}
}

func (h *Handler) ReverseProxy(c *gin.Context) {
	if h.config.USE_CACHE &&
		(c.Request.Method == "GET" || c.Request.Method == "HEAD") {
		h.UseCache(c)
	} else {
		h.FetchData(c)
	}
}
