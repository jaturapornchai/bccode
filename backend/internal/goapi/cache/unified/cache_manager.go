package cache

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"smlcloudplatform/internal/goapi/logger"
)

// Cache TTL Configurations
var CacheTTL = map[string]time.Duration{
	"ai_chat":        15 * time.Minute,
	"product_search": 30 * time.Minute,
	"stockrealtime":  5 * time.Minute,
	"document":       60 * time.Minute,
	"analytics":      24 * time.Hour,
}

// CacheKeyPatterns สำหรับการสร้าง key
var CacheKeyPatterns = map[string]string{
	"ai_chat":        "ai:chat:{holdingcode}:{hash}",
	"product_search": "product:search:{holdingcode}:{hash}",
	"stockrealtime":  "stock:realtime:{holdingcode}:{itemcode}:{warehouse}:{location}",
	"document":       "doc:{holdingcode}:{docno}:{version}",
	"analytics":      "analytics:{holdingcode}:{type}:{period}",
}

// CacheEntry แสดงข้อมูลใน cache
type CacheEntry struct {
	Data       interface{}   `json:"data"`
	Timestamp  time.Time     `json:"timestamp"`
	TTL        time.Duration `json:"ttl"`
	HitCount   int64         `json:"hitcount"`
	LastAccess time.Time     `json:"lastaccess"`
}

// CacheStats สถิติการใช้งาน cache
type CacheStats struct {
	TotalEntries int64            `json:"totalentries"`
	TotalHits    int64            `json:"totalhits"`
	TotalMisses  int64            `json:"totalmisses"`
	HitRate      float64          `json:"hitrate"`
	MemoryUsage  int64            `json:"memoryusage"`
	ByType       map[string]int64 `json:"bytype"`
	LastUpdated  time.Time        `json:"lastupdated"`
}

// UnifiedCacheManager จัดการ cache หลายระดับ
type UnifiedCacheManager struct {
	// L1: In-Memory Cache (Go sync.Map)
	l1Cache *sync.Map

	// L2: Redis Cache (Global) - จะเพิ่มในการใช้งานจริง
	redisEnabled bool

	// L3: Database Cache (PostgreSQL materialized views) - จะเพิ่มในการใช้งานจริง
	dbEnabled bool

	// Statistics
	statsMutex sync.RWMutex
	stats      *CacheStats

	// Configuration
	config *CacheConfig
}

// CacheConfig การตั้งค่า Cache
type CacheConfig struct {
	L1CacheSize     int                      // ขนาด L1 cache
	EnableRedis     bool                     // เปิดใช้ Redis
	RedisURL        string                   // URL Redis
	EnableDatabase  bool                     // เปิดใช้ Database cache
	DatabaseURL     string                   // URL Database
	TTLOverrides    map[string]time.Duration // override TTL ตามประเภท
	AutoCleanup     bool                     // เปิดการทำความสะอาดอัตโนมัติ
	CleanupInterval time.Duration            // ช่วงเวลาในการทำความสะอาด
}

// DefaultCacheConfig การตั้งค่าเริ่มต้น
var DefaultCacheConfig = &CacheConfig{
	L1CacheSize:     10000,
	EnableRedis:     false, // จะเปิดในการใช้งานจริง
	EnableDatabase:  false, // จะเปิดในการใช้งานจริง
	TTLOverrides:    make(map[string]time.Duration),
	AutoCleanup:     true,
	CleanupInterval: 10 * time.Minute,
}

// NewUnifiedCacheManager สร้าง UnifiedCacheManager ใหม่
func NewUnifiedCacheManager(config *CacheConfig) *UnifiedCacheManager {
	if config == nil {
		config = DefaultCacheConfig
	}

	// Apply TTL overrides
	for cacheType, overrideTTL := range config.TTLOverrides {
		if _, exists := CacheTTL[cacheType]; exists {
			CacheTTL[cacheType] = overrideTTL
		}
	}

	cache := &UnifiedCacheManager{
		l1Cache:      &sync.Map{},
		redisEnabled: config.EnableRedis,
		dbEnabled:    config.EnableDatabase,
		config:       config,
		stats: &CacheStats{
			ByType: make(map[string]int64),
		},
	}

	// เริ่มการทำความสะอาดอัตโนมัติ
	if config.AutoCleanup {
		go cache.startAutoCleanup()
	}

	logger.Success("✅ UnifiedCacheManager เริ่มต้นเรียบร้อย (L1: %d entries, Redis: %v, DB: %v)",
		config.L1CacheSize, config.EnableRedis, config.EnableDatabase)

	return cache
}

