package mypg

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/models"
	"smlcloudplatform/internal/goapi/myglobal"

	"github.com/lib/pq"
)

func InsertDocListToPostgreSql(ctx context.Context, db *sql.DB, data []models.DocStruct, docRefData []models.DocRefStruct, docPaymentData []models.DocPaymentStruct) error {
	if db == nil {
		return fmt.Errorf("database connection is nil")
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	if err := insertDocListWithTx(ctx, tx, data, docRefData, docPaymentData); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func InsertDocListTx(ctx context.Context, tx *sql.Tx, data []models.DocStruct, docRefData []models.DocRefStruct, docPaymentData []models.DocPaymentStruct) error {
	if tx == nil {
		return fmt.Errorf("transaction is nil")
	}
	return insertDocListWithTx(ctx, tx, data, docRefData, docPaymentData)
}

func insertDocListWithTx(ctx context.Context, tx *sql.Tx, data []models.DocStruct, docRefData []models.DocRefStruct, docPaymentData []models.DocPaymentStruct) error {
	if len(data) == 0 && len(docRefData) == 0 && len(docPaymentData) == 0 {
		return nil
	}

	totalInserted := 0
	totalErrors := 0
	totalSkipped := 0

	// 1. Insert Doc Records
	if len(data) > 0 {

		// Data validation
		validData := make([]models.DocStruct, 0, len(data))
		for _, doc := range data {
			if doc.DocNo == "" {
				totalSkipped++
				continue
			}
			if doc.DocDateTime.IsZero() {
				totalSkipped++
				continue
			}
			validData = append(validData, doc)
		}

		if len(validData) > 0 {
			// เตรียม columns สำหรับ COPY doc
			columns := []string{
				"transflag", "docno", "custcode", "docdatetime", "perioddatetime", "taxdocno",
				"totalamount", "roundamount", "paytype", "paycashamount", "paycashchange",
				"paycashbalance", "deliverycode", "checksum", "branchid", "slipurl",
				"salechannelcode", "deliveryamount", "iscancel", "cancelreason",
				"guidpos", "guidbranch", "guidfixed",
				"creatorcode", "creatorname", "createdat",
				// Multi-Currency Fields
				"currency", "currencysymbol",
				"doccurrency", "doccurrencysymbol",
				"exchangerate", "totalamountdoc",
				// Soft Delete
				"isdelete",
				// สถานะการอนุมัติ
				"approvalstatus",
			}

			// แปลง Doc เป็น [][]any สำหรับ COPY
			rows := make([][]any, len(validData))
			for i, data := range validData {
				// กำหนดค่า createdat ถ้ายังไม่มี
				createdAt := data.CreatedAt
				if createdAt.IsZero() {
					createdAt = data.DocDateTime // ใช้ DocDateTime เป็น default
				}

				rows[i] = []any{
					data.TransFlag,
					data.DocNo,
					data.CustCode,
					data.DocDateTime,
					data.PeriodDateTime,
					data.TaxDocNo,
					data.TotalAmount,
					data.RoundAmount,
					data.PayType,
					data.PayCashAmount,
					data.PayCashChange,
					data.PayCashBalance,
					data.DeliveryCode,
					data.Checksum,
					data.BranchID,
					data.SlipURL,
					data.SaleChannelCode,
					data.DeliveryAmount,
					data.IsCancel,
					data.CancelReason,
					data.GuidPOS,
					data.GuidBranch,
					data.GuidFixed,
					data.CreatorCode,
					data.CreatorName,
					createdAt,
					// Multi-Currency Fields
					data.Currency,
					data.CurrencySymbol,
					data.DocCurrency,
					data.DocCurrencySymbol,
					data.ExchangeRate,
					data.TotalAmountDoc,
					// Soft Delete
					data.IsDelete,
					// สถานะการอนุมัติ
					data.ApprovalStatus,
				}
			}

			// ใช้ BulkInsertWithCopy จาก mypg package
			if err := BulkInsertWithCopy(ctx, tx, "doc", columns, rows); err != nil {
				return fmt.Errorf("insert doc: %w", err)
			}
			totalInserted += len(validData)
		}
	}

	// 2. Insert DocRef Records
	if len(docRefData) > 0 {
		// Data validation
		validRefData := make([]models.DocRefStruct, 0, len(docRefData))
		for _, item := range docRefData {
			if item.DocNo == "" || item.DocRefNo == "" {
				totalErrors++
				continue
			}
			validRefData = append(validRefData, item)
		}

		if len(validRefData) > 0 {
			// เตรียม columns สำหรับ COPY doc ref
			columns := []string{
				"docno", "docnotransflag", "docnoref", "docnoreftransflag",
			}

			// แปลง DocRef เป็น [][]any สำหรับ COPY
			rows := make([][]any, len(validRefData))
			for i, item := range validRefData {
				rows[i] = []any{
					item.DocNo,
					item.DocNoTransFlag,
					item.DocRefNo,
					item.DocRefNoTransFlag,
				}
			}

			if err := BulkInsertWithCopy(ctx, tx, "docref", columns, rows); err != nil {
				return fmt.Errorf("insert docref: %w", err)
			}
			totalInserted += len(validRefData)
		}
	}

	// 3. Insert DocPayment Records (skip for PO/SO/Transfer)
	if len(docPaymentData) > 0 {

		// สร้าง map เก็บ transflag จากหัวเอกสาร (doc)
		docTransFlagMap := make(map[string]int32)
		for _, doc := range data {
			docTransFlagMap[doc.DocNo] = int32(doc.TransFlag)
		}

		// Data validation และกรองเฉพาะเอกสารที่มีการชำระเงิน
		validPaymentData := make([]models.DocPaymentStruct, 0, len(docPaymentData))
		for _, doc := range docPaymentData {
			if doc.DocNo == "" {
				totalErrors++
				continue
			}

			// ดึง transflag จากหัวเอกสาร
			transflag := doc.TransFlag
			if transflag == 0 {
				if headerTransFlag, exists := docTransFlagMap[doc.DocNo]; exists {
					transflag = headerTransFlag
				} else {
					totalErrors++
					continue
				}
			}

			// ข้ามเอกสารที่ยังไม่มีการชำระเงิน (PO=6, SO=36, Transfer=72)
			if transflag == 6 || transflag == 36 || transflag == 72 {
				totalSkipped++
				continue
			}

			// อัปเดต transflag ให้ถูกต้อง
			doc.TransFlag = transflag
			validPaymentData = append(validPaymentData, doc)
		}

		if len(validPaymentData) > 0 {
			// เตรียม columns สำหรับ COPY doc payment
			columns := []string{
				"branchid", "docdatetime", "perioddatetime", "providername", "amount",
				"description", "docno", "transflag", "guidfixed", "guidbranch",
			}

			// แปลง DocPayment เป็น [][]any สำหรับ COPY
			rows := make([][]any, len(validPaymentData))
			for i, payItem := range validPaymentData {
				rows[i] = []any{
					payItem.BranchID,
					payItem.DocDateTime,
					payItem.PeriodDateTime,
					payItem.ProviderName,
					payItem.Amount,
					payItem.Description,
					payItem.DocNo,
					payItem.TransFlag,
					payItem.GuidFixed,
					payItem.GuidBranch,
				}
			}

			if err := BulkInsertWithCopy(ctx, tx, "docpayment", columns, rows); err != nil {
				return fmt.Errorf("insert docpayment: %w", err)
			}
			totalInserted += len(validPaymentData)
		}
	}

	// Log only errors
	if totalErrors > 0 {
		logger.Warn("PG insert: %d inserted, %d errors", totalInserted, totalErrors)
	}
	return nil
}

func InsertDocDetailListToPostgreSql(ctx context.Context, pgdb *sql.DB, holdingCode string, data []models.DocDetailStruct) error {
	if pgdb == nil {
		return fmt.Errorf("database connection is nil")
	}

	tx, err := pgdb.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	if err := insertDocDetailListWithTx(ctx, tx, holdingCode, data); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func InsertDocDetailListToPostgreSqlTx(ctx context.Context, tx *sql.Tx, holdingCode string, data []models.DocDetailStruct) error {
	if tx == nil {
		return fmt.Errorf("transaction is nil")
	}
	return insertDocDetailListWithTx(ctx, tx, holdingCode, data)
}

func insertDocDetailListWithTx(ctx context.Context, tx *sql.Tx, holdingCode string, data []models.DocDetailStruct) error {
	if len(data) == 0 {
		return nil
	}

	validData := make([]models.DocDetailStruct, 0, len(data))
	skippedCount := 0

	for _, docDetail := range data {
		if docDetail.DocNo == "" || (docDetail.ItemCode == "" && docDetail.Barcode == "") {
			skippedCount++
			continue
		}
		validData = append(validData, docDetail)
	}

	if len(validData) == 0 {
		return nil
	}

	columns := []string{
		"docdatetime", "docno", "linenumber", "transflag", "calcflag", "calcseq",
		"itemcode", "description", "barcodemain", "barcode", "unitcode",
		"whcode", "locationcode", "totalqty", "price", "priceexcludevat",
		"unitstand", "unitdivide", "docref", "sumamount", "iscancel",
		"pricedoc", "sumamountdoc",
		"discountamountdoc", "priceexcludevatdoc", "sumamountexcludevatdoc", "totalvaluevatdoc",
	}

	// Pre-fetch product info to avoid post-insert UPDATE
	uniqueBarcodes := make(map[string]struct{})
	for _, item := range validData {
		if item.Barcode != "" {
			uniqueBarcodes[item.Barcode] = struct{}{}
		}
	}

	if len(uniqueBarcodes) > 0 {
		barcodes := make([]string, 0, len(uniqueBarcodes))
		for barcode := range uniqueBarcodes {
			barcodes = append(barcodes, barcode)
		}

		// Query productbarcode
		queryProduct := `
			SELECT barcode, itemcode, barcoderefunitstand, barcoderefunitdivide
			FROM productbarcode
			WHERE barcode = ANY($1)
		`

		rows, err := tx.QueryContext(ctx, queryProduct, pq.Array(barcodes))
		if err == nil {
			defer rows.Close()

			productMap := make(map[string]struct {
				ItemCode   string
				UnitStand  float64
				UnitDivide float64
			})

			for rows.Next() {
				var barcode, itemCode string
				var unitStand, unitDivide float64
				if err := rows.Scan(&barcode, &itemCode, &unitStand, &unitDivide); err == nil {
					productMap[barcode] = struct {
						ItemCode   string
						UnitStand  float64
						UnitDivide float64
					}{itemCode, unitStand, unitDivide}
				}
			}

			// Update validData in memory
			for i := range validData {
				if info, ok := productMap[validData[i].Barcode]; ok {
					if validData[i].ItemCode == "" || validData[i].ItemCode != info.ItemCode {
						validData[i].ItemCode = info.ItemCode
					}
					validData[i].UnitStand = info.UnitStand
					validData[i].UnitDivide = info.UnitDivide
				}
			}
		}
	}

	rows := make([][]any, len(validData))
	deleteTargets := make(map[string]struct{})
	docNos := make(map[string]struct{})
	lineNumbers := 0

	for i, item := range validData {
		whCode := item.WhCode
		locationCode := item.LocationCode
		if whCode == "" {
			whCode = "X"
		}
		if locationCode == "" {
			locationCode = "X"
		}

		description := item.Description
		if description == "" {
			description = item.ItemCode
			if description == "" {
				description = "UNKNOWN"
			}
		}

		barcode := item.Barcode
		if barcode == "" {
			barcode = item.ItemCode
			if barcode == "" {
				barcode = "UNKNOWN"
			}
		}

		barcodeMain := item.BarcodeMain
		if barcodeMain == "" {
			barcodeMain = barcode
		}

		unitStand := item.UnitStand
		unitDivide := item.UnitDivide
		if unitStand == 0 {
			unitStand = 1.0
		}
		if unitDivide == 0 {
			unitDivide = 1.0
		}

		calcFlag := item.CalcFlag
		if calcFlag == 0 {
			calcFlag = myglobal.GetTransactionMultiplier(item.TransFlag)
		}

		lineNumbers++

		rows[i] = []any{
			item.DocDateTime,
			item.DocNo,
			lineNumbers,
			item.TransFlag,
			int(calcFlag),
			item.CalcSeq,
			item.ItemCode,
			description,
			barcodeMain,
			barcode,
			item.UnitCode,
			whCode,
			locationCode,
			item.TotalQty,
			item.Price,
			item.PriceExcludeVat,
			unitStand,
			unitDivide,
			item.DocRef,
			item.SumAmount,
			false, // iscancel
			item.PriceDoc,
			item.SumAmountDoc,
			item.DiscountAmountDoc,
			item.PriceExcludeVatDoc,
			item.SumAmountExcludeVatDoc,
			item.TotalValueVatDoc,
		}

		key := fmt.Sprintf("%s|%d", item.DocNo, item.TransFlag)
		deleteTargets[key] = struct{}{}
		docNos[item.DocNo] = struct{}{}
	}

	// Group docNos by TransFlag for batch delete
	docNosByTransFlag := make(map[int][]string)
	for key := range deleteTargets {
		parts := strings.Split(key, "|")
		if len(parts) != 2 {
			continue
		}
		transFlag, _ := strconv.Atoi(parts[1])
		docNo := parts[0]
		docNosByTransFlag[transFlag] = append(docNosByTransFlag[transFlag], docNo)
	}

	// Execute batch delete
	for transFlag, docNos := range docNosByTransFlag {
		if len(docNos) == 0 {
			continue
		}

		// Use ANY for bulk delete
		query := "DELETE FROM docdetail WHERE transflag = $1 AND docno = ANY($2)"
		if _, err := tx.ExecContext(ctx, query, transFlag, pq.Array(docNos)); err != nil {
			return fmt.Errorf("failed to batch delete existing docdetail for transflag=%d: %w", transFlag, err)
		}
	}

	if err := BulkInsertWithCopy(ctx, tx, "docdetail", columns, rows); err != nil {
		return err
	}

	return nil
}
