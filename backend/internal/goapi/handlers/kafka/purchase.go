package kafka

import (
	"context"
	"fmt"
	"runtime/debug"

	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/models"
	"smlcloudplatform/internal/goapi/myglobal"
	"smlcloudplatform/internal/goapi/mypg"
	"smlcloudplatform/internal/goapi/process/build"
)

// OnConsumeMessagePurchaseCreateOrUpdate - handles purchase create/update messages
func OnConsumeMessagePurchaseCreateOrUpdate(msg string) error {
	return SafeConsumerWrapper("PURCHASE", func(msg string) error {
		return ProcessPurchaseDocument(msg)
	})(msg)
}

// OnConsumeMessagePurchaseDelete - handles purchase delete messages
func OnConsumeMessagePurchaseDelete(msg string) error {
	logger.Info("OnConsumeMessagePurchaseDelete: Processing deletion message")

	docData := TransPurchaseDecode(msg)
	if docData.HoldingCode == "" || docData.DocNo == "" {
		logger.Error("OnConsumeMessagePurchaseDelete: Invalid data - missing HoldingCode or DocNo")
		return fmt.Errorf("invalid purchase data")
	}

	// Delete from both databases
	err := DeleteDocumentFromDatabases(context.Background(), docData.HoldingCode, docData.BusinessCode, docData.DocNo, TRANS_FLAG_PURCHASE)
	if err != nil {
		logger.Error("OnConsumeMessagePurchaseDelete: Failed to delete: %v", err)
		return err
	}

	logger.Info("OnConsumeMessagePurchaseDelete: Successfully processed deletion for DocNo=%s", docData.DocNo)
	return nil
}

