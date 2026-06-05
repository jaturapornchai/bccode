package myglobal

import (
	"encoding/json"
	"fmt"
	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/models"
	"strings"
)

// ==================== DocDetail Functions (รวมจาก build-doc-detail.go) ====================

func MapDocDetailFromMongo(docData models.MongoDocModel, detail any, transFlag int, calcFlag float64, calcSeq int, barcodeMain string, unitStand, unitDivide float64, whCode, locationCode string, totalQty float64, lineNumber int) models.DocDetailStruct {
	// Type assertion for detail
	var detailStruct struct {
		LineNumber             int
		ItemCode               string
		ItemName               string
		Barcode                string
		UnitCode               string
		Price                  float64
		PriceExcludeVat        float64
		DocRef                 string
		SumAmount              float64
		PriceDoc               float64
		SumAmountDoc           float64
		DiscountAmountDoc      float64
		PriceExcludeVatDoc     float64
		SumAmountExcludeVatDoc float64
		TotalValueVatDoc       float64
	}

	// ดึงข้อมูลจาก detail struct
	if d, ok := detail.(models.MongoDocDetailModel); ok {
		var itemCode = strings.TrimSpace(d.ItemCode)
		if itemCode == "" {
			itemCode = d.Barcode
			if itemCode == "" {
				itemCode = "UNKNOWN"
			}
		}
		var itemName string = ""
		if len(d.ItemNames) > 0 {
			itemName = d.ItemNames[0].Name
		} else {
			itemName = "UNKNOWN"
		}
		detailStruct.ItemCode = itemCode
		detailStruct.LineNumber = lineNumber
		detailStruct.ItemName = itemName
		detailStruct.Barcode = d.Barcode
		detailStruct.UnitCode = d.UnitCode
		detailStruct.Price = d.Price
		detailStruct.PriceExcludeVat = d.PriceExcludeVat
		detailStruct.DocRef = d.DocRef
		detailStruct.SumAmount = d.SumAmount
		detailStruct.PriceDoc = d.PriceDoc
		detailStruct.SumAmountDoc = d.SumAmountDoc
		detailStruct.DiscountAmountDoc = d.DiscountAmountDoc
		detailStruct.PriceExcludeVatDoc = d.PriceExcludeVatDoc
		detailStruct.SumAmountExcludeVatDoc = d.SumAmountExcludeVatDoc
		detailStruct.TotalValueVatDoc = d.TotalValueVatDoc
	} else {
		logger.Error("Type assertion failed for detail in docNo %s: type is %T", docData.DocNo, detail)
		logger.Fatal("Type assertion failed for detail: %v", detail)
	}

	if whCode == "" {
		whCode = "X"
	}
	if locationCode == "" {
		locationCode = "X"
	}

	return models.DocDetailStruct{
		DocDateTime:            docData.DocDateTime,
		DocNo:                  docData.DocNo,
		LineNumber:             detailStruct.LineNumber,
		TransFlag:              transFlag,
		CalcFlag:               calcFlag,
		CalcSeq:                calcSeq,
		ItemCode:               detailStruct.ItemCode,
		Description:            detailStruct.ItemName,
		BarcodeMain:            barcodeMain,
		Barcode:                detailStruct.Barcode,
		UnitCode:               detailStruct.UnitCode,
		WhCode:                 whCode,
		LocationCode:           locationCode,
		TotalQty:               totalQty,
		Price:                  detailStruct.Price,
		PriceExcludeVat:        detailStruct.PriceExcludeVat,
		UnitStand:              unitStand,
		UnitDivide:             unitDivide,
		DocRef:                 detailStruct.DocRef,
		SumAmount:              detailStruct.SumAmount,
		PriceDoc:               detailStruct.PriceDoc,
		SumAmountDoc:           detailStruct.SumAmountDoc,
		DiscountAmountDoc:      detailStruct.DiscountAmountDoc,
		PriceExcludeVatDoc:     detailStruct.PriceExcludeVatDoc,
		SumAmountExcludeVatDoc: detailStruct.SumAmountExcludeVatDoc,
		TotalValueVatDoc:       detailStruct.TotalValueVatDoc,
	}
}

