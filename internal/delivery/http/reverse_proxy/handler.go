package http_reverseproxy_handler

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"go-waf/config"
	"go-waf/internal/interface/service"
	"go-waf/pkg/logger"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

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

func (h *Handler) ReverseProxy(c *gin.Context, cacheHandler service.CacheInterface) {
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

	// var data CacheHandler
	data, ok := getCache.(map[string]interface{})
	if !ok {
		logger.Logger("[warn] cache is not a map[string]interface{}. This is: ", reflect.TypeOf(getCache)).Warn()
		h.cacheDriver.Remove(url)
		h.FetchData(c)
		return
	}

	var cacheData CacheHandler
	cacheHeaders, _ := data["headers"].(map[string]interface{})
	cacheData.CacheData, _ = base64.StdEncoding.DecodeString(data["data"].(string))

	for key, headers := range cacheHeaders {
		header := headers.([]interface{})
		if len(header) > 0 {
			c.Header(key, header[0].(string))
		}
	}

	ttl, _ := h.cacheDriver.GetTTL(url)
	ttl = time.Duration(h.config.CACHE_TTL) - (ttl / time.Second)
	if ttl < 0 {
		h.cacheDriver.Remove(url)
	}

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
	}

	proxy.ModifyResponse = func(r *http.Response) error {
		defer r.Body.Close()
		body, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Logger(err).Warn()
			return err
		}

		hostDestination, err := url.Parse(h.config.HOST_DESTINATION)
		if err != nil {
			logger.Logger(err).Warn()
			return err
		}

		body = bytes.ReplaceAll(
			body,
			[]byte(h.config.HOST_DESTINATION),
			[]byte(fmt.Sprintf("%s://%s", hostDestination.Scheme, h.config.HOST)),
		)

		r.Body = io.NopCloser(bytes.NewReader(body))
		r.ContentLength = int64(len(body))
		r.Header.Set("Content-Length", strconv.Itoa(len(body)))

		if h.config.USE_CACHE && r.StatusCode == 200 &&
			(c.Request.Method == "GET" || c.Request.Method == "HEAD") &&
			!strings.Contains(r.Header.Get("Cache-Control"), "max-age=0") {
			cacheData := &CacheHandler{
				CacheURL:     r.Request.URL.String(),
				CacheHeaders: r.Header,
				CacheData:    body,
			}
			logger.Logger("[debug] Set new cache", r.Request.URL.String()).Debug()
			h.cacheDriver.Set(r.Request.URL.String(), cacheData, time.Duration(h.config.CACHE_TTL)*time.Second)
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
