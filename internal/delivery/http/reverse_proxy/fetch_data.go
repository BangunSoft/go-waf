package http_reverseproxy_handler

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jahrulnr/go-waf/pkg/logger"
)

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
			[]byte(fmt.Sprintf("\"//%s", h.config.HOST_DESTINATION)),
			[]byte(fmt.Sprintf("\"%s://%s", scheme, c.Request.Host)),
		)
		body = bytes.ReplaceAll(
			body,
			[]byte(fmt.Sprintf("'//%s", h.config.HOST_DESTINATION)),
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
			if deviceKey := c.GetHeader("X-Device"); deviceKey != "" && h.config.DETECT_DEVICE && h.config.SPLIT_CACHE_BY_DEVICE {
				h.cacheDriver.SetKey(deviceKey)
			}
			cacheData := &CacheHandler{
				CacheURL:     r.Request.URL.String(),
				CacheHeaders: r.Header,
				CacheData:    body,
			}
			data, _ := json.Marshal(cacheData)
			logger.Logger("[debug]", "Set new cache"+r.Request.URL.String()).Debug()
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