// ProcessPurchaseDocument - processes purchase document following standardized 10-step pattern
func ProcessPurchaseDocument(msg string) error {
	defer func() {
		if r := recover(); r != nil {
			logger.Info("ProcessPurchaseDocument: Panic recovered: %v", r)
			logger.Info("ProcessPurchaseDocument: Stack trace: %s", debug.Stack())
		}
	}()

	// Step 1: Decode JSON message
	logger.Info("ProcessPurchaseDocument: Step 1 - Decoding JSON message")
	purchaseData := TransPurchaseDecode(msg)

	if purchaseData.HoldingCode == "" || purchaseData.DocNo == "" {
		logger.Error("ProcessPurchaseDocument: Invalid data - HoldingCode='%s', DocNo='%s'", purchaseData.HoldingCode, purchaseData.DocNo)
		return fmt.Errorf("invalid purchase data - missing HoldingCode or DocNo")
	}

	logger.Info("ProcessPurchaseDocument: Starting processing for DocNo=%s", purchaseData.DocNo)
	build.DatabaseChecker(purchaseData.HoldingCode, false)

	// Step 2: Convert Kafka message to ProcessMongoTransModel
	logger.Info("ProcessPurchaseDocument: Step 2 - Converting Kafka message to ProcessMongoTransModel")
	processData := ConvertPurchaseMongoDocToProcessModel(purchaseData)

	// Step 3: Connect to PostgreSQL
	logger.Info("ProcessPurchaseDocument: Step 3 - Connecting to PostgreSQL for holdingCode=%s", purchaseData.HoldingCode)
	db, err := mypg.PgSqlFastConnect(purchaseData.HoldingCode)
	if err != nil {
		logger.Error("ProcessPurchaseDocument: Failed to connect to PostgreSQL: %v", err)
		return fmt.Errorf("failed to connect to PostgreSQL: %v", err)
	}
	// ใช้ connection pool ไม่ต้อง close
	ctx := context.Background()

	// Step 4: Delete existing documents from PostgreSQL
	logger.Info("ProcessPurchaseDocument: Step 4 - Deleting existing documents from PostgreSQL for DocNo=%s", purchaseData.DocNo)
	mypg.DeleteDocPgSql(ctx, db, purchaseData.BusinessCode, purchaseData.DocNo, TRANS_FLAG_PURCHASE)

	// Step 5: Convert process model to build-doc structs
	logger.Info("ProcessPurchaseDocument: Step 5 - Converting process model to build-doc structs")
	docStruct, docPaymentStruct := myglobal.MapDocStructFromMongo(processData, purchaseData.HoldingCode)
	docDetailStructs := MapPurchaseToDocDetailStructs(processData, purchaseData.HoldingCode)
	setDocumentCompany(&docStruct, docDetailStructs, purchaseData.BusinessCode)

	// Step 6: Create document references จาก docDetailStructs ที่มี DocRef
	logger.Info("ProcessPurchaseDocument: Step 6 - Creating document references")
	logger.Info("ProcessPurchaseDocument: Total details count: %d", len(docDetailStructs))
	var docRefStructs []models.DocRefStruct
	docRefMap := make(map[string]bool) // ป้องกัน duplicate
	for i, detail := range docDetailStructs {
		logger.Info("ProcessPurchaseDocument: Detail[%d] DocRef='%s', ItemCode='%s'", i, detail.DocRef, detail.ItemCode)
		if detail.DocRef != "" && !docRefMap[detail.DocRef] {
			docRefMap[detail.DocRef] = true
			// สร้าง docref record: Purchase (12) อ้างอิง Purchase Order (6)
			docRefStructs = append(docRefStructs, models.DocRefStruct{
				DocNo:             processData.DocNo,   // เลขที่เอกสาร Purchase
				DocNoTransFlag:    TRANS_FLAG_PURCHASE, // 12
				DocRefNo:          detail.DocRef,       // เลขที่เอกสาร Purchase Order ที่อ้างอิง
				DocRefNoTransFlag: 6,                   // Purchase Order transflag
			})
			logger.Info("ProcessPurchaseDocument: Added docref %s -> %s", processData.DocNo, detail.DocRef)
		}
	}
	logger.Info("ProcessPurchaseDocument: Total docRefStructs created: %d", len(docRefStructs))

	// Step 7-10: Insert to PostgreSQL and ClickHouse using shared functions
	err = InsertDocumentToPostgreSQL(ctx, db, docStruct, docRefStructs, docPaymentStruct, 7)
	if err != nil {
		return fmt.Errorf("failed to insert document to PostgreSQL: %w", err)
	}

	err = InsertDocDetailToPostgreSQL(ctx, db, purchaseData.HoldingCode, docDetailStructs, 8)
	if err != nil {
		return fmt.Errorf("failed to insert doc details to PostgreSQL: %w", err)
	}

	err = InsertDocumentToClickHouse(ctx, purchaseData.HoldingCode, docStruct, docRefStructs, docPaymentStruct, docDetailStructs, 9)
	if err != nil {
		return fmt.Errorf("failed to insert to ClickHouse: %w", err)
	}

	// Step 10: Calculate stock cost using global config
	err = ProcessDocumentStockCalculation(db, purchaseData.HoldingCode, docDetailStructs, 10)
	if err != nil {
		logger.Error("Failed to calculate stock cost: %v", err)
		// Don't return error, just log it
	}

	// Step 11: Add to document wait process queues
	logger.Info("ProcessPurchaseDocument: Step 11 - Adding to document wait process queues")
	mypg.AddToDocWaitProcessQueues(ctx, db, purchaseData.DocNo, TRANS_FLAG_PURCHASE)

	// Step 11.1: Add purchase-order ที่ถูกอ้างอิงเข้า queue เพื่อ update isref
	for _, docRef := range docRefStructs {
		if docRef.DocRefNo != "" && docRef.DocRefNoTransFlag == 6 {
			logger.Info("ProcessPurchaseDocument: Adding referenced PO to queue: %s", docRef.DocRefNo)
			mypg.AddToDocWaitProcessQueues(ctx, db, docRef.DocRefNo, 6) // 6 = Purchase Order
		}
	}

	// Step 12: Process document status immediately after insert
	logger.Info("ProcessPurchaseDocument: Step 12 - Processing document status for holdingCode=%s", purchaseData.HoldingCode)
	ProcessDocumentStatusAsync(purchaseData.HoldingCode)

	logger.Success("ProcessPurchaseDocument: Successfully processed DocNo=%s", purchaseData.DocNo)
	return nil
}

