package purchasereturn

import (
	pkgModels "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/models"
)

type PurchaseReturnTransactionStockPhaser struct{}

func (p PurchaseReturnTransactionStockPhaser) PhaseSingleDoc(doc models.PurchaseReturnTransactionPG) (*models.StockTransaction, error) {

	transaction, err := p.PhasePurchaseReturnStock(doc)
	if err != nil {
		return nil, err
	}

	return transaction, nil
}

func (p PurchaseReturnTransactionStockPhaser) PhasePurchaseReturnStock(doc models.PurchaseReturnTransactionPG) (*models.StockTransaction, error) {

	details := []models.StockTransactionDetail{}

	if doc.Items != nil {
		details = make([]models.StockTransactionDetail, len(*doc.Items))

		for i, detail := range *doc.Items {

			stockDetail := models.StockTransactionDetail{
				HoldingCode:         doc.HoldingCode,
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
				CalcFlag:            -1,
				LineNumber:          int8(detail.LineNumber),
				DocRef:              doc.DocRefNo,
			}
			details[i] = stockDetail
		}
	}

	stockTransaction := models.StockTransaction{
		GuidFixed: doc.GuidFixed,
		HoldingCodeentity: pkgModels.HoldingCodeentity{
			HoldingCode: doc.HoldingCode,
		},
		TransFlag:      16,
		InquiryType:    doc.InquiryType,
		GuidRef:        doc.GuidRef,
		DocNo:          doc.DocNo,
		DocDate:        doc.DocDate,
		DocRefType:     doc.DocRefType,
		DocRefNo:       doc.DocRefNo,
		DocRefDate:     doc.DocRefDate,
		VatType:        doc.VatType,
		VatRate:        doc.VatRate,
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
