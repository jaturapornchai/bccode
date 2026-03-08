package kafka

import (
	"context"
	"fmt"
	"smlcloudplatform/internal/goapi/logger"
	"runtime/debug"

	"smlcloudplatform/internal/goapi/models"
	"smlcloudplatform/internal/goapi/myglobal"
	"smlcloudplatform/internal/goapi/mypg"
	"smlcloudplatform/internal/goapi/process/build"
)

// OnConsumeMessagePurchasePartialCreateOrUpdate - handles purchase partial create/update messages
func OnConsumeMessagePurchasePartialCreateOrUpdate(msg string) error {
	return SafeConsumerWrapper("PURCHASE_PARTIAL", func(msg string) error {
		return ProcessPurchasePartialDocument(msg)
	})(msg)
}

// OnConsumeMessagePurchasePartialDelete - handles purchase partial delete messages
func OnConsumeMessagePurchasePartialDelete(msg string) error {
	// รับข้อความจาก Kafka และลบข้อมูลใน PostgreSQL
	logger.Info("OnConsumeMessagePurchasePartialDelete: %s", msg)

	docData := TransPurchasePartialDecode(msg)
	build.DatabaseChecker(docData.ShopId, false)

	if docData.ShopId == "" || docData.DocNo == "" {
		logger.Error("Invalid purchase partial data - missing ShopId or DocNo")
		return fmt.Errorf("invalid purchase partial data")
	}

	return DeleteDocumentFromDatabases(context.Background(), docData.ShopId, docData.DocNo, TRANS_FLAG_PURCHASE_PARTIAL)
}

// ProcessPurchasePartialDocument - processes purchase partial using build-doc system
func ProcessPurchasePartialDocument(msg string) error {
	logger.Info("--- ProcessPurchasePartialDocument START ---")

	defer func() {
		if r := recover(); r != nil {
			logger.Info("PANIC recovered in ProcessPurchasePartialDocument: %v", r)
			logger.Info("Stack trace: %s", debug.Stack())
			logger.Info("Message length: %d characters", len(msg))
		}
	}()

	// Decode purchase partial from JSON message
	logger.Debug("Step 1: Decoding JSON message...")
	purchasePartialData := TransPurchasePartialDecode(msg)
	logger.Info("Step 1: Purchase Partial decoded successfully - ShopID=%s, DocNo=%s, TotalAmount=%.2f, Details=%d",
		purchasePartialData.ShopId, purchasePartialData.DocNo, purchasePartialData.TotalAmount, len(purchasePartialData.Details))

	if purchasePartialData.ShopId == "" || purchasePartialData.DocNo == "" {
		logger.Error("Invalid purchase partial data - ShopId='%s', DocNo='%s'", purchasePartialData.ShopId, purchasePartialData.DocNo)
		return fmt.Errorf("invalid purchase partial data - missing ShopId or DocNo")
	}

	// Convert MongoDocModel to ProcessMongoTransModel
	logger.Debug("Step 2: Converting to process model...")
	processData := ConvertPurchasePartialMongoDocToProcessModel(purchasePartialData)
	logger.Info("Step 2 completed: ProcessData - DocNo=%s, TransFlag=%d, Details=%d",
		processData.DocNo, processData.TransFlag, len(processData.Details))

	// Connect to database
	logger.Debug("Step 3: Connecting to PostgreSQL...")
	db, err := mypg.PgSqlFastConnect(purchasePartialData.ShopId)
	if err != nil {
		logger.Error("Failed to connect to database: %v", err)
		return fmt.Errorf("failed to connect to database: %v", err)
	}
	logger.Debug("Step 3 completed: Database connected successfully")

	ctx := context.Background()

	// Delete existing documents (upsert behavior)
	logger.Debug("Step 4: Deleting existing documents...")
	mypg.DeleteDocPgSql(ctx, db, purchasePartialData.DocNo, TRANS_FLAG_PURCHASE_PARTIAL)
	logger.Debug("Step 4 completed: Existing documents deleted")

	// Convert to build-doc structs
	logger.Debug("Step 5: Converting to build-doc structs...")
	docStruct, docPaymentStruct := myglobal.MapDocStructFromMongo(processData, purchasePartialData.ShopId)
	docDetailStructs := MapPurchasePartialToDocDetailStructs(processData, purchasePartialData.ShopId)
	logger.Debug("Step 5 completed: DocStruct created with %d details", len(docDetailStructs))

	// Create doc references if any
	logger.Debug("Step 6: Creating document references...")
	var docRefStructs []models.DocRefStruct
	for _, refNo := range processData.DocReferences {
		docRefStructs = append(docRefStructs,
			myglobal.MapDocRefStruct(processData.DocNo, processData.TransFlag, refNo))
	}
	logger.Debug("Step 6 completed: Created %d document references", len(docRefStructs))

	// Step 7-9: Insert to PostgreSQL and ClickHouse using shared functions
	err = InsertDocumentToPostgreSQL(ctx, db, docStruct, docRefStructs, docPaymentStruct, 7)
	if err != nil {
		return fmt.Errorf("failed to insert document to PostgreSQL: %w", err)
	}

	err = InsertDocDetailToPostgreSQL(ctx, db, purchasePartialData.ShopId, docDetailStructs, 8)
	if err != nil {
		return fmt.Errorf("failed to insert doc details to PostgreSQL: %w", err)
	}

	err = InsertDocumentToClickHouse(ctx, purchasePartialData.ShopId, docStruct, docRefStructs, docPaymentStruct, docDetailStructs, 9)
	if err != nil {
		return fmt.Errorf("failed to insert to ClickHouse: %w", err)
	}

	// Step 10: Calculate stock cost using global config
	err = ProcessDocumentStockCalculation(db, purchasePartialData.ShopId, docDetailStructs, 10)
	if err != nil {
		logger.Error("Failed to calculate stock cost: %v", err)
		// Don't return error, just log it
	}

	// Step 11: Add to processing queues
	logger.Debug("Step 11: Adding to processing queues...")
	mypg.AddToDocWaitProcessQueues(ctx, db, purchasePartialData.DocNo, TRANS_FLAG_PURCHASE_PARTIAL)
	logger.Debug("Step 11 completed: Added to processing queues")

	logger.Info("--- ProcessPurchasePartialDocument COMPLETED SUCCESSFULLY: %s ---", purchasePartialData.DocNo)
	return nil
}

