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

// OnConsumeMessageRFQCreateOrUpdate - handles RFQ create/update messages
func OnConsumeMessageRFQCreateOrUpdate(msg string) error {
	return SafeConsumerWrapper("RFQ", func(msg string) error {
		return ProcessRFQDocument(msg)
	})(msg)
}

// OnConsumeMessageRFQDelete - handles RFQ delete messages
func OnConsumeMessageRFQDelete(msg string) error {
	logger.Info("OnConsumeMessageRFQDelete: %s", msg)

	docData := TransRFQDecode(msg)

	if docData.HoldingCode == "" || docData.DocNo == "" {
		logger.Error("Invalid RFQ data - missing HoldingCode or DocNo")
		return fmt.Errorf("invalid RFQ data")
	}

	return DeleteDocumentFromDatabases(context.Background(), docData.HoldingCode, docData.DocNo, TRANS_FLAG_RFQ)
}

// ProcessRFQDocument - processes RFQ using build-doc system
func ProcessRFQDocument(msg string) error {
	logger.Info("--- ProcessRFQDocument START ---")

	defer func() {
		if r := recover(); r != nil {
			logger.Error("PANIC recovered in ProcessRFQDocument: %v", r)
			logger.Error("Stack trace: %s", debug.Stack())
		}
	}()

	// Decode RFQ from JSON message
	rfqData := TransRFQDecode(msg)
	logger.Info("RFQ decoded - HoldingCode=%s, DocNo=%s, TotalAmount=%.2f, Details=%d",
		rfqData.HoldingCode, rfqData.DocNo, rfqData.TotalAmount, len(rfqData.Details))

	if rfqData.HoldingCode == "" || rfqData.DocNo == "" {
		return fmt.Errorf("invalid RFQ data - missing HoldingCode or DocNo")
	}

	// Convert MongoDocModel to ProcessMongoTransModel
	processData := ConvertRFQMongoDocToProcessModel(rfqData)

	// Connect to database
	db, err := mypg.PgSqlFastConnect(rfqData.HoldingCode)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	ctx := context.Background()

	// Delete existing documents (upsert behavior)
	mypg.DeleteDocPgSql(ctx, db, rfqData.DocNo, TRANS_FLAG_RFQ)

	// Convert to build-doc structs
	docStruct, docPaymentStruct := myglobal.MapDocStructFromMongo(processData, rfqData.HoldingCode)
	docDetailStructs := MapRFQToDocDetailStructs(processData, rfqData.HoldingCode)

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

	err = InsertDocDetailToPostgreSQL(ctx, db, rfqData.HoldingCode, docDetailStructs, 0)
	if err != nil {
		return fmt.Errorf("failed to insert doc details to PostgreSQL: %w", err)
	}

	// Insert to ClickHouse
	err = InsertDocumentToClickHouse(ctx, rfqData.HoldingCode, docStruct, docRefStructs, docPaymentStruct, docDetailStructs, 0)
	if err != nil {
		return fmt.Errorf("failed to insert to ClickHouse: %w", err)
	}

	// Add to processing queues
	mypg.AddToDocWaitProcessQueues(ctx, db, rfqData.DocNo, TRANS_FLAG_RFQ)

	logger.Info("--- ProcessRFQDocument COMPLETED SUCCESSFULLY: %s ---", rfqData.DocNo)
	return nil
}

// MapRFQToDocDetailStructs - converts RFQ to document detail structs
func MapRFQToDocDetailStructs(processData models.ProcessMongoTransModel, holdingCode string) []models.DocDetailStruct {
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

		// RFQ = no stock movement (CalcFlag = 0)
		docDetailStruct := models.DocDetailStruct{
			DocDateTime:     processData.DocDateTime,
			DocNo:           processData.DocNo,
			LineNumber:      i + 1,
			TransFlag:       TRANS_FLAG_RFQ,
			CalcFlag:        0, // RFQ = no stock movement
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

// TransRFQDecode - decodes transaction RFQ JSON data
func TransRFQDecode(jsonData string) models.MongoDocModel {
	var docData models.MongoDocModel
	DecodeKafkaMessage(jsonData, &docData, "RFQ")
	return docData
}

// ConvertRFQMongoDocToProcessModel - converts MongoDocModel to ProcessMongoTransModel for RFQ
func ConvertRFQMongoDocToProcessModel(mongoDoc models.MongoDocModel) models.ProcessMongoTransModel {
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
		TransFlag:        TRANS_FLAG_RFQ,
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
