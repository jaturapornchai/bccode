package handlers

import (
	"net/http"
	"time"

	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/mypg"

	"github.com/labstack/echo/v4"
)

// StartConnectionHealthChecker - starts background health checker for PostgreSQL connections
func StartConnectionHealthChecker() {
	logger.Info("Starting PostgreSQL connection health checker (interval: 5 minutes)")

	go func() {
		ticker := time.NewTicker(5 * time.Minute) // ตรวจสอบทุก 5 นาที
		defer ticker.Stop()

		for range ticker.C {
			logger.Info("Running PostgreSQL connection health check...")
			mypg.HealthCheckAndReconnect()

			// แสดงสถิติ connection
			stats := mypg.GetConnectionStats()
			for dbName, stat := range stats {
				logger.Info("DB: %s - Open: %d, Idle: %d, InUse: %d, WaitCount: %d",
					dbName, stat.OpenConnections, stat.Idle, stat.InUse, stat.WaitCount)
			}
		}
	}()
}

// QueueStatusHandler - ตรวจสอบสถานะ queue ทั้งหมด (PostgreSQL-based)
func QueueStatusHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":  "ok",
		"message": "Queue now managed by PostgreSQL - use SystemHealthHandler for details",
	})
}

// QueueShopStatusHandler - ตรวจสอบสถานะ queue ของ shop ที่ระบุ (PostgreSQL-based)
func QueueShopStatusHandler(c echo.Context) error {
	holdingCode := c.Param("holdingcode")
	if holdingCode == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"message": "holdingcode parameter is required",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":      "ok",
		"holdingcode": holdingCode,
		"message":     "Queue now managed by PostgreSQL - use SystemHealthHandler for details",
	})
}

// DatabaseHealthHandler - ตรวจสอบสถานะ database connections
func DatabaseHealthHandler(c echo.Context) error {
	// ตรวจสอบ PostgreSQL connections
	pgStats := mypg.GetConnectionStats()

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":     "ok",
		"message":    "Database health check completed",
		"postgresql": pgStats,
	})
}

// SystemHealthHandler - ตรวจสอบสถานะระบบทั้งหมด (comprehensive)
func SystemHealthHandler(c echo.Context) error {
	health := map[string]interface{}{
		"timestamp": time.Now(),
		"status":    "ok",
	}

	// PostgreSQL health
	pgStats := mypg.GetConnectionStats()
	health["postgresql"] = map[string]interface{}{
		"databases": len(pgStats),
		"stats":     pgStats,
	}

	return c.JSON(http.StatusOK, health)
}
