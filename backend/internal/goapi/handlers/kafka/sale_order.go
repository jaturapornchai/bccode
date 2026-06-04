package kafka

import (
	"context"
	"fmt"
	"smlcloudplatform/internal/goapi/logger"

	"runtime/debug"

	"smlcloudplatform/internal/goapi/models"
	"smlcloudplatform/internal/goapi/myglobal"
	"smlcloudplatform/internal/goapi/mypg"
)

// OnConsumeMessageSaleOrderCreateOrUpdate - รับ message จาก Kafka สำหรับสร้าง/แก้ไขใบสั่งขาย
func OnConsumeMessageSaleOrderCreateOrUpdate(msg string) error {
	return SafeConsumerWrapper("SALE_ORDER", func(msg string) error {
		return ProcessSaleOrderDocument(msg)
	})(msg)
}

// OnConsumeMessageSaleOrderDelete - รับ message จาก Kafka สำหรับลบใบสั่งขาย
// msg คือ JSON string เช่น {"holding_code": "shop123", "docno": "SO2024001"}
func OnConsumeMessageSaleOrderDelete(msg string) error {

	logger.Info("OnConsumeMessageSaleOrderDelete: %s", msg)

	docData := TransSaleOrderDecode(msg)

	if docData.HoldingCode == "" || docData.DocNo == "" {
		logger.Error("Invalid sale order data - missing HoldingCode or DocNo")
		return fmt.Errorf("invalid sale order data")
	}

	return DeleteDocumentFromDatabases(context.Background(), docData.HoldingCode, docData.DocNo, TRANS_FLAG_SALE_ORDER)
}

// ProcessSaleOrderDocument - ประมวลผลเอกสาร Sale Order (ใบสั่งขาย)
// รับ JSON message จาก Kafka แล้วบันทึกลง PostgreSQL และ ClickHouse
// ระบบนี้ไม่มีการตรวจสต็อก เพราะ MongoDB มี → PostgreSQL, ClickHouse ก็ต้องมีด้วย
func ProcessSaleOrderDocument(msg string) error {
	logger.Info("--- ProcessSaleOrderDocument START ---")

	defer func() {
		if r := recover(); r != nil {
			logger.Info("PANIC recovered in ProcessSaleOrderDocument: %v", r)
			logger.Info("Stack trace: %s", debug.Stack())
			logger.Info("Message length: %d characters", len(msg))
		}
	}()

	// Step 1: แปลง JSON message เป็น struct
	logger.Debug("Step 1: Decoding JSON message... ")
	saleOrderData := TransSaleOrderDecode(msg)
	logger.Info("Step 1: Sale Order decoded successfully - HoldingCode=%s, DocNo=%s, TotalAmount=%.2f, Details=%d",
		saleOrderData.HoldingCode, saleOrderData.DocNo, saleOrderData.TotalAmount, len(saleOrderData.Details))

	if saleOrderData.HoldingCode == "" || saleOrderData.DocNo == "" {
		logger.Error("Invalid sale order data - HoldingCode='%s', DocNo='%s'", saleOrderData.HoldingCode, saleOrderData.DocNo)
		return fmt.Errorf("invalid sale order data - missing HoldingCode or DocNo")
	}

	// Step 2: แปลง MongoDocModel เป็น ProcessMongoTransModel
	logger.Debug("Step 2: Converting to process model...")
	processData := ConvertSaleOrderMongoDocToProcessModel(saleOrderData)
	logger.Info("Step 2: Conversion completed - Process Model DocNo=%s, Details=%d",
		processData.DocNo, len(processData.Details))

	// Step 3: เชื่อมต่อ PostgreSQL
	logger.Debug("Step 3: Connecting to PostgreSQL...")
	db, err := mypg.PgSqlFastConnect(saleOrderData.HoldingCode)
	if err != nil {
		logger.Error("Failed to connect to database: %v", err)
		return fmt.Errorf("failed to connect to database: %v", err)
	}
	logger.Debug("Step 3 completed: Connected to PostgreSQL")

	ctx := context.Background()

	// Step 4: ลบเอกสารเดิม (upsert behavior)
	logger.Debug("Step 4: Deleting existing documents...")
	mypg.DeleteDocPgSql(ctx, db, saleOrderData.DocNo, TRANS_FLAG_SALE_ORDER)
	logger.Debug("Step 4 completed: Existing documents deleted (if any)")

	// Step 5: แปลงเป็น build-doc structs
	logger.Debug("Step 5: Converting to build-doc structs...")
	docStruct, docPaymentStruct := myglobal.MapDocStructFromMongo(processData, saleOrderData.HoldingCode)
	docDetailStructs := MapSaleOrderToDocDetailStructs(processData, saleOrderData.HoldingCode)
	logger.Debug("Step 5 completed: Converted to %d doc details", len(docDetailStructs))

	// Step 6: สร้าง document references (ถ้ามี)
	logger.Debug("Step 6: Creating document references...")
	var docRefStructs []models.DocRefStruct
	for _, refNo := range processData.DocReferences {
		docRefStructs = append(docRefStructs,
			myglobal.MapDocRefStruct(processData.DocNo, processData.TransFlag, refNo))
	}
	logger.Debug("Step 6 completed: Created %d document references", len(docRefStructs))

	// Step 7-9: บันทึกลง PostgreSQL และ ClickHouse
	err = InsertDocumentToPostgreSQL(ctx, db, docStruct, docRefStructs, docPaymentStruct, 7)
	if err != nil {
		return fmt.Errorf("failed to insert document to PostgreSQL: %w", err)
	}

	err = InsertDocDetailToPostgreSQL(ctx, db, saleOrderData.HoldingCode, docDetailStructs, 8)
	if err != nil {
		return fmt.Errorf("failed to insert doc details to PostgreSQL: %w", err)
	}

	err = InsertDocumentToClickHouse(ctx, saleOrderData.HoldingCode, docStruct, docRefStructs, docPaymentStruct, docDetailStructs, 9)
	if err != nil {
		return fmt.Errorf("failed to insert to ClickHouse: %w", err)
	}

	// Step 10: คำนวณต้นทุนสต็อก
	err = ProcessDocumentStockCalculation(db, saleOrderData.HoldingCode, docDetailStructs, 10)
	if err != nil {
		logger.Error("Failed to calculate stock cost: %v", err)
		// Don't return error, just log it
	}

	// Step 11: เพิ่มเข้า queue รอประมวลผล
	logger.Debug("Step 11: Adding to processing queues...")
	mypg.AddToDocWaitProcessQueues(ctx, db, saleOrderData.DocNo, TRANS_FLAG_SALE_ORDER)
	logger.Debug("Step 11 completed: Added to processing queues")

	// Step 12: ประมวลผลสถานะเอกสาร
	logger.Debug("Step 12: Processing document status for holdingCode=%s", saleOrderData.HoldingCode)
	ProcessDocumentStatusAsync(saleOrderData.HoldingCode)

	logger.Info("--- ProcessSaleOrderDocument COMPLETED SUCCESSFULLY: %s ---", saleOrderData.DocNo)
	return nil
}