// Get ดึงข้อมูลจาก cache
func (c *UnifiedCacheManager) Get(ctx context.Context, cacheType, holdingCode string, key interface{}) (interface{}, bool) {
	cacheKey := c.generateCacheKey(cacheType, holdingCode, key)

	startTime := time.Now()

	// ลอง L1 Cache ก่อน
	if value, exists := c.l1Cache.Load(cacheKey); exists {
		entry, ok := value.(*CacheEntry)
		if ok && !c.isExpired(entry) {
			c.recordHit(cacheType)
			entry.HitCount++
			entry.LastAccess = time.Now()
			c.l1Cache.Store(cacheKey, entry)

			logger.Debug("Cache HIT (L1): %s (shop: %s, duration: %v)", cacheType, holdingCode, time.Since(startTime))
			return entry.Data, true
		}
		// ลบ entry ที่หมดอายุ
		c.l1Cache.Delete(cacheKey)
	}

	// TODO: เพิ่ม L2 (Redis) และ L3 (Database) cache
	// ในการใช้งานจริง จะเพิ่มการดึงจาก Redis และ Database

	c.recordMiss(cacheType)
	logger.Debug("Cache MISS: %s (shop: %s, duration: %v)", cacheType, holdingCode, time.Since(startTime))

	return nil, false
}

// Set บันทึกข้อมูลลง cache
func (c *UnifiedCacheManager) Set(ctx context.Context, cacheType, holdingCode string, key interface{}, data interface{}, ttl time.Duration) error {
	cacheKey := c.generateCacheKey(cacheType, holdingCode, key)

	// ใช้ TTL เริ่มต้นถ้าไม่ได้ระบุ
	if ttl == 0 {
		if defaultTTL, exists := CacheTTL[cacheType]; exists {
			ttl = defaultTTL
		} else {
			ttl = 30 * time.Minute // default TTL
		}
	}

	entry := &CacheEntry{
		Data:       data,
		Timestamp:  time.Now(),
		TTL:        ttl,
		HitCount:   0,
		LastAccess: time.Now(),
	}

	// บันทึกใน L1 Cache
	c.l1Cache.Store(cacheKey, entry)

	// TODO: เพิ่มการบันทึกใน Redis และ Database
	// ในการใช้งานจริง จะเพิ่มการบันทึกแบบ async

	logger.Debug("Cache SET: %s (shop: %s, key: %s, ttl: %v)", cacheType, holdingCode, cacheKey, ttl)
	return nil
}

// Delete ลบข้อมูลจาก cache
func (c *UnifiedCacheManager) Delete(ctx context.Context, cacheType, holdingCode string, key interface{}) error {
	cacheKey := c.generateCacheKey(cacheType, holdingCode, key)

	c.l1Cache.Delete(cacheKey)

	// TODO: เพิ่มการลบจาก Redis และ Database

	logger.Debug("Cache DELETE: %s (shop: %s, key: %s)", cacheType, holdingCode, cacheKey)
	return nil
}

// Clear ล้าง cache ทั้งหมด
func (c *UnifiedCacheManager) Clear(ctx context.Context, cacheType string) error {
	if cacheType == "" {
		// ล้างทั้งหมด
		c.l1Cache = &sync.Map{}
		logger.Info("🧹 Cache ทั้งหมดถูกล้างเรียบร้อย")
	} else {
		// ล้างตามประเภท
		c.l1Cache.Range(func(key, value interface{}) bool {
			keyStr, ok := key.(string)
			if ok && c.isCacheTypeMatch(keyStr, cacheType) {
				c.l1Cache.Delete(key)
			}
			return true
		})
		logger.Info("🧹 Cache type %s ถูกล้างเรียบร้อย", cacheType)
	}

	return nil
}

// GetStats ดึงสถิติ cache
func (c *UnifiedCacheManager) GetStats() *CacheStats {
	c.statsMutex.RLock()
	defer c.statsMutex.RUnlock()

	// อัปเดตสถิติปัจจุบัน
	c.stats.TotalEntries = c.getTotalEntries()
	c.stats.MemoryUsage = c.estimateMemoryUsage()
	c.stats.HitRate = c.calculateHitRate()
	c.stats.LastUpdated = time.Now()

	// สร้าง deep copy
	statsCopy := *c.stats
	statsCopy.ByType = make(map[string]int64)
	for k, v := range c.stats.ByType {
		statsCopy.ByType[k] = v
	}

	return &statsCopy
}

// generateCacheKey สร้าง cache key
func (c *UnifiedCacheManager) generateCacheKey(cacheType, holdingCode string, key interface{}) string {
	keyStr := fmt.Sprintf("%v", key)
	hash := md5.Sum([]byte(keyStr))
	hashStr := hex.EncodeToString(hash[:8])

	return fmt.Sprintf("%s:%s:%s", cacheType, holdingCode, hashStr)
}

// isExpired ตรวจสอบว่า entry หมดอายุหรือไม่
func (c *UnifiedCacheManager) isExpired(entry *CacheEntry) bool {
	return time.Since(entry.Timestamp) > entry.TTL
}

