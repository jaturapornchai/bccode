package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime/debug"
	"smlcloudplatform/internal/goapi/handlers/approval"
	"smlcloudplatform/internal/goapi/handlers/datahistory"
	"smlcloudplatform/internal/goapi/logger"
	"time"

	"smlcloudplatform/internal/goapi/models"
	"smlcloudplatform/internal/goapi/myglobal"
	"smlcloudplatform/internal/goapi/mypg"
	processdoc "smlcloudplatform/internal/goapi/process/process-doc"
)

// OnConsumeMessagePurchaseOrderCreateOrUpdate - handles purchase order create/update messages
func OnConsumeMessagePurchaseOrderCreateOrUpdate(msg string) error {
	return SafeConsumerWrapper("PURCHASE_ORDER", func(msg string) error {
		return ProcessPurchaseOrderDocument(msg)
	})(msg)
}

// OnConsumeMessagePurchaseOrderDelete - handles purchase order delete messages
func OnConsumeMessagePurchaseOrderDelete(msg string) error {
	docData := TransPurchaseOrderDecode(msg)
	if docData.HoldingCode == "" || docData.DocNo == "" {
		return fmt.Errorf("invalid purchase order data")
	}

	// บันทึก history ก่อนลบ
	dataBefore := convertPOToMap(docData)
	go func() {
		// ใช้ CreatorCode/CreatorName จาก document ถ้ามี ไม่งั้นใช้ system
		userCode := docData.CreatorCode
		userName := docData.CreatorName
		if userCode == "" {
			userCode = "system"
		}
		if userName == "" {
			userName = "System"
		}
		if err := datahistory.SavePOHistory(
			docData.HoldingCode,
			docData.DocNo,
			docData.GuidFixed,
			userCode,
			userName,
			datahistory.ActionDelete,
			dataBefore,
			nil,
		); err != nil {
			logger.Error("[DataHistory] Failed to save PO delete history: %v", err)
		}
	}()

	return DeleteDocumentFromDatabases(context.Background(), docData.HoldingCode, docData.DocNo, TRANS_FLAG_PURCHASE_ORDER)
}

