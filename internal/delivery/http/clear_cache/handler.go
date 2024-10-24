package http_clearcache_handler

import (
	"go-waf/config"
	"go-waf/internal/interface/service"
	service_allow_ip "go-waf/internal/service/allow_ip"
	"go-waf/pkg/logger"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	config *config.Config

	cacheDriver service.CacheInterface
	ipService   service.AllowIPInterface
}

func NewHttpHandler(config *config.Config, handler *gin.Engine, cacheDriver service.CacheInterface) *Handler {
	httpHandler := &Handler{
		config:      config,
		cacheDriver: cacheDriver,
	}

	httpHandler.ipService = service_allow_ip.NewAllowIP(config)
	return httpHandler
}

func (h *Handler) isAllowed(c *gin.Context) bool {
	clientIp := c.ClientIP()
	remoteIp := c.RemoteIP()

	return h.ipService.Check(clientIp) || h.ipService.Check(remoteIp)
}

func (h *Handler) Clear(c *gin.Context) {
	fullUrl := h.config.HOST_DESTINATION + c.Request.URL.String()
	if !h.isAllowed(c) {
		logger.Logger("[warn] IP ", c.ClientIP(), "|", c.RemoteIP(), " trying to clear ", fullUrl).Warn()
		c.JSON(400, map[string]interface{}{
			"status": "Bad Request",
		})
		return
	}

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
