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

// OnConsumeMessageStockTransferCreateOrUpdate - handles stock transfer create/update messages
func OnConsumeMessageStockTransferCreateOrUpdate(msg string) error {
	return SafeConsumerWrapper("STOCK_TRANSFER", func(msg string) error {
		return ProcessStockTransferDocument(msg)
	})(msg)
}

// OnConsumeMessageStockTransferDelete - handles stock transfer delete messages
func OnConsumeMessageStockTransferDelete(msg string) error {
	logger.Info("OnConsumeMessageStockTransferDelete: Processing deletion message")

	docData := TransStockTransferDecode(msg)
	if docData.HoldingCode == "" || docData.DocNo == "" {
		logger.Error("OnConsumeMessageStockTransferDelete: Invalid data - missing HoldingCode or DocNo")
		return fmt.Errorf("invalid stock transfer data")
	}

	return DeleteDocumentFromDatabases(context.Background(), docData.HoldingCode, docData.DocNo, TRANS_FLAG_STOCK_TRANSFER)
}

// ProcessStockTransferDocument - processes stock transfer document following standardized 10-step pattern
func ProcessStockTransferDocument(msg string) error {
	defer func() {
		if r := recover(); r != nil {
			logger.Info("ProcessStockTransferDocument: Panic recovered: %v", r)
			logger.Info("ProcessStockTransferDocument: Stack trace: %s", debug.Stack())
		}
	}()

	// Step 1: Decode JSON message
	logger.Info("ProcessStockTransferDocument: Step 1 - Decoding JSON message")
	docData := TransStockTransferDecode(msg)

	if docData.HoldingCode == "" || docData.DocNo == "" {
		logger.Error("ProcessStockTransferDocument: Invalid data - HoldingCode='%s', DocNo='%s'", docData.HoldingCode, docData.DocNo)
		return fmt.Errorf("invalid stock transfer data - missing HoldingCode or DocNo")
	}

	logger.Info("ProcessStockTransferDocument: Starting processing for DocNo=%s", docData.DocNo)
	build.DatabaseChecker(docData.HoldingCode, false)

	// Step 2: Convert Kafka message to ProcessMongoTransModel
	logger.Info("ProcessStockTransferDocument: Step 2 - Converting Kafka message to ProcessMongoTransModel")
	processData := ConvertStockTransferMongoDocToProcessModel(docData)

	// Step 3: Connect to PostgreSQL
	logger.Info("ProcessStockTransferDocument: Step 3 - Connecting to PostgreSQL for holdingCode=%s", docData.HoldingCode)
	db, err := mypg.PgSqlFastConnect(docData.HoldingCode)
	if err != nil {
		logger.Error("ProcessStockTransferDocument: Failed to connect to PostgreSQL: %v", err)
		return fmt.Errorf("failed to connect to PostgreSQL: %v", err)
	}
	// ใช้ connection pool ไม่ต้อง close
	ctx := context.Background()

	// Step 4: Delete existing documents from PostgreSQL
	logger.Info("ProcessStockTransferDocument: Step 4 - Deleting existing documents from PostgreSQL for DocNo=%s", docData.DocNo)
	mypg.DeleteDocPgSql(ctx, db, docData.DocNo, TRANS_FLAG_STOCK_TRANSFER)

	// Step 5: Convert process model to build-doc structs
	logger.Info("ProcessStockTransferDocument: Step 5 - Converting process model to build-doc structs")
	docStruct, docPaymentStruct := myglobal.MapDocStructFromMongo(processData, docData.HoldingCode)
	docDetailStructs := MapStockTransferToDocDetailStructs(processData, docData.HoldingCode)

	// Step 6: Create document references
	logger.Info("ProcessStockTransferDocument: Step 6 - Creating document references")
	var docRefStructs []models.DocRefStruct
	for _, refNo := range processData.DocReferences {
		docRefStructs = append(docRefStructs,
			myglobal.MapDocRefStruct(processData.DocNo, processData.TransFlag, refNo))
	}

	// Step 7-10: Insert to PostgreSQL and ClickHouse using shared functions
	err = InsertDocumentToPostgreSQL(ctx, db, docStruct, docRefStructs, docPaymentStruct, 7)
	if err != nil {
		return fmt.Errorf("failed to insert document to PostgreSQL: %w", err)
	}

	err = InsertDocDetailToPostgreSQL(ctx, db, docData.HoldingCode, docDetailStructs, 8)
	if err != nil {
		return fmt.Errorf("failed to insert doc details to PostgreSQL: %w", err)
	}

	err = InsertDocumentToClickHouse(ctx, docData.HoldingCode, docStruct, docRefStructs, docPaymentStruct, docDetailStructs, 9)
	if err != nil {
		return fmt.Errorf("failed to insert to ClickHouse: %w", err)
	}

	// Step 10: Calculate stock cost using global config
	err = ProcessDocumentStockCalculation(db, docData.HoldingCode, docDetailStructs, 10)
	if err != nil {
		logger.Error("Failed to calculate stock cost: %v", err)
		// Don't return error, just log it
	}

	// Step 11: Add to document wait process queues
	logger.Info("ProcessStockTransferDocument: Step 11 - Adding to document wait process queues")
	mypg.AddToDocWaitProcessQueues(ctx, db, docData.DocNo, TRANS_FLAG_STOCK_TRANSFER)

	logger.Success("ProcessStockTransferDocument: Successfully processed DocNo=%s", docData.DocNo)
	return nil
}

