package memory_cache

import (
	"sync"
	"time"

	"github.com/jahrulnr/go-waf/internal/interface/repository"
)

// reference
// https://www.alexedwards.net/blog/implementing-an-in-memory-cache-in-go

// item represents a cache item with a value and an expiration time.
type item[V any] struct {
	value  V
	expiry time.Time
}

// isExpired checks if the cache item has expired.
func (i item[V]) isExpired() bool {
	return time.Now().After(i.expiry)
}

// TTLCache is a generic cache implementation with support for time-to-live
// (TTL) expiration.
type TTLCache struct {
	items map[string]item[[]byte] // The map storing cache items.
	mu    sync.Mutex              // Mutex for controlling concurrent access to the cache.
}

// NewTTL creates a new TTLCache instance and starts a goroutine to periodically
// remove expired items every 5 seconds.
func NewCache() repository.CacheInterface {
	c := &TTLCache{
		items: make(map[string]item[[]byte]),
	}

	go func() {
		for range time.Tick(5 * time.Second) {
			c.mu.Lock()

			// Iterate over the cache items and delete expired ones.
			for key, item := range c.items {
				if item.isExpired() {
					delete(c.items, key)
				}
			}

			c.mu.Unlock()
		}
	}()

	return c
}

// Set adds a new item to the cache with the specified key, value, and
// time-to-live (TTL).
func (c *TTLCache) Set(key string, value []byte, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = item[[]byte]{
		value:  value,
		expiry: time.Now().Add(ttl),
	}
}

// Get retrieves the value associated with the given key from the cache.
func (c *TTLCache) Get(key string) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, found := c.items[key]
	if !found {
		// If the key is not found, return the zero value for V and false.
		return item.value, false
	}

	if item.isExpired() {
		// If the item has expired, remove it from the cache and return the
		// value and false.
		delete(c.items, key)
		return item.value, false
	}

	// Otherwise return the value and true.
	return item.value, true
}

// Remove removes the item with the specified key from the cache.
func (c *TTLCache) Remove(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Delete the item with the given key from the cache.
	delete(c.items, key)
}

func (c *TTLCache) RemoveByPrefix(prefix string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	prefixN := len(prefix)
	for key := range c.items {
		// matching key with prefix
		if len(key) >= prefixN && key[:prefixN] == prefix {
			// Delete the item with the given prefix from the cache.
			delete(c.items, key)
		}
	}
}

// Pop removes and returns the item with the specified key from the cache.
func (c *TTLCache) Pop(key string) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, found := c.items[key]
	if !found {
		// If the key is not found, return the zero value for V and false.
		return item.value, false
	}

	// If the key is found, delete the item from the cache.
	delete(c.items, key)

	if item.isExpired() {
		// If the item has expired, return the value and false.
		return item.value, false
	}

	// Otherwise return the value and true.
	return item.value, true
}

// GetTTL returns the remaining time before the specified key expires.
func (c *TTLCache) GetTTL(key string) (time.Duration, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, found := c.items[key]
	if !found {
		return 0, false // Key does not exist
	}

	if item.isExpired() {
		delete(c.items, key) // Optionally remove expired item
		return 0, false
	}

	// Calculate remaining TTL
	remaining := time.Until(item.expiry)
	return remaining, true
}
