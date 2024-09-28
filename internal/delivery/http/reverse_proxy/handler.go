package http_reverseproxy_handler

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"go-waf/config"
	"go-waf/pkg/logger"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	config *config.Config
}

func NewHttpHandler(config *config.Config, handler *gin.Engine) *Handler {
	return &Handler{
		config: config,
	}
}

func (h *Handler) ReverseProxy(c *gin.Context) {
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
