package aichat

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"smlcloudplatform/internal/goapi/logger"
)

var (
	promptCacheStore = make(map[string]*PromptCache)
)

// CalculateDataHash creates a hash of data for cache comparison
func CalculateDataHash(stockData []StockData) string {
	// Create a string representation of stock data
	var builder strings.Builder
	for _, item := range stockData {
		builder.WriteString(fmt.Sprintf("%s|%s|%s|%s|%s;",
			item.ProductCode,
			item.ProductName,
			item.BarcodeList,
			item.UnitStruct,
			item.StockQty,
		))
	}

	// Calculate SHA256 hash
	hash := sha256.Sum256([]byte(builder.String()))
	return hex.EncodeToString(hash[:])
}

// CheckCache checks if cached data is still valid
func CheckCache(cacheKey string, dataHash string) (bool, *PromptCache) {
	cachedPrompt, exists := promptCacheStore[cacheKey]
	if !exists {
		return false, nil
	}

	if cachedPrompt.Hash != dataHash {
		return false, nil
	}

	// Check if cache is still valid (within duration)
	if time.Since(cachedPrompt.Timestamp) > CacheDuration {
		logger.Info("Cache expired for %s (age: %v), refreshing...", cacheKey, time.Since(cachedPrompt.Timestamp).Round(time.Second))
		return false, nil
	}

	logger.Info("Using cached prompt for %s (hash: %s, age: %v)", cacheKey, dataHash[:8], time.Since(cachedPrompt.Timestamp).Round(time.Second))
	return true, cachedPrompt
}

// UpdateCache updates the cache with new data
func UpdateCache(cacheKey string, dataHash string, stockData []StockData) {
	promptCacheStore[cacheKey] = &PromptCache{
		Hash:      dataHash,
		StockData: stockData,
		Timestamp: time.Now(),
	}
	logger.Info("Updated prompt cache for %s (hash: %s)", cacheKey, dataHash[:8])
}
