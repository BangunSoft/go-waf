package waf

import (
	"bytes"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/jahrulnr/go-waf/internal/interface/service"
	"github.com/jahrulnr/go-waf/pkg/logger"
)

type WAFMiddleware struct {
}

func NewWAFMiddleware(wafService service.WAFInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		request := &service.Request{
			IP:      c.Request.RemoteAddr,
			Path:    c.Request.RequestURI,
			Headers: make(map[string]string),
			Body:    []byte{}, // You can read the body if needed
		}

		// Read the request body
		if c.Request.Body != nil {
			body, err := io.ReadAll(c.Request.Body)
			if err == nil {
				request.Body = body
				// Reset the body so it can be read again later
				c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
			}
		}

		for key, value := range c.Request.Header {
			request.Headers[key] = value[0]
		}

		request.Headers["RequestURI"] = request.Path
		logger.Logger(request.Headers).Debug()
		response, err := wafService.HandleRequest(request)
		if err != nil {
			c.String(500, "Error: Internal Server Error")
			c.Abort()
			return
		}

		if response != nil {
			c.String(403, string(response.Body))
			c.Abort()
			return
		}

		c.Next()
	}
}
