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

// StockPickupProductCreated - handles stock pickup product creation events
func StockPickupProductCreated(message string, offset int64, partition int32) {
	logger.Info("StockPickupProductCreated: Processing message at offset=%d, partition=%d", offset, partition)

	docData := TransStockPickupProductDecode(message)
	if docData.DocNo == "" {
		logger.Info("StockPickupProductCreated: Invalid or empty document data, skipping")
		return
	}

	err := ProcessStockPickupProductDocument(docData)
	if err != nil {
		logger.Info("StockPickupProductCreated: Error processing DocNo=%s: %v", docData.DocNo, err)
		return
	}

	logger.Info("StockPickupProductCreated: Successfully processed DocNo=%s", docData.DocNo)
}

// StockPickupProductUpdated - handles stock pickup product update events
func StockPickupProductUpdated(message string, offset int64, partition int32) {
	logger.Info("StockPickupProductUpdated: Processing message at offset=%d, partition=%d", offset, partition)

	docData := TransStockPickupProductDecode(message)
	if docData.DocNo == "" {
		logger.Info("StockPickupProductUpdated: Invalid or empty document data, skipping")
		return
	}

	err := ProcessStockPickupProductDocument(docData)
	if err != nil {
		logger.Info("StockPickupProductUpdated: Error processing DocNo=%s: %v", docData.DocNo, err)
		return
	}

	logger.Info("StockPickupProductUpdated: Successfully processed DocNo=%s", docData.DocNo)
}

// StockPickupProductDeleted - handles stock pickup product deletion events
func StockPickupProductDeleted(message string, offset int64, partition int32) {
	logger.Info("StockPickupProductDeleted: Processing message at offset=%d, partition=%d", offset, partition)

	docData := TransStockPickupProductDecode(message)
	if docData.DocNo == "" {
		logger.Info("StockPickupProductDeleted: Invalid or empty document data, skipping")
		return
	}

	err := OnConsumeMessageStockPickupProductDelete(message)
	if err != nil {
		logger.Info("StockPickupProductDeleted: Error deleting DocNo=%s: %v", docData.DocNo, err)
		return
	}

	logger.Info("StockPickupProductDeleted: Successfully deleted DocNo=%s", docData.DocNo)
}

// OnConsumeMessageStockPickupProductCreateOrUpdate - handles stock pickup product create/update messages
func OnConsumeMessageStockPickupProductCreateOrUpdate(msg string) error {
	return SafeConsumerWrapper("STOCK_PICKUP_PRODUCT", func(msg string) error {
		return ProcessStockPickupProductDocument(TransStockPickupProductDecode(msg))
	})(msg)
}

// OnConsumeMessageStockPickupProductDelete - handles stock pickup product delete messages
func OnConsumeMessageStockPickupProductDelete(msg string) error {
	logger.Info("OnConsumeMessageStockPickupProductDelete: Processing deletion message")

	docData := TransStockPickupProductDecode(msg)
	if docData.HoldingCode == "" || docData.DocNo == "" {
		logger.Error("OnConsumeMessageStockPickupProductDelete: Invalid data - missing HoldingCode or DocNo")
		return fmt.Errorf("invalid stock pickup product data")
	}

	return DeleteDocumentFromDatabases(context.Background(), docData.HoldingCode, docData.BusinessCode, docData.DocNo, TRANS_FLAG_STOCK_PICKUP_PRODUCT)
}

