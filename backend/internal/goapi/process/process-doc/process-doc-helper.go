package processdoc

import (
	"smlcloudplatform/internal/goapi/logger"
)

// ProcessDocumentStatusByShop - ประมวลผลสถานะเอกสารสำหรับ shop เดียว
// เรียกใช้หลังจาก Kafka แต่ละเอกสารเสร็จ หรือหลังจาก rebuild เสร็จ
// ⚡ ใช้ BATCH MODE เพื่อความเร็วสูงสุด (3 queries แทน 1:1)
func ProcessDocumentStatusByShop(holdingCode string) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("❌ Panic in ProcessDocumentStatusByShop for shop %s: %v", holdingCode, r)
		}
	}()

	logger.Debug("Processing document status for shop %s (Kafka mode: BATCH)...", holdingCode)

	// ประมวลผลสถานะเอกสารใบสั่งซื้อ (transflag=6) แบบ Batch (ดึงจาก queue)
	ProcessDocPurchaseBatch(holdingCode)
}

// ProcessDocumentStatusByDocNo - ประมวลผลสถานะเอกสารเฉพาะ docno (สำหรับ Kafka)
// ⚡ DIRECT MODE: ไม่ต้องรอ queue, process ทันที
func ProcessDocumentStatusByDocNo(holdingCode string, docNo string) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("❌ Panic in ProcessDocumentStatusByDocNo for docno %s: %v", docNo, r)
		}
	}()

	logger.Debug("Processing document status for docno %s (Kafka mode: DIRECT)...", docNo)

	// ประมวลผลเฉพาะ docno นี้โดยตรง ไม่ต้องรอ queue
	ProcessDocPurchaseDirectBatchByDocNos(holdingCode, []string{docNo})
}

// ProcessDocumentStatusByShopBatch - ประมวลผลสถานะเอกสารแบบ batch สำหรับ rebuild
// ⚡ ULTRA-FAST MODE: ใช้ Direct Batch (ไม่ผ่าน queue) + Mega-Query (1 query แทน 3)
func ProcessDocumentStatusByShopBatch(holdingCode string) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("❌ Panic in ProcessDocumentStatusByShopBatch for shop %s: %v", holdingCode, r)
		}
	}()

	logger.Debug("Processing document status for shop %s (Rebuild mode: ULTRA-FAST DIRECT!)...", holdingCode)

	// ประมวลผลสถานะเอกสารใบสั่งซื้อ (transflag=6) แบบ Direct Batch (ไม่ผ่าน queue)
	ProcessDocPurchaseDirectBatch(holdingCode)
}
