package http_clearcache_handler

import (
	"go-waf/config"
	"go-waf/internal/interface/service"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	config *config.Config

	cacheDriver service.CacheInterface
}

func NewHttpHandler(config *config.Config, handler *gin.Engine, cacheDriver service.CacheInterface) *Handler {
	return &Handler{
		config:      config,
		cacheDriver: cacheDriver,
	}
}

func (h *Handler) Clear(c *gin.Context) {
	fullUrl := h.config.HOST_DESTINATION + c.Request.URL.String()
	parsedURL, _ := url.Parse(fullUrl)
	query := parsedURL.Query()
	query.Del("is_prefix")
	parsedURL.RawQuery = query.Encode()
	fullUrl = parsedURL.String()
	isPrefix := strings.ToLower(c.Query("is_prefix"))
	if isPrefix == "true" {
		h.cacheDriver.RemoveByPrefix(fullUrl)
	} else {
		h.cacheDriver.Remove(fullUrl)
	}

	c.JSON(200, map[string]interface{}{
		"status": "OK",
	})
}
