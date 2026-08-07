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

// OnConsumeMessageStockBalanceCreateOrUpdate - handles stock balance create/update messages
func OnConsumeMessageStockBalanceCreateOrUpdate(msg string) error {
	return SafeConsumerWrapper("STOCK_BALANCE", func(msg string) error {
		return ProcessStockBalanceDocument(TransStockBalanceDecode(msg))
	})(msg)
}

// StockBalanceCreated - handles stock balance creation events
func StockBalanceCreated(message string, offset int64, partition int32) {
	logger.Info("StockBalanceCreated: Processing message at offset=%d, partition=%d", offset, partition)

	docData := TransStockBalanceDecode(message)
	if docData.DocNo == "" {
		logger.Info("StockBalanceCreated: Invalid or empty document data, skipping")
		return
	}

	err := ProcessStockBalanceDocument(docData)
	if err != nil {
		logger.Info("StockBalanceCreated: Error processing DocNo=%s: %v", docData.DocNo, err)
		return
	}

	logger.Info("StockBalanceCreated: Successfully processed DocNo=%s", docData.DocNo)
}

// StockBalanceUpdated - handles stock balance update events
func StockBalanceUpdated(message string, offset int64, partition int32) {
	logger.Info("StockBalanceUpdated: Processing message at offset=%d, partition=%d", offset, partition)

	docData := TransStockBalanceDecode(message)
	if docData.DocNo == "" {
		logger.Info("StockBalanceUpdated: Invalid or empty document data, skipping")
		return
	}

	err := ProcessStockBalanceDocument(docData)
	if err != nil {
		logger.Info("StockBalanceUpdated: Error processing DocNo=%s: %v", docData.DocNo, err)
		return
	}

	logger.Info("StockBalanceUpdated: Successfully processed DocNo=%s", docData.DocNo)
}

// StockBalanceDeleted - handles stock balance deletion events
func StockBalanceDeleted(message string, offset int64, partition int32) {
	logger.Info("StockBalanceDeleted: Processing message at offset=%d, partition=%d", offset, partition)

	docData := TransStockBalanceDecode(message)
	if docData.DocNo == "" {
		logger.Info("StockBalanceDeleted: Invalid or empty document data, skipping")
		return
	}

	err := OnConsumeMessageStockBalanceDelete(message)
	if err != nil {
		logger.Info("StockBalanceDeleted: Error processing deletion for DocNo=%s: %v", docData.DocNo, err)
		return
	}

	logger.Info("StockBalanceDeleted: Successfully processed deletion for DocNo=%s", docData.DocNo)
}

// TransStockBalanceDecode - decode JSON message to StockBalanceStruct
func TransStockBalanceDecode(message string) models.StockBalanceStruct {
	defer func() {
		if r := recover(); r != nil {
			logger.Info("TransStockBalanceDecode: Panic recovered: %v", r)
			logger.Info("TransStockBalanceDecode: Stack trace: %s", debug.Stack())
		}
	}()

	var data models.StockBalanceStruct
	DecodeKafkaMessage(message, &data, "StockBalance")

	logger.Info("TransStockBalanceDecode: Successfully decoded DocNo=%s, TransFlag=%d", data.DocNo, data.TransFlag)
	return data
}

// ProcessStockBalanceDocument - processes stock balance document following standardized 10-step pattern
func ProcessStockBalanceDocument(docData models.StockBalanceStruct) error {
	defer func() {
		if r := recover(); r != nil {
			logger.Info("ProcessStockBalanceDocument: Panic recovered: %v", r)
			logger.Info("ProcessStockBalanceDocument: Stack trace: %s", debug.Stack())
		}
	}()

	logger.Info("ProcessStockBalanceDocument: Starting processing for DocNo=%s", docData.DocNo)
	build.DatabaseChecker(docData.HoldingCode, false)

	// Step 1: Convert Kafka message to ProcessMongoTransModel
	logger.Info("ProcessStockBalanceDocument: Step 1 - Converting Kafka message to ProcessMongoTransModel")
	processData := ConvertStockBalanceMongoDocToProcessModel(docData)

	// Step 2: Connect to PostgreSQL
	logger.Info("ProcessStockBalanceDocument: Step 2 - Connecting to PostgreSQL for holdingCode=%s", docData.HoldingCode)
	db, err := mypg.PgSqlFastConnect(docData.HoldingCode)
	if err != nil {
		logger.Error("ProcessStockBalanceDocument: Failed to connect to PostgreSQL: %v", err)
		return fmt.Errorf("failed to connect to PostgreSQL: %v", err)
	}
	// ใช้ connection pool ไม่ต้อง close
	ctx := context.Background()

	// Step 3: Delete existing documents from PostgreSQL
	logger.Info("ProcessStockBalanceDocument: Step 3 - Deleting existing documents from PostgreSQL for DocNo=%s", docData.DocNo)
	mypg.DeleteDocPgSql(ctx, db, docData.DocNo, TRANS_FLAG_STOCK_BALANCE)

	// Step 4: Convert process model to build-doc structs
	logger.Info("ProcessStockBalanceDocument: Step 4 - Converting process model to build-doc structs")
	docStruct, docPaymentStruct := myglobal.MapDocStructFromMongo(processData, docData.HoldingCode)
	docDetailStructs := MapStockBalanceToDocDetailStructs(processData, docData.HoldingCode)

	// Step 5: Create document references
	logger.Info("ProcessStockBalanceDocument: Step 5 - Creating document references")
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
	logger.Info("ProcessStockBalanceDocument: Step 10 - Adding to document wait process queues")
	mypg.AddToDocWaitProcessQueues(ctx, db, docData.DocNo, TRANS_FLAG_STOCK_BALANCE)

	logger.Success("ProcessStockBalanceDocument: Successfully processed DocNo=%s", docData.DocNo)
	return nil
}