// TransStockTransferDecode - decodes stock transfer message
func TransStockTransferDecode(msg string) models.StockTransferStruct {
	var result models.StockTransferStruct
	DecodeKafkaMessage(msg, &result, "StockTransfer")
	return result
}

// ConvertStockTransferMongoDocToProcessModel - converts StockTransferStruct to ProcessMongoTransModel
func ConvertStockTransferMongoDocToProcessModel(stockTransfer models.StockTransferStruct) models.ProcessMongoTransModel {
	logger.Info("ConvertStockTransferMongoDocToProcessModel: Starting conversion for DocNo=%s", stockTransfer.DocNo)

	// Convert details with nil checks
	var details []models.ProcessMongoTransDetailModel
	if stockTransfer.Details != nil {
		logger.Info("ConvertStockTransferMongoDocToProcessModel: Processing %d details", len(stockTransfer.Details))
		for i, detail := range stockTransfer.Details {
			logger.Info("ConvertStockTransferMongoDocToProcessModel: Processing detail %d: ItemCode=%s", i+1, detail.ItemCode)

			// For stock transfer, we don't have ItemNames in the model, so create empty
			var itemNames []models.LanguageModel

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
		logger.Info("ConvertStockTransferMongoDocToProcessModel: WARNING - stockTransfer.Details is nil")
	}

	// Convert branch names from models.LanguageModel to models.LanguageModel with nil checks
	var branchNames []models.LanguageModel
	if stockTransfer.Branch.Names != nil {
		logger.Info("ConvertStockTransferMongoDocToProcessModel: Processing %d branch names", len(stockTransfer.Branch.Names))
		for _, name := range stockTransfer.Branch.Names {
			branchNames = append(branchNames, models.LanguageModel{
				Code:     name.Code,
				Name:     name.Name,
				IsAuto:   name.IsAuto,
				IsDelete: name.IsDelete,
			})
		}
	} else {
		logger.Info("ConvertStockTransferMongoDocToProcessModel: WARNING - stockTransfer.Branch.Names is nil")
	}

	// Convert branch with nil checks
	logger.Info("ConvertStockTransferMongoDocToProcessModel: Converting branch - Code=%s, GuidFixed=%s", stockTransfer.Branch.Code, stockTransfer.Branch.GuidFixed)
	branch := models.BranchModel{
		Code:      stockTransfer.Branch.Code,
		GuidFixed: stockTransfer.Branch.GuidFixed,
		Names:     branchNames,
	}

	logger.Info("ConvertStockTransferMongoDocToProcessModel: Creating final ProcessMongoTransModel")
	result := models.ProcessMongoTransModel{
		HoldingCode:      stockTransfer.HoldingCode,
		BranchId:         stockTransfer.BranchCode,
		GuidFixed:        stockTransfer.Guid,
		DocNo:            stockTransfer.DocNo,
		DocDateTime:      stockTransfer.DocDateTime,
		TotalAmount:      stockTransfer.TotalAmount,
		RoundAmount:      0.0,  // Stock transfers usually don't have rounding
		PayCashAmount:    0.0,  // Stock transfers don't have payment
		PayCashChange:    0.0,  // Stock transfers don't have payment
		PaymentDetailRaw: "[]", // Empty payment details
		SlipUrl:          "",   // Stock transfers don't have slips
		SaleChannelCode:  "",   // Not applicable for stock transfers
		DeliveryAmount:   0.0,  // Stock transfers don't have delivery amount
		IsCancel:         stockTransfer.IsCancel,
		CancelReason:     "",
		GuidPos:          "",
		Branch:           branch,
		TransFlag:        TRANS_FLAG_STOCK_TRANSFER,
		Details:          details,
		DocReferences:    []models.ProcessMongoDocReferenceModel{},
	}

	logger.Info("ConvertStockTransferMongoDocToProcessModel: Conversion completed successfully")
	return result
}

// MapStockTransferToDocDetailStructs - converts stock transfer to document detail structs
func MapStockTransferToDocDetailStructs(processData models.ProcessMongoTransModel, holdingCode string) []models.DocDetailStruct {
	var docDetailStructs []models.DocDetailStruct

	for i, detail := range processData.Details {
		// For stock transfer, we need to create two entries:
		// 1. Negative entry for source warehouse (reduce stock)
		// 2. Positive entry for destination warehouse (increase stock)

		// Note: Stock transfer logic will be handled by the existing detail processing
		// For now, create a single entry and let the stock calculation handle the transfer
		docDetailStruct := models.DocDetailStruct{
			DocDateTime:     processData.DocDateTime,
			DocNo:           processData.DocNo,
			LineNumber:      i + 1,
			TransFlag:       TRANS_FLAG_STOCK_TRANSFER,
			CalcFlag:        1, // Will be processed by stock calculation
			CalcSeq:         1,
			ItemCode:        detail.ItemCode,
			Description:     detail.ItemCode, // Use ItemCode as description if no name available
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

	return docDetailStructs
}