// ProcessPurchaseOrderDocument - processes purchase order using build-doc system
func ProcessPurchaseOrderDocument(msg string) error {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("PANIC in ProcessPurchaseOrderDocument: %v\n%s", r, debug.Stack())
		}
	}()

	// Decode purchase order
	purchaseOrderData := TransPurchaseOrderDecode(msg)
	if purchaseOrderData.HoldingCode == "" || purchaseOrderData.DocNo == "" {
		return fmt.Errorf("invalid purchase order data - missing HoldingCode or DocNo")
	}

	// Convert to process model
	processData := ConvertPurchaseOrderMongoDocToProcessModel(purchaseOrderData)

	// Connect to database
	db, err := mypg.PgSqlFastConnect(purchaseOrderData.HoldingCode)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	ctx := context.Background()

	// อ่านค่า isclosedmanual ก่อน DELETE เพื่อ preserve (เพราะ field นี้ set จาก goapi ไม่ใช่จาก MongoDB)
	var preservedIsClosedManual bool
	var preservedClosedManualByCode, preservedClosedManualByName, preservedClosedManualReason string
	var preservedClosedManualAt time.Time
	_ = db.QueryRowContext(ctx,
		"SELECT COALESCE(isclosedmanual, false), COALESCE(closedmanual_by_code, ''), COALESCE(closedmanual_by_name, ''), COALESCE(closedmanual_at, '1970-01-01'), COALESCE(closedmanual_reason, '') FROM doc WHERE docno = $1 AND transflag = $2",
		purchaseOrderData.DocNo, TRANS_FLAG_PURCHASE_ORDER,
	).Scan(&preservedIsClosedManual, &preservedClosedManualByCode, &preservedClosedManualByName, &preservedClosedManualAt, &preservedClosedManualReason)

	// Begin transaction
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	txCommitted := false
	defer func() {
		if !txCommitted {
			tx.Rollback()
		}
	}()

	// Delete existing documents (upsert behavior)
	if err := mypg.DeleteDocPgSqlTx(ctx, tx, purchaseOrderData.DocNo, TRANS_FLAG_PURCHASE_ORDER); err != nil {
		return fmt.Errorf("failed to delete existing document: %w", err)
	}

	// Convert to build-doc structs
	docStruct, docPaymentStruct := myglobal.MapDocStructFromMongo(processData, purchaseOrderData.HoldingCode)
	docDetailStructs := MapPurchaseOrderToDocDetailStructs(processData, purchaseOrderData.HoldingCode)

	// Create doc references
	var docRefStructs []models.DocRefStruct
	for _, refNo := range processData.DocReferences {
		docRefStructs = append(docRefStructs, myglobal.MapDocRefStruct(processData.DocNo, processData.TransFlag, refNo))
	}

	// Insert to PostgreSQL
	if err = InsertDocumentToPostgreSQLTx(ctx, tx, docStruct, docRefStructs, docPaymentStruct, 0); err != nil {
		return fmt.Errorf("failed to insert document to PostgreSQL: %v", err)
	}

	// Insert document details
	if err = InsertDocDetailToPostgreSQLTx(ctx, tx, purchaseOrderData.HoldingCode, docDetailStructs, 0); err != nil {
		return err
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	txCommitted = true

	// Restore isclosedmanual หลัง INSERT (เพราะ DELETE+INSERT จะ reset ค่า)
	if preservedIsClosedManual {
		_, _ = db.ExecContext(ctx,
			"UPDATE doc SET isclosedmanual = $1, closedmanual_by_code = $2, closedmanual_by_name = $3, closedmanual_at = $4, closedmanual_reason = $5 WHERE docno = $6 AND transflag = $7",
			preservedIsClosedManual, preservedClosedManualByCode, preservedClosedManualByName, preservedClosedManualAt, preservedClosedManualReason,
			purchaseOrderData.DocNo, TRANS_FLAG_PURCHASE_ORDER,
		)
	}

	// Insert to ClickHouse (async)
	go func() {
		InsertDocumentToClickHouse(ctx, purchaseOrderData.HoldingCode, docStruct, docRefStructs, docPaymentStruct, docDetailStructs, 0)
	}()

	// Add to processing queue
	if err = AddDocToProcessQueue(purchaseOrderData.HoldingCode, purchaseOrderData.DocNo, TRANS_FLAG_PURCHASE_ORDER); err != nil {
		mypg.AddToDocWaitProcessQueues(ctx, db, purchaseOrderData.DocNo, TRANS_FLAG_PURCHASE_ORDER)
	}

	// Process document status
	processdoc.ProcessDocumentStatusByDocNo(purchaseOrderData.HoldingCode, purchaseOrderData.DocNo)

	// บันทึก history (create หรือ update) — เฉพาะเมื่อมีการเปลี่ยนแปลงจริง
	dataAfter := convertPOToMap(purchaseOrderData)
	go func() {
		// ตรวจสอบว่าเป็น create หรือ update โดยดูจาก history ที่มีอยู่
		lastSnapshot := datahistory.GetLastPOSnapshot(purchaseOrderData.HoldingCode, purchaseOrderData.GuidFixed)

		var action datahistory.ActionType
		if lastSnapshot == nil {
			action = datahistory.ActionCreate
		} else {
			action = datahistory.ActionUpdate
			// ถ้าเป็น update ให้ตรวจสอบว่ามีการเปลี่ยนแปลง field สำคัญจริงหรือไม่
			if !datahistory.HasMeaningfulChanges(lastSnapshot, dataAfter) {
				logger.Debug("[DataHistory] ข้ามการบันทึก PO history — ไม่มีการเปลี่ยนแปลง field สำคัญ (docno=%s)", purchaseOrderData.DocNo)
				return
			}
		}

		// เลือก user ตาม action:
		// - create → ใช้ CreatorCode/CreatorName
		// - update → ใช้ ModifierCode/ModifierName (ถ้ามี) ไม่งั้น fallback เป็น Creator
		var userCode, userName string
		if action == datahistory.ActionUpdate && purchaseOrderData.ModifierCode != "" {
			userCode = purchaseOrderData.ModifierCode
			userName = purchaseOrderData.ModifierName
		} else {
			userCode = purchaseOrderData.CreatorCode
			userName = purchaseOrderData.CreatorName
		}
		if userCode == "" {
			userCode = "system"
		}
		if userName == "" {
			userName = userCode // fallback เป็น userCode แทน "System"
		}

		if err := datahistory.SavePOHistory(
			purchaseOrderData.HoldingCode,
			purchaseOrderData.DocNo,
			purchaseOrderData.GuidFixed,
			userCode,
			userName,
			action,
			lastSnapshot, // dataBefore — ข้อมูลเก่าจาก history ล่าสุด
			dataAfter,
		); err != nil {
			logger.Error("[DataHistory] Failed to save PO history: %v", err)
		}
	}()

	// ถ้า PO ถูกยกเลิก ให้อัปเดตสถานะ approval เป็น cancelled
	if purchaseOrderData.IsCancel {
		logger.Info("[PO] Document %s is cancelled - updating approval status", purchaseOrderData.DocNo)
		if err := approval.UpdatePOApprovalStatusToCancelled(
			purchaseOrderData.HoldingCode,
			purchaseOrderData.DocNo,
			purchaseOrderData.CancelReason,
			"", // ไม่มีข้อมูล cancel user code ใน MongoDocModel
			"", // ไม่มีข้อมูล cancel user name ใน MongoDocModel
		); err != nil {
			logger.Error("[PO] Failed to update approval status to cancelled: %v", err)
			// ไม่ return error เพราะเป็น side effect ไม่ใช่ core functionality
		}
	}

	return nil
}

