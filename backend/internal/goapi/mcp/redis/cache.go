package redis

import (
	"context"
	"fmt"
	"sync"
	"time"

	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/mcp/mongodb"
)

const (
	// Key prefixes for cache
	PrefixAPIKey     = "mcp:apikey:"
	PrefixRateLimit  = "mcp:ratelimit:"
	PrefixShopData   = "mcp:shopdata:"
	PrefixToolResult = "mcp:toolresult:"

	// Default TTL values
	TTLAPIKey     = 5 * time.Minute
	TTLRateLimit  = 1 * time.Minute
	TTLShopData   = 10 * time.Minute
	TTLToolResult = 2 * time.Minute
)

// Cache handles caching for MCP using in-memory store
// Can be extended to use Redis when the dependency is added
type Cache struct {
	data    map[string]interface{}
	expires map[string]time.Time
	mutex   sync.RWMutex
	ctx     context.Context
}

// NewCache creates a new cache instance
func NewCache() *Cache {
	cache := &Cache{
		data:    make(map[string]interface{}),
		expires: make(map[string]time.Time),
		ctx:     context.Background(),
	}

	// Start cleanup goroutine
	go cache.startCleanup()

	logger.Success("MCP Cache initialized (in-memory)")
	return cache
}

// startCleanup periodically removes expired entries
func (c *Cache) startCleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.cleanup()
	}
}

// cleanup removes expired entries
func (c *Cache) cleanup() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := time.Now()
	for key, expiry := range c.expires {
		if now.After(expiry) {
			delete(c.data, key)
			delete(c.expires, key)
		}
	}
}

// set stores a value in cache with TTL
func (c *Cache) set(key string, value interface{}, ttl time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.data[key] = value
	c.expires[key] = time.Now().Add(ttl)
}

// get retrieves a value from cache
func (c *Cache) get(key string) (interface{}, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	expiry, exists := c.expires[key]
	if !exists {
		return nil, false
	}

	if time.Now().After(expiry) {
		return nil, false
	}

	data, exists := c.data[key]
	return data, exists
}

// delete removes a key from cache
func (c *Cache) delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.data, key)
	delete(c.expires, key)
}

// IsAvailable returns true if cache is available
func (c *Cache) IsAvailable() bool {
	return true
}

// GetAPIKey retrieves an API key from cache
func (c *Cache) GetAPIKey(apiKey string) (*mongodb.APIKey, error) {
	key := PrefixAPIKey + apiKey
	data, exists := c.get(key)
	if !exists {
		return nil, nil
	}

	apiKeyData, ok := data.(*mongodb.APIKey)
	if !ok {
		return nil, nil
	}

	return apiKeyData, nil
}

// SetAPIKey caches an API key
func (c *Cache) SetAPIKey(apiKey string, keyData *mongodb.APIKey) error {
	key := PrefixAPIKey + apiKey
	c.set(key, keyData, TTLAPIKey)
	return nil
}

// DeleteAPIKey removes an API key from cache
func (c *Cache) DeleteAPIKey(apiKey string) error {
	key := PrefixAPIKey + apiKey
	c.delete(key)
	return nil
}

// RateLimitEntry tracks request timestamps
type RateLimitEntry struct {
	Requests  []time.Time
	CreatedAt time.Time
}

