package ratelimit

import (
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
	config *config.Config

	driver string
	store  ratelimit.Store
	prefix string

	rate  time.Duration
	limit uint
}

func NewRateLimit(config *config.Config) *RateLimit {
	return &RateLimit{
		config: config,
	}
}

func (s *RateLimit) initialize() {
	s.rate = time.Duration(s.config.RATELIMIT_SECOND) * time.Second
	s.limit = s.config.RATELIMIT_MAX
}

func (s *RateLimit) Driver(driver string) {
	s.driver = strings.ToLower(driver)
}

func (s *RateLimit) keyFunc(c *gin.Context) string {
	return fmt.Sprintf("%s_%s", s.prefix, c.ClientIP())
}

func (s *RateLimit) errorHandler(c *gin.Context, info ratelimit.Info) {
	file, err := os.OpenFile("views/429.html", os.O_RDONLY, 0600)
	if err != nil {
		logger.Logger(err).Warn()
		c.String(http.StatusTooManyRequests, "429 | Too many request.")
		return
	}
	defer file.Close()

	page, err := io.ReadAll(file)
	logger.Logger(err).Fatal()

	c.Data(http.StatusTooManyRequests, "text/html", page)
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
