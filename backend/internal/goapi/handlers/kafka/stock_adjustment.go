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

// StockAdjustmentCreated - handles stock adjustment creation events
func StockAdjustmentCreated(message string, offset int64, partition int32) {
	logger.Info("StockAdjustmentCreated: Processing message at offset=%d, partition=%d", offset, partition)

	err := OnConsumeMessageStockAdjustmentCreateOrUpdate(message)
	if err != nil {
		logger.Info("StockAdjustmentCreated: Error: %v", err)
		return
	}

	logger.Info("StockAdjustmentCreated: Successfully processed")
}

// StockAdjustmentUpdated - handles stock adjustment update events
func StockAdjustmentUpdated(message string, offset int64, partition int32) {
	logger.Info("StockAdjustmentUpdated: Processing message at offset=%d, partition=%d", offset, partition)

	err := OnConsumeMessageStockAdjustmentCreateOrUpdate(message)
	if err != nil {
		logger.Info("StockAdjustmentUpdated: Error: %v", err)
		return
	}

	logger.Info("StockAdjustmentUpdated: Successfully processed")
}

// StockAdjustmentDeleted - handles stock adjustment deletion events
func StockAdjustmentDeleted(message string, offset int64, partition int32) {
	logger.Info("StockAdjustmentDeleted: Processing message at offset=%d, partition=%d", offset, partition)

	err := OnConsumeMessageStockAdjustmentDelete(message)
	if err != nil {
		logger.Info("StockAdjustmentDeleted: Error: %v", err)
		return
	}

	logger.Info("StockAdjustmentDeleted: Successfully processed")
}

// TransStockAdjustmentDecode - decode JSON message to StockAdjustmentStruct
func TransStockAdjustmentDecode(message string) models.StockAdjustmentStruct {
	defer func() {
		if r := recover(); r != nil {
			logger.Info("TransStockAdjustmentDecode: Panic recovered: %v", r)
			logger.Info("TransStockAdjustmentDecode: Stack trace: %s", debug.Stack())
		}
	}()

	var data models.StockAdjustmentStruct
	DecodeKafkaMessage(message, &data, "StockAdjustment")

	logger.Info("TransStockAdjustmentDecode: Successfully decoded DocNo=%s, TransFlag=%d", data.DocNo, data.TransFlag)
	return data
}

// ProcessStockAdjustmentDocument - processes stock adjustment using build-doc system
func ProcessStockAdjustmentDocument(msg string) error {
	// ประมวลผลเอกสาร Stock Adjustment (ปรับปรุงสต็อก)
	logger.Info("--- ProcessStockAdjustmentDocument START ---")

	defer func() {
		if r := recover(); r != nil {
			logger.Info("PANIC recovered in ProcessStockAdjustmentDocument: %v", r)
			logger.Info("Stack trace: %s", debug.Stack())
			logger.Info("Message length: %d characters", len(msg))
		}
	}()

	// Decode stock adjustment from JSON message
	logger.Debug("Step 1: Decoding JSON message... ")
	docData := TransStockAdjustmentDecode(msg)
	logger.Info("Step 1: Stock Adjustment decoded successfully - HoldingCode=%s, DocNo=%s, TransFlag=%d",
		docData.HoldingCode, docData.DocNo, docData.TransFlag)

	if docData.HoldingCode == "" || docData.DocNo == "" {
		logger.Error("Invalid stock adjustment data - HoldingCode='%s', DocNo='%s'", docData.HoldingCode, docData.DocNo)
		return fmt.Errorf("invalid stock adjustment data - missing HoldingCode or DocNo")
	}

	// Convert StockAdjustmentStruct to ProcessMongoTransModel
	logger.Debug("Step 2: Converting to process model...")
	processData := ConvertStockAdjustmentMongoDocToProcessModel(docData)
	logger.Info("Step 2: Conversion completed - Process Model DocNo=%s, Details=%d",
		processData.DocNo, len(processData.Details))

	// Connect to database
	logger.Debug("Step 3: Connecting to PostgreSQL...")
	db, err := mypg.PgSqlFastConnect(docData.HoldingCode)
	if err != nil {
		logger.Error("Failed to connect to database: %v", err)
		return fmt.Errorf("failed to connect to database: %v", err)
	}
	logger.Debug("Step 3 completed: Connected to PostgreSQL")

	ctx := context.Background()

	// Delete existing documents (upsert behavior)
	logger.Debug("Step 4: Deleting existing documents...")
	mypg.DeleteDocPgSql(ctx, db, docData.BusinessCode, docData.DocNo, docData.TransFlag)
	logger.Debug("Step 4 completed: Existing documents deleted")

	// Convert to build-doc structs
	logger.Debug("Step 5: Converting to build-doc structs...")
	docStruct, docPaymentStruct := myglobal.MapDocStructFromMongo(processData, docData.HoldingCode)
	docDetailStructs := MapStockAdjustmentToDocDetailStructs(processData, docData.HoldingCode)
	setDocumentCompany(&docStruct, docDetailStructs, docData.BusinessCode)
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

	// Step 11: Add to processing queues
	logger.Debug("Step 11: Adding to processing queues...")
	mypg.AddToDocWaitProcessQueues(ctx, db, docData.DocNo, docData.TransFlag)
	logger.Debug("Step 11 completed: Added to processing queues")

	logger.Info("--- ProcessStockAdjustmentDocument COMPLETED SUCCESSFULLY: %s ---", docData.DocNo)
	return nil
}

