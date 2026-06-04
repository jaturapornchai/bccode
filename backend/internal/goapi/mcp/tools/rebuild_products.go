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
	HoldingCode string    `json:"holding_code"`
	Duration    string    `json:"duration"`
	GeneratedAt time.Time `json:"generated_at"`
}

// RebuildProducts ทำ full rebuild สินค้าจาก MongoDB ลง PostgreSQL + ClickHouse
// flow เหมือน frontend: POST /api/report { command_id: "rebuild-products" }
func RebuildProducts(ctx context.Context, holdingCode string) (*RebuildProductsResponse, error) {
	if holdingCode == "" {
		return nil, fmt.Errorf("holding_code is required")
	}

	logger.Info("[MCP RebuildProducts] เริ่ม full rebuild สำหรับ shop=%s", holdingCode)
	start := time.Now()

	err := build.RebuildProductsOnly(holdingCode)
	if err != nil {
		logger.Error("[MCP RebuildProducts] rebuild ล้มเหลว (shop=%s): %v", holdingCode, err)
		return nil, fmt.Errorf("rebuild products ล้มเหลว: %w", err)
	}

	duration := time.Since(start).Round(time.Millisecond)
	logger.Info("[MCP RebuildProducts] rebuild สำเร็จ (shop=%s, duration=%s)", holdingCode, duration)

	return &RebuildProductsResponse{
		Success:     true,
		Message:     fmt.Sprintf("Rebuild products สำเร็จ — ข้อมูลจาก MongoDB sync ลง PostgreSQL + ClickHouse แล้ว (ใช้เวลา %s)", duration),
		HoldingCode: holdingCode,
		Duration:    duration.String(),
		GeneratedAt: time.Now(),
	}, nil
}