// CheckRateLimit checks if the request is within rate limit
func (c *Cache) CheckRateLimit(apiKey string, limitPerMinute int) (bool, int, error) {
	key := PrefixRateLimit + apiKey

	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := time.Now()
	windowStart := now.Add(-1 * time.Minute)

	// Get or create rate limit entry
	var entry *RateLimitEntry
	data, exists := c.data[key]
	if exists {
		if expiry, ok := c.expires[key]; ok && now.Before(expiry) {
			if existingEntry, ok := data.(*RateLimitEntry); ok {
				entry = existingEntry
			}
		}
	}

	if entry == nil {
		entry = &RateLimitEntry{
			Requests:  []time.Time{},
			CreatedAt: now,
		}
	}

	// Filter requests within the time window
	validRequests := make([]time.Time, 0)
	for _, reqTime := range entry.Requests {
		if reqTime.After(windowStart) {
			validRequests = append(validRequests, reqTime)
		}
	}

	// Check if limit exceeded
	if len(validRequests) >= limitPerMinute {
		// Update entry
		entry.Requests = validRequests
		c.data[key] = entry
		c.expires[key] = now.Add(TTLRateLimit)
		return false, 0, nil
	}

	// Add current request
	validRequests = append(validRequests, now)
	entry.Requests = validRequests

	// Update entry
	c.data[key] = entry
	c.expires[key] = now.Add(TTLRateLimit)

	remaining := limitPerMinute - len(validRequests)
	return true, remaining, nil
}

// GetRateLimitRemaining returns remaining rate limit for an API key
func (c *Cache) GetRateLimitRemaining(apiKey string, limitPerMinute int) (int, int, error) {
	key := PrefixRateLimit + apiKey

	c.mutex.RLock()
	defer c.mutex.RUnlock()

	now := time.Now()
	windowStart := now.Add(-1 * time.Minute)

	data, exists := c.data[key]
	if !exists {
		return 0, limitPerMinute, nil
	}

	if expiry, ok := c.expires[key]; !ok || now.After(expiry) {
		return 0, limitPerMinute, nil
	}

	rateEntry, ok := data.(*RateLimitEntry)
	if !ok {
		return 0, limitPerMinute, nil
	}

	// Count valid requests
	validCount := 0
	for _, reqTime := range rateEntry.Requests {
		if reqTime.After(windowStart) {
			validCount++
		}
	}

	remaining := limitPerMinute - validCount
	if remaining < 0 {
		remaining = 0
	}

	return validCount, remaining, nil
}

// SetShopData caches shop-related data
func (c *Cache) SetShopData(shopID string, dataType string, data interface{}, ttl time.Duration) error {
	key := fmt.Sprintf("%s%s:%s", PrefixShopData, shopID, dataType)
	if ttl == 0 {
		ttl = TTLShopData
	}
	c.set(key, data, ttl)
	return nil
}

// GetShopData retrieves cached shop data
func (c *Cache) GetShopData(shopID string, dataType string, result interface{}) error {
	key := fmt.Sprintf("%s%s:%s", PrefixShopData, shopID, dataType)
	data, exists := c.get(key)
	if !exists {
		return fmt.Errorf("cache miss")
	}

	// Type assertion would need to be handled by caller
	// This is a simplified implementation
	_ = data
	return nil
}

// SetToolResult caches tool execution result
func (c *Cache) SetToolResult(toolName string, params map[string]interface{}, result interface{}) error {
	// Create cache key from tool name and params hash
	keyData := fmt.Sprintf("%v", params)
	key := fmt.Sprintf("%s%s:%s", PrefixToolResult, toolName, keyData)

	c.set(key, result, TTLToolResult)
	return nil
}

// GetToolResult retrieves cached tool result
func (c *Cache) GetToolResult(toolName string, params map[string]interface{}, result interface{}) error {
	keyData := fmt.Sprintf("%v", params)
	key := fmt.Sprintf("%s%s:%s", PrefixToolResult, toolName, keyData)

	data, exists := c.get(key)
	if !exists {
		return fmt.Errorf("cache miss")
	}

	_ = data
	return nil
}

// InvalidateShopCache clears all cached data for a shop
func (c *Cache) InvalidateShopCache(shopID string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	prefix := PrefixShopData + shopID + ":"
	for key := range c.data {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			delete(c.data, key)
			delete(c.expires, key)
		}
	}

	return nil
}

// Close closes the cache
func (c *Cache) Close() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.data = make(map[string]interface{})
	c.expires = make(map[string]time.Time)
	return nil
}

// HealthCheck checks cache health
func (c *Cache) HealthCheck() error {
	return nil
}
