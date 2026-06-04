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

// OnConsumeMessagePurchaseRequisitionCreateOrUpdate - handles purchase requisition create/update messages
func OnConsumeMessagePurchaseRequisitionCreateOrUpdate(msg string) error {
	return SafeConsumerWrapper("PURCHASE_REQUISITION", func(msg string) error {
		return ProcessPurchaseRequisitionDocument(msg)
	})(msg)
}

// OnConsumeMessagePurchaseRequisitionDelete - handles purchase requisition delete messages
func OnConsumeMessagePurchaseRequisitionDelete(msg string) error {
	logger.Info("OnConsumeMessagePurchaseRequisitionDelete: %s", msg)

	docData := TransPurchaseRequisitionDecode(msg)

	if docData.HoldingCode == "" || docData.DocNo == "" {
		logger.Error("Invalid purchase requisition data - missing HoldingCode or DocNo")
		return fmt.Errorf("invalid purchase requisition data")
	}

	return DeleteDocumentFromDatabases(context.Background(), docData.HoldingCode, docData.DocNo, TRANS_FLAG_PURCHASE_REQUISITION)
}

// ProcessPurchaseRequisitionDocument - processes purchase requisition using build-doc system
func ProcessPurchaseRequisitionDocument(msg string) error {
	logger.Info("--- ProcessPurchaseRequisitionDocument START ---")

	defer func() {
		if r := recover(); r != nil {
			logger.Error("PANIC recovered in ProcessPurchaseRequisitionDocument: %v", r)
			logger.Error("Stack trace: %s", debug.Stack())
		}
	}()

	// Decode purchase requisition from JSON message
	purchaseRequisitionData := TransPurchaseRequisitionDecode(msg)
	logger.Info("Purchase Requisition decoded - HoldingCode=%s, DocNo=%s, TotalAmount=%.2f, Details=%d",
		purchaseRequisitionData.HoldingCode, purchaseRequisitionData.DocNo, purchaseRequisitionData.TotalAmount, len(purchaseRequisitionData.Details))

	if purchaseRequisitionData.HoldingCode == "" || purchaseRequisitionData.DocNo == "" {
		return fmt.Errorf("invalid purchase requisition data - missing HoldingCode or DocNo")
	}

	// Convert MongoDocModel to ProcessMongoTransModel
	processData := ConvertPurchaseRequisitionMongoDocToProcessModel(purchaseRequisitionData)

	// Connect to database
	db, err := mypg.PgSqlFastConnect(purchaseRequisitionData.HoldingCode)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	ctx := context.Background()

	// Delete existing documents (upsert behavior)
	mypg.DeleteDocPgSql(ctx, db, purchaseRequisitionData.DocNo, TRANS_FLAG_PURCHASE_REQUISITION)

	// Convert to build-doc structs
	docStruct, docPaymentStruct := myglobal.MapDocStructFromMongo(processData, purchaseRequisitionData.HoldingCode)
	docDetailStructs := MapPurchaseRequisitionToDocDetailStructs(processData, purchaseRequisitionData.HoldingCode)

	// Create doc references if any
	var docRefStructs []models.DocRefStruct
	for _, refNo := range processData.DocReferences {
		docRefStructs = append(docRefStructs,
			myglobal.MapDocRefStruct(processData.DocNo, processData.TransFlag, refNo))
	}

	// Insert to PostgreSQL
	err = InsertDocumentToPostgreSQL(ctx, db, docStruct, docRefStructs, docPaymentStruct, 0)
	if err != nil {
		return fmt.Errorf("failed to insert document to PostgreSQL: %w", err)
	}

	err = InsertDocDetailToPostgreSQL(ctx, db, purchaseRequisitionData.HoldingCode, docDetailStructs, 0)
	if err != nil {
		return fmt.Errorf("failed to insert doc details to PostgreSQL: %w", err)
	}

	// Insert to ClickHouse
	err = InsertDocumentToClickHouse(ctx, purchaseRequisitionData.HoldingCode, docStruct, docRefStructs, docPaymentStruct, docDetailStructs, 0)
	if err != nil {
		return fmt.Errorf("failed to insert to ClickHouse: %w", err)
	}

	// Add to processing queues
	mypg.AddToDocWaitProcessQueues(ctx, db, purchaseRequisitionData.DocNo, TRANS_FLAG_PURCHASE_REQUISITION)

	logger.Info("--- ProcessPurchaseRequisitionDocument COMPLETED SUCCESSFULLY: %s ---", purchaseRequisitionData.DocNo)
	return nil
}

// MapPurchaseRequisitionToDocDetailStructs - converts purchase requisition to document detail structs
func MapPurchaseRequisitionToDocDetailStructs(processData models.ProcessMongoTransModel, holdingCode string) []models.DocDetailStruct {
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

		// Purchase Requisition = no stock movement (CalcFlag = 0)
		docDetailStruct := models.DocDetailStruct{
			DocDateTime:     processData.DocDateTime,
			DocNo:           processData.DocNo,
			LineNumber:      i + 1,
			TransFlag:       TRANS_FLAG_PURCHASE_REQUISITION,
			CalcFlag:        0, // Purchase Requisition = no stock movement
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

// TransPurchaseRequisitionDecode - decodes transaction purchase requisition JSON data
func TransPurchaseRequisitionDecode(jsonData string) models.MongoDocModel {
	var docData models.MongoDocModel
	DecodeKafkaMessage(jsonData, &docData, "PurchaseRequisition")
	return docData
}

// ConvertPurchaseRequisitionMongoDocToProcessModel - converts MongoDocModel to ProcessMongoTransModel for purchase requisition
func ConvertPurchaseRequisitionMongoDocToProcessModel(mongoDoc models.MongoDocModel) models.ProcessMongoTransModel {
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
		TransFlag:        TRANS_FLAG_PURCHASE_REQUISITION,
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