// ConvertSaleOrderMongoDocToProcessModel - แปลง MongoDocModel เป็น ProcessMongoTransModel สำหรับใบสั่งขาย
func ConvertSaleOrderMongoDocToProcessModel(mongoDoc models.MongoDocModel) models.ProcessMongoTransModel {
	logger.Info("ConvertSaleOrderMongoDocToProcessModel: Starting conversion for DocNo=%s", mongoDoc.DocNo)

	// แปลง details พร้อมตรวจสอบ nil
	var details []models.ProcessMongoTransDetailModel
	if mongoDoc.Details != nil {
		logger.Info("ConvertSaleOrderMongoDocToProcessModel: Processing %d details", len(mongoDoc.Details))
		for i, detail := range mongoDoc.Details {
			logger.Info("ConvertSaleOrderMongoDocToProcessModel: Processing detail %d: ItemCode=%s", i+1, detail.ItemCode)

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
		logger.Info("ConvertSaleOrderMongoDocToProcessModel: WARNING - mongoDoc.Details is nil")
	}

	branchNames := ConvertLanguageModels(mongoDoc.Branch.Names)

	// แปลง branch พร้อมตรวจสอบ nil
	logger.Info("ConvertSaleOrderMongoDocToProcessModel: Converting branch - Code=%s, GuidFixed=%s", mongoDoc.Branch.Code, mongoDoc.Branch.GuidFixed)
	branch := models.BranchModel{
		Code:      mongoDoc.Branch.Code,
		GuidFixed: mongoDoc.Branch.GuidFixed,
		Names:     branchNames,
	}

	logger.Info("ConvertSaleOrderMongoDocToProcessModel: Creating final ProcessMongoTransModel")
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
		TransFlag:        TRANS_FLAG_SALE_ORDER, // Sale Order transaction flag
		Details:          details,
		DocReferences:    []models.ProcessMongoDocReferenceModel{}, // Initialize empty
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

	logger.Info("ConvertSaleOrderMongoDocToProcessModel: Conversion completed successfully")
	return result
}

// MapSaleOrderToDocDetailStructs - แปลงข้อมูลใบสั่งขายเป็น DocDetailStruct สำหรับบันทึกลง database
// CalcFlag = 0 หมายถึงไม่เคลื่อนไหวสต็อก (รอแปลงเป็นใบขายก่อน)
func MapSaleOrderToDocDetailStructs(processData models.ProcessMongoTransModel, holdingCode string) []models.DocDetailStruct {
	var docDetailStructs []models.DocDetailStruct

	for i, detail := range processData.Details {
		// ใบสั่งขายไม่เคลื่อนไหวสต็อก (CalcFlag = 0)
		docDetailStruct := models.DocDetailStruct{
			DocDateTime:     processData.DocDateTime,
			DocNo:           processData.DocNo,
			LineNumber:      i + 1,
			TransFlag:       TRANS_FLAG_SALE_ORDER,
			CalcFlag:        0, // Sale Order = no stock movement
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

// TransSaleOrderDecode - แปลง JSON string เป็น MongoDocModel สำหรับใบสั่งขาย
func TransSaleOrderDecode(jsonData string) models.MongoDocModel {
	var docData models.MongoDocModel
	DecodeKafkaMessage(jsonData, &docData, "SaleOrder")
	return docData
}
