package http_reverseproxy_handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jahrulnr/go-waf/pkg/logger"
	"github.com/vmihailenco/msgpack"
)

// cacheResponse caches the response data using MessagePack.
func (h *Handler) cacheResponse(c *gin.Context, url string, headers http.Header, body []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if deviceKey := h.getDeviceKey(c); deviceKey != "" {
		h.cacheDriver.SetKey(deviceKey)
	}

	cacheData := &CacheHandler{
		CacheURL:     url,
		CacheHeaders: headers,
		CacheData:    body,
	}
	data, err := msgpack.Marshal(cacheData) // Use MessagePack for serialization
	if err != nil {
		logger.Logger("[error] Failed to marshal cache data: ", err).Error()
		return
	}

	logger.Logger("[debug]", "Set new cache "+url).Debug()
	h.cacheDriver.Set(url, data, time.Duration(h.config.CACHE_TTL)*time.Second)
}

// UseCache retrieves cached data or fetches it if not found.
func (h *Handler) UseCache(c *gin.Context) {
	url := h.config.HOST_DESTINATION + c.Request.URL.String()

	if deviceKey := h.getDeviceKey(c); deviceKey != "" {
		h.cacheDriver.SetKey(deviceKey)
	}

	getCache, ok := h.cacheDriver.Get(url)
	if !ok {
		logger.Logger("[debug] cache not found", url).Debug()
		h.FetchData(c)
		return
	}

	var cacheData CacheHandler
	if err := msgpack.Unmarshal(getCache, &cacheData); err != nil { // Use MessagePack for deserialization
		logger.Logger("[error] Failed to unmarshal cache data: ", err).Error()
		go h.cacheDriver.Remove(url)
		h.FetchData(c)
		return
	}

	// Set headers from cacheData
	for key, headers := range cacheData.CacheHeaders {
		if len(headers) > 0 {
			c.Header(key, headers[0])
		}
	}

	// Check and manage TTL
	ttl, _ := h.cacheDriver.GetTTL(url)
	ttl = time.Duration(h.config.CACHE_TTL) - (ttl / time.Second)
	go func() {
		if ttl <= 0 {
			h.cacheDriver.Remove(url)
		}
	}()

	// Manage headers
	if h.config.ENABLE_GZIP {
		c.Header("Accept-Encoding", "")
		c.Header("Vary", "")
	}
	c.Header("Via", "")
	c.Header("Server", "")
	c.Header("X-Varnish", "")
	c.Header("X-Cache", "HIT")
	c.Header("X-Age", fmt.Sprintf("%d", ttl))

	// Send cached data
	c.Data(http.StatusOK, c.GetHeader("Content-Type"), cacheData.CacheData)
}
