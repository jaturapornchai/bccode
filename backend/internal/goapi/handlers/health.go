package handlers

import (
	"context"
	"net/http"
	"os"
	"sync"
	"time"

	serviceConfig "smlcloudplatform/internal/goapi/config"
	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/mypg"

	"github.com/labstack/echo/v4"
	"github.com/segmentio/kafka-go"
)

var (
	backgroundTaskRunning = false
	backgroundTaskMutex   sync.Mutex
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

// StartBackgroundTask - เริ่ม background task สำหรับ ClickHouse clone
// ⚡ OPTIMIZED: เปลี่ยนจาก 1 นาที → 10 นาที (ลด overhead 90%)
func StartBackgroundTask() {
	logger.Info("🔄 Starting background task (interval: 10 minutes)")

	go func() {
		for {
			// รอ 10 นาที แทนที่ 1 นาที (ลด overhead)
			time.Sleep(10 * time.Minute)

			// ตรวจสอบว่า task กำลังทำงานอยู่หรือไม่
			backgroundTaskMutex.Lock()
			if backgroundTaskRunning {
				logger.Info("⏳ Background task is still running, skipping this cycle...")
				backgroundTaskMutex.Unlock()
				continue
			}

			// เริ่มทำงาน
			backgroundTaskRunning = true
			backgroundTaskMutex.Unlock()

			logger.Info("🚀 Starting background task execution...")
			startTime := time.Now()

			// ทำงานจริง
			err := performBackgroundTask()

			duration := time.Since(startTime)

			// เสร็จแล้ว - ปลดล็อค
			backgroundTaskMutex.Lock()
			backgroundTaskRunning = false
			backgroundTaskMutex.Unlock()

			if err != nil {
				logger.Error("❌ Background task failed (duration: %v): %v", duration, err)
			} else {
				logger.Info("✅ Background task completed successfully (duration: %v)", duration)
			}
		}
	}()
}

// performBackgroundTask - ฟังก์ชันที่ทำงานจริง (เปลี่ยนตามต้องการ)
func performBackgroundTask() error {
	// ClickHouse is permanently deprecated and disabled.
	return nil
} // GetBackgroundTaskStatus - ตรวจสอบสถานะ background task
func GetBackgroundTaskStatus() map[string]interface{} {
	backgroundTaskMutex.Lock()
	defer backgroundTaskMutex.Unlock()

	return map[string]interface{}{
		"running":   backgroundTaskRunning,
		"timestamp": time.Now().Format(time.RFC3339),
	}
}

// KafkaHealthHandler - ตรวจสอบสถานะการเชื่อมต่อ Kafka
func KafkaHealthHandler(c echo.Context) error {
	svcConfig := serviceConfig.NewServiceConfig()

	// ตรวจสอบว่า Kafka URL ถูกตั้งค่าหรือไม่
	kafkaURL := svcConfig.KafkaURI()
	if kafkaURL == "" {
		return c.JSON(http.StatusServiceUnavailable, map[string]any{
			"status":  "error",
			"message": "KAFKA_SERVER_URL not configured",
			"kafka":   false,
		})
	}

	// ตรวจสอบว่า Kafka ถูกเปิดใช้งานหรือไม่
	enableKafka := os.Getenv("ENABLE_KAFKA")
	if enableKafka != "true" {
		return c.JSON(http.StatusOK, map[string]any{
			"status":  "disabled",
			"message": "Kafka is disabled in configuration",
			"kafka":   false,
		})
	}

	// สร้าง connection ชั่วคราวเพื่อทดสอบ
	conn, err := kafka.DialContext(context.Background(), "tcp", kafkaURL)
	if err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]any{
			"status":  "error",
			"message": "Failed to connect to Kafka broker",
			"error":   err.Error(),
			"broker":  kafkaURL,
			"kafka":   false,
		})
	}
	defer conn.Close()

	// ดึง controller info
	controller, err := conn.Controller()
	if err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]any{
			"status":  "error",
			"message": "Failed to get Kafka controller",
			"error":   err.Error(),
			"broker":  kafkaURL,
			"kafka":   false,
		})
	}

	// ดึงรายชื่อ topics
	partitions, err := conn.ReadPartitions()
	if err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]any{
			"status":  "warning",
			"message": "Connected but failed to read partitions",
			"error":   err.Error(),
			"broker":  kafkaURL,
			"kafka":   true,
		})
	}

	// สร้างรายการ topics ที่ไม่ซ้ำกัน
	topicsMap := make(map[string]bool)
	for _, p := range partitions {
		topicsMap[p.Topic] = true
	}

	topics := make([]string, 0, len(topicsMap))
	for topic := range topicsMap {
		topics = append(topics, topic)
	}

	// รายการ consumer groups ที่ระบบใช้งาน
	consumerGroups := []string{
		"consumer.sale-invoice",
		"consumer.sale-return",
		"consumer.sale-order",
		"consumer.inventory",
		"consumer.inventory-bulk",
		"consumer.warehouse",
		"consumer.purchase",
	}

	return c.JSON(http.StatusOK, map[string]any{
		"status":          "ok",
		"message":         "Kafka is healthy and connected",
		"kafka":           true,
		"broker":          kafkaURL,
		"controller_id":   controller.ID,
		"controller_host": controller.Host,
		"controller_port": controller.Port,
		"topics_count":    len(topics),
		"topics":          topics,
		"consumer_groups": consumerGroups,
		"ssl_enabled":     svcConfig.KafKaSecurityProtocol() == "SSL",
	})
}

// BackgroundTaskStatusHandler - ตรวจสอบสถานะ background task
func BackgroundTaskStatusHandler(c echo.Context) error {
	status := GetBackgroundTaskStatus()
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":          "ok",
		"background_task": status,
	})
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
	shopId := c.Param("shopid")
	if shopId == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"message": "shopid parameter is required",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":  "ok",
		"shop_id": shopId,
		"message": "Queue now managed by PostgreSQL - use SystemHealthHandler for details",
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

	// Background task status
	health["background_task"] = GetBackgroundTaskStatus()

	return c.JSON(http.StatusOK, health)
}
