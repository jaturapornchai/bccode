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

// PayKafkaDoc represents incoming Kafka payload for creditor payment
type PayKafkaDoc struct {
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
	PayCashChange      float64                 `json:"paycashchange"`
	PaymentDetailRaw   string                  `json:"paymentdetailraw"`
	RoundAmount        float64                 `json:"roundamount"`
	IsCancel           bool                    `json:"iscancel"`
	Branch             models.MongoBranchModel `json:"branch"`
	IsDelete           bool                    `json:"isdelete"`
	CreatedAt          time.Time               `json:"createdat"`
	CreatedBy          string                  `json:"createdby"`
	UpdatedAt          time.Time               `json:"updatedat"`
	UpdatedBy          string                  `json:"updatedby"`
}

// OnConsumeMessagePayCreateOrUpdate handles creditor payment create/update messages
func OnConsumeMessagePayCreateOrUpdate(msg string) error {
	return SafeConsumerWrapper("CREDITOR_PAYMENT", func(msg string) error {
		return ProcessPayDocument(msg)
	})(msg)
}

// OnConsumeMessagePayDelete handles creditor payment deletion messages
func OnConsumeMessagePayDelete(msg string) error {
	logger.Info("OnConsumeMessagePayDelete: %s", msg)

	docData := TransPayDecode(msg)
	if docData.HoldingCode == "" || docData.DocNo == "" {
		logger.Error("Invalid pay data - missing HoldingCode or DocNo")
		return fmt.Errorf("invalid pay data")
	}

	transFlag := docData.TransFlag
	if transFlag == 0 {
		transFlag = TRANS_FLAG_CREDITOR_PAYMENT
	}

	return DeleteDocumentFromDatabases(context.Background(), docData.HoldingCode, docData.BusinessCode, docData.DocNo, transFlag)
}

// ProcessPayDocument processes creditor payment document into PostgreSQL
func ProcessPayDocument(msg string) error {
	logger.Info("--- ProcessPayDocument START ---")

	defer func() {
		if r := recover(); r != nil {
			logger.Error("PANIC recovered in ProcessPayDocument: %v", r)
			logger.Error("Stack trace: %s", debug.Stack())
		}
	}()

	docData := TransPayDecode(msg)
	if docData.HoldingCode == "" || docData.DocNo == "" {
		return fmt.Errorf("invalid pay data - missing HoldingCode or DocNo")
	}

	build.DatabaseChecker(docData.HoldingCode, false)

	transFlag := docData.TransFlag
	if transFlag == 0 {
		transFlag = TRANS_FLAG_CREDITOR_PAYMENT
	}

	processData := ConvertPayDocToProcessModel(docData)

	db, err := mypg.PgSqlFastConnect(docData.HoldingCode)
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	ctx := context.Background()

	// Delete existing documents (upsert behavior)
	mypg.DeleteDocPgSql(ctx, db, docData.BusinessCode, docData.DocNo, transFlag)

	docStruct, docPaymentStruct := myglobal.MapDocStructFromMongo(processData, docData.HoldingCode)
	setDocumentCompany(&docStruct, nil, docData.BusinessCode)

	err = InsertDocumentToPostgreSQL(ctx, db, docStruct, nil, docPaymentStruct, 1)
	if err != nil {
		return fmt.Errorf("failed to insert pay document to PostgreSQL: %w", err)
	}

	err = InsertDocumentToClickHouse(ctx, docData.HoldingCode, docStruct, nil, docPaymentStruct, nil, 2)
	if err != nil {
		return fmt.Errorf("failed to insert to ClickHouse: %w", err)
	}

	mypg.AddToDocWaitProcessQueues(ctx, db, docData.DocNo, transFlag)

	logger.Success("ProcessPayDocument: Successfully processed DocNo=%s, TransFlag=%d", docData.DocNo, transFlag)
	return nil
}

// ConvertPayDocToProcessModel converts PayKafkaDoc to ProcessMongoTransModel
func ConvertPayDocToProcessModel(docData PayKafkaDoc) models.ProcessMongoTransModel {
	transFlag := docData.TransFlag
	if transFlag == 0 {
		transFlag = TRANS_FLAG_CREDITOR_PAYMENT
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
		PayCashChange:    docData.PayCashChange,
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

// TransPayDecode decodes JSON Kafka message to PayKafkaDoc
func TransPayDecode(jsonData string) PayKafkaDoc {
	var docData PayKafkaDoc
	DecodeKafkaMessage(jsonData, &docData, "Pay")
	return docData
}
