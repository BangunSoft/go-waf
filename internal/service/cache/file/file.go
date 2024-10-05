package file_cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type FileCache struct {
	cacheDir string
	mu       sync.RWMutex
}

// NewFileCache creates a new FileCache instance with the specified directory.
func NewFileCache(cacheDir string) *FileCache {
	return &FileCache{cacheDir: cacheDir}
}

// Set adds a new item to the file cache with the specified key, value, and TTL.
func (c *FileCache) Set(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	serializedValue, err := json.Marshal(value)
	if err != nil {
		fmt.Printf("Error serializing value: %v\n", err)
		return
	}

	cacheFilePath := c.getFilePath(key)
	err = os.WriteFile(cacheFilePath, serializedValue, 0644)
	if err != nil {
		fmt.Printf("Error writing to cache file: %v\n", err)
	}

	// Optionally, you can implement a mechanism to clean up expired files
	go c.scheduleCleanup(key, ttl)
}

// Get retrieves the value associated with the given key from the file cache.
func (c *FileCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cacheFilePath := c.getFilePath(key)
	data, err := os.ReadFile(cacheFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false // Key does not exist
		}
		fmt.Printf("Error reading cache file: %v\n", err)
		return nil, false
	}

	var value interface{}
	err = json.Unmarshal(data, &value)
	if err != nil {
		fmt.Printf("Error deserializing value: %v\n", err)
		return nil, false
	}

	return value, true
}

// Pop removes and returns the item with the specified key from the file cache.
func (c *FileCache) Pop(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	cacheFilePath := c.getFilePath(key)
	data, err := os.ReadFile(cacheFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false // Key does not exist
		}
		fmt.Printf("Error reading cache file: %v\n", err)
		return nil, false
	}

	err = os.Remove(cacheFilePath)
	if err != nil {
		fmt.Printf("Error removing cache file: %v\n", err)
	}

	var value interface{}
	err = json.Unmarshal(data, &value)
	if err != nil {
		fmt.Printf("Error deserializing value: %v\n", err)
		return nil, false
	}

	return value, true
}

// Remove removes the item with the specified key from the file cache.
func (c *FileCache) Remove(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	cacheFilePath := c.getFilePath(key)
	err := os.Remove(cacheFilePath)
	if err != nil {
		fmt.Printf("Error removing cache file: %v\n", err)
	}
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
