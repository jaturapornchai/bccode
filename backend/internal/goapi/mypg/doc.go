package mypg

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"smlcloudplatform/internal/goapi/logger"

	"smlcloudplatform/internal/goapi/models"
	"smlcloudplatform/internal/goapi/myglobal"

	"github.com/google/uuid"
)

// DocUpdate ทำหน้าที่อัพเดทข้อมูลเอกสารในฐานข้อมูล PostgreSQL
func xxDocUpdate(docData models.MongoDocModel) error {
	db, err := PgSqlFastConnect(docData.ShopId)
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	ctx := context.Background()

	// สร้าง checksum สำหรับเช็คการเปลี่ยนแปลงข้อมูล
	checkSumMongodb := myglobal.CalculateMD5(fmt.Sprintf("%v", docData))
	if len(checkSumMongodb) != 32 {
		checkSumMongodb = uuid.NewString()
	}

	// เช็คว่ามีข้อมูลเอกสารนี้อยู่แล้วหรือไม่
	query := "SELECT checksum FROM doc WHERE docno = $1"
	dataRows, err := QuerySelectAll(db, query, docData.DocNo)
	if err != nil {
		return fmt.Errorf("failed to check existing document: %w", err)
	}

	// ถ้ามีข้อมูลอยู่แล้วและ checksum ตรงกัน ไม่ต้องอัพเดท
	if len(dataRows) > 0 {
		existingChecksum := dataRows[0]["checksum"].(string)
		if existingChecksum == checkSumMongodb {
			logger.Info("Document %s already up to date", docData.DocNo)
			return nil
		}
	}

	// เริ่ม transaction
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// ลบข้อมูลเก่า
	deleteQueries := []string{
		"DELETE FROM doc WHERE docno = $1",
		"DELETE FROM docdetail WHERE docno = $1",
		"DELETE FROM docpayment WHERE docno = $1",
	}

	for _, deleteQuery := range deleteQueries {
		_, err := tx.ExecContext(ctx, deleteQuery, docData.DocNo)
		if err != nil {
			return fmt.Errorf("failed to delete existing data: %w", err)
		}
	}

	logger.Info("Deleted existing document: %s", docData.DocNo)

	// แปลงวันที่เป็นรูปแบบที่ต้องการ
	payCashBalance := docData.PayCashAmount - docData.PayCashChange

	// Insert ข้อมูลหลักเอกสาร
	// สร้างตัวแปรเพิ่มเติมสำหรับ field ที่ขาดหาย
	var transFlag int = 1        // TODO: ต้องกำหนดค่าจริง
	var taxDocNo string = ""     // TODO: ต้องกำหนดค่าจริง
	var payType int = 1          // TODO: ต้องกำหนดค่าจริง
	var deliveryCode string = "" // TODO: ต้องกำหนดค่าจริง

	logger.Info("Inserting document %s %s", docData.DocNo, docData.CustCode)

	insertDocQuery := `
		INSERT INTO doc (custcode,
			transflag, docno, docdatetime, perioddatetime, taxdocno,
			totalamount, roundamount, paytype, paycashamount, paycashchange,
			paycashbalance, deliverycode, checksum, branchid, slipurl,
			salechannelcode, deliveryamount, iscancel, cancelreason,
			guidpos, guidbranch, guidfixed
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23
		)`

	_, err = tx.ExecContext(ctx, insertDocQuery,
		docData.CustCode,
		transFlag, docData.DocNo, docData.DocDateTime, docData.DocDateTime, taxDocNo,
		docData.TotalAmount, docData.RoundAmount, payType, docData.PayCashAmount, docData.PayCashChange,
		payCashBalance, deliveryCode, checkSumMongodb, docData.BranchId, docData.SlipUrl,
		docData.SaleChannelCode, docData.DeliveryAmount, docData.IsCancel, docData.CancelReason,
		docData.GuidPos, docData.Branch.GuidFixed, docData.GuidFixed)

	if err != nil {
		return fmt.Errorf("failed to insert doc: %w", err)
	}

	// Insert รายการสินค้า (docdetail)
	for i, detail := range docData.Details {
		// สร้างตัวแปรเพิ่มเติมสำหรับ field ที่ขาดหาย
		var docRef string = ""                     // TODO: ต้องกำหนดค่าจริง
		var transFlag int = 1                      // TODO: ต้องกำหนดค่าจริง
		var calcFlag int = 0                       // TODO: ต้องกำหนดค่าจริง
		var calcSeq int = 0                        // TODO: ต้องกำหนดค่าจริง
		var isCancel bool = false                  // TODO: ต้องกำหนดค่าจริง
		var itemCode string = detail.ItemCode      // TODO: ต้องกำหนดค่าจริง
		var barcodeMain string = detail.Barcode    // TODO: ต้องกำหนดค่าจริง
		var unitCode string = ""                   // TODO: ต้องกำหนดค่าจริง
		var whCode string = "X"                    // TODO: ต้องกำหนดค่าจริง
		var locationCode string = "X"              // TODO: ต้องกำหนดค่าจริง
		var totalQty float64 = detail.Qty          // TODO: ต้องกำหนดค่าจริง
		var unitStand float64 = 1.0                // TODO: ต้องกำหนดค่าจริง
		var unitDivide float64 = 1.0               // TODO: ต้องกำหนดค่าจริง
		var priceExcludeVat float64 = detail.Price // TODO: ต้องกำหนดค่าจริง

		insertDetailQuery := `
			INSERT INTO docdetail (
				docdatetime, docno, docref, linenumber, transflag, calcflag,
				calcseq, iscancel, itemcode, barcodemain, barcode, unitcode, whcode,
				locationcode, totalqty, unitstand, unitdivide, price, priceexcludevat
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19
			)`

		_, err = tx.ExecContext(ctx, insertDetailQuery,
			docData.DocDateTime, docData.DocNo, docRef, i+1, transFlag, calcFlag,
			calcSeq, isCancel, itemCode, barcodeMain, detail.Barcode, unitCode, whCode,
			locationCode, totalQty, unitStand, unitDivide, detail.Price, priceExcludeVat)

		if err != nil {
			return fmt.Errorf("failed to insert docdetail line %d: %w", i+1, err)
		}
	}

	// Insert การชำระเงิน (docpayment)
	if docData.PaymentDetailRaw != "" {
		var payments []map[string]any
		if err := json.Unmarshal([]byte(docData.PaymentDetailRaw), &payments); err == nil {
			for _, payment := range payments {
				amount, _ := payment["amount"].(float64)
				providerName := payment["provider_name"].(string)
				transFlag, _ := payment["trans_flag"].(float64)

				insertPaymentQuery := `
					INSERT INTO docpayment (
						branchid, docno, docdatetime, perioddatetime, 
						description, amount, trans_flag, guidfixed, guidbranch
					) VALUES (
						$1, $2, $3, $4, $5, $6, $7, $8, $9
					)`

				_, err = tx.ExecContext(ctx, insertPaymentQuery,
					docData.BranchId, docData.DocNo, docData.DocDateTime, docData.DocDateTime,
					providerName, amount, transFlag, docData.GuidFixed, docData.Branch.GuidFixed)

				if err != nil {
					return fmt.Errorf("failed to insert docpayment: %w", err)
				}
			}
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	logger.Success("updated document %s", docData.DocNo)
	return nil
}

type execContext interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func DeleteDocPgSql(ctx context.Context, db *sql.DB, docNo string, transFlag int) {
	if err := deleteDocPgSqlExec(ctx, db, docNo, transFlag); err != nil {
		logger.Warn("DeleteDocPgSql failed: %v", err)
	}
}

func DeleteDocPgSqlTx(ctx context.Context, tx *sql.Tx, docNo string, transFlag int) error {
	return deleteDocPgSqlExec(ctx, tx, docNo, transFlag)
}

func deleteDocPgSqlExec(ctx context.Context, exec execContext, docNo string, transFlag int) error {
	queries := []string{
		"DELETE FROM doc WHERE docno = $1 AND transflag = $2",
		"DELETE FROM docdetail WHERE docno = $1 AND transflag = $2",
	}

	for _, query := range queries {
		if _, err := exec.ExecContext(ctx, query, docNo, transFlag); err != nil {
			return fmt.Errorf("execute delete query '%s': %w", query, err)
		}
	}

	if _, err := exec.ExecContext(ctx, "DELETE FROM docpayment WHERE docno = $1", docNo); err != nil {
		return fmt.Errorf("delete docpayment: %w", err)
	}

	return nil
}

// addToDocWaitProcessQueues - adds documents to processing queues
func AddToDocWaitProcessQueues(ctx context.Context, db *sql.DB, docNo string, transFlag int) {

	// Add to doc wait process
	_, err := db.ExecContext(ctx,
		"INSERT INTO docwaitprocess (docno, transflag) VALUES ($1, $2)",
		docNo, transFlag)
	if err != nil {
		logger.Warn("Could not add to doc queue: %v", err)
	}

	// Add to stock wait process for all items in the document
	insertStockWaitQuery := `
		INSERT INTO stockwaitprocess (itemcode)
		SELECT DISTINCT itemcode
		FROM docdetail 
		WHERE docno = $1 AND transflag = $2`

	_, err = db.ExecContext(ctx, insertStockWaitQuery, docNo, transFlag)
	if err != nil {
		logger.Warn("Could not add to stock queue: %v", err)
	}
}
