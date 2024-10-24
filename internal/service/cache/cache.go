package service_cache

import (
	"context"
	"crypto/md5"
	"fmt"
	"go-waf/config"
	"go-waf/internal/interface/repository"
	service "go-waf/internal/interface/service"
	file_cache "go-waf/internal/repository/file"
	memory_cache "go-waf/internal/repository/memory"
	redis_cache "go-waf/internal/repository/redis"
	"go-waf/pkg/logger"
	"os"
	"regexp"
	"time"

	"github.com/redis/go-redis/v9"
)

type CacheService struct {
	config *config.Config
	driver repository.CacheInterface
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
	}
}

func (s *CacheService) generateKey(key string) string {
	// Define a regex that matches illegal characters
	re := regexp.MustCompile(`[\/\\\?\*\:\<\>\|\"\s\&]`)
	// Replace illegal characters with an underscore
	newKey := re.ReplaceAllString(key, "_")

	// linux limiting file name. So, we make it short.
	if len(newKey) > 100 {
		newKey = newKey[:100] + "---md5hash---" + fmt.Sprintf("%x", md5.Sum([]byte(key[100:])))
	}

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