// MapPurchaseOrderToDocDetailStructs - converts purchase order to document detail structs
func MapPurchaseOrderToDocDetailStructs(processData models.ProcessMongoTransModel, holdingCode string) []models.DocDetailStruct {
	var docDetailStructs []models.DocDetailStruct

	for i, detail := range processData.Details {
		totalQty := detail.TotalQty
		if totalQty == 0 {
			totalQty = detail.Qty
		}

		unitStand := detail.StandValue
		if unitStand == 0 {
			unitStand = 1.0
		}

		unitDivide := detail.DivideValue
		if unitDivide == 0 {
			unitDivide = 1.0
		}

		// For purchase order transactions, no stock movement (CalcFlag = 0)
		docDetailStruct := models.DocDetailStruct{
			DocDateTime:     processData.DocDateTime,
			DocNo:           processData.DocNo,
			LineNumber:      i + 1,
			TransFlag:       TRANS_FLAG_PURCHASE_ORDER,
			CalcFlag:        0, // Purchase Order = no stock movement
			CalcSeq:         1,
			ItemCode:        detail.ItemCode,
			Description:     GetItemName(detail.ItemNames),
			BarcodeMain:     detail.Barcode,
			Barcode:         detail.Barcode,
			UnitCode:        detail.UnitCode,
			WhCode:          detail.WhCode,
			LocationCode:    detail.LocationCode,
			TotalQty:        totalQty,
			Price:           detail.Price,
			PriceExcludeVat: detail.PriceExcludeVat,
			UnitStand:       unitStand,
			UnitDivide:      unitDivide,
			DocRef:          detail.DocRef,
			SumAmount:       detail.SumAmount,
		}
		docDetailStructs = append(docDetailStructs, docDetailStruct)
	}

	return docDetailStructs
}

