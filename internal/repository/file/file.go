package file_cache

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/jahrulnr/go-waf/internal/interface/repository"
	"github.com/jahrulnr/go-waf/pkg/logger"
	"github.com/vmihailenco/msgpack"
)

type FileCache struct {
	cacheDir      string
	mu            sync.RWMutex
	cleanupTicker *time.Ticker
	stopChan      chan struct{}
}

// CacheItem represents an item stored in the cache, along with its expiration time.
type CacheItem struct {
	Value      []byte
	Expiration int64 // Unix timestamp
}

// NewFileCache creates a new FileCache instance with the specified directory.
func NewFileCache(cacheDir string) repository.CacheInterface {
	c := &FileCache{
		cacheDir: cacheDir,
		stopChan: make(chan struct{}),
	}

	// Start the cleanup goroutine
	c.cleanupTicker = time.NewTicker(5 * time.Second) // Set cleanup interval
	go c.scheduleCleanup()

	return c
}

// scheduleCleanup removes expired items from the file cache periodically.
func (c *FileCache) scheduleCleanup() {
	for {
		select {
		case <-c.cleanupTicker.C:
			c.cleanupExpiredItems()
		case <-c.stopChan:
			c.cleanupTicker.Stop()
			return
		}
	}
}

// cleanupExpiredItems removes expired items from the cache.
func (c *FileCache) cleanupExpiredItems() {
	c.mu.Lock()
	defer c.mu.Unlock()

	files, err := os.ReadDir(c.cacheDir)
	if err != nil {
		logger.Logger("[warn] Error reading cache directory: ", err).Warn()
		return
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		cacheFilePath := filepath.Join(c.cacheDir, file.Name())
		if c.isExpired(cacheFilePath) {
			if err := os.Remove(cacheFilePath); err != nil {
				logger.Logger("Error removing expired cache file: "+file.Name(), err).Warn()
			}
		}
	}
}

// isExpired checks if the cache item is expired.
func (c *FileCache) isExpired(cacheFilePath string) bool {
	item, err := c.readCacheItem(cacheFilePath)
	if err != nil {
		return false
	}
	return time.Now().Unix() > item.Expiration
}

// Set adds a new item to the file cache with the specified key, value, and TTL.
func (c *FileCache) Set(key string, value []byte, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	item := CacheItem{
		Value:      value,
		Expiration: time.Now().Add(ttl).Unix(),
	}

	cacheFilePath := c.getFilePath(key)
	if err := c.writeCacheItem(cacheFilePath, item); err != nil {
		logger.Logger("Error writing to cache file for key: "+key, err).Warn()
	}
}

// writeCacheItem writes the CacheItem to the specified file using MessagePack.
func (c *FileCache) writeCacheItem(cacheFilePath string, item CacheItem) error {
	file, err := os.Create(cacheFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := msgpack.NewEncoder(file)
	return encoder.Encode(item)
}

// Get retrieves the value associated with the given key from the file cache.
func (c *FileCache) Get(key string) ([]byte, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cacheFilePath := c.getFilePath(key)
	item, err := c.readCacheItem(cacheFilePath)
	if err != nil {
		return nil, false
	}

	// Check if the item is expired
	if time.Now().Unix() > item.Expiration {
		c.Remove(key) // Optionally remove expired item
		return nil, false
	}

	return item.Value, true
}

// readCacheItem reads the CacheItem from the specified file using MessagePack.
func (c *FileCache) readCacheItem(cacheFilePath string) (CacheItem, error) {
	file, err := os.Open(cacheFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return CacheItem{}, err // Key does not exist
		}
		logger.Logger("Error reading cache file: "+cacheFilePath, err).Warn()
		return CacheItem{}, err
	}
	defer file.Close()

	var item CacheItem
	decoder := msgpack.NewDecoder(file)
	if err := decoder.Decode(&item); err != nil {
		logger.Logger("Error deserializing cache item: "+cacheFilePath, err).Error()
		return CacheItem{}, err
	}

	return item, nil
}

// Pop removes and returns the item with the specified key from the file cache.
func (c *FileCache) Pop(key string) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	cacheFilePath := c.getFilePath(key)
	item, err := c.readCacheItem(cacheFilePath)
	if err != nil {
		return nil, false
	}

	// Remove the cache file
	if err := os.Remove(cacheFilePath); err != nil {
		logger.Logger("Error removing cache file for key: "+key, err).Error()
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
	if err := os.Remove(cacheFilePath); err != nil {
		logger.Logger("Error removing cache file for key: "+key, err).Warn()
	}
}

// RemoveByPrefix removes all items with the specified prefix from the file cache.
func (c *FileCache) RemoveByPrefix(prefix string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	files, err := os.ReadDir(c.cacheDir)
	if err != nil {
		logger.Logger("[warn] Error reading cache directory: ", err).Warn()
		return
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if len(file.Name()) >= len(prefix) && file.Name()[:len(prefix)] == prefix {
			if err := os.Remove(filepath.Join(c.cacheDir, file.Name())); err != nil {
				logger.Logger("[warn] Error deleting cache file for key: "+file.Name(), err).Warn()
			}
		}
	}
}

// GetTTL returns the remaining time before the specified key expires.
func (c *FileCache) GetTTL(key string) (time.Duration, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cacheFilePath := c.getFilePath(key)
	item, err := c.readCacheItem(cacheFilePath)
	if err != nil {
		return 0, false // Key does not exist
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
	return filepath.Join(c.cacheDir, key+".cache")
}

// Stop stops the cleanup goroutine.
func (c *FileCache) Stop() {
	close(c.stopChan)
}