// MapPurchasePartialToDocDetailStructs - converts purchase partial to document detail structs
func MapPurchasePartialToDocDetailStructs(processData models.ProcessMongoTransModel, shopId string) []models.DocDetailStruct {
	var docDetailStructs []models.DocDetailStruct

	for i, detail := range processData.Details {
		// For purchase partial transactions, use positive value to increase stock
		docDetailStruct := models.DocDetailStruct{
			DocDateTime:     processData.DocDateTime,
			DocNo:           processData.DocNo,
			LineNumber:      i + 1,
			TransFlag:       TRANS_FLAG_PURCHASE_PARTIAL,
			CalcFlag:        1, // Purchase Partial = increase stock
			CalcSeq:         1,
			ItemCode:        detail.ItemCode,
			Description:     GetItemName(detail.ItemNames),
			BarcodeMain:     detail.Barcode,
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

func TransPurchasePartialDecode(jsonData string) models.MongoDocModel {
	var docData models.MongoDocModel
	DecodeKafkaMessage(jsonData, &docData, "PurchasePartial")
	return docData
}

// ConvertPurchasePartialMongoDocToProcessModel - converts MongoDocModel to ProcessMongoTransModel for purchase partial
func ConvertPurchasePartialMongoDocToProcessModel(mongoDoc models.MongoDocModel) models.ProcessMongoTransModel {
	logger.Info("ConvertPurchasePartialMongoDocToProcessModel: Starting conversion for DocNo=%s", mongoDoc.DocNo)

	// Convert details with nil checks
	var details []models.ProcessMongoTransDetailModel
	if mongoDoc.Details != nil {
		logger.Info("ConvertPurchasePartialMongoDocToProcessModel: Processing %d details", len(mongoDoc.Details))
		for i, detail := range mongoDoc.Details {
			logger.Info("ConvertPurchasePartialMongoDocToProcessModel: Processing detail %d: ItemCode=%s", i+1, detail.ItemCode)

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
	} else {
		logger.Info("ConvertPurchasePartialMongoDocToProcessModel: WARNING - mongoDoc.Details is nil")
	}

	branchNames := ConvertLanguageModels(mongoDoc.Branch.Names)

	// Convert branch with nil checks
	logger.Info("ConvertPurchasePartialMongoDocToProcessModel: Converting branch - Code=%s, GuidFixed=%s", mongoDoc.Branch.Code, mongoDoc.Branch.GuidFixed)
	branch := models.BranchModel{
		Code:      mongoDoc.Branch.Code,
		GuidFixed: mongoDoc.Branch.GuidFixed,
		Names:     branchNames,
	}

	logger.Info("ConvertPurchasePartialMongoDocToProcessModel: Creating final ProcessMongoTransModel")
	result := models.ProcessMongoTransModel{
		ShopId:           mongoDoc.ShopId,
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
		TransFlag:        TRANS_FLAG_PURCHASE_PARTIAL,
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

	logger.Info("ConvertPurchasePartialMongoDocToProcessModel: Conversion completed successfully")
	return result
}

// preparePurchasePartialDocDetailRow - prepares a row for bulk insert
func preparePurchasePartialDocDetailRow(detail models.DocDetailStruct) []any {
	// Set default values for missing fields
	whCode := detail.WhCode
	if whCode == "" {
		whCode = "X"
	}
	locationCode := detail.LocationCode
	if locationCode == "" {
		locationCode = "X"
	}
	description := detail.Description
	if description == "" {
		description = detail.ItemCode
	}
	barcode := detail.Barcode
	if barcode == "" {
		barcode = detail.ItemCode
	}
	barcodeMain := detail.BarcodeMain
	if barcodeMain == "" {
		barcodeMain = barcode
	}

	return []any{
		detail.DocDateTime,
		detail.DocNo,
		detail.LineNumber,
		detail.TransFlag,
		int(detail.CalcFlag),
		detail.CalcSeq,
		detail.ItemCode,
		description,
		barcodeMain,
		barcode,
		detail.UnitCode,
		whCode,
		locationCode,
		detail.TotalQty,
		detail.Price,
		detail.PriceExcludeVat,
		detail.UnitStand,
		detail.UnitDivide,
		detail.DocRef,
		detail.SumAmount,
	}
}
