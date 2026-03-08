package processdoc

import (
	"context"
	"database/sql"
	"smlcloudplatform/internal/goapi/logger"

	mypg "smlcloudplatform/internal/goapi/mypg"
)

func ProcessDocPurchase(db *sql.DB, docNo string) {
	logger.Info("Processing purchase document: %s", docNo)
	// ประมวลผลเอกสารใบสั่งซื้อ
	/*
		 	ค้นหาเอกสารที่เชื่อมโยง (เอกสารซื้อ, เอกสารรับสินค้า)
			- ถ้ามีเอกสารเชื่อมโยง isref = true
			- ถ้าไม่มีเอกสารเชื่อมโยง isref = false
			- ดึงรายการสินค้าในเอกสารใบเสนอราคา
			- เปรียบเทียบจำนวนสินค้าในเอกสารใบเสนอราคา ว่าซื้อครบหรือไม่ หรือรับสินค้าครบหรือไม่
			- ถ้ายังไม่มีอ้างอิง iscomparedsuccess = 0
			- ถ้ารับครบ iscomparedsuccess = 1
			- ถ้ายังไม่ครบ iscomparedsuccess = 2
			- ถ้าเกิน iscomparedsuccess = 3
			เสร็จแล้ว update table doc,docdetail ใช้ upsert
			* หมายเหตุ
			- การอ้างอิงเอกสาร เป็นแบบ one-to-many (เอกสารซื้อ 1 ใบ อ้างอิงเอกสารรับสินค้าได้หลายใบได้, หรืออ้างอิงเอกสารซื้อหลายใบได้)
	*/
	// ดึงเอกสารต้นทาง (ใบสั่งซื้อ, transflag=6)
	var isrefOld, islockedOld, isclosedOld bool
	var iscomparedsuccessOld int
	isref := false
	islocked := false
	isclosed := false
	iscomparedsuccess := 0
	{
		queryDoc := `
			SELECT isref, iscomparedsuccess, islocked, isclosed
			FROM doc
			WHERE docno = $1 AND transflag = 6
		`
		rows, err := db.Query(queryDoc, docNo)
		if err != nil {
			logger.Error("querying doc for %s: %v", docNo, err)
			return
		}

		if rows.Next() {
			if err := rows.Scan(&isrefOld, &iscomparedsuccessOld, &islockedOld, &isclosedOld); err != nil {
				logger.Error("scanning doc for %s: %v", docNo, err)
				return
			}
		}
		rows.Close()
	}
	{
		// ค้นหาเอกสารที่อ้างอิง (310,12)
		queryFindDocRef := `
			SELECT count(*) as countdoc
			FROM docref
			WHERE docnoref = $1 AND docnotransflag IN (310,12) group by docno
		`
		rows, err := db.Query(queryFindDocRef, docNo)
		if err != nil {
			logger.Error("querying doc references for %s: %v", docNo, err)
			return
		}
		if rows.Next() {
			var countDoc int
			if err := rows.Scan(&countDoc); err != nil {
				logger.Error("scanning doc reference count for %s: %v", docNo, err)
			}
			if countDoc > 0 {
				isref = true
			}
		}
		rows.Close()
	}
	{
		// ตรวจสอบรายการสินค้าในเอกสารใบสั่งซื้อ ว่ารับสินค้าครบหรือไม่ (auto packing)
		// เอกสารใบสั่งซื้อ docdetail.transflag = 6
		// เอกสารรับสินค้า docdetail.transflag = 310,12
		// totalqty = จำนวน อัตราส่วน 1:1
		// unitstand = หน่วยตัวตั้ง
		// unitdivide = หน่วยตัวแยก
		queryCompare := `
            SELECT 
                SUM(COALESCE(dds.totalqty * (dds.unitstand / NULLIF(dds.unitdivide, 0)), dds.totalqty)) AS totalordered,
                COALESCE(SUM(drc.totalqty * (drc.unitstand / NULLIF(drc.unitdivide, 0))), 0) AS totalreceived
            FROM docdetail dds
            LEFT JOIN (
                SELECT
                    dd.itemcode,
                    SUM(dd.totalqty) AS totalqty,
                    MAX(dd.unitstand) AS unitstand,
                    MAX(dd.unitdivide) AS unitdivide
                FROM docdetail dd
                INNER JOIN docref dr ON dd.docno = dr.docno
                WHERE dd.transflag IN (310,12) AND dr.docnoref = $1
                GROUP BY dd.itemcode
            ) AS drc ON dds.itemcode = drc.itemcode
            WHERE dds.docno = $1 AND dds.transflag = 6
        `
		// เพิ่ม query ที่พร้อมใช้ใน pgAdmin (แทนที่ $1 ด้วยค่าจริง)
		/*queryForPgAdmin := strings.ReplaceAll(queryCompare, "$1", "'"+docNo+"'")
		logger.Info("Query for pgAdmin: %s", queryForPgAdmin)*/

		rows, err := db.Query(queryCompare, docNo)
		if err != nil {
			logger.Error("querying docdetail comparison for %s: %v", docNo, err)
			return
		}
		if rows.Next() {
			var totalOrdered, totalReceived float64
			if err := rows.Scan(&totalOrdered, &totalReceived); err != nil {
				logger.Error("scanning docdetail comparison for %s: %v", docNo, err)
			}
			logger.Info("Document %s: totalOrdered=%.2f, totalReceived=%.2f", docNo, totalOrdered, totalReceived)
			if totalReceived == 0 {
				iscomparedsuccess = 0 // ยังไม่ได้รับสินค้า
			} else if totalReceived == totalOrdered {
				iscomparedsuccess = 1 // รับครบ
				isclosed = true       // ปิดเอกสารเมื่อรับครบ
			} else if totalReceived < totalOrdered {
				iscomparedsuccess = 2 // รับไม่ครบ
			} else if totalReceived > totalOrdered {
				iscomparedsuccess = 3 // รับเกิน
				isclosed = true       // ปิดเอกสารเมื่อรับครบ
			}
		}
		rows.Close()
	}
	{
		if isref != isrefOld || iscomparedsuccess != iscomparedsuccessOld || islocked != islockedOld || isclosed != isclosedOld {
			// update isdocref in doc table
			queryUpdateDoc := `
			UPDATE doc
			SET isref = $1, iscomparedsuccess = $2, islocked = $3, isclosed = $4
			WHERE docno = $5 AND transflag = 6
		`
			_, err := db.Exec(queryUpdateDoc, isref, iscomparedsuccess, islocked, isclosed, docNo)
			if err != nil {
				logger.Error("updating isref for %s: %v", docNo, err)
				return
			}
		}
	}
}

