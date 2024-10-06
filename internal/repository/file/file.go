package file_cache

import (
	"encoding/json"
	"fmt"
	"go-waf/pkg/logger"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type FileCache struct {
	cacheDir string
	mu       sync.RWMutex
}

// CacheItem represents an item stored in the cache, along with its expiration time.
type CacheItem struct {
	Value      []byte `json:"value"`
	Expiration int64  `json:"expiration"` // Unix timestamp
}

// NewFileCache creates a new FileCache instance with the specified directory.
func NewFileCache(cacheDir string) *FileCache {
	return &FileCache{cacheDir: cacheDir}
}

// Set adds a new item to the file cache with the specified key, value, and TTL.
func (c *FileCache) Set(key string, value []byte, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	serializedValue, err := json.Marshal(CacheItem{
		Value:      value,
		Expiration: time.Now().Add(ttl).Unix(), // Set the expiration time
	})
	if err != nil {
		logger.Logger("Error serializing value: ", err).Warn()
		return
	}

	cacheFilePath := c.getFilePath(key)
	err = os.WriteFile(cacheFilePath, serializedValue, 0644)
	if err != nil {
		logger.Logger("Error writing to cache file: ", err).Warn()
	}

	// Optionally, you can implement a mechanism to clean up expired files
	go c.scheduleCleanup(key, ttl)
}

// Get retrieves the value associated with the given key from the file cache.
func (c *FileCache) Get(key string) ([]byte, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cacheFilePath := c.getFilePath(key)
	data, err := os.ReadFile(cacheFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false // Key does not exist
		}
		logger.Logger("Error reading cache file: ", err).Warn()
		return nil, false
	}

	var item CacheItem
	err = json.Unmarshal(data, &item)
	if err != nil {
		logger.Logger("Error deserializing value: ", err).Error()
		return nil, false
	}

	// Check if the item is expired
	if time.Now().Unix() > item.Expiration {
		c.Remove(key) // Optionally remove expired item
		return nil, false
	}

	return item.Value, true
}

// Pop removes and returns the item with the specified key from the file cache.
func (c *FileCache) Pop(key string) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	cacheFilePath := c.getFilePath(key)
	data, err := os.ReadFile(cacheFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false // Key does not exist
		}
		logger.Logger("Error reading cache file: ", err).Error()
		return nil, false
	}

	err = os.Remove(cacheFilePath)
	if err != nil {
		logger.Logger("Error removing cache file: ", err).Error()
	}

	var item CacheItem
	err = json.Unmarshal(data, &item)
	if err != nil {
		logger.Logger("Error deserializing value: ", err).Error()
		return nil, false
	}

	// Check if the item is expired
	if time.Now().Unix() > item.Expiration {
		return nil, false
	}

	return item.Value, true
}

// Remove removes the item with the specified key from the file cache.
func (c *FileCache) Remove(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	cacheFilePath := c.getFilePath(key)
	err := os.Remove(cacheFilePath)
	if err != nil {
		logger.Logger("Error removing cache file: ", err).Warn()
	}
}

func (c *FileCache) RemoveByPrefix(prefix string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// For file cache, you need to implement logic to iterate through the files
	// Assuming the file cache uses a specific directory to store cache files
	files, err := os.ReadDir(c.cacheDir)
	if err != nil {
		logger.Logger("[warn] Error reading cache directory: ", err).Warn()
		return
	}

	prefixN := len(prefix)
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if len(file.Name()) >= prefixN && file.Name()[:prefixN] == prefix {
			err = os.Remove(c.cacheDir + file.Name())
			if err != nil {
				logger.Logger("[warn] Error deleting cache file: ", err).Warn()
			}
		}
	}
}

// GetTTL returns the remaining time before the specified key expires.
func (c *FileCache) GetTTL(key string) (time.Duration, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cacheFilePath := c.getFilePath(key)
	data, err := os.ReadFile(cacheFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, false // Key does not exist
		}
		logger.Logger("Error reading cache file: ", err).Error()
		return 0, false
	}

	var item CacheItem
	err = json.Unmarshal(data, &item)
	if err != nil {
		logger.Logger("Error deserializing value: ", err).Error()
		return 0, false
	}

	// Calculate remaining TTL
	remaining := time.Until(time.Unix(item.Expiration, 0))
	if remaining < 0 {
		return 0, false // Item is expired
	}

	return remaining, true
}

// getFilePath constructs the file path for a given key.
func (c *FileCache) getFilePath(key string) string {
	return filepath.Join(c.cacheDir, fmt.Sprintf("%s.cache", key))
}

// scheduleCleanup removes the file after the TTL expires.
func (c *FileCache) scheduleCleanup(key string, ttl time.Duration) {
	time.Sleep(ttl)
	c.Remove(key)
}
