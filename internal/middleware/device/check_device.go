package device

import (
	"github.com/gamebtc/devicedetector"
	"github.com/gin-gonic/gin"
	"github.com/jahrulnr/go-waf/config"
	"github.com/jahrulnr/go-waf/pkg/logger"
)

type Device struct {
	config *config.Config
}

func NewCheckDevice(config *config.Config) *Device {
	return &Device{
		config: config,
	}
}

func (m *Device) SendHeader() gin.HandlerFunc {
	detector, err := devicedetector.NewDeviceDetector("config/devices")

	return func(c *gin.Context) {
		if err != nil {
			logger.Logger("warn", err.Error()).Warn()
			c.Next()
			return
		}

		userAgent := c.Request.Header.Get("User-Agent")
		if len(userAgent) == 0 {
			userAgent = "desktop"
			c.Request.Header.Add("X-Device", "desktop")
			c.Next()
			return
		}

		info := detector.Parse(userAgent)
		if info.IsMobile() {
			c.Request.Header.Add("X-Device", "mobile")
		} else {
			c.Request.Header.Add("X-Device", "desktop")
		}

		c.Next()
	}
}
