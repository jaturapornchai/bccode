package purchase

import (
	"errors"
	pkgModels "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/models"
)

type PurchaseTransactionStockPhaser struct{}

func (p PurchaseTransactionStockPhaser) PhaseSingleDoc(doc models.PurchaseTransactionPG) (*models.StockTransaction, error) {

	transaction, err := p.PhaseStockTransactionPurchaseDoc(doc)
	if err != nil {
		return nil, errors.New("Error on Convert PurchaseDoc to StockTransaction : " + err.Error())
	}
	return transaction, err
}

func (p *PurchaseTransactionStockPhaser) PhaseStockTransactionPurchaseDoc(doc models.PurchaseTransactionPG) (*models.StockTransaction, error) {
	details := make([]models.StockTransactionDetail, len(*doc.Items))

	for i, detail := range *doc.Items {
		stockDetail := models.StockTransactionDetail{
			DocRef:              detail.DocRef,
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
			CalcFlag:            1,
			LineNumber:          int8(detail.LineNumber),
		}
		details[i] = stockDetail
	}

	stockTransaction := models.StockTransaction{
		HoldingCodeentity: pkgModels.HoldingCodeentity{
			HoldingCode: doc.HoldingCode,
		},
		GuidFixed:      doc.GuidFixed,
		GuidRef:        doc.GuidRef,
		DocRefType:     doc.DocRefType,
		DocRefNo:       doc.DocRefNo,
		DocRefDate:     doc.DocRefDate,
		VatType:        doc.VatType,
		TransFlag:      12,
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
