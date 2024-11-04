package http_reverseproxy_handler

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/jahrulnr/go-waf/config"
	"github.com/jahrulnr/go-waf/internal/interface/service"
	"github.com/jahrulnr/go-waf/pkg/logger"

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

func (h *Handler) UseCache(c *gin.Context) {
	url := h.config.HOST_DESTINATION + c.Request.URL.String()
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

func (h *Handler) FetchData(c *gin.Context) {
	remote, err := url.Parse(h.config.HOST_DESTINATION)
	if err != nil {
		panic(err)
	}

	host := h.config.HOST
	if host == "" {
		host = c.Request.Host
	}

	proxy := httputil.NewSingleHostReverseProxy(remote)
	proxy.Director = func(req *http.Request) {
		req.Header = c.Request.Header
		req.Host = host
		req.URL.Scheme = remote.Scheme
		req.URL.Host = remote.Host
		req.URL.Path = c.Param("path")
		req.Header.Del("Accept-Encoding")
	}

	proxy.ModifyResponse = func(r *http.Response) error {
		defer r.Body.Close()
		body, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Logger(err).Warn()
			return err
		}

		scheme := c.Request.URL.Scheme
		if scheme == "" {
			scheme = "http"
		}

		// replace scheme://host to local host
		body = bytes.ReplaceAll(
			body,
			[]byte(h.config.HOST_DESTINATION),
			[]byte(fmt.Sprintf("%s://%s", scheme, c.Request.Host)),
		)

		// replace //host to local host
		body = bytes.ReplaceAll(
			body,
			[]byte(fmt.Sprintf("\"//%s", c.Request.Host)),
			[]byte(fmt.Sprintf("\"%s://%s", scheme, c.Request.Host)),
		)
		body = bytes.ReplaceAll(
			body,
			[]byte(fmt.Sprintf("'//%s", c.Request.Host)),
			[]byte(fmt.Sprintf("'%s://%s", scheme, c.Request.Host)),
		)

		r.Body = io.NopCloser(bytes.NewReader(body))
		r.ContentLength = int64(len(body))

		// remove duplicate header
		if h.config.ENABLE_GZIP {
			r.Header.Del("Accept-Encoding")
			r.Header.Del("Vary")
		}

		// modify header
		r.Header.Set("Content-Length", strconv.Itoa(len(body)))
		r.Header.Del("Via")
		r.Header.Del("Server")
		r.Header.Del("X-Varnish")

		if h.config.USE_CACHE && r.StatusCode == 200 &&
			(c.Request.Method == "GET" || c.Request.Method == "HEAD") &&
			!strings.Contains(r.Header.Get("Cache-Control"), "max-age=0") &&
			!strings.Contains(r.Header.Get("Cache-Control"), "no-cache") {
			cacheData := &CacheHandler{
				CacheURL:     r.Request.URL.String(),
				CacheHeaders: r.Header,
				CacheData:    body,
			}
			data, _ := json.Marshal(cacheData)
			logger.Logger("[debug] Set new cache", r.Request.URL.String()).Debug()
			h.cacheDriver.Set(r.Request.URL.String(), data, time.Duration(h.config.CACHE_TTL)*time.Second)
			r.Header.Set("X-Cache", "MISS")
		}

		return nil
	}

	proxy.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: h.config.IGNORE_SSL_VERIFY,
			MinVersion:         tls.VersionTLS10,
		},
	}

	proxy.ServeHTTP(c.Writer, c.Request)
}
