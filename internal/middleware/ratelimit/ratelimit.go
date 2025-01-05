package ratelimit

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jahrulnr/go-waf/config"
	"github.com/jahrulnr/go-waf/pkg/logger"

	ratelimit "github.com/JGLTechnologies/gin-rate-limit"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type RateLimit struct {
	config  *config.Config
	page429 []byte

	driver string
	store  ratelimit.Store
	prefix string

	rate  time.Duration
	limit uint
}

func NewRateLimit(config *config.Config) *RateLimit {
	return &RateLimit{
		config:  config,
		page429: nil,
	}
}

func (s *RateLimit) initialize() {
	s.rate = time.Duration(s.config.RATELIMIT_SECOND) * time.Second
	s.limit = s.config.RATELIMIT_MAX

	file, err := os.OpenFile("views/429.html", os.O_RDONLY, 0600)
	if err != nil {
		logger.Logger(err).Warn()
		return
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	var page bytes.Buffer
	_, err = io.Copy(&page, reader)
	if err != nil {
		logger.Logger(err).Warn()
		return
	}

	s.page429 = page.Bytes()
}

func (s *RateLimit) Driver(driver string) {
	s.driver = strings.ToLower(driver)
}

func (s *RateLimit) keyFunc(c *gin.Context) string {
	return fmt.Sprintf("%s_%s", s.prefix, c.ClientIP())
}

func (s *RateLimit) errorHandler(c *gin.Context, info ratelimit.Info) {
	if s.page429 == nil {
		c.String(http.StatusTooManyRequests, "429 | Too many request.")
		c.Abort()
		return
	}

	c.Data(http.StatusTooManyRequests, "text/html", s.page429)
	c.Abort()
}

func (s *RateLimit) RateLimit() gin.HandlerFunc {
	s.initialize()
	switch s.driver {
	case "redis":
		s.store = ratelimit.RedisStore(&ratelimit.RedisOptions{
			Rate:  s.rate,
			Limit: s.limit,
			RedisClient: redis.NewClient(&redis.Options{
				Addr:     s.config.REDIS_ADDR,
				Username: s.config.REDIS_USER,
				Password: s.config.REDIS_PASS,
				DB:       s.config.REDIS_DB, // use default DB
			}),
			PanicOnErr: false,
		})
	default: // default in memory
		s.store = ratelimit.InMemoryStore(&ratelimit.InMemoryOptions{
			Rate:  s.rate,
			Limit: s.limit,
		})
	}

	middleware := ratelimit.RateLimiter(s.store, &ratelimit.Options{
		ErrorHandler: s.errorHandler,
		KeyFunc:      s.keyFunc,
	})

	return middleware
}
