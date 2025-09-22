package main

import (
	"sync"
	"time"
)

// CacheItem represents a single item in the cache
type CacheItem struct {
	Value     []Member
	ExpiresAt time.Time
}

// Cache is a simple in-memory cache with TTL
type Cache struct {
	items map[string]CacheItem
	mu    sync.RWMutex
}

// NewCache creates a new Cache instance
func NewCache() *Cache {
	return &Cache{
		items: make(map[string]CacheItem),
	}
}

// Get retrieves an item from the cache.
// It returns the item and a boolean indicating if the item was found and is not expired.
func (c *Cache) Get(key string) ([]Member, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, found := c.items[key]
	if !found || time.Now().After(item.ExpiresAt) {
		return nil, false
	}
	return item.Value, true
}

// Set adds an item to the cache with a specified TTL.
func (c *Cache) Set(key string, value []Member, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = CacheItem{
		Value:     value,
		ExpiresAt: time.Now().Add(ttl),
	}
}
