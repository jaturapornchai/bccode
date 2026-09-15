package kafka

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/models"
	"smlcloudplatform/internal/goapi/mypg"
	"smlcloudplatform/internal/goapi/process/stockengine"
)

// UseStockEngineV2 บอกว่าจะให้ consumer เข้าคิวให้ worker คำนวณ (ค่าเริ่มต้น)
// หรือย้อนกลับไปคำนวณทันทีในตัว consumer แบบเดิมด้วย BCAI_STOCK_ENGINE=v1
//
// การคำนวณทันทีทำให้ consumer ค้างรอไปกับสินค้าทุกตัวในเอกสาร และคิดซ้ำทุกครั้งที่เอกสารเข้ามา
// แม้จะเป็นสินค้าตัวเดิม การเข้าคิวรวมงานให้เหลือชิ้นเดียวต่อสินค้าจึงเร็วกว่าและกู้คืนได้เมื่อพัง
func UseStockEngineV2() bool {
	return !strings.EqualFold(strings.TrimSpace(os.Getenv("BCAI_STOCK_ENGINE")), "v1")
}

// EnqueueStockRecalculation บันทึกว่าสินค้าในเอกสารนี้ต้องคำนวณต้นทุนใหม่
//
// เก็บเป็นงานค้างหนึ่งชิ้นต่อสินค้าหนึ่งตัว พร้อมวันที่เก่าสุดที่ถูกกระทบ
// เอกสารที่แก้ย้อนหลังจึงทำให้คิดใหม่ตั้งแต่วันนั้น ไม่ใช่ตั้งแต่รายการแรกของสินค้า
func EnqueueStockRecalculation(ctx context.Context, db *sql.DB, holdingCode string, docDetails []models.DocDetailStruct) error {
	type target struct {
		businessCode string
		itemCode     string
	}

	earliest := make(map[target]time.Time, len(docDetails))
	for _, detail := range docDetails {
		businessCode := strings.ToUpper(strings.TrimSpace(detail.BusinessCode))
		itemCode := strings.ToUpper(strings.TrimSpace(detail.ItemCode))
		if businessCode == "" || itemCode == "" {
			return fmt.Errorf("stock recalculation requires businesscode and itemcode")
		}
		if _, ok := stockengine.DirectionOf(detail.TransFlag); !ok {
			continue // เอกสารที่ไม่กระทบสต็อก เช่น ใบสั่งซื้อ ใบสั่งขาย
		}

		key := target{businessCode: businessCode, itemCode: itemCode}
		if current, ok := earliest[key]; !ok || detail.DocDateTime.Before(current) {
			earliest[key] = detail.DocDateTime
		}
	}

	for key, from := range earliest {
		if err := stockengine.MarkDirty(ctx, db, key.businessCode, key.itemCode, from, "doc"); err != nil {
			return err
		}
	}

	// เปิด worker ของกลุ่มกิจการนี้ตอนมีงานเข้ามาครั้งแรก ไม่ต้องตั้งรายชื่อล่วงหน้า
	if len(earliest) > 0 && holdingCode != "" {
		stockengine.EnsureWorker(holdingCode, db, mypg.ListenerDSN(holdingCode))
	}

	if len(earliest) > 0 {
		logger.Debug("stock engine queued %d item(s) for recalculation", len(earliest))
	}
	return nil
}
