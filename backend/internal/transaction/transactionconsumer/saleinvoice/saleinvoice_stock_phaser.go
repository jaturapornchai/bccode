package saleinvoice

import (
	"errors"
	pkgModels "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/models"
)

type SaleInvoiceTransactionStockPhaser struct{}

func (p SaleInvoiceTransactionStockPhaser) PhaseSingleDoc(doc models.SaleInvoiceTransactionPG) (*models.StockTransaction, error) {

	transaction, err := p.PhaseSaleInvoiceDoc(doc)
	if err != nil {
		return nil, errors.New("Error on Convert PurchaseDoc to StockTransaction : " + err.Error())
	}
	return transaction, err
}

func (p SaleInvoiceTransactionStockPhaser) PhaseSaleInvoiceDoc(doc models.SaleInvoiceTransactionPG) (*models.StockTransaction, error) {

	details := []models.StockTransactionDetail{}

	if doc.Items != nil {
		details = make([]models.StockTransactionDetail, len(*doc.Items))

		for i, detail := range *doc.Items {
			stockDetail := models.StockTransactionDetail{
				CalcFlag:            -1,
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
				LineNumber:          int8(detail.LineNumber),
			}
			details[i] = stockDetail
		}
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
		TransFlag:      44,
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
