package kafka

import (
	"context"
	"fmt"
	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/myglobal"

	"runtime/debug"

	"smlcloudplatform/internal/goapi/models"
	"smlcloudplatform/internal/goapi/mypg"
)

// OnConsumeMessagePurchaseReturnCreateOrUpdate - handles purchase return create/update messages
func OnConsumeMessagePurchaseReturnCreateOrUpdate(msg string) error {
	return SafeConsumerWrapper("PURCHASE_RETURN", func(msg string) error {
		return ProcessPurchaseReturnDocument(msg)
	})(msg)
}

// OnConsumeMessagePurchaseReturnDelete - handles purchase return delete messages
func OnConsumeMessagePurchaseReturnDelete(msg string) error {
	// รับ Message จาก Kafka และแปลงเป็น struct ก่อนประมวลผล
	// Message ควรเป็น JSON string ที่มีข้อมูลของเอกสารที่ต้องการลบ

	logger.Info("OnConsumeMessagePurchaseReturnDelete: %s", msg)

	docData := TransPurchaseReturnDecode(msg)

	if docData.HoldingCode == "" || docData.DocNo == "" {
		logger.Error("Invalid purchase return data - missing HoldingCode or DocNo")
		return fmt.Errorf("invalid purchase return data")
	}

	return DeleteDocumentFromDatabases(context.Background(), docData.HoldingCode, docData.DocNo, TRANS_FLAG_PURCHASE_RETURN)
}

// ProcessPurchaseReturnDocument - processes purchase return using build-doc system
func ProcessPurchaseReturnDocument(msg string) error {
	logger.Info("--- ProcessPurchaseReturnDocument START ---")

	defer func() {
		if r := recover(); r != nil {
			logger.Info("PANIC recovered in ProcessPurchaseReturnDocument: %v", r)
			logger.Info("Stack trace: %s", debug.Stack())
			logger.Info("Message length: %d characters", len(msg))
		}
	}()

	// Decode purchase return from JSON message
	logger.Debug("Step 1: Decoding JSON message...")
	purchaseReturnData := TransPurchaseReturnDecode(msg)
	logger.Info("Step 1 completed: Decoded PurchaseReturn - DocNo=%s, HoldingCode=%s, Amount=%.2f",
		purchaseReturnData.DocNo, purchaseReturnData.HoldingCode, purchaseReturnData.TotalAmount)

	if purchaseReturnData.HoldingCode == "" || purchaseReturnData.DocNo == "" {
		logger.Error("Invalid purchase return data - HoldingCode='%s', DocNo='%s'", purchaseReturnData.HoldingCode, purchaseReturnData.DocNo)
		return fmt.Errorf("invalid purchase return data - missing HoldingCode or DocNo")
	}

	// Convert MongoDocModel to ProcessMongoTransModel
	logger.Debug("Step 2: Converting to process model...")
	processData := ConvertPurchaseReturnMongoDocToProcessModel(purchaseReturnData)
	logger.Info("Step 2 completed: ProcessData - DocNo=%s, TransFlag=%d, Details=%d",
		processData.DocNo, processData.TransFlag, len(processData.Details))

	// Connect to database
	logger.Debug("Step 3: Connecting to PostgreSQL...")
	db, err := mypg.PgSqlFastConnect(purchaseReturnData.HoldingCode)
	if err != nil {
		logger.Error("Failed to connect to database: %v", err)
		return fmt.Errorf("failed to connect to database: %v", err)
	}
	logger.Debug("Step 3 completed: Database connected successfully")

	ctx := context.Background()

	// Delete existing documents (upsert behavior)
	logger.Debug("Step 4: Deleting existing documents...")
	mypg.DeleteDocPgSql(ctx, db, purchaseReturnData.DocNo, TRANS_FLAG_PURCHASE_RETURN)
	logger.Debug("Step 4 completed: Existing documents deleted")

	// Convert to build-doc structs
	logger.Debug("Step 5: Converting to build-doc structs...")
	docStruct, docPaymentStruct := myglobal.MapDocStructFromMongo(processData, purchaseReturnData.HoldingCode)
	docDetailStructs := MapPurchaseReturnToDocDetailStructs(processData, purchaseReturnData.HoldingCode)
	logger.Debug("Step 5 completed: DocStruct created with %d details", len(docDetailStructs))

	// Create doc references if any
	var docRefStructs []models.DocRefStruct
	for _, refNo := range processData.DocReferences {
		docRefStructs = append(docRefStructs,
			myglobal.MapDocRefStruct(processData.DocNo, processData.TransFlag, refNo))
	}
	logger.Debug("Step 6: Created %d document references", len(docRefStructs))

	// Step 7-9: Insert to PostgreSQL and ClickHouse using shared functions
	err = InsertDocumentToPostgreSQL(ctx, db, docStruct, docRefStructs, docPaymentStruct, 7)
	if err != nil {
		return fmt.Errorf("failed to insert document to PostgreSQL: %w", err)
	}

	err = InsertDocDetailToPostgreSQL(ctx, db, purchaseReturnData.HoldingCode, docDetailStructs, 8)
	if err != nil {
		return fmt.Errorf("failed to insert doc details to PostgreSQL: %w", err)
	}

	err = InsertDocumentToClickHouse(ctx, purchaseReturnData.HoldingCode, docStruct, docRefStructs, docPaymentStruct, docDetailStructs, 9)
	if err != nil {
		return fmt.Errorf("failed to insert to ClickHouse: %w", err)
	}

	// Step 10: Calculate stock cost using global config
	err = ProcessDocumentStockCalculation(db, purchaseReturnData.HoldingCode, docDetailStructs, 10)
	if err != nil {
		logger.Error("Failed to calculate stock cost: %v", err)
		// Don't return error, just log it
	}

	// Step 11: Add to processing queues
	logger.Debug("Step 11: Adding to processing queues...")
	mypg.AddToDocWaitProcessQueues(ctx, db, purchaseReturnData.DocNo, TRANS_FLAG_PURCHASE_RETURN)
	logger.Debug("Step 11 completed: Added to processing queues")

	logger.Info("--- ProcessPurchaseReturnDocument COMPLETED SUCCESSFULLY: %s ---", purchaseReturnData.DocNo)
	return nil
}

