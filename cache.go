// cache.go
package simplehttp

import (
	"net/http"
	"sync"
	"time"
)

// Cache defines the interface for cache implementations
type CacheStore interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{}, ttl time.Duration) error
	Delete(key string) error
	Clear() error
}

// Cache middleware configuration
type CacheConfig struct {
	TTL           time.Duration
	KeyPrefix     string
	KeyFunc       func(MedaContext) string
	Store         CacheStore
	IgnoreHeaders []string
}

func MiddlewareCache(config CacheConfig) MedaMiddleware {
	return WithName("cache", SimpleCache(config))
}

// SimpleCache returns a caching middleware
func SimpleCache(config CacheConfig) MedaMiddlewareFunc {
	return func(next MedaHandlerFunc) MedaHandlerFunc {
		return func(c MedaContext) error {
			// fmt.Println("--- cache middleware")

			key := config.KeyFunc(c)
			if cached, found := config.Store.Get(key); found {
				return c.JSON(http.StatusOK, cached)
			}

			// Continue with request
			return next(c)
		}
	}
}

// MemoryCache provides a simple in-memory cache implementation
type MemoryCache struct {
	sync.RWMutex
	data map[string]cacheItem
}

type cacheItem struct {
	value      interface{}
	expiration time.Time
}

func NewMemoryCache() CacheStore {
	return &MemoryCache{
		data: make(map[string]cacheItem),
	}
}

func (c *MemoryCache) Get(key string) (interface{}, bool) {
	item, exists := c.data[key]
	if !exists || time.Now().After(item.expiration) {
		return nil, false
	}
	return item.value, true
}

func (c *MemoryCache) Set(key string, value interface{}, ttl time.Duration) error {
	c.data[key] = cacheItem{
		value:      value,
		expiration: time.Now().Add(ttl),
	}
	return nil
}

func (c *MemoryCache) Delete(key string) error {
	delete(c.data, key)
	return nil
}

func (c *MemoryCache) Clear() error {
	c.data = make(map[string]cacheItem)
	return nil
}
