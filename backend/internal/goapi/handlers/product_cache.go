package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"sync"
	"time"

	"smlcloudplatform/internal/goapi/mypg"

	"github.com/labstack/echo/v4"
)

// ProductCache - Centralized product cache
type ProductCache struct {
	mu       sync.RWMutex
	cache    map[string]*ProductCacheEntry
	maxSize  int
	ttl      time.Duration
}

// ProductCacheEntry - Single cache entry
type ProductCacheEntry struct {
	Data      []map[string]any
	CreatedAt time.Time
	HitCount  int64
}

// Global product cache instance
var productCache = NewProductCache(1000, 5*time.Minute)

// NewProductCache - Create new product cache
func NewProductCache(maxSize int, ttl time.Duration) *ProductCache {
	cache := &ProductCache{
		cache:   make(map[string]*ProductCacheEntry),
		maxSize: maxSize,
		ttl:     ttl,
	}

	// Start cleanup goroutine
	go cache.cleanupLoop()

	return cache
}

// cleanupLoop - Periodically cleanup expired entries
func (pc *ProductCache) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	for range ticker.C {
		pc.cleanup()
	}
}

// cleanup - Remove expired entries
func (pc *ProductCache) cleanup() {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	now := time.Now()
	for key, entry := range pc.cache {
		if now.Sub(entry.CreatedAt) > pc.ttl {
			delete(pc.cache, key)
		}
	}
}

// Get - Get from cache
func (pc *ProductCache) Get(key string) ([]map[string]any, bool) {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	entry, exists := pc.cache[key]
	if !exists {
		return nil, false
	}

	// Check if expired
	if time.Since(entry.CreatedAt) > pc.ttl {
		return nil, false
	}

	entry.HitCount++
	return entry.Data, true
}

// Set - Set cache entry
func (pc *ProductCache) Set(key string, data []map[string]any) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	// Check max size and evict if needed
	if len(pc.cache) >= pc.maxSize {
		pc.evictLRU()
	}

	pc.cache[key] = &ProductCacheEntry{
		Data:      data,
		CreatedAt: time.Now(),
		HitCount:  0,
	}
}

// evictLRU - Evict least recently used entries (20%)
func (pc *ProductCache) evictLRU() {
	type cacheItem struct {
		key       string
		createdAt time.Time
	}

	items := make([]cacheItem, 0, len(pc.cache))
	for key, entry := range pc.cache {
		items = append(items, cacheItem{key: key, createdAt: entry.CreatedAt})
	}

	// Sort by creation time (oldest first)
	for i := 0; i < len(items)-1; i++ {
		for j := i + 1; j < len(items); j++ {
			if items[j].createdAt.Before(items[i].createdAt) {
				items[i], items[j] = items[j], items[i]
			}
		}
	}

	// Remove oldest 20%
	removeCount := len(items) / 5
	if removeCount < 1 {
		removeCount = 1
	}

	for i := 0; i < removeCount && i < len(items); i++ {
		delete(pc.cache, items[i].key)
	}
}

// Clear - Clear all cache
func (pc *ProductCache) Clear() {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	pc.cache = make(map[string]*ProductCacheEntry)
}

// ClearByPrefix - Clear cache by key prefix
func (pc *ProductCache) ClearByPrefix(prefix string) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	for key := range pc.cache {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			delete(pc.cache, key)
		}
	}
}

// Stats - Get cache statistics
func (pc *ProductCache) Stats() map[string]any {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	var totalHits int64
	for _, entry := range pc.cache {
		totalHits += entry.HitCount
	}

	return map[string]any{
		"size":       len(pc.cache),
		"max_size":   pc.maxSize,
		"ttl_seconds": int(pc.ttl.Seconds()),
		"total_hits":  totalHits,
	}
}

