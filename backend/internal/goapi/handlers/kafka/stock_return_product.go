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

// StockReturnProductCreated - handles stock return product creation events
func StockReturnProductCreated(message string, offset int64, partition int32) {
	logger.Info("StockReturnProductCreated: Processing message at offset=%d, partition=%d", offset, partition)

	docData := TransStockReturnProductDecode(message)
	if docData.DocNo == "" {
		logger.Info("StockReturnProductCreated: Invalid or empty document data, skipping")
		return
	}

	err := ProcessStockReturnProductDocument(docData)
	if err != nil {
		logger.Info("StockReturnProductCreated: Error processing DocNo=%s: %v", docData.DocNo, err)
		return
	}

	logger.Info("StockReturnProductCreated: Successfully processed DocNo=%s", docData.DocNo)
}

// StockReturnProductUpdated - handles stock return product update events
func StockReturnProductUpdated(message string, offset int64, partition int32) {
	logger.Info("StockReturnProductUpdated: Processing message at offset=%d, partition=%d", offset, partition)

	docData := TransStockReturnProductDecode(message)
	if docData.DocNo == "" {
		logger.Info("StockReturnProductUpdated: Invalid or empty document data, skipping")
		return
	}

	err := ProcessStockReturnProductDocument(docData)
	if err != nil {
		logger.Info("StockReturnProductUpdated: Error processing DocNo=%s: %v", docData.DocNo, err)
		return
	}

	logger.Info("StockReturnProductUpdated: Successfully processed DocNo=%s", docData.DocNo)
}

// StockReturnProductDeleted - handles stock return product deletion events
func StockReturnProductDeleted(message string, offset int64, partition int32) {
	logger.Info("StockReturnProductDeleted: Processing message at offset=%d, partition=%d", offset, partition)

	docData := TransStockReturnProductDecode(message)
	if docData.DocNo == "" {
		logger.Info("StockReturnProductDeleted: Invalid or empty document data, skipping")
		return
	}

	err := OnConsumeMessageStockReturnProductDelete(message)
	if err != nil {
		logger.Info("StockReturnProductDeleted: Error processing DocNo=%s: %v", docData.DocNo, err)
		return
	}

	logger.Info("StockReturnProductDeleted: Successfully processed DocNo=%s", docData.DocNo)
}

// OnConsumeMessageStockReturnProductCreateOrUpdate - main entry point for stock return product create/update events
func OnConsumeMessageStockReturnProductCreateOrUpdate(msg string) error {
	return SafeConsumerWrapper("STOCK_RETURN_PRODUCT", func(msg string) error {
		return ProcessStockReturnProductDocument(TransStockReturnProductDecode(msg))
	})(msg)
}

// OnConsumeMessageStockReturnProductDelete - main entry point for stock return product delete events
func OnConsumeMessageStockReturnProductDelete(msg string) error {
	logger.Info("OnConsumeMessageStockReturnProductDelete: Processing deletion message")

	docData := TransStockReturnProductDecode(msg)
	if docData.HoldingCode == "" || docData.DocNo == "" {
		logger.Error("OnConsumeMessageStockReturnProductDelete: Invalid data - missing HoldingCode or DocNo")
		return fmt.Errorf("invalid stock return product data")
	}

	return DeleteDocumentFromDatabases(context.Background(), docData.HoldingCode, docData.BusinessCode, docData.DocNo, TRANS_FLAG_STOCK_RETURN_PRODUCT)
}

// TransStockReturnProductDecode - decode JSON message to StockReturnProductStruct
func TransStockReturnProductDecode(message string) models.StockReturnProductStruct {
	var data models.StockReturnProductStruct
	DecodeKafkaMessage(message, &data, "StockReturnProduct")
	return data
}

