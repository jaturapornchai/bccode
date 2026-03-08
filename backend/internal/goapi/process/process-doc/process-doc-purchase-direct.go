package processdoc

import (
	"context"
	"fmt"
	"smlcloudplatform/internal/goapi/logger"

	mypg "smlcloudplatform/internal/goapi/mypg"
)

// ProcessDocPurchaseDirectBatch - ประมวลผลสถานะเอกสารแบบ Direct (ไม่ผ่าน docwaitprocess)
// ⚡ ULTRA-FAST MODE สำหรับ Rebuild - ใช้ 1 Mega-Query แทน 3 Queries
// คาดการณ์: เร็วกว่า ProcessDocPurchaseBatch อีก 30-40%
func ProcessDocPurchaseDirectBatch(shopId string) {
	logger.Info("Starting ProcessDocPurchaseDirectBatch for shop %s (ULTRA-FAST DIRECT MODE)", shopId)
	ctx := context.Background()

	db, err := mypg.PgSqlFastConnect(shopId)
	if err != nil {
		logger.Error("Failed to connect to PostgreSQL: %v", err)
		return
	}

	// 🚀 MEGA-QUERY: รวม 3 steps เป็น 1 query ด้วย nested CTE
	// Step 1: คำนวณ isref
	// Step 2: คำนวณ iscomparedsuccess + isclosed
	// Step 3: ไม่ต้อง DELETE queue (เพราะไม่มี queue ตั้งแต่แรก!)
	logger.Debug("Executing MEGA-QUERY: Direct batch update (no queue)")

	megaQuery := `
		WITH 
		-- CTE 1: คำนวณ isref สำหรับทุกเอกสาร
		doc_isref AS (
			SELECT 
				doc.docno,
				CASE 
					WHEN EXISTS (
						SELECT 1 FROM docref 
						WHERE docref.docnoref = doc.docno 
						AND docref.docnotransflag IN (310, 12)
					) THEN true
					ELSE false
				END AS new_isref
			FROM doc
			WHERE doc.transflag = 6
		),
		-- CTE 2: คำนวณ totalordered และ totalreceived
		doc_comparison AS (
			SELECT 
				dds.docno,
				SUM(COALESCE(dds.totalqty * (dds.unitstand / NULLIF(dds.unitdivide, 0)), dds.totalqty)) AS totalordered,
				COALESCE(SUM(drc.totalqty * (drc.unitstand / NULLIF(drc.unitdivide, 0))), 0) AS totalreceived
			FROM docdetail dds
			LEFT JOIN (
				SELECT
					dr.docnoref,
					dd.itemcode,
					SUM(dd.totalqty) AS totalqty,
					MAX(dd.unitstand) AS unitstand,
					MAX(dd.unitdivide) AS unitdivide
				FROM docdetail dd
				INNER JOIN docref dr ON dd.docno = dr.docno
				WHERE dd.transflag IN (310, 12)
				GROUP BY dr.docnoref, dd.itemcode
			) AS drc ON dds.docno = drc.docnoref AND dds.itemcode = drc.itemcode
			WHERE dds.transflag = 6
			GROUP BY dds.docno
		)
		-- UPDATE ทั้งหมดในครั้งเดียว
		UPDATE doc
		SET 
			isref = COALESCE(dir.new_isref, doc.isref),
			iscomparedsuccess = CASE
				WHEN dc.totalreceived IS NULL THEN doc.iscomparedsuccess
				WHEN dc.totalreceived = 0 THEN 0
				WHEN dc.totalreceived = dc.totalordered THEN 1
				WHEN dc.totalreceived < dc.totalordered THEN 2
				WHEN dc.totalreceived > dc.totalordered THEN 3
				ELSE doc.iscomparedsuccess
			END,
			isclosed = CASE
				WHEN dc.totalreceived IS NULL THEN doc.isclosed
				WHEN dc.totalreceived = 0 THEN false
				WHEN dc.totalreceived = dc.totalordered THEN true
				WHEN dc.totalreceived > dc.totalordered THEN true
				ELSE false
			END
		FROM doc_isref dir
		LEFT JOIN doc_comparison dc ON dir.docno = dc.docno
		WHERE doc.docno = dir.docno AND doc.transflag = 6
	`

	result, err := db.ExecContext(ctx, megaQuery)
	if err != nil {
		logger.Error("Failed to execute mega-query: %v", err)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	logger.Success("✅ MEGA-QUERY completed: Updated %d documents (isref + iscomparedsuccess + isclosed)", rowsAffected)
	logger.Success("ProcessDocPurchaseDirectBatch completed for shop %s", shopId)
}

// ProcessDocPurchaseDirectBatchByDocNos - ประมวลผลเฉพาะเอกสารที่ระบุ (Direct mode)
// ⚡ ULTRA-FAST MODE สำหรับ Rebuild เฉพาะบางเอกสาร
func ProcessDocPurchaseDirectBatchByDocNos(shopId string, docNos []string) {
	if len(docNos) == 0 {
		logger.Warn("No documents to process")
		return
	}

	logger.Info("Starting ProcessDocPurchaseDirectBatchByDocNos for shop %s (%d documents)", shopId, len(docNos))
	ctx := context.Background()

	db, err := mypg.PgSqlFastConnect(shopId)
	if err != nil {
		logger.Error("Failed to connect to PostgreSQL: %v", err)
		return
	}

	// สร้าง IN clause จาก docNos
	docNoList := "'" + joinStringSlice(docNos, "','") + "'"

	// 🚀 MEGA-QUERY สำหรับเอกสารเฉพาะ
	megaQuery := fmt.Sprintf(`
		WITH 
		-- CTE 1: คำนวณ isref
		doc_isref AS (
			SELECT 
				doc.docno,
				CASE 
					WHEN EXISTS (
						SELECT 1 FROM docref 
						WHERE docref.docnoref = doc.docno 
						AND docref.docnotransflag IN (310, 12)
					) THEN true
					ELSE false
				END AS new_isref
			FROM doc
			WHERE doc.transflag = 6
			AND doc.docno IN (%s)
		),
		-- CTE 2: คำนวณ totalordered และ totalreceived
		doc_comparison AS (
			SELECT 
				dds.docno,
				SUM(COALESCE(dds.totalqty * (dds.unitstand / NULLIF(dds.unitdivide, 0)), dds.totalqty)) AS totalordered,
				COALESCE(SUM(drc.totalqty * (drc.unitstand / NULLIF(drc.unitdivide, 0))), 0) AS totalreceived
			FROM docdetail dds
			LEFT JOIN (
				SELECT
					dr.docnoref,
					dd.itemcode,
					SUM(dd.totalqty) AS totalqty,
					MAX(dd.unitstand) AS unitstand,
					MAX(dd.unitdivide) AS unitdivide
				FROM docdetail dd
				INNER JOIN docref dr ON dd.docno = dr.docno
				WHERE dd.transflag IN (310, 12)
				GROUP BY dr.docnoref, dd.itemcode
			) AS drc ON dds.docno = drc.docnoref AND dds.itemcode = drc.itemcode
			WHERE dds.transflag = 6
			AND dds.docno IN (%s)
			GROUP BY dds.docno
		)
		-- UPDATE ทั้งหมดในครั้งเดียว
		UPDATE doc
		SET 
			isref = COALESCE(dir.new_isref, doc.isref),
			iscomparedsuccess = CASE
				WHEN dc.totalreceived IS NULL THEN doc.iscomparedsuccess
				WHEN dc.totalreceived = 0 THEN 0
				WHEN dc.totalreceived = dc.totalordered THEN 1
				WHEN dc.totalreceived < dc.totalordered THEN 2
				WHEN dc.totalreceived > dc.totalordered THEN 3
				ELSE doc.iscomparedsuccess
			END,
			isclosed = CASE
				WHEN dc.totalreceived IS NULL THEN doc.isclosed
				WHEN dc.totalreceived = 0 THEN false
				WHEN dc.totalreceived = dc.totalordered THEN true
				WHEN dc.totalreceived > dc.totalordered THEN true
				ELSE false
			END
		FROM doc_isref dir
		LEFT JOIN doc_comparison dc ON dir.docno = dc.docno
		WHERE doc.docno = dir.docno
		AND doc.transflag = 6
		AND doc.docno IN (%s)
	`, docNoList, docNoList, docNoList)

	result, err := db.ExecContext(ctx, megaQuery)
	if err != nil {
		logger.Error("Failed to execute mega-query for specific docs: %v", err)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	logger.Success("✅ MEGA-QUERY completed: Updated %d documents", rowsAffected)
	logger.Success("ProcessDocPurchaseDirectBatchByDocNos completed for shop %s", shopId)
}

// joinStringSlice - helper function to join string slice
func joinStringSlice(arr []string, separator string) string {
	if len(arr) == 0 {
		return ""
	}
	result := arr[0]
	for i := 1; i < len(arr); i++ {
		result += separator + arr[i]
	}
	return result
}
