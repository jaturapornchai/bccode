package kafka

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/models"
	"smlcloudplatform/internal/goapi/mypg"
	"smlcloudplatform/internal/goapi/process/stockengine"
)

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
		if _, ok := stockengine.DirectionOf(detail.TransFlag); !ok {
			continue // เอกสารที่ไม่กระทบสต็อก เช่น ใบสั่งซื้อ ใบสั่งขาย
		}

		// ตรวจความครบถ้วนเฉพาะบรรทัดที่กระทบสต็อกจริง
		// ใบสั่งซื้อ/ใบสั่งขายมีบรรทัดที่ไม่ได้ระบุรหัสสินค้าได้ และไม่ควรทำให้ทั้งใบล้มเหลว
		businessCode := strings.ToUpper(strings.TrimSpace(detail.BusinessCode))
		itemCode := strings.ToUpper(strings.TrimSpace(detail.ItemCode))
		if businessCode == "" || itemCode == "" {
			return fmt.Errorf("stock recalculation requires businesscode and itemcode")
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
