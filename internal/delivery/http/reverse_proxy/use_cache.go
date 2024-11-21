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

	if deviceKey := c.GetHeader("X-Device"); deviceKey != "" && h.config.DETECT_DEVICE && h.config.SPLIT_CACHE_BY_DEVICE {
		h.cacheDriver.SetKey(deviceKey)
	}
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
		cacheData = CacheHandler{}
		var data map[string]interface{}
		err = json.Unmarshal(getCache, &data)
		if err != nil {
			logger.Logger("[debug] I can't explain this error. err: ", err).Warn()
		}

		cacheHeaders, _ := data["headers"].(map[string]interface{})
		cacheData.CacheData, _ = base64.StdEncoding.DecodeString(data["data"].(string))

		// set header
		for key, headers := range cacheHeaders {
			header := headers.([]interface{})
			if len(header) > 0 {
				c.Header(key, header[0].(string))
			}
		}
	} else {
		logger.Logger("[debug] cast cache data to CacheHandler").Debug()
		// set header
		for key, headers := range cacheData.CacheHeaders {
			if len(headers) > 0 {
				c.Header(key, headers[0])
			}
		}
	}

	ttl, _ := h.cacheDriver.GetTTL(url)
	ttl = time.Duration(h.config.CACHE_TTL) - (ttl / time.Second)
	if ttl < 0 {
		h.cacheDriver.Remove(url)
	}

	// remove duplicate header
	if h.config.ENABLE_GZIP {
		c.Header("Accept-Encoding", "")
		c.Header("Vary", "")
	}

	// remove some header
	c.Header("Via", "")
	c.Header("Server", "")
	c.Header("X-Varnish", "")
	c.Header("X-Cache", "HIT")
	c.Header("X-Age", fmt.Sprintf("%d", ttl))
	c.Data(200, c.GetHeader("Content-Type"), cacheData.CacheData)
}
