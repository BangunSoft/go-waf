package memory_cache

import (
	"runtime"
	"sync"
	"time"

	"github.com/jahrulnr/go-waf/internal/interface/repository"
	"github.com/jahrulnr/go-waf/pkg/logger"
)

// item represents a cache item with a value and an expiration time.
type item struct {
	value  []byte
	expiry time.Time
}

// isExpired checks if the cache item has expired.
func (i item) isExpired() bool {
	return time.Now().After(i.expiry)
}

// TTLCache is a generic cache implementation with support for time-to-live (TTL) expiration.
type TTLCache struct {
	items map[string]item // The map storing cache items.
	mu    sync.RWMutex    // RWMutex for controlling concurrent access to the cache.
	limit int             // Maximum size of the cache in bytes.
}

// NewCache creates a new TTLCache instance with a limit set to 80% of available memory.
func NewCache() repository.CacheInterface {
	// Get total memory
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Set cache limit to 80% of total memory
	limit := int(float64(m.Sys) * 0.8) // 80% of total memory

	return &TTLCache{
		items: make(map[string]item),
		limit: limit,
	}
}

// cleanupExpiredItems removes expired items from the cache.
func (c *TTLCache) cleanupExpiredItems() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for key, item := range c.items {
		if item.isExpired() {
			delete(c.items, key)
			logger.Logger("Removed expired item from cache", key).Debug()
		}
	}
}

// Set adds a new item to the cache with the specified key, value, and time-to-live (TTL).
func (c *TTLCache) Set(key string, value []byte, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if the cache has reached its limit
	if c.limit > 0 && c.currentSize()+len(value) > c.limit {
		c.evict() // Evict an item if the limit is reached
	}

	c.items[key] = item{
		value:  value,
		expiry: time.Now().Add(ttl),
	}
	logger.Logger("Set item in cache", key).Debug()
}

// currentSize calculates the current size of the cache in bytes.
func (c *TTLCache) currentSize() int {
	size := 0
	for _, item := range c.items {
		size += len(item.value)
	}
	return size
}

// evict removes the least recently used item from the cache.
func (c *TTLCache) evict() {
	var oldestKey string
	var oldestExpiry time.Time

	for key, item := range c.items {
		if oldestKey == "" || item.expiry.Before(oldestExpiry) {
			oldestKey = key
			oldestExpiry = item.expiry
		}
	}

	if oldestKey != "" {
		delete(c.items, oldestKey)
		logger.Logger("Evicted item from cache", oldestKey).Debug()
	}
}

// Get retrieves the value associated with the given key from the cache.
func (c *TTLCache) Get(key string) ([]byte, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, found := c.items[key]
	if !found || item.isExpired() {
		if found {
			delete(c.items, key) // Remove expired item
			logger.Logger("Removed expired item from cache", key).Debug()
		}
		return nil, false
	}

	go c.cleanupExpiredItems()

	return item.value, true
}

// Remove removes the item with the specified key from the cache.
func (c *TTLCache) Remove(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
	logger.Logger("Removed item from cache", key).Debug()
}

// RemoveByPrefix removes all items with the specified prefix from the cache.
func (c *TTLCache) RemoveByPrefix(prefix string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	prefixN := len(prefix)
	for key := range c.items {
		if len(key) >= prefixN && key[:prefixN] == prefix {
			delete(c.items, key)
			logger.Logger("Removed item with prefix from cache", key).Debug()
		}
	}
}

// Pop removes and returns the item with the specified key from the cache.
func (c *TTLCache) Pop(key string) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, found := c.items[key]
	if !found || item.isExpired() {
		if found {
			delete(c.items, key) // Remove expired item
			logger.Logger("Removed expired item from cache", key).Debug()
		}
		return nil, false
	}

	delete(c.items, key)
	logger.Logger("Popped item from cache", key).Debug()
	return item.value, true
}

// GetTTL returns the remaining time before the specified key expires.
func (c *TTLCache) GetTTL(key string) (time.Duration, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, found := c.items[key]
	if !found || item.isExpired() {
		if found {
			delete(c.items, key) // Optionally remove expired item
			logger.Logger("Removed expired item from cache", key).Debug()
		}
		return 0, false
	}

	// Calculate remaining TTL
	remaining := time.Until(item.expiry)
	return remaining, true
}