// ProductSearchRequest - Request for product search
type ProductSearchRequest struct {
	ShopID string `json:"shop_id"`
	Search string `json:"search"`
	BranchCode string `json:"branch_code"`
	BusinessTypeCode string `json:"business_type_code"`
	Limit int    `json:"limit"`
	Offset int    `json:"offset"`
	UseCache bool   `json:"use_cache"`
}

// generateCacheKey - Generate cache key from request
func generateCacheKey(req ProductSearchRequest) string {
	data := fmt.Sprintf("%s_%s_%s_%s_%d_%d",
		req.ShopID, req.Search, req.BranchCode,
		req.BusinessTypeCode, req.Limit, req.Offset)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:16]) // Use first 16 bytes
}

// ProductSearchHandler - Search products with caching
func ProductSearchHandler(c echo.Context) error {
	var req ProductSearchRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request payload",
			"code":  "INVALID_PAYLOAD",
		})
	}

	if req.ShopID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "shop_id is required",
			"code":  "MISSING_SHOP_ID",
		})
	}

	// Set defaults
	if req.Limit <= 0 || req.Limit > 500 {
		req.Limit = 50
	}
	if req.Offset < 0 {
		req.Offset = 0
	}

	// Default to use cache
	if !req.UseCache {
		req.UseCache = true
	}

	cacheKey := generateCacheKey(req)

	// Check cache first
	if req.UseCache {
		if cachedData, found := productCache.Get(cacheKey); found {
			return c.JSON(http.StatusOK, map[string]any{
				"status":  "success",
				"data":    cachedData,
				"count":   len(cachedData),
				"cached":  true,
				"limit":   req.Limit,
				"offset":  req.Offset,
			})
		}
	}

	// Connect to database
	db, err := mypg.PgSqlFastConnect(req.ShopID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Database connection failed",
			"code":  "DB_CONNECTION_ERROR",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build parameterized query
	query, args := buildProductSearchQuery(req)

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Query execution failed",
			"code":  "QUERY_ERROR",
		})
	}
	defer rows.Close()

	// Scan results
	results := make([]map[string]any, 0)
	columns, err := rows.Columns()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get columns",
			"code":  "COLUMN_ERROR",
		})
	}

	for rows.Next() {
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			continue
		}

		row := make(map[string]any)
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}
		results = append(results, row)
	}

	// Update cache
	productCache.Set(cacheKey, results)

	return c.JSON(http.StatusOK, map[string]any{
		"status": "success",
		"data":   results,
		"count":  len(results),
		"cached": false,
		"limit":  req.Limit,
		"offset": req.Offset,
	})
}

