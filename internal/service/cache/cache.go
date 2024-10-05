package service_cache

import (
	"context"
	"go-waf/config"
	"go-waf/internal/interface/service"
	file_cache "go-waf/internal/service/cache/file"
	memory_cache "go-waf/internal/service/cache/memory"
	redis_cache "go-waf/internal/service/cache/redis"
	"go-waf/pkg/logger"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

type CacheService struct {
	config config.Config
	driver service.CacheInterface
}

func NewCacheService(config config.Config) *CacheService {
	var driver service.CacheInterface
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
		cachePath := "cache"
		cacheStat, err := os.Stat(cachePath)
		if err != nil {
			logger.Logger("Cache path does'nt exists. Create cache path...").Info()
			err = os.MkdirAll(cachePath, 0755)
			if err != nil {
				logger.Logger("Create cache path error.", err).Fatal()
			}
		}

		if !cacheStat.IsDir() {
			logger.Logger("Cache path is file!").Fatal()
		}

		driver = file_cache.NewFileCache(cachePath)
	default:
		driver = memory_cache.NewCache[string, interface{}]()
	}

	return &CacheService{
		config: config,
		driver: driver,
	}
}

func (s *CacheService) Set(key string, value interface{}, duration time.Duration) {
	s.driver.Set(key, value, duration)
}

func (s *CacheService) Get(key string) (interface{}, bool) {
	return s.driver.Get(key)
}

func (s *CacheService) Pop(key string) (interface{}, bool) {
	return s.driver.Pop(key)
}

func (s *CacheService) Remove(key string) {
	s.driver.Remove(key)
}
