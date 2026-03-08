package processdoc

import (
	"context"
	"fmt"
	"smlcloudplatform/internal/goapi/logger"
	"strings"

	mypg "smlcloudplatform/internal/goapi/mypg"
)

// ProcessDocPurchaseBatch - ประมวลผลสถานะเอกสารแบบ batch สำหรับ rebuild (เร็วกว่าแบบ 1:1)
// ใช้ SQL batch queries แทนการ loop ทีละเอกสาร
func ProcessDocPurchaseBatch(shopId string) {
	logger.Info("Starting ProcessDocPurchaseBatch for shop %s (OPTIMIZED FOR REBUILD)", shopId)
	ctx := context.Background()

	db, err := mypg.PgSqlFastConnect(shopId)
	if err != nil {
		logger.Error("Failed to connect to PostgreSQL: %v", err)
		return
	}

	// Step 1: UPDATE isref ทั้งหมดในครั้งเดียว (Batch Update)
	logger.Debug("Step 1: Batch updating isref for all purchase orders...")
	queryUpdateIsRef := `
		UPDATE doc
		SET isref = CASE 
			WHEN EXISTS (
				SELECT 1 FROM docref 
				WHERE docref.docnoref = doc.docno 
				AND docref.docnotransflag IN (310, 12)
			) THEN true
			ELSE false
		END
		WHERE doc.transflag = 6
		AND EXISTS (SELECT 1 FROM docwaitprocess WHERE docwaitprocess.docno = doc.docno AND docwaitprocess.transflag = 6)
	`
	result, err := db.ExecContext(ctx, queryUpdateIsRef)
	if err != nil {
		logger.Error("Failed to batch update isref: %v", err)
	} else {
		rowsAffected, _ := result.RowsAffected()
		logger.Success("Step 1 completed: Updated isref for %d documents", rowsAffected)
	}

	// Step 2: UPDATE iscomparedsuccess และ isclosed ทั้งหมดในครั้งเดียว (Batch Update)
	logger.Debug("Step 2: Batch updating iscomparedsuccess and isclosed...")
	queryUpdateStatus := `
		WITH doc_comparison AS (
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
			AND EXISTS (SELECT 1 FROM docwaitprocess WHERE docwaitprocess.docno = dds.docno AND docwaitprocess.transflag = 6)
			GROUP BY dds.docno
		)
		UPDATE doc
		SET 
			iscomparedsuccess = CASE
				WHEN dc.totalreceived = 0 THEN 0
				WHEN dc.totalreceived = dc.totalordered THEN 1
				WHEN dc.totalreceived < dc.totalordered THEN 2
				WHEN dc.totalreceived > dc.totalordered THEN 3
				ELSE 0
			END,
			isclosed = CASE
				WHEN dc.totalreceived = 0 THEN false
				WHEN dc.totalreceived = dc.totalordered THEN true
				WHEN dc.totalreceived > dc.totalordered THEN true
				ELSE false
			END
		FROM doc_comparison dc
		WHERE doc.docno = dc.docno AND doc.transflag = 6
	`
	result, err = db.ExecContext(ctx, queryUpdateStatus)
	if err != nil {
		logger.Error("Failed to batch update status: %v", err)
	} else {
		rowsAffected, _ := result.RowsAffected()
		logger.Success("Step 2 completed: Updated status for %d documents", rowsAffected)
	}

	// Step 3: ลบจาก docwaitprocess ทั้งหมดในครั้งเดียว (Batch Delete)
	logger.Debug("Step 3: Batch deleting from docwaitprocess...")
	queryDelete := "DELETE FROM docwaitprocess WHERE transflag = 6"
	result, err = db.ExecContext(ctx, queryDelete)
	if err != nil {
		logger.Error("Failed to batch delete from docwaitprocess: %v", err)
	} else {
		rowsAffected, _ := result.RowsAffected()
		logger.Success("Step 3 completed: Deleted %d records from docwaitprocess", rowsAffected)
	}

	logger.Success("ProcessDocPurchaseBatch completed for shop %s (BATCH MODE - FAST!)", shopId)
}

