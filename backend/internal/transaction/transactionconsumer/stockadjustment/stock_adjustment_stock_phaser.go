package stockadjustment

import (
	pkgModels "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/models"
)

type StockAdjustmentStockPhaser struct{}

func (p StockAdjustmentStockPhaser) PhaseSingleDoc(doc models.StockAdjustmentTransactionPG) (*models.StockTransaction, error) {

	calcFlag := 1
	if doc.TransFlag == 68 {
		calcFlag = -1
	}
	details := make([]models.StockTransactionDetail, len(*doc.Items))

	for i, detail := range *doc.Items {
		stockDetail := models.StockTransactionDetail{
			CalcFlag:            int8(calcFlag),
			DocRef:              detail.DocRef,
			ShopID:              doc.ShopID,
			DocNo:               doc.DocNo,
			Barcode:             detail.Barcode,
			ItemType:            detail.ItemType,
			ItemGuid:            detail.ItemGuid,
			VatType:             detail.VatType,
			TaxType:             detail.TaxType,
			UnitCode:            detail.UnitCode,
			StandValue:          detail.StandValue,
			DivideValue:         detail.DivideValue,
			WhCode:              detail.WhCode,
			LocationCode:        detail.LocationCode,
			Qty:                 detail.Qty,
			Price:               detail.Price,
			PriceExcludeVat:     detail.PriceExcludeVat,
			TotalValueVat:       detail.TotalValueVat,
			SumAmount:           detail.SumAmount,
			SumAmountExcludeVat: detail.SumAmountExcludeVat,
			Discount:            detail.Discount,
			DiscountAmount:      detail.DiscountAmount,
			LineNumber:          int8(detail.LineNumber),
		}
		details[i] = stockDetail
	}

	stockTransaction := models.StockTransaction{
		ShopIdentity: pkgModels.ShopIdentity{
			ShopID: doc.ShopID,
		},
		GuidFixed:      doc.GuidFixed,
		GuidRef:        doc.GuidRef,
		DocRefType:     doc.DocRefType,
		DocRefNo:       doc.DocRefNo,
		DocRefDate:     doc.DocRefDate,
		VatType:        doc.VatType,
		TransFlag:      doc.TransFlag,
		InquiryType:    doc.InquiryType,
		DocNo:          doc.DocNo,
		DocDate:        doc.DocDate,
		Details:        &details,
		TotalValue:     doc.TotalValue,
		DiscountWord:   doc.DiscountWord,
		TotalDiscount:  doc.TotalDiscount,
		TotalBeforeVat: doc.TotalBeforeVat,
		TotalVatValue:  doc.TotalVatValue,
		TotalExceptVat: doc.TotalExceptVat,
		TotalAfterVat:  doc.TotalAfterVat,
		TotalAmount:    doc.TotalAmount,
		IsCancel:       doc.IsCancel,
	}
	return &stockTransaction, nil
}