// OnConsumeMessageStockBalanceDelete - handles stock balance deletion events
func OnConsumeMessageStockBalanceDelete(message string) error {
	logger.Info("OnConsumeMessageStockBalanceDelete: Processing deletion message")

	docData := TransStockBalanceDecode(message)
	if docData.HoldingCode == "" || docData.DocNo == "" {
		logger.Error("OnConsumeMessageStockBalanceDelete: Invalid data - missing HoldingCode or DocNo")
		return fmt.Errorf("invalid stock balance data")
	}

	return DeleteDocumentFromDatabases(context.Background(), docData.HoldingCode, docData.DocNo, TRANS_FLAG_STOCK_BALANCE)
}

// ConvertStockBalanceMongoDocToProcessModel - converts StockBalanceStruct to ProcessModel format
func ConvertStockBalanceMongoDocToProcessModel(docData models.StockBalanceStruct) models.ProcessMongoTransModel {
	logger.Info("ConvertStockBalanceMongoDocToProcessModel: Starting conversion for DocNo=%s", docData.DocNo)

	// Convert details with nil checks
	var details []models.ProcessMongoTransDetailModel
	if docData.Details != nil {
		logger.Info("ConvertStockBalanceToProcessModel: Processing %d details", len(docData.Details))
		for i, detail := range docData.Details {
			logger.Info("ConvertStockBalanceToProcessModel: Processing detail %d: ItemCode=%s", i+1, detail.ItemCode)

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
		logger.Info("ConvertStockBalanceToProcessModel: WARNING - docData.Details is nil")
	}

	// Convert branch names from models.LanguageModel to models.LanguageModel with nil checks
	var branchNames []models.LanguageModel
	if docData.Branch.Names != nil {
		logger.Info("ConvertStockBalanceToProcessModel: Processing %d branch names", len(docData.Branch.Names))
		for _, name := range docData.Branch.Names {
			branchNames = append(branchNames, models.LanguageModel{
				Code: name.Code,
				Name: name.Name,
			})
		}
	} else {
		logger.Info("ConvertStockBalanceToProcessModel: WARNING - docData.Branch.Names is nil")
	}

	// Convert branch with nil checks
	logger.Info("ConvertStockBalanceToProcessModel: Converting branch - Code=%s, GuidFixed=%s", docData.Branch.Code, docData.Branch.GuidFixed)
	branch := models.BranchModel{
		Code:      docData.Branch.Code,
		GuidFixed: docData.Branch.GuidFixed,
		Names:     branchNames,
	}

	logger.Info("ConvertStockBalanceToProcessModel: Creating final ProcessMongoTransModel")
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
		TransFlag:        TRANS_FLAG_STOCK_BALANCE,
		Details:          details,
		DocReferences:    []models.ProcessMongoDocReferenceModel{},
	}
	logger.Info("ConvertStockBalanceToProcessModel: Conversion completed successfully")
	return result
}

// MapStockBalanceToDocDetailStructs - converts stock balance to document detail structs
func MapStockBalanceToDocDetailStructs(processData models.ProcessMongoTransModel, holdingCode string) []models.DocDetailStruct {
	var docDetailStructs []models.DocDetailStruct

	for i, detail := range processData.Details {
		// For stock balance transactions, use positive value to set initial stock (ยอดยกมา)
		docDetailStruct := models.DocDetailStruct{
			DocDateTime:     processData.DocDateTime,
			DocNo:           processData.DocNo,
			LineNumber:      i + 1,
			TransFlag:       TRANS_FLAG_STOCK_BALANCE,
			CalcFlag:        1, // Stock Balance = set initial stock (ยอดยกมา)
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

	return docDetailStructs
}