// ProcessStockPickupProductDocument - processes stock pickup product document following standardized 10-step pattern
func ProcessStockPickupProductDocument(docData models.StockPickupProductStruct) error {
	defer func() {
		if r := recover(); r != nil {
			logger.Info("ProcessStockPickupProductDocument: Panic recovered: %v", r)
			logger.Info("ProcessStockPickupProductDocument: Stack trace: %s", debug.Stack())
		}
	}()

	logger.Info("ProcessStockPickupProductDocument: Starting processing for DocNo=%s", docData.DocNo)
	build.DatabaseChecker(docData.HoldingCode, false)

	// Step 1: Convert Kafka message to ProcessMongoTransModel
	logger.Info("ProcessStockPickupProductDocument: Step 1 - Converting Kafka message to ProcessMongoTransModel")
	processData := ConvertStockPickupProductMongoDocToProcessModel(docData)

	// Step 2: Connect to PostgreSQL
	logger.Info("ProcessStockPickupProductDocument: Step 2 - Connecting to PostgreSQL for holdingCode=%s", docData.HoldingCode)
	db, err := mypg.PgSqlFastConnect(docData.HoldingCode)
	if err != nil {
		logger.Error("ProcessStockPickupProductDocument: Failed to connect to PostgreSQL: %v", err)
		return fmt.Errorf("failed to connect to PostgreSQL: %v", err)
	}
	// ใช้ connection pool ไม่ต้อง close
	ctx := context.Background()

	// Step 3: Delete existing documents from PostgreSQL
	logger.Info("ProcessStockPickupProductDocument: Step 3 - Deleting existing documents from PostgreSQL for DocNo=%s", docData.DocNo)
	mypg.DeleteDocPgSql(ctx, db, docData.BusinessCode, docData.DocNo, TRANS_FLAG_STOCK_PICKUP_PRODUCT)

	// Step 4: Convert process model to build-doc structs
	logger.Info("ProcessStockPickupProductDocument: Step 4 - Converting process model to build-doc structs")
	docStruct, docPaymentStruct := myglobal.MapDocStructFromMongo(processData, docData.HoldingCode)
	docDetailStructs := MapStockPickupProductToDocDetailStructs(processData, docData.HoldingCode)
	setDocumentCompany(&docStruct, docDetailStructs, docData.BusinessCode)

	// Step 5: Create document references
	logger.Info("ProcessStockPickupProductDocument: Step 5 - Creating document references")
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
	logger.Info("ProcessStockPickupProductDocument: Step 10 - Adding to document wait process queues")
	mypg.AddToDocWaitProcessQueues(ctx, db, docData.DocNo, TRANS_FLAG_STOCK_PICKUP_PRODUCT)

	logger.Success("ProcessStockPickupProductDocument: Successfully processed DocNo=%s", docData.DocNo)
	return nil
}

// MapStockPickupProductToDocDetailStructs - converts stock pickup product to document detail structs
func MapStockPickupProductToDocDetailStructs(processData models.ProcessMongoTransModel, holdingCode string) []models.DocDetailStruct {
	var docDetailStructs []models.DocDetailStruct

	for i, detail := range processData.Details {
		// For stock pickup product transactions, use negative value to decrease stock
		docDetailStruct := models.DocDetailStruct{
			DocDateTime:     processData.DocDateTime,
			DocNo:           processData.DocNo,
			LineNumber:      i + 1,
			TransFlag:       TRANS_FLAG_STOCK_PICKUP_PRODUCT,
			CalcFlag:        -1, // Stock Pickup Product = decrease stock (เบิกออก)
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

// ConvertStockPickupProductMongoDocToProcessModel - converts StockPickupProductStruct to ProcessMongoTransModel
func ConvertStockPickupProductMongoDocToProcessModel(docData models.StockPickupProductStruct) models.ProcessMongoTransModel {
	logger.Info("ConvertStockPickupProductMongoDocToProcessModel: Starting conversion for DocNo=%s", docData.DocNo)

	// Convert details with nil checks
	var details []models.ProcessMongoTransDetailModel
	if docData.Details != nil {
		logger.Info("ConvertStockPickupProductToProcessModel: Processing %d details", len(docData.Details))
		for i, detail := range docData.Details {
			logger.Info("ConvertStockPickupProductToProcessModel: Processing detail %d: ItemCode=%s", i+1, detail.ItemCode)

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
		logger.Info("ConvertStockPickupProductToProcessModel: WARNING - docData.Details is nil")
	}

	// Convert branch names from models.LanguageModel to models.LanguageModel with nil checks
	var branchNames []models.LanguageModel
	if docData.Branch.Names != nil {
		logger.Info("ConvertStockPickupProductToProcessModel: Processing %d branch names", len(docData.Branch.Names))
		for _, name := range docData.Branch.Names {
			branchNames = append(branchNames, models.LanguageModel{
				Code: name.Code,
				Name: name.Name,
			})
		}
	} else {
		logger.Info("ConvertStockPickupProductToProcessModel: WARNING - docData.Branch.Names is nil")
	}

	// Convert branch with nil checks
	logger.Info("ConvertStockPickupProductToProcessModel: Converting branch - Code=%s, GuidFixed=%s", docData.Branch.Code, docData.Branch.GuidFixed)
	branch := models.BranchModel{
		Code:      docData.Branch.Code,
		GuidFixed: docData.Branch.GuidFixed,
		Names:     branchNames,
	}

	logger.Info("ConvertStockPickupProductToProcessModel: Creating final ProcessMongoTransModel")
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
		TransFlag:        TRANS_FLAG_STOCK_PICKUP_PRODUCT,
		Details:          details,
		DocReferences:    []models.ProcessMongoDocReferenceModel{},
	}
	logger.Info("ConvertStockPickupProductToProcessModel: Conversion completed successfully")
	return result
}

// TransStockPickupProductDecode - decodes transaction stock pickup product JSON data
func TransStockPickupProductDecode(jsonData string) models.StockPickupProductStruct {
	var data models.StockPickupProductStruct
	DecodeKafkaMessage(jsonData, &data, "StockPickupProduct")
	return data
}