// ProcessStockReturnProductThroughBuildDoc - processes stock return product document through build-doc system
func ProcessStockReturnProductDocument(docData models.StockReturnProductStruct) error {
	defer func() {
		if r := recover(); r != nil {
			logger.Info("ProcessStockReturnProductDocument: Panic recovered: %v", r)
			logger.Info("ProcessStockReturnProductDocument: Stack trace: %s", debug.Stack())
		}
	}()

	logger.Info("ProcessStockReturnProductDocument: Starting processing for DocNo=%s", docData.DocNo)
	build.DatabaseChecker(docData.HoldingCode, false)

	// Step 1: Convert Kafka message to ProcessMongoTransModel
	logger.Info("ProcessStockReturnProductDocument: Step 1 - Converting Kafka message to ProcessMongoTransModel")
	processData := ConvertStockReturnProductMongoDocToProcessModel(docData)

	// Step 2: Connect to PostgreSQL
	logger.Info("ProcessStockReturnProductDocument: Step 2 - Connecting to PostgreSQL for holdingCode=%s", docData.HoldingCode)
	db, err := mypg.PgSqlFastConnect(docData.HoldingCode)
	if err != nil {
		logger.Error("ProcessStockReturnProductDocument: Failed to connect to PostgreSQL: %v", err)
		return fmt.Errorf("failed to connect to PostgreSQL: %v", err)
	}
	// ใช้ connection pool ไม่ต้อง close
	ctx := context.Background()

	// Step 3: Delete existing documents from PostgreSQL
	logger.Info("ProcessStockReturnProductDocument: Step 3 - Deleting existing documents from PostgreSQL for DocNo=%s", docData.DocNo)
	mypg.DeleteDocPgSql(ctx, db, docData.BusinessCode, docData.DocNo, TRANS_FLAG_STOCK_RETURN_PRODUCT)

	// Step 4: Convert process model to build-doc structs
	logger.Info("ProcessStockReturnProductDocument: Step 4 - Converting process model to build-doc structs")
	docStruct, docPaymentStruct := myglobal.MapDocStructFromMongo(processData, docData.HoldingCode)
	docDetailStructs := MapStockReturnProductToDocDetailStructs(processData, docData.HoldingCode)
	setDocumentCompany(&docStruct, docDetailStructs, docData.BusinessCode)

	// Step 5: Create document references
	logger.Info("ProcessStockReturnProductDocument: Step 5 - Creating document references")
	var docRefStructs []models.DocRefStruct
	for _, refNo := range processData.DocReferences {
		docRefStructs = append(docRefStructs,
			myglobal.MapDocRefStruct(processData.DocNo, processData.TransFlag, refNo))
	}

	// Step 6-9: Insert to PostgreSQL and ClickHouse using shared functions
	err = InsertDocumentToPostgreSQL(ctx, db, docStruct, docRefStructs, docPaymentStruct, 6)
	if err != nil {
		return fmt.Errorf("failed to insert document to PostgreSQL: %w", err)
	}

	err = InsertDocDetailToPostgreSQL(ctx, db, docData.HoldingCode, docDetailStructs, 7)
	if err != nil {
		return fmt.Errorf("failed to insert doc details to PostgreSQL: %w", err)
	}

	err = InsertDocumentToClickHouse(ctx, docData.HoldingCode, docStruct, docRefStructs, docPaymentStruct, docDetailStructs, 8)
	if err != nil {
		return fmt.Errorf("failed to insert to ClickHouse: %w", err)
	}

	// Step 9: Calculate stock cost using global config
	err = ProcessDocumentStockCalculation(db, docData.HoldingCode, docDetailStructs, 9)
	if err != nil {
		logger.Error("Failed to calculate stock cost: %v", err)
		// Don't return error, just log it
	}

	// Step 10: Add to document wait process queues
	logger.Info("ProcessStockReturnProductDocument: Step 10 - Adding to document wait process queues")
	mypg.AddToDocWaitProcessQueues(ctx, db, docData.DocNo, TRANS_FLAG_STOCK_RETURN_PRODUCT)

	logger.Success("ProcessStockReturnProductDocument: Successfully processed DocNo=%s", docData.DocNo)
	return nil
}

