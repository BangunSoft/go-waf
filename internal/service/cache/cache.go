package service_cache

import (
	"context"
	"crypto/md5"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"time"

	"github.com/jahrulnr/go-waf/config"
	"github.com/jahrulnr/go-waf/internal/interface/repository"
	service "github.com/jahrulnr/go-waf/internal/interface/service"
	file_cache "github.com/jahrulnr/go-waf/internal/repository/file"
	memory_cache "github.com/jahrulnr/go-waf/internal/repository/memory"
	redis_cache "github.com/jahrulnr/go-waf/internal/repository/redis"
	"github.com/jahrulnr/go-waf/pkg/logger"

	"github.com/redis/go-redis/v9"
)

type CacheService struct {
	config *config.Config
	driver repository.CacheInterface
	key    string
}

// NewCacheService initializes a new CacheService based on the provided configuration.
func NewCacheService(config *config.Config) service.CacheInterface {
	var driver repository.CacheInterface
	ctx := context.Background() // Create a context for Redis operations

	switch config.CACHE_DRIVER {
	case "redis":
		rds := redis.NewClient(&redis.Options{
			Addr:         config.REDIS_ADDR,
			Username:     config.REDIS_USER,
			Password:     config.REDIS_PASS,
			DB:           config.REDIS_DB, // use default DB
			PoolSize:     50,
			DialTimeout:  30 * time.Second,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
		})
		driver = redis_cache.NewCache(ctx, rds)
	case "file":
		cachePath := "cache/"
		if err := os.MkdirAll(cachePath, 0755); err != nil {
			logger.Logger("[Fatal] Create cache path error.", err).Fatal()
		}
		driver = file_cache.NewFileCache(cachePath)
	default:
		driver = memory_cache.NewCache()
	}

	return &CacheService{
		config: config,
		driver: driver,
		key:    "gowaf-",
	}
}

// SetKey sets the base key prefix for cache keys.
func (s *CacheService) SetKey(key string) {
	s.key = "gowaf-" + key + "-"
}

// generateKey generates a cache key based on the provided key.
func (s *CacheService) generateKey(key string) string {
	parseUrl, err := url.Parse(key)
	if err != nil {
		logger.Logger("[error] Failed to parse URL: ", err).Error()
		return key // Return the original key if parsing fails
	}

	key = s.key + parseUrl.Path
	if query := parseUrl.Query().Encode(); query != "" {
		key = key + "?" + query
	}

	// Define a regex that matches illegal characters
	re := regexp.MustCompile(`[\/\\\?\*\:\<\>\|\"\s\&]`)
	// Replace illegal characters with an underscore
	newKey := re.ReplaceAllString(key, "_")

	// Limit the length of the key
	if len(newKey) > 100 {
		newKey = newKey[:100] + "---md5hash---" + fmt.Sprintf("%x", md5.Sum([]byte(key[100:])))
	}

	logger.Logger("debug", "generated key: "+newKey).Debug()
	return newKey
}

// Set stores a value in the cache with a specified duration.
func (s *CacheService) Set(key string, value []byte, duration time.Duration) {
	generatedKey := s.generateKey(key)
	s.driver.Set(generatedKey, value, duration) // No error handling to avoid duplicate logging
}

// Get retrieves a value from the cache.
func (s *CacheService) Get(key string) ([]byte, bool) {
	generatedKey := s.generateKey(key)
	value, ok := s.driver.Get(generatedKey)
	if !ok {
		logger.Logger("[debug] Cache miss for key: " + generatedKey).Debug()
	}
	return value, ok
}

// Pop removes and returns a value from the cache.
func (s *CacheService) Pop(key string) ([]byte, bool) {
	generatedKey := s.generateKey(key)
	return s.driver.Pop(generatedKey)
}

// Remove deletes a value from the cache.
func (s *CacheService) Remove(key string) {
	generatedKey := s.generateKey(key)
	s.driver.Remove(generatedKey) // No error handling to avoid duplicate logging
}

// RemoveByPrefix deletes all values with the specified prefix from the cache.
func (s *CacheService) RemoveByPrefix(prefix string) {
	generatedKey := s.generateKey(prefix)
	s.driver.RemoveByPrefix(generatedKey) // No error handling to avoid duplicate logging
}

// GetTTL retrieves the remaining time-to-live for a cache key.
func (s *CacheService) GetTTL(key string) (time.Duration, bool) {
	generatedKey := s.generateKey(key)
	return s.driver.GetTTL(generatedKey)
}
