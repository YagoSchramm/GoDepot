package cache

import (
	"strings"
	"sync"
	"time"
)

type Cache interface {
	Get(key string) ([]byte, bool)
	Set(key string, data []byte)
	Invalidate(name string)
}

func NewMemoryCache(defaultTTL time.Duration) Cache {
	return &memoryCache{
		defaultTTL: defaultTTL,
		items:      make(map[string]cacheItem),
	}
}

type memoryCache struct {
	mu         sync.RWMutex
	defaultTTL time.Duration
	items      map[string]cacheItem
}

type cacheItem struct {
	data      []byte
	expiresAt time.Time
}

func (c *memoryCache) Get(key string) ([]byte, bool) {
	c.mu.RLock()
	item, ok := c.items[key]
	c.mu.RUnlock()
	if !ok {
		return nil, false
	}

	if !item.expiresAt.IsZero() && time.Now().After(item.expiresAt) {
		c.mu.Lock()
		delete(c.items, key)
		c.mu.Unlock()
		return nil, false
	}

	data := make([]byte, len(item.data))
	copy(data, item.data)
	return data, true
}

func (c *memoryCache) Set(key string, data []byte) {
	expiresAt := time.Time{}
	if c.defaultTTL > 0 {
		expiresAt = time.Now().Add(c.defaultTTL)
	}

	copied := make([]byte, len(data))
	copy(copied, data)

	c.mu.Lock()
	c.items[key] = cacheItem{
		data:      copied,
		expiresAt: expiresAt,
	}
	c.mu.Unlock()
}

func (c *memoryCache) Invalidate(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for key := range c.items {
		if key == name || strings.HasPrefix(key, name) {
			delete(c.items, key)
		}
	}
}
