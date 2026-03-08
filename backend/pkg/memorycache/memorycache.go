package memorycache

import (
	"time"

	"github.com/patrickmn/go-cache"
)

type IMemoryCache interface {
	Set(key string, value interface{}, d time.Duration)
	Get(key string) (interface{}, bool)
	Delete(key string)
}

type MemoryCache struct {
	cache *cache.Cache
}

func NewMemoryCache() *MemoryCache {
	cache := cache.New(5*time.Minute, 10*time.Minute)
	return &MemoryCache{
		cache: cache,
	}
}

func (c *MemoryCache) Set(key string, value interface{}, d time.Duration) {
	c.cache.Set(key, value, d)
}

func (c *MemoryCache) Get(key string) (interface{}, bool) {
	return c.cache.Get(key)
}

func (c *MemoryCache) Delete(key string) {
	c.cache.Delete(key)
}
