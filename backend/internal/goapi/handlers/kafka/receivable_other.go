package kafka

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"

	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/models"
	"smlcloudplatform/internal/goapi/myglobal"
	"smlcloudplatform/internal/goapi/mypg"
	"smlcloudplatform/internal/goapi/process/build"
)

// ReceivableOtherKafkaDoc represents incoming Kafka payload for receivable other
type ReceivableOtherKafkaDoc struct {
	HoldingCode        string                  `json:"holdingcode"`
	BusinessCode       string                  `json:"businesscode"`
	GuidFixed          string                  `json:"guidfixed"`
	DocNo              string                  `json:"docno"`
	DocDatetime        time.Time               `json:"docdatetime"`
	DocType            int8                    `json:"doctype"`
	TransFlag          int                     `json:"transflag"`
	CustCode           string                  `json:"custcode"`
	SaleCode           string                  `json:"salecode"`
	SaleName           string                  `json:"salename"`
	TotalPaymentAmount float64                 `json:"totalpaymentamount"`
	TotalAmount        float64                 `json:"totalamount"`
	TotalBalance       float64                 `json:"totalbalance"`
	TotalValue         float64                 `json:"totalvalue"`
	PayCashAmount      float64                 `json:"paycashamount"`
	PaymentDetailRaw   string                  `json:"paymentdetailraw"`
	RoundAmount        float64                 `json:"roundamount"`
	RefDocNo           string                  `json:"refdocno"`
	RefDocDate         time.Time               `json:"refdocdate"`
	IsCancel           bool                    `json:"iscancel"`
	Branch             models.MongoBranchModel `json:"branch"`
	IsDelete           bool                    `json:"isdelete"`
	CreatedAt          time.Time               `json:"createdat"`
	CreatedBy          string                  `json:"createdby"`
	UpdatedAt          time.Time               `json:"updatedat"`
	UpdatedBy          string                  `json:"updatedby"`
}

// OnConsumeMessageReceivableOtherCreateOrUpdate handles receivable other create/update messages
func OnConsumeMessageReceivableOtherCreateOrUpdate(msg string) error {
	return SafeConsumerWrapper("RECEIVABLE_OTHER", func(msg string) error {
		return ProcessReceivableOtherDocument(msg)
	})(msg)
}

// OnConsumeMessageReceivableOtherDelete handles receivable other deletion messages
func OnConsumeMessageReceivableOtherDelete(msg string) error {
	logger.Info("OnConsumeMessageReceivableOtherDelete: %s", msg)

	docData := TransReceivableOtherDecode(msg)
	if docData.HoldingCode == "" || docData.DocNo == "" {
		logger.Error("Invalid receivable other data - missing HoldingCode or DocNo")
		return fmt.Errorf("invalid receivable other data")
	}

	transFlag := docData.TransFlag
	if transFlag == 0 {
		transFlag = TRANS_FLAG_RECEIVABLE_OTHER
	}

	return DeleteDocumentFromDatabases(context.Background(), docData.HoldingCode, docData.BusinessCode, docData.DocNo, transFlag)
}

// ProcessReceivableOtherDocument processes receivable other document into PostgreSQL
func ProcessReceivableOtherDocument(msg string) error {
	logger.Info("--- ProcessReceivableOtherDocument START ---")

	defer func() {
		if r := recover(); r != nil {
			logger.Error("PANIC recovered in ProcessReceivableOtherDocument: %v", r)
			logger.Error("Stack trace: %s", debug.Stack())
		}
	}()

	docData := TransReceivableOtherDecode(msg)
	if docData.HoldingCode == "" || docData.DocNo == "" {
		return fmt.Errorf("invalid receivable other data - missing HoldingCode or DocNo")
	}

	build.DatabaseChecker(docData.HoldingCode, false)

	transFlag := docData.TransFlag
	if transFlag == 0 {
		transFlag = TRANS_FLAG_RECEIVABLE_OTHER
	}

	processData := ConvertReceivableOtherDocToProcessModel(docData)

	db, err := mypg.PgSqlFastConnect(docData.HoldingCode)
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	ctx := context.Background()

	// Delete existing documents (upsert behavior)
	mypg.DeleteDocPgSql(ctx, db, docData.BusinessCode, docData.DocNo, transFlag)

	docStruct, docPaymentStruct := myglobal.MapDocStructFromMongo(processData, docData.HoldingCode)
	setDocumentCompany(&docStruct, nil, docData.BusinessCode)

	var docRefStructs []models.DocRefStruct
	if docData.RefDocNo != "" {
		docRefStructs = append(docRefStructs, models.DocRefStruct{
			DocNo:          docData.DocNo,
			DocNoTransFlag: transFlag,
			DocRefNo:       docData.RefDocNo,
		})
	}

	err = InsertDocumentToPostgreSQL(ctx, db, docStruct, docRefStructs, docPaymentStruct, 1)
	if err != nil {
		return fmt.Errorf("failed to insert receivable other document to PostgreSQL: %w", err)
	}

	err = InsertDocumentToClickHouse(ctx, docData.HoldingCode, docStruct, docRefStructs, docPaymentStruct, nil, 2)
	if err != nil {
		return fmt.Errorf("failed to insert to ClickHouse: %w", err)
	}

	mypg.AddToDocWaitProcessQueues(ctx, db, docData.DocNo, transFlag)

	logger.Success("ProcessReceivableOtherDocument: Successfully processed DocNo=%s, TransFlag=%d", docData.DocNo, transFlag)
	return nil
}

// ConvertReceivableOtherDocToProcessModel converts ReceivableOtherKafkaDoc to ProcessMongoTransModel
func ConvertReceivableOtherDocToProcessModel(docData ReceivableOtherKafkaDoc) models.ProcessMongoTransModel {
	transFlag := docData.TransFlag
	if transFlag == 0 {
		transFlag = TRANS_FLAG_RECEIVABLE_OTHER
	}

	totalAmount := docData.TotalAmount
	if totalAmount == 0 && docData.TotalPaymentAmount != 0 {
		totalAmount = docData.TotalPaymentAmount
	}

	return models.ProcessMongoTransModel{
		HoldingCode:      docData.HoldingCode,
		BusinessCode:     docData.BusinessCode,
		GuidFixed:        docData.GuidFixed,
		DocNo:            docData.DocNo,
		DocDateTime:      docData.DocDatetime,
		TransFlag:        transFlag,
		CustCode:         docData.CustCode,
		TotalAmount:      totalAmount,
		TotalValue:       docData.TotalValue,
		PayCashAmount:    docData.PayCashAmount,
		RoundAmount:      docData.RoundAmount,
		PaymentDetailRaw: docData.PaymentDetailRaw,
		IsCancel:         docData.IsCancel,
		Branch:           docData.Branch,
		CreatorCode:      docData.CreatedBy,
		CreatedAt:        docData.CreatedAt,
		ModifierCode:     docData.UpdatedBy,
		ModifiedAt:       docData.UpdatedAt,
		IsDelete:         docData.IsDelete,
	}
}

// TransReceivableOtherDecode decodes JSON Kafka message to ReceivableOtherKafkaDoc
func TransReceivableOtherDecode(jsonData string) ReceivableOtherKafkaDoc {
	var docData ReceivableOtherKafkaDoc
	DecodeKafkaMessage(jsonData, &docData, "ReceivableOther")
	return docData
}