func MapDocStructFromMongo(docData models.MongoDocModel, holdingCode string) (models.DocStruct, models.DocPaymentStruct) {
	// คำนวณ checksum สำหรับข้อมูลเอกสาร - ใช้ MD5 เพื่อให้ได้ 32 ตัวอักษร
	checksumData := fmt.Sprintf("%v", docData)
	checksum := CalculateMD5(checksumData)

	// กำหนดค่า creator จาก MongoDocModel
	creatorCode := docData.CreatorCode
	creatorName := docData.CreatorName
	createdAt := docData.CreatedAt
	if creatorCode == "" {
		creatorCode = "system"
	}
	if creatorName == "" {
		creatorName = "System"
	}
	// ถ้าไม่มี createdAt ให้ใช้ DocDateTime แทน
	if createdAt.IsZero() {
		createdAt = docData.DocDateTime
	}

	// ตรวจสอบ IsDelete จาก field isdelete หรือ deletedat (soft delete)
	isDelete := docData.IsDelete || docData.DeletedAt != nil

	s := models.DocStruct{
		HoldingCode:     holdingCode,
		TransFlag:       docData.TransFlag,
		DocNo:           docData.DocNo,
		DocDateTime:     docData.DocDateTime,
		CustCode:        docData.CustCode,
		PeriodDateTime:  docData.DocDateTime,
		TaxDocNo:        "",
		TotalAmount:     docData.TotalAmount,
		RoundAmount:     docData.RoundAmount,
		PayType:         1,
		PayCashAmount:   docData.PayCashAmount,
		PayCashChange:   docData.PayCashChange,
		PayCashBalance:  0,
		DeliveryCode:    "",
		Checksum:        checksum,
		BranchID:        docData.Branch.GuidFixed,
		SlipURL:         docData.SlipUrl,
		SaleChannelCode: docData.SaleChannelCode,
		DeliveryAmount:  docData.DeliveryAmount,
		IsCancel:        docData.IsCancel,
		CancelReason:    docData.CancelReason,
		GuidPOS:         docData.GuidPos,
		GuidBranch:      docData.Branch.GuidFixed,
		GuidFixed:       docData.GuidFixed,
		// ข้อมูลผู้สร้าง
		CreatorCode: creatorCode,
		CreatorName: creatorName,
		CreatedAt:   createdAt,

		// ============ Multi-Currency Fields ============
		// Currency (field เดิม) = สกุลเงินหลัก (Base)
		Currency:       docData.Currency,
		CurrencySymbol: docData.CurrencySymbol,
		// DocCurrency (field ใหม่) = สกุลเงินเอกสาร
		DocCurrency:       docData.DocCurrency,
		DocCurrencySymbol: docData.DocCurrencySymbol,
		ExchangeRate:      docData.ExchangeRate,
		TotalAmountDoc:    docData.TotalAmountDoc,
		// Soft Delete
		IsDelete: isDelete,
	}

	// docpayment
	p := models.DocPaymentStruct{}
	jsonPaymentRaw := strings.Replace(docData.PaymentDetailRaw, "'", "''", -1)

	if jsonPaymentRaw != "" {
		var payments []map[string]interface{}
		if err := json.Unmarshal([]byte(jsonPaymentRaw), &payments); err != nil {
			// logger.Warn("Failed to unmarshal payment detail for DocNo %s: %v", docData.DocNo, err)
		} else {
			for _, payment := range payments {
				amount, _ := payment["amount"].(float64)
				providerName, _ := payment["providername"].(string)
				transflag, _ := payment["transflag"].(float64)
				p = models.DocPaymentStruct{
					HoldingCode:    holdingCode,
					BranchID:       docData.Branch.GuidFixed,
					DocDateTime:    docData.DocDateTime,
					PeriodDateTime: docData.DocDateTime,
					ProviderName:   providerName,
					Amount:         amount,
					Description:    providerName,
					DocNo:          docData.DocNo,
					TransFlag:      int32(transflag),
					GuidFixed:      docData.GuidFixed,
					GuidBranch:     docData.Branch.GuidFixed,
				}
			}
		}
	}

	return s, p
}

func MapDocRefStruct(docNo string, transFlag int, docRef models.MongoDocReferenceModel) models.DocRefStruct {
	return models.DocRefStruct{
		DocNo:             docNo,
		DocNoTransFlag:    transFlag,
		DocRefNo:          docRef.DocNo,
		DocRefNoTransFlag: 0, // TODO: ต้องกำหนดค่าจริง
	}
}