// ConvertStockReturnProductMongoDocToProcessModel - converts StockReturnProductStruct to ProcessModel format
func ConvertStockReturnProductMongoDocToProcessModel(docData models.StockReturnProductStruct) models.ProcessMongoTransModel {
	logger.Info("ConvertStockReturnProductMongoDocToProcessModel: Starting conversion for DocNo=%s", docData.DocNo)

	// Convert details with nil checks
	var details []models.ProcessMongoTransDetailModel
	if docData.Details != nil {
		logger.Info("ConvertStockReturnProductMongoDocToProcessModel: Processing %d details", len(docData.Details))
		for i, detail := range docData.Details {
			logger.Info("ConvertStockReturnProductMongoDocToProcessModel: Processing detail %d: ItemCode=%s", i+1, detail.ItemCode)

			// Convert ItemNames from models.LanguageModel to models.LanguageModel with nil check
			var itemNames []models.LanguageModel
			if detail.ItemNames != nil {
				for _, name := range detail.ItemNames {
					itemNames = append(itemNames, models.LanguageModel{
						Code: name.Code,
						Name: name.Name,
					})
				}
			}

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
		logger.Info("ConvertStockReturnProductMongoDocToProcessModel: WARNING - docData.Details is nil")
	}

	// Convert branch names from models.LanguageModel to models.LanguageModel with nil checks
	var branchNames []models.LanguageModel
	if docData.Branch.Names != nil {
		logger.Info("ConvertStockReturnProductMongoDocToProcessModel: Processing %d branch names", len(docData.Branch.Names))
		for _, name := range docData.Branch.Names {
			branchNames = append(branchNames, models.LanguageModel{
				Code: name.Code,
				Name: name.Name,
			})
		}
	} else {
		logger.Info("ConvertStockReturnProductMongoDocToProcessModel: WARNING - docData.Branch.Names is nil")
	}

	// Convert branch with nil checks
	logger.Info("ConvertStockReturnProductMongoDocToProcessModel: Converting branch - Code=%s, GuidFixed=%s", docData.Branch.Code, docData.Branch.GuidFixed)
	branch := models.BranchModel{
		Code:      docData.Branch.Code,
		GuidFixed: docData.Branch.GuidFixed,
		Names:     branchNames,
	}

	logger.Info("ConvertStockReturnProductMongoDocToProcessModel: Creating final ProcessMongoTransModel")
	result := models.ProcessMongoTransModel{
		HoldingCode:      docData.HoldingCode,
		BranchId:         docData.BranchId,
		GuidFixed:        docData.GuidFixed,
		DocNo:            docData.DocNo,
		DocDateTime:      docData.DocDateTime,
		TotalAmount:      docData.TotalAmount,
		RoundAmount:      docData.RoundAmount,
		PayCashAmount:    docData.PayCashAmount,
		PayCashChange:    docData.PayCashChange,
		PaymentDetailRaw: docData.PaymentDetailRaw,
		SlipUrl:          docData.SlipUrl,
		SaleChannelCode:  docData.SaleChannelCode,
		DeliveryAmount:   docData.DeliveryAmount,
		IsCancel:         docData.IsCancel,
		CancelReason:     docData.CancelReason,
		GuidPos:          docData.GuidPos,
		Branch:           branch,
		TransFlag:        TRANS_FLAG_STOCK_RETURN_PRODUCT,
		Details:          details,
		DocReferences:    []models.ProcessMongoDocReferenceModel{},
	}
	logger.Info("ConvertStockReturnProductMongoDocToProcessModel: Conversion completed successfully")
	return result
}

// MapStockReturnProductToDocDetailStructs - converts stock return product to document detail structs
func MapStockReturnProductToDocDetailStructs(processData models.ProcessMongoTransModel, holdingCode string) []models.DocDetailStruct {
	var docDetailStructs []models.DocDetailStruct

	for i, detail := range processData.Details {
		// For stock return product transactions, use positive value to increase stock
		docDetailStruct := models.DocDetailStruct{
			DocDateTime:     processData.DocDateTime,
			DocNo:           processData.DocNo,
			LineNumber:      i + 1,
			TransFlag:       TRANS_FLAG_STOCK_RETURN_PRODUCT,
			CalcFlag:        1, // Stock Return Product = increase stock (รับคืน)
			CalcSeq:         1,
			ItemCode:        detail.ItemCode,
			Description:     GetItemName(detail.ItemNames),
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
