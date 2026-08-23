package kafka

import (
	"context"
	"fmt"
	"runtime/debug"
	"smlcloudplatform/internal/goapi/logger"

	"smlcloudplatform/internal/goapi/models"
	"smlcloudplatform/internal/goapi/myglobal"
	"smlcloudplatform/internal/goapi/mypg"
)

// OnConsumeMessageSaleInvoiceReturnCreateOrUpdate - handles sale invoice return create/update messages
func OnConsumeMessageSaleInvoiceReturnCreateOrUpdate(msg string) error {
	return SafeConsumerWrapper("SALE_INVOICE_RETURN", func(msg string) error {
		return ProcessSaleInvoiceReturnDocument(msg)
	})(msg)
}

// OnConsumeMessageSaleInvoiceReturnDelete - handles sale invoice return delete messages
func OnConsumeMessageSaleInvoiceReturnDelete(msg string) error {
	// รับ Message จาก Kafka ที่เป็นการลบเอกสาร Sale Invoice Return
	// msg จะเป็น JSON string ที่มีข้อมูลของเอกสารที่ต้องการลบ เช่น {"holdingcode": "shop123", "docno": "SR2024001"}

	logger.Info("OnConsumeMessageSaleInvoiceReturnDelete: %s", msg)

	docData := TransSaleInvoiceReturnDecode(msg)

	if docData.HoldingCode == "" || docData.DocNo == "" {
		logger.Error("Invalid sale invoice return data - missing HoldingCode or DocNo")
		return fmt.Errorf("invalid sale invoice return data")
	}

	return DeleteDocumentFromDatabases(context.Background(), docData.HoldingCode, docData.BusinessCode, docData.DocNo, TRANS_FLAG_SALE_INVOICE_RETURN)
}

// ProcessSaleInvoiceReturnDocument - processes sale invoice return using build-doc system
func ProcessSaleInvoiceReturnDocument(msg string) error {
	// ประมวลผลเอกสาร Sale Invoice Return (ใบรับคืนสินค้า)
	logger.Info("--- ProcessSaleInvoiceReturnDocument START ---")

	defer func() {
		if r := recover(); r != nil {
			logger.Info("PANIC recovered in ProcessSaleInvoiceReturnDocument: %v", r)
			logger.Info("Stack trace: %s", debug.Stack())
			logger.Info("Message length: %d characters", len(msg))
		}
	}()

	// Decode sale invoice return from JSON message
	logger.Debug("Step 1: Decoding JSON message... ")
	saleInvoiceReturnData := TransSaleInvoiceReturnDecode(msg)
	logger.Info("Step 1: Sale Invoice Return decoded successfully - HoldingCode=%s, DocNo=%s, TotalAmount=%.2f, Details=%d",
		saleInvoiceReturnData.HoldingCode, saleInvoiceReturnData.DocNo, saleInvoiceReturnData.TotalAmount, len(saleInvoiceReturnData.Details))

	if saleInvoiceReturnData.HoldingCode == "" || saleInvoiceReturnData.DocNo == "" {
		logger.Error("Invalid sale invoice return data - HoldingCode='%s', DocNo='%s'", saleInvoiceReturnData.HoldingCode, saleInvoiceReturnData.DocNo)
		return fmt.Errorf("invalid sale invoice return data - missing HoldingCode or DocNo")
	}

	// Convert MongoDocModel to ProcessMongoTransModel
	logger.Debug("Step 2: Converting to process model...")
	processData := ConvertSaleInvoiceReturnMongoDocToProcessModel(saleInvoiceReturnData)
	logger.Info("Step 2: Conversion completed - Process Model DocNo=%s, Details=%d",
		processData.DocNo, len(processData.Details))

	// Connect to database
	logger.Debug("Step 3: Connecting to PostgreSQL...")
	db, err := mypg.PgSqlFastConnect(saleInvoiceReturnData.HoldingCode)
	if err != nil {
		logger.Error("Failed to connect to database: %v", err)
		return fmt.Errorf("failed to connect to database: %v", err)
	}
	logger.Debug("Step 3 completed: Connected to PostgreSQL")

	ctx := context.Background()

	// Delete existing documents (upsert behavior)
	logger.Debug("Step 4: Deleting existing documents...")
	mypg.DeleteDocPgSql(ctx, db, saleInvoiceReturnData.BusinessCode, saleInvoiceReturnData.DocNo, TRANS_FLAG_SALE_INVOICE_RETURN)
	logger.Debug("Step 4 completed: Existing documents deleted")

	// Convert to build-doc structs
	logger.Debug("Step 5: Converting to build-doc structs...")
	docStruct, docPaymentStruct := myglobal.MapDocStructFromMongo(processData, saleInvoiceReturnData.HoldingCode)
	docDetailStructs := MapSaleInvoiceReturnToDocDetailStructs(processData, saleInvoiceReturnData.HoldingCode)
	setDocumentCompany(&docStruct, docDetailStructs, saleInvoiceReturnData.BusinessCode)
	logger.Debug("Step 5 completed: Converted to %d doc details", len(docDetailStructs))

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

	err = InsertDocDetailToPostgreSQL(ctx, db, saleInvoiceReturnData.HoldingCode, docDetailStructs, 8)
	if err != nil {
		return fmt.Errorf("failed to insert doc details to PostgreSQL: %w", err)
	}

	err = InsertDocumentToClickHouse(ctx, saleInvoiceReturnData.HoldingCode, docStruct, docRefStructs, docPaymentStruct, docDetailStructs, 9)
	if err != nil {
		return fmt.Errorf("failed to insert to ClickHouse: %w", err)
	}

	// Step 10: Calculate stock cost using global config
	err = ProcessDocumentStockCalculation(db, saleInvoiceReturnData.HoldingCode, docDetailStructs, 10)
	if err != nil {
		logger.Error("Failed to calculate stock cost: %v", err)
		// Don't return error, just log it
	}

	// Step 11: Add to processing queues
	logger.Debug("Step 11: Adding to processing queues...")
	mypg.AddToDocWaitProcessQueues(ctx, db, saleInvoiceReturnData.DocNo, TRANS_FLAG_SALE_INVOICE_RETURN)
	logger.Debug("Step 11 completed: Added to processing queues")

	// Step 12: Process document status immediately after insert
	logger.Debug("Step 12: Processing document status for holdingCode=%s", saleInvoiceReturnData.HoldingCode)
	ProcessDocumentStatusAsync(saleInvoiceReturnData.HoldingCode)

	logger.Info("--- ProcessSaleInvoiceReturnDocument COMPLETED SUCCESSFULLY: %s ---", saleInvoiceReturnData.DocNo)
	return nil
}