func ProcessDocPurchaseAll(shopId string) {
	logger.Info("Starting ProcessDocPurchaseAll for shop %s", shopId)
	ctx := context.Background()

	db, err := mypg.PgSqlFastConnect(shopId)
	if err != nil {
		logger.Info("Failed to connect to PostgreSQL: %v", err)
		return
	}

	// ดึง docno ทั้งหมดที่ยังไม่ถูกประมวลผล ใบสั่งซื้อ (transflag = 6)
	query := "SELECT DISTINCT docno FROM docwaitprocess WHERE transflag = 6"
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		logger.Error("querying docno: %v", err)
		return
	}

	// เก็บรายการ docno ทั้งหมดก่อน
	var docList []string
	for rows.Next() {
		var docNo string
		if err := rows.Scan(&docNo); err != nil {
			logger.Error("scanning docno: %v", err)
			continue
		}
		docList = append(docList, docNo)
	}
	rows.Close()

	if len(docList) == 0 {
		logger.Debug("No purchase documents to process for shop %s", shopId)
		return
	}

	logger.Info("Processing %d purchase documents for shop %s", len(docList), shopId)

	// ประมวลผลแบบ batch (ไม่ใช้ goroutine เพื่อความรวดเร็ว)
	successCount := 0
	errorCount := 0

	for _, docNo := range docList {
		// ประมวลผลเอกสาร
		ProcessDocPurchase(db, docNo)

		// ลบ docNo จาก docwaitprocess หลังประมวลผลเสร็จ
		queryDelete := "DELETE FROM docwaitprocess WHERE docno=$1 AND transflag=6"
		_, err := db.ExecContext(ctx, queryDelete, docNo)
		if err != nil {
			logger.Error("deleting docno %s from docwaitprocess: %v", docNo, err)
			errorCount++
		} else {
			successCount++
		}
	}

	logger.Success("ProcessDocPurchaseAll completed for shop %s: processed %d docs (%d success, %d errors)",
		shopId, len(docList), successCount, errorCount)
}