// MapPurchaseToDocDetailStructs - converts purchase to document detail structs
func MapPurchaseToDocDetailStructs(processData models.ProcessMongoTransModel, holdingCode string) []models.DocDetailStruct {
	var docDetailStructs []models.DocDetailStruct

	for i, detail := range processData.Details {
		// For purchase transactions, use positive value to increase stock
		docDetailStruct := models.DocDetailStruct{
			DocDateTime:     processData.DocDateTime,
			DocNo:           processData.DocNo,
			LineNumber:      i + 1,
			TransFlag:       TRANS_FLAG_PURCHASE,
			CalcFlag:        1, // Purchase = increase stock
			CalcSeq:         1,
			ItemCode:        detail.ItemCode,
			Description:     GetItemName(detail.ItemNames),
			Barcode:         detail.Barcode,
			UnitCode:        detail.UnitCode,
			WhCode:          detail.WhCode,
			LocationCode:    detail.LocationCode,
			TotalQty:        detail.Qty, // Use original Qty from MongoDB
			Price:           detail.Price,
			PriceExcludeVat: detail.PriceExcludeVat,
			UnitStand:       1.0,
			UnitDivide:      1.0,
			DocRef:          detail.DocRef,
			SumAmount:       detail.SumAmount,
		}
		docDetailStructs = append(docDetailStructs, docDetailStruct)
	}

	return docDetailStructs
}

// TransPurchaseDecode - decodes transaction purchase JSON data
func TransPurchaseDecode(jsonData string) models.MongoDocModel {
	var docData models.MongoDocModel
	err := DecodeKafkaMessage(jsonData, &docData, "Purchase")
	if err != nil {
		return models.MongoDocModel{}
	}
	return docData
}

// ConvertPurchaseMongoDocToProcessModel - converts MongoDocModel to ProcessMongoTransModel for purchase
func ConvertPurchaseMongoDocToProcessModel(mongoDoc models.MongoDocModel) models.ProcessMongoTransModel {
	logger.Info("ConvertPurchaseMongoDocToProcessModel: Starting conversion for DocNo=%s", mongoDoc.DocNo)

	// Convert details with nil checks
	var details []models.ProcessMongoTransDetailModel
	if mongoDoc.Details != nil {
		logger.Info("ConvertPurchaseMongoDocToProcessModel: Processing %d details", len(mongoDoc.Details))
		for i, detail := range mongoDoc.Details {
			logger.Info("ConvertPurchaseMongoDocToProcessModel: Processing detail %d: ItemCode=%s", i+1, detail.ItemCode)

			processDetail := models.ProcessMongoTransDetailModel{
				ItemCode:        detail.ItemCode,
				ItemNames:       ConvertLanguageModels(detail.ItemNames),
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
	} else {
		logger.Info("ConvertPurchaseMongoDocToProcessModel: WARNING - mongoDoc.Details is nil")
	}

	// Convert branch names
	logger.Info("ConvertPurchaseMongoDocToProcessModel: Processing branch names")
	branchNames := ConvertLanguageModels(mongoDoc.Branch.Names)

	// Convert branch with nil checks
	logger.Info("ConvertPurchaseMongoDocToProcessModel: Converting branch - Code=%s, GuidFixed=%s", mongoDoc.Branch.Code, mongoDoc.Branch.GuidFixed)
	branch := models.BranchModel{
		Code:      mongoDoc.Branch.Code,
		GuidFixed: mongoDoc.Branch.GuidFixed,
		Names:     branchNames,
	}

	logger.Info("ConvertPurchaseMongoDocToProcessModel: Creating final ProcessMongoTransModel")
	result := models.ProcessMongoTransModel{
		HoldingCode:      mongoDoc.HoldingCode,
		BranchId:         mongoDoc.BranchId,
		GuidFixed:        mongoDoc.GuidFixed,
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
		TransFlag:        TRANS_FLAG_PURCHASE,
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

	logger.Info("ConvertPurchaseMongoDocToProcessModel: Conversion completed successfully")
	return result
}
