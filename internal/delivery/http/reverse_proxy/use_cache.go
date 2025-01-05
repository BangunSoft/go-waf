package http_reverseproxy_handler

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jahrulnr/go-waf/pkg/logger"
)

func (h *Handler) UseCache(c *gin.Context) {
	url := h.config.HOST_DESTINATION + c.Request.URL.String()

	// Set device key for cache if applicable
	if deviceKey := c.GetHeader("X-Device"); deviceKey != "" && h.config.DETECT_DEVICE && h.config.SPLIT_CACHE_BY_DEVICE {
		h.cacheDriver.SetKey(deviceKey)
	}

	// Attempt to get cache
	getCache, ok := h.cacheDriver.Get(url)
	if !ok {
		logger.Logger("[debug] cache not found", url).Debug()
		h.FetchData(c)
		return
	}

	var cacheData CacheHandler
	err := json.Unmarshal(getCache, &cacheData)
	if err != nil {
		logger.Logger("[debug] cannot cast cache data to CacheHandler, cache data type is ", reflect.TypeOf(getCache), ". Trying with map[string]interface{}").Debug()

		// Attempt to unmarshal into a map
		var data map[string]interface{}
		if err = json.Unmarshal(getCache, &data); err != nil {
			logger.Logger("[debug] I can't explain this error. err: ", err).Warn()
			return
		}

		// Decode cache data
		if encodedData, ok := data["data"].(string); ok {
			cacheData.CacheData, _ = base64.StdEncoding.DecodeString(encodedData)
		}

		// Set headers from cache
		if cacheHeaders, ok := data["headers"].(map[string]interface{}); ok {
			for key, headers := range cacheHeaders {
				if headerList, ok := headers.([]interface{}); ok && len(headerList) > 0 {
					c.Header(key, headerList[0].(string))
				}
			}
		}
	} else {
		logger.Logger("[debug] cast cache data to CacheHandler").Debug()

		// Set headers from cacheData
		for key, headers := range cacheData.CacheHeaders {
			if len(headers) > 0 {
				c.Header(key, headers[0])
			}
		}
	}

	// Check and manage TTL
	ttl, _ := h.cacheDriver.GetTTL(url)
	ttl = time.Duration(h.config.CACHE_TTL) - (ttl / time.Second)
	if ttl < 0 {
		h.cacheDriver.Remove(url)
	}

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
	c.Data(200, c.GetHeader("Content-Type"), cacheData.CacheData)
}