// TransPurchaseOrderDecode - decodes transaction purchase order JSON data
func TransPurchaseOrderDecode(jsonData string) models.MongoDocModel {
	var docData models.MongoDocModel
	DecodeKafkaMessage(jsonData, &docData, "PurchaseOrder")
	return docData
}

func ConvertPurchaseOrderMongoDocToProcessModel(mongoDoc models.MongoDocModel) models.ProcessMongoTransModel {
	var details []models.ProcessMongoTransDetailModel
	for _, detail := range mongoDoc.Details {
		itemNames := ConvertLanguageModels(detail.ItemNames)
		processDetail := models.ProcessMongoTransDetailModel{
			ItemCode:        detail.ItemCode,
			ItemNames:       itemNames,
			Barcode:         detail.Barcode,
			UnitCode:        detail.UnitCode,
			LineNumber:      detail.LineNumber,
			WhCode:          detail.WhCode,
			LocationCode:    detail.LocationCode,
			Qty:             detail.Qty,
			Price:           detail.Price,
			PriceExcludeVat: detail.PriceExcludeVat,
			DocRef:          detail.DocRef,
			SumAmount:       detail.SumAmount,
		}
		details = append(details, processDetail)
	}

	branchNames := ConvertLanguageModels(mongoDoc.Branch.Names)
	branch := models.BranchModel{
		Code:      mongoDoc.Branch.Code,
		GuidFixed: mongoDoc.Branch.GuidFixed,
		Names:     branchNames,
	}

	return models.ProcessMongoTransModel{
		HoldingCode:      mongoDoc.HoldingCode,
		BranchId:         mongoDoc.BranchId,
		GuidFixed:        mongoDoc.GuidFixed,
		CustCode:         mongoDoc.CustCode,
		DocNo:            mongoDoc.DocNo,
		DocDateTime:      mongoDoc.DocDateTime,
		TotalAmount:      mongoDoc.TotalAmount,
		RoundAmount:      mongoDoc.RoundAmount,
		PayCashAmount:    mongoDoc.PayCashAmount,
		PayCashChange:    mongoDoc.PayCashChange,
		PaymentDetailRaw: mongoDoc.PaymentDetailRaw,
		SlipUrl:          mongoDoc.SlipUrl,
		SaleChannelCode:  mongoDoc.SaleChannelCode,
		DeliveryAmount:   mongoDoc.DeliveryAmount,
		IsCancel:         mongoDoc.IsCancel,
		CancelReason:     mongoDoc.CancelReason,
		GuidPos:          mongoDoc.GuidPos,
		Branch:           branch,
		TransFlag:        TRANS_FLAG_PURCHASE_ORDER,
		Details:          details,
		DocReferences:    []models.ProcessMongoDocReferenceModel{},
		// Multi-Currency Fields
		Currency:          mongoDoc.Currency,
		CurrencySymbol:    mongoDoc.CurrencySymbol,
		DocCurrency:       mongoDoc.DocCurrency,
		DocCurrencySymbol: mongoDoc.DocCurrencySymbol,
		ExchangeRate:      mongoDoc.ExchangeRate,
		TotalAmountDoc:    mongoDoc.TotalAmountDoc,
		// Creator Fields
		CreatorCode: mongoDoc.CreatorCode,
		CreatorName: mongoDoc.CreatorName,
		CreatedAt:   mongoDoc.CreatedAt,
	}
}

// convertPOToMap - แปลง MongoDocModel เป็น map สำหรับเก็บ history
func convertPOToMap(doc models.MongoDocModel) map[string]interface{} {
	var result map[string]interface{}
	jsonBytes, err := json.Marshal(doc)
	if err != nil {
		logger.Error("[DataHistory] Failed to marshal doc: %v", err)
		return nil
	}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		logger.Error("[DataHistory] Failed to unmarshal doc: %v", err)
		return nil
	}
	return result
}