// ConvertStockAdjustmentMongoDocToProcessModel - converts StockAdjustmentStruct to ProcessModel format
func ConvertStockAdjustmentMongoDocToProcessModel(docData models.StockAdjustmentStruct) models.ProcessMongoTransModel {
	logger.Info("ConvertStockAdjustmentMongoDocToProcessModel: Starting conversion for DocNo=%s", docData.DocNo)

	// Convert details with nil checks
	var details []models.ProcessMongoTransDetailModel
	if docData.Details != nil {
		logger.Info("ConvertStockAdjustmentMongoDocToProcessModel: Processing %d details", len(docData.Details))
		for i, detail := range docData.Details {
			logger.Info("ConvertStockAdjustmentMongoDocToProcessModel: Processing detail %d: ItemCode=%s", i+1, detail.ItemCode)

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
		logger.Info("ConvertStockAdjustmentMongoDocToProcessModel: WARNING - docData.Details is nil")
	}

	branchNames := ConvertLanguageModels(docData.Branch.Names)

	// Convert branch with nil checks
	logger.Info("ConvertStockAdjustmentMongoDocToProcessModel: Converting branch - Code=%s, GuidFixed=%s", docData.Branch.Code, docData.Branch.GuidFixed)
	branch := models.BranchModel{
		Code:      docData.Branch.Code,
		GuidFixed: docData.Branch.GuidFixed,
		Names:     branchNames,
	}

	logger.Info("ConvertStockAdjustmentMongoDocToProcessModel: Creating final ProcessMongoTransModel")
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
		TransFlag:        docData.TransFlag, // 66 or 68
		Details:          details,
		DocReferences:    []models.ProcessMongoDocReferenceModel{},
	}
	logger.Info("ConvertStockAdjustmentMongoDocToProcessModel: Conversion completed successfully")
	return result
}

// MapStockAdjustmentToDocDetailStructs - converts stock adjustment to document detail structs
func MapStockAdjustmentToDocDetailStructs(processData models.ProcessMongoTransModel, holdingCode string) []models.DocDetailStruct {
	var docDetailStructs []models.DocDetailStruct

	for i, detail := range processData.Details {
		// Determine CalcFlag based on TransFlag
		var calcFlag float64
		if processData.TransFlag == TRANS_FLAG_STOCK_ADJUSTMENT_INCREASE {
			calcFlag = 1 // Stock Adjustment Increase = increase stock (+1)
		} else if processData.TransFlag == TRANS_FLAG_STOCK_ADJUSTMENT_DECREASE {
			calcFlag = -1 // Stock Adjustment Decrease = decrease stock (-1)
		} else {
			// Use myglobal.GetTransactionMultiplier for safety
			calcFlag = myglobal.GetTransactionMultiplier(processData.TransFlag)
		}

		docDetailStruct := models.DocDetailStruct{
			DocDateTime:     processData.DocDateTime,
			DocNo:           processData.DocNo,
			LineNumber:      i + 1,
			TransFlag:       processData.TransFlag,
			CalcFlag:        calcFlag,
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

// OnConsumeMessageStockAdjustmentCreateOrUpdate - handles stock adjustment create/update messages
func OnConsumeMessageStockAdjustmentCreateOrUpdate(msg string) error {
	return SafeConsumerWrapper("STOCK_ADJUSTMENT", func(msg string) error {
		return ProcessStockAdjustmentDocument(msg)
	})(msg)
}

// OnConsumeMessageStockAdjustmentDelete - handles stock adjustment delete messages
func OnConsumeMessageStockAdjustmentDelete(msg string) error {
	// รับ Message จาก Kafka ที่เป็นการลบเอกสาร Stock Adjustment
	// msg จะเป็น JSON string ที่มีข้อมูลของเอกสารที่ต้องการลบ

	logger.Info("OnConsumeMessageStockAdjustmentDelete: %s", msg)

	docData := TransStockAdjustmentDecode(msg)

	if docData.HoldingCode == "" || docData.DocNo == "" {
		logger.Error("Invalid stock adjustment data - missing HoldingCode or DocNo")
		return fmt.Errorf("invalid stock adjustment data")
	}

	return DeleteDocumentFromDatabases(context.Background(), docData.HoldingCode, docData.BusinessCode, docData.DocNo, docData.TransFlag)
}