// buildProductSearchQuery - Build parameterized product search query
func buildProductSearchQuery(req ProductSearchRequest) (string, []any) {
	args := make([]any, 0)
	argIndex := 1

	query := `
SELECT
	pb.itemcode,
	pb.barcode,
	pb.name0 as itemname,
	pb.unitcode,
	pb.unitname,
	pb.price,
	pb.barcoderefunitstand as unitstand,
	pb.barcoderefunitdivide as unitdivide,
	p.categorycode,
	p.vattype,
	p.costprice
FROM public.productbarcode pb
LEFT JOIN public.product p ON pb.itemcode = p.code
WHERE 1=1`

	// Add search filter (parameterized)
	if req.Search != "" {
		query += fmt.Sprintf(`
  AND (
    pb.itemcode ILIKE $%d
    OR pb.barcode ILIKE $%d
    OR pb.name0 ILIKE $%d
  )`, argIndex, argIndex+1, argIndex+2)
		searchPattern := "%" + req.Search + "%"
		args = append(args, searchPattern, searchPattern, searchPattern)
		argIndex += 3
	}

	// Add branch filter if provided
	if req.BranchCode != "" {
		// This would typically filter by branch-specific pricing or availability
		// For now, just add as a placeholder for future implementation
	}

	// Add business type filter if provided
	if req.BusinessTypeCode != "" {
		// This would filter by business type specific products
		// For now, just add as a placeholder for future implementation
	}

	query += "\nORDER BY pb.name0"
	query += fmt.Sprintf("\nLIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, req.Limit, req.Offset)

	return query, args
}

// ProductCacheStatsHandler - Get cache statistics
func ProductCacheStatsHandler(c echo.Context) error {
	stats := productCache.Stats()
	return c.JSON(http.StatusOK, map[string]any{
		"status": "success",
		"data":   stats,
	})
}

// ProductCacheClearHandler - Clear product cache
func ProductCacheClearHandler(c echo.Context) error {
	var req struct {
		ShopID string `json:"shop_id"`
		Prefix string `json:"prefix"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request payload",
			"code":  "INVALID_PAYLOAD",
		})
	}

	if req.Prefix != "" {
		productCache.ClearByPrefix(req.Prefix)
	} else {
		productCache.Clear()
	}

	return c.JSON(http.StatusOK, map[string]any{
		"status":  "success",
		"message": "Cache cleared",
	})
}

// ProductBarcodeSearchHandler - Search by barcode (optimized single lookup)
func ProductBarcodeSearchHandler(c echo.Context) error {
	var req struct {
		ShopID string `json:"shop_id"`
		Barcode string `json:"barcode"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request payload",
			"code":  "INVALID_PAYLOAD",
		})
	}

	if req.ShopID == "" || req.Barcode == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "shop_id and barcode are required",
			"code":  "MISSING_REQUIRED_FIELDS",
		})
	}

	// Check cache first
	cacheKey := fmt.Sprintf("barcode_%s_%s", req.ShopID, req.Barcode)
	if cachedData, found := productCache.Get(cacheKey); found && len(cachedData) > 0 {
		return c.JSON(http.StatusOK, map[string]any{
			"status": "success",
			"data":   cachedData[0],
			"cached": true,
		})
	}

	// Connect to database
	db, err := mypg.PgSqlFastConnect(req.ShopID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Database connection failed",
			"code":  "DB_CONNECTION_ERROR",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := `
SELECT
	pb.itemcode,
	pb.barcode,
	pb.name0 as itemname,
	pb.unitcode,
	pb.unitname,
	pb.price,
	pb.barcoderefunitstand as unitstand,
	pb.barcoderefunitdivide as unitdivide,
	p.categorycode,
	p.vattype,
	p.costprice
FROM public.productbarcode pb
LEFT JOIN public.product p ON pb.itemcode = p.code
WHERE pb.barcode = $1
LIMIT 1`

	row := db.QueryRowContext(ctx, query, req.Barcode)

	var result struct {
		ItemCode    string
		Barcode     string
		ItemName    string
		UnitCode    string
		UnitName    string
		Price       float64
		UnitStand   float64
		UnitDivide  float64
		CategoryCode *string
		VatType     *int
		CostPrice   *float64
	}

	err = row.Scan(
		&result.ItemCode, &result.Barcode, &result.ItemName,
		&result.UnitCode, &result.UnitName, &result.Price,
		&result.UnitStand, &result.UnitDivide,
		&result.CategoryCode, &result.VatType, &result.CostPrice,
	)

	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Product not found",
			"code":  "NOT_FOUND",
		})
	}

	data := map[string]any{
		"itemcode":     result.ItemCode,
		"barcode":      result.Barcode,
		"item_name":     result.ItemName,
		"unitcode":     result.UnitCode,
		"unit_name":     result.UnitName,
		"price":        result.Price,
		"unitstand":    result.UnitStand,
		"unitdivide":   result.UnitDivide,
		"categorycode": result.CategoryCode,
		"vat_type":      result.VatType,
		"costprice":    result.CostPrice,
	}

	// Cache the result
	productCache.Set(cacheKey, []map[string]any{data})

	return c.JSON(http.StatusOK, map[string]any{
		"status": "success",
		"data":   data,
		"cached": false,
	})
}