// isCacheTypeMatch ตรวจสอบว่า key ตรงกับ cache type หรือไม่
func (c *UnifiedCacheManager) isCacheTypeMatch(keyStr, cacheType string) bool {
	for pattern := range CacheTTL {
		if _, exists := CacheTTL[cacheType]; !exists {
			continue
		}
		expectedPattern := CacheKeyPatterns[pattern]
		if expectedPattern == "" {
			continue
		}
		// Simple pattern matching (ในการใช้งานจริง ควรใช้ regex ที่ซับซ้อนกว่า)
		if len(keyStr) > len(cacheType) && keyStr[:len(cacheType)] == cacheType {
			return true
		}
	}
	return false
}

// recordHit บันทึก cache hit
func (c *UnifiedCacheManager) recordHit(cacheType string) {
	c.statsMutex.Lock()
	defer c.statsMutex.Unlock()

	c.stats.TotalHits++
	c.stats.ByType[cacheType]++
}

// recordMiss บันทึก cache miss
func (c *UnifiedCacheManager) recordMiss(cacheType string) {
	c.statsMutex.Lock()
	defer c.statsMutex.Unlock()

	c.stats.TotalMisses++
}

// getTotalEntries นับจำนวน entries ทั้งหมด
func (c *UnifiedCacheManager) getTotalEntries() int64 {
	var count int64
	c.l1Cache.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}

// estimateMemoryUsage ประมาณการใช้งาน memory
func (c *UnifiedCacheManager) estimateMemoryUsage() int64 {
	// การประมาณการแบบง่าย ๆ (ในการใช้งานจริง ควรใช้ runtime.MemStats)
	return c.getTotalEntries() * 1024 // 1KB ต่อ entry โดยประมาณ
}

// calculateHitRate คำนวณ hit rate
func (c *UnifiedCacheManager) calculateHitRate() float64 {
	total := c.stats.TotalHits + c.stats.TotalMisses
	if total == 0 {
		return 0.0
	}
	return float64(c.stats.TotalHits) / float64(total)
}

// startAutoCleanup เริ่มการทำความสะอาดอัตโนมัติ
func (c *UnifiedCacheManager) startAutoCleanup() {
	ticker := time.NewTicker(c.config.CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		c.performCleanup()
	}
}

// performCleanup ทำความสะอาด cache
func (c *UnifiedCacheManager) performCleanup() {
	cleaned := 0

	c.l1Cache.Range(func(key, value interface{}) bool {
		entry, ok := value.(*CacheEntry)
		if ok && c.isExpired(entry) {
			c.l1Cache.Delete(key)
			cleaned++
		}
		return true
	})

	if cleaned > 0 {
		logger.Info("🧹 Cache cleanup: ลบ %d entries ที่หมดอายุ", cleaned)
	}
}

// Specialized cache methods

// AI Chat Cache Methods
func (c *UnifiedCacheManager) GetAIChat(ctx context.Context, holdingCode, questionHash string) (interface{}, bool) {
	return c.Get(ctx, "ai_chat", holdingCode, questionHash)
}

func (c *UnifiedCacheManager) SetAIChat(ctx context.Context, holdingCode, questionHash string, response interface{}) error {
	return c.Set(ctx, "ai_chat", holdingCode, questionHash, response, 0)
}

// Product Search Cache Methods
func (c *UnifiedCacheManager) GetProductSearch(ctx context.Context, holdingCode, queryHash string) (interface{}, bool) {
	return c.Get(ctx, "product_search", holdingCode, queryHash)
}

func (c *UnifiedCacheManager) SetProductSearch(ctx context.Context, holdingCode, queryHash string, results interface{}) error {
	return c.Set(ctx, "product_search", holdingCode, queryHash, results, 0)
}

// Real-time Stock Cache Methods
func (c *UnifiedCacheManager) GetStock(ctx context.Context, holdingCode, itemCode, warehouse, location string) (interface{}, bool) {
	key := fmt.Sprintf("%s:%s:%s", itemCode, warehouse, location)
	return c.Get(ctx, "stockrealtime", holdingCode, key)
}

func (c *UnifiedCacheManager) SetStock(ctx context.Context, holdingCode, itemCode, warehouse, location string, balance interface{}) error {
	key := fmt.Sprintf("%s:%s:%s", itemCode, warehouse, location)
	return c.Set(ctx, "stockrealtime", holdingCode, key, balance, 0)
}

func (c *UnifiedCacheManager) UpdateStock(ctx context.Context, holdingCode, itemCode, warehouse, location string, balance interface{}) error {
	key := fmt.Sprintf("%s:%s:%s", itemCode, warehouse, location)
	return c.Set(ctx, "stockrealtime", holdingCode, key, balance, 0)
}
