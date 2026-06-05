package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/mcp/mongodb"
	"smlcloudplatform/internal/goapi/mcp/redis"

	"github.com/labstack/echo/v4"
)

const (
	HeaderAPIKey = "X-API-Key"
	ContextKey   = "mcp_apikey"
)

// AuthMiddleware handles API key validation
type AuthMiddleware struct {
	keysRepo *mongodb.KeysRepository
	cache    *redis.Cache
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(keysRepo *mongodb.KeysRepository, cache *redis.Cache) *AuthMiddleware {
	return &AuthMiddleware{
		keysRepo: keysRepo,
		cache:    cache,
	}
}

// APIKeyAuth validates the API key from request header
func (am *AuthMiddleware) APIKeyAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Extract API key from header
		apiKey := c.Request().Header.Get(HeaderAPIKey)
		if apiKey == "" {
			// Also check query parameter
			apiKey = c.QueryParam("apikey")
		}

		if apiKey == "" {
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{
				"error":   "API key is required",
				"code":    "MISSING_API_KEY",
				"message": "Please provide an API key via X-API-Key header or apikey query parameter",
			})
		}

		// Validate API key
		keyData, err := am.validateAPIKey(c.Request().Context(), apiKey)
		if err != nil {
			logger.Error("API key validation error: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"error":   "Internal server error",
				"code":    "INTERNAL_ERROR",
				"message": "Failed to validate API key",
			})
		}

		if keyData == nil {
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{
				"error":   "Invalid API key",
				"code":    "INVALID_API_KEY",
				"message": "The provided API key is invalid or has been revoked",
			})
		}

		// Check rate limit
		allowed, remaining, err := am.cache.CheckRateLimit(apiKey, keyData.RateLimitPerMinute)
		if err != nil {
			logger.Error("Rate limit check error: %v", err)
		}

		if !allowed {
			return c.JSON(http.StatusTooManyRequests, map[string]interface{}{
				"error":      "Rate limit exceeded",
				"code":       "RATE_LIMIT_EXCEEDED",
				"message":    fmt.Sprintf("Rate limit of %d requests per minute exceeded", keyData.RateLimitPerMinute),
				"retryafter": 60,
			})
		}

		// Set rate limit headers
		c.Response().Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", keyData.RateLimitPerMinute))
		c.Response().Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))

		// Update last used timestamp (async)
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := am.keysRepo.UpdateLastUsedAt(ctx, apiKey); err != nil {
				logger.Error("Failed to update last used at: %v", err)
			}
		}()

		// Store API key data in context
		c.Set(ContextKey, keyData)

		return next(c)
	}
}

// validateAPIKey validates an API key using cache first, then database
func (am *AuthMiddleware) validateAPIKey(ctx context.Context, apiKey string) (*mongodb.APIKey, error) {
	// Try cache first
	if am.cache != nil {
		cachedKey, err := am.cache.GetAPIKey(apiKey)
		if err != nil {
			logger.Error("Cache get error: %v", err)
		}
		if cachedKey != nil {
			return cachedKey, nil
		}
	}

	// Try database
	keyData, err := am.keysRepo.GetAPIKeyByKey(ctx, apiKey)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if keyData != nil && am.cache != nil {
		if err := am.cache.SetAPIKey(apiKey, keyData); err != nil {
			logger.Error("Cache set error: %v", err)
		}
	}

	return keyData, nil
}

// GetAPIKeyFromContext retrieves API key data from echo context
func GetAPIKeyFromContext(c echo.Context) (*mongodb.APIKey, bool) {
	data, ok := c.Get(ContextKey).(*mongodb.APIKey)
	return data, ok
}

// RequireToolPermission checks if the API key has permission to use a specific tool
func (am *AuthMiddleware) RequireToolPermission(toolName string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			keyData, ok := GetAPIKeyFromContext(c)
			if !ok {
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{
					"error":   "Unauthorized",
					"code":    "UNAUTHORIZED",
					"message": "API key not found in context",
				})
			}

			if !keyData.IsToolAllowed(toolName) {
				return c.JSON(http.StatusForbidden, map[string]interface{}{
					"error":   "Forbidden",
					"code":    "TOOL_NOT_ALLOWED",
					"message": fmt.Sprintf("API key does not have permission to use tool: %s", toolName),
				})
			}

			return next(c)
		}
	}
}

// ExtractBearerToken extracts Bearer token from Authorization header
func ExtractBearerToken(c echo.Context) string {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}

	return parts[1]
}

// GenerateAPIKey generates a new API key with prefix
func GenerateAPIKey() string {
	// Format: bc_live_{random}
	prefix := "bc_live_"
	random := generateRandomString(32)
	return prefix + random
}

// generateRandomString generates a random string of specified length
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		// Simple random generation - in production use crypto/rand
		result[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(result)
}