// ProcessDocPurchaseBatchByDocNos - ประมวลผลเฉพาะ docno ที่ระบุแบบ batch (สำหรับ Kafka ที่มีหลายเอกสาร)
func ProcessDocPurchaseBatchByDocNos(shopId string, docNos []string) {
	if len(docNos) == 0 {
		return
	}

	logger.Info("Starting ProcessDocPurchaseBatchByDocNos for shop %s with %d documents", shopId, len(docNos))
	ctx := context.Background()

	db, err := mypg.PgSqlFastConnect(shopId)
	if err != nil {
		logger.Error("Failed to connect to PostgreSQL: %v", err)
		return
	}

	// สร้าง IN clause
	placeholders := make([]string, len(docNos))
	args := make([]interface{}, len(docNos))
	for i, docNo := range docNos {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = docNo
	}
	inClause := strings.Join(placeholders, ",")

	// Step 1: UPDATE isref สำหรับ docno ที่ระบุ
	queryUpdateIsRef := fmt.Sprintf(`
		UPDATE doc
		SET isref = CASE 
			WHEN EXISTS (
				SELECT 1 FROM docref 
				WHERE docref.docnoref = doc.docno 
				AND docref.docnotransflag IN (310, 12)
			) THEN true
			ELSE false
		END
		WHERE doc.transflag = 6 AND doc.docno IN (%s)
	`, inClause)

	result, err := db.ExecContext(ctx, queryUpdateIsRef, args...)
	if err != nil {
		logger.Error("Failed to batch update isref: %v", err)
	} else {
		rowsAffected, _ := result.RowsAffected()
		logger.Debug("Updated isref for %d documents", rowsAffected)
	}

	// Step 2: UPDATE iscomparedsuccess และ isclosed สำหรับ docno ที่ระบุ
	queryUpdateStatus := fmt.Sprintf(`
		WITH doc_comparison AS (
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
			WHERE dds.transflag = 6 AND dds.docno IN (%s)
			GROUP BY dds.docno
		)
		UPDATE doc
		SET 
			iscomparedsuccess = CASE
				WHEN dc.totalreceived = 0 THEN 0
				WHEN dc.totalreceived = dc.totalordered THEN 1
				WHEN dc.totalreceived < dc.totalordered THEN 2
				WHEN dc.totalreceived > dc.totalordered THEN 3
				ELSE 0
			END,
			isclosed = CASE
				WHEN dc.totalreceived = 0 THEN false
				WHEN dc.totalreceived = dc.totalordered THEN true
				WHEN dc.totalreceived > dc.totalordered THEN true
				ELSE false
			END
		FROM doc_comparison dc
		WHERE doc.docno = dc.docno AND doc.transflag = 6
	`, inClause)

	result, err = db.ExecContext(ctx, queryUpdateStatus, args...)
	if err != nil {
		logger.Error("Failed to batch update status: %v", err)
	} else {
		rowsAffected, _ := result.RowsAffected()
		logger.Debug("Updated status for %d documents", rowsAffected)
	}

	// Step 3: ลบจาก docwaitprocess สำหรับ docno ที่ระบุ
	queryDelete := fmt.Sprintf("DELETE FROM docwaitprocess WHERE transflag = 6 AND docno IN (%s)", inClause)
	result, err = db.ExecContext(ctx, queryDelete, args...)
	if err != nil {
		logger.Error("Failed to batch delete from docwaitprocess: %v", err)
	} else {
		rowsAffected, _ := result.RowsAffected()
		logger.Debug("Deleted %d records from docwaitprocess", rowsAffected)
	}

	logger.Success("ProcessDocPurchaseBatchByDocNos completed for %d documents", len(docNos))
}