// ConvertSaleInvoiceReturnMongoDocToProcessModel - converts MongoDocModel to ProcessMongoTransModel for Sale Invoice Returns
func ConvertSaleInvoiceReturnMongoDocToProcessModel(mongoDoc models.MongoDocModel) models.ProcessMongoTransModel {
	logger.Info("ConvertSaleInvoiceReturnMongoDocToProcessModel: Starting conversion for DocNo=%s", mongoDoc.DocNo)

	// Convert details with nil checks
	var details []models.ProcessMongoTransDetailModel
	if mongoDoc.Details != nil {
		logger.Info("ConvertSaleInvoiceReturnMongoDocToProcessModel: Processing %d details", len(mongoDoc.Details))
		for i, detail := range mongoDoc.Details {
			logger.Info("ConvertSaleInvoiceReturnMongoDocToProcessModel: Processing detail %d: ItemCode=%s", i+1, detail.ItemCode)

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
		logger.Info("ConvertSaleInvoiceReturnMongoDocToProcessModel: WARNING - mongoDoc.Details is nil")
	}

	branchNames := ConvertLanguageModels(mongoDoc.Branch.Names)

	// Convert branch with nil checks
	logger.Info("ConvertSaleInvoiceReturnMongoDocToProcessModel: Converting branch - Code=%s, GuidFixed=%s", mongoDoc.Branch.Code, mongoDoc.Branch.GuidFixed)
	branch := models.BranchModel{
		Code:      mongoDoc.Branch.Code,
		GuidFixed: mongoDoc.Branch.GuidFixed,
		Names:     branchNames,
	}

	logger.Info("ConvertSaleInvoiceReturnMongoDocToProcessModel: Creating final ProcessMongoTransModel")
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
		TransFlag:        TRANS_FLAG_SALE_INVOICE_RETURN,
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

	logger.Info("ConvertSaleInvoiceReturnMongoDocToProcessModel: Conversion completed successfully")
	return result
}

// MapSaleInvoiceReturnToDocDetailStructs - converts sale invoice return to document detail structs
func MapSaleInvoiceReturnToDocDetailStructs(processData models.ProcessMongoTransModel, holdingCode string) []models.DocDetailStruct {
	var docDetailStructs []models.DocDetailStruct

	for i, detail := range processData.Details {
		// For return transactions, use positive quantity to increase stock
		docDetailStruct := models.DocDetailStruct{
			DocDateTime:     processData.DocDateTime,
			DocNo:           processData.DocNo,
			LineNumber:      i + 1,
			TransFlag:       TRANS_FLAG_SALE_INVOICE_RETURN,
			CalcFlag:        1, // Return = increase stock (positive)
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

// TransSaleInvoiceReturnDecode - decodes transaction sale invoice return JSON data
func TransSaleInvoiceReturnDecode(jsonData string) models.MongoDocModel {
	var docData models.MongoDocModel
	DecodeKafkaMessage(jsonData, &docData, "SaleInvoiceReturn")
	return docData
}