// MapPurchaseReturnToDocDetailStructs - converts purchase return to document detail structs
func MapPurchaseReturnToDocDetailStructs(processData models.ProcessMongoTransModel, holdingCode string) []models.DocDetailStruct {
	var docDetailStructs []models.DocDetailStruct

	for i, detail := range processData.Details {
		// For purchase return transactions, use negative value to decrease stock
		docDetailStruct := models.DocDetailStruct{
			DocDateTime:     processData.DocDateTime,
			DocNo:           processData.DocNo,
			LineNumber:      i + 1,
			TransFlag:       TRANS_FLAG_PURCHASE_RETURN,
			CalcFlag:        -1, // Purchase Return = decrease stock
			CalcSeq:         1,
			ItemCode:        detail.ItemCode,
			Description:     GetItemName(detail.ItemNames),
			BarcodeMain:     detail.Barcode,
			Barcode:         detail.Barcode,
			UnitCode:        detail.UnitCode,
			WhCode:          detail.WhCode,
			LocationCode:    detail.LocationCode,
			TotalQty:        detail.Qty,
			Price:           detail.Price,
			PriceExcludeVat: detail.PriceExcludeVat,
			UnitStand:       1.0,
			UnitDivide:      1.0,
			DocRef:          detail.DocRef,
			SumAmount:       detail.SumAmount,
		}
		docDetailStructs = append(docDetailStructs, docDetailStruct)
	}

	logger.Info("MapPurchaseReturnToDocDetailStructs: Converted %d details", len(docDetailStructs))
	return docDetailStructs
}

// TransPurchaseReturnDecode - decodes transaction purchase return JSON data
func TransPurchaseReturnDecode(jsonData string) models.MongoDocModel {
	var docData models.MongoDocModel
	DecodeKafkaMessage(jsonData, &docData, "PurchaseReturn")
	return docData
}

// ConvertPurchaseReturnMongoDocToProcessModel - converts MongoDocModel to ProcessMongoTransModel for purchase return
func ConvertPurchaseReturnMongoDocToProcessModel(mongoDoc models.MongoDocModel) models.ProcessMongoTransModel {
	logger.Info("ConvertPurchaseReturnMongoDocToProcessModel: Starting conversion for DocNo: %s", mongoDoc.DocNo)
	// Convert details with nil checks
	var details []models.ProcessMongoTransDetailModel
	if mongoDoc.Details != nil {
		logger.Info("ConvertPurchaseReturnMongoDocToProcessModel: Processing %d details", len(mongoDoc.Details))
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
	} else {
		logger.Info("ConvertPurchaseReturnMongoDocToProcessModel: WARNING - mongoDoc.Details is nil")
	}

	branchNames := ConvertLanguageModels(mongoDoc.Branch.Names)

	// Convert branch with nil checks
	logger.Info("ConvertPurchaseReturnMongoDocToProcessModel: Converting branch - Code=%s, GuidFixed=%s", mongoDoc.Branch.Code, mongoDoc.Branch.GuidFixed)
	branch := models.BranchModel{
		Code:      mongoDoc.Branch.Code,
		GuidFixed: mongoDoc.Branch.GuidFixed,
		Names:     branchNames,
	}

	logger.Info("ConvertPurchaseReturnMongoDocToProcessModel: Creating final ProcessMongoTransModel")
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
		TransFlag:        TRANS_FLAG_PURCHASE_RETURN,
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

	logger.Info("ConvertPurchaseReturnMongoDocToProcessModel: Conversion completed successfully")
	return result
}
