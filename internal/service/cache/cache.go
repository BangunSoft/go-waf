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

func NewCacheService(config *config.Config) service.CacheInterface {
	var driver repository.CacheInterface
	switch config.CACHE_DRIVER {
	case "redis":
		rds := redis.NewClient(&redis.Options{
			Addr:     config.REDIS_ADDR,
			Username: config.REDIS_USER,
			Password: config.REDIS_PASS,
			DB:       config.REDIS_DB, // use default DB
		})
		driver = redis_cache.NewCache(context.Background(), rds)
	case "file":
		cachePath := "cache/"
		_, err := os.Stat(cachePath)
		if err != nil {
			logger.Logger("[debug] Cache path does'nt exists. Create cache path...").Debug()
			err = os.MkdirAll(cachePath, 0755)
			if err != nil {
				logger.Logger("[Fatal] Create cache path error.", err).Fatal()
			}
		}

		cacheStat, _ := os.Stat(cachePath)
		if !cacheStat.IsDir() {
			logger.Logger("Cache path is file!").Fatal()
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

func (s *CacheService) SetKey(key string) {
	s.key = "gowaf-" + key + "-"
}

func (s *CacheService) generateKey(key string) string {
	parseUrl, err := url.Parse(key)
	if err == nil {
		key = s.key + parseUrl.Path
		if query := parseUrl.Query().Encode(); query != "" {
			key = key + "?" + query
		}
	}

	// Define a regex that matches illegal characters
	re := regexp.MustCompile(`[\/\\\?\*\:\<\>\|\"\s\&]`)
	// Replace illegal characters with an underscore
	newKey := re.ReplaceAllString(key, "_")

	// linux limiting file name. So, we make it short.
	if len(newKey) > 100 {
		newKey = newKey[:100] + "---md5hash---" + fmt.Sprintf("%x", md5.Sum([]byte(key[100:])))
	}

	logger.Logger("debug", "generated key: "+newKey).Debug()
	return newKey
}

func (s *CacheService) Set(key string, value []byte, duration time.Duration) {
	generatedKey := s.generateKey(key)
	s.driver.Set(generatedKey, value, duration)
}

func (s *CacheService) Get(key string) ([]byte, bool) {
	generatedKey := s.generateKey(key)
	return s.driver.Get(generatedKey)
}

func (s *CacheService) Pop(key string) ([]byte, bool) {
	generatedKey := s.generateKey(key)
	return s.driver.Pop(generatedKey)
}

func (s *CacheService) Remove(key string) {
	generatedKey := s.generateKey(key)
	s.driver.Remove(generatedKey)
}

func (s *CacheService) RemoveByPrefix(prefix string) {
	generatedKey := s.generateKey(prefix)
	s.driver.RemoveByPrefix(generatedKey)
}

func (s *CacheService) GetTTL(key string) (time.Duration, bool) {
	generatedKey := s.generateKey(key)
	return s.driver.GetTTL(generatedKey)
}
