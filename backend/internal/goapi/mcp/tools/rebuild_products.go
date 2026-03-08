package tools

import (
	"context"
	"fmt"
	"time"

	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/process/build"
)

// ==================== Rebuild Products ====================
// Full rebuild: MongoDB → PostgreSQL + ClickHouse (ทั้งยวง)
// ใช้กรณี Kafka sync ผิดพลาด หรือต้องการ force sync ข้อมูลทั้งหมด
// เหมือนปุ่ม "สร้างสินค้าใหม่" ใน frontend ทุกประการ

// RebuildProductsResponse ผลลัพธ์จาก rebuild
type RebuildProductsResponse struct {
	Success     bool      `json:"success"`
	Message     string    `json:"message"`
	ShopID      string    `json:"shop_id"`
	Duration    string    `json:"duration"`
	GeneratedAt time.Time `json:"generated_at"`
}

// RebuildProducts ทำ full rebuild สินค้าจาก MongoDB ลง PostgreSQL + ClickHouse
// flow เหมือน frontend: POST /api/report { command_id: "rebuild-products" }
func RebuildProducts(ctx context.Context, shopID string) (*RebuildProductsResponse, error) {
	if shopID == "" {
		return nil, fmt.Errorf("shop_id is required")
	}

	logger.Info("[MCP RebuildProducts] เริ่ม full rebuild สำหรับ shop=%s", shopID)
	start := time.Now()

	err := build.RebuildProductsOnly(shopID)
	if err != nil {
		logger.Error("[MCP RebuildProducts] rebuild ล้มเหลว (shop=%s): %v", shopID, err)
		return nil, fmt.Errorf("rebuild products ล้มเหลว: %w", err)
	}

	duration := time.Since(start).Round(time.Millisecond)
	logger.Info("[MCP RebuildProducts] rebuild สำเร็จ (shop=%s, duration=%s)", shopID, duration)

	return &RebuildProductsResponse{
		Success:     true,
		Message:     fmt.Sprintf("Rebuild products สำเร็จ — ข้อมูลจาก MongoDB sync ลง PostgreSQL + ClickHouse แล้ว (ใช้เวลา %s)", duration),
		ShopID:      shopID,
		Duration:    duration.String(),
		GeneratedAt: time.Now(),
	}, nil
}
