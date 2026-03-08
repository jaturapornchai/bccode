package stockadjustment

import (
	"encoding/json"
	"errors"
	pkgModels "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/models"
	stockAdjustmentModels "smlcloudplatform/internal/transaction/stockadjustment/models"
)

type StockAdjustmentTransactionPhaser struct{}

func (p StockAdjustmentTransactionPhaser) PhaseSingleDoc(input string) (*models.StockAdjustmentTransactionPG, error) {
	doc := stockAdjustmentModels.StockAdjustmentDoc{}
	err := json.Unmarshal([]byte(input), &doc)
	if err != nil {
		return nil, errors.New("Cannot Unmarshal StockAdjustment Doc Error: " + err.Error())
	}

	transaction, err := p.PhaseStockAdjustmentTransactionPGDoc(doc)
	if err != nil {
		return nil, errors.New("Cannot Phase StockAdjustment Doc Error: " + err.Error())
	}
	return transaction, nil
}

func (p StockAdjustmentTransactionPhaser) PhaseMultipleDoc(input string) (*[]models.StockAdjustmentTransactionPG, error) {

	docs := []stockAdjustmentModels.StockAdjustmentDoc{}
	err := json.Unmarshal([]byte(input), &docs)
	if err != nil {
		return nil, errors.New("Cannot Unmarshal StockAdjustment Doc Error: " + err.Error())
	}

	transactions := make([]models.StockAdjustmentTransactionPG, len(docs))
	for i, doc := range docs {
		transaction, err := p.PhaseStockAdjustmentTransactionPGDoc(doc)
		if err != nil {
			return nil, errors.New("Cannot Phase StockAdjustment Doc Error: " + err.Error())
		}
		transactions[i] = *transaction
	}
	return &transactions, nil
}

func (p StockAdjustmentTransactionPhaser) PhaseStockAdjustmentTransactionPGDoc(doc stockAdjustmentModels.StockAdjustmentDoc) (*models.StockAdjustmentTransactionPG, error) {
	details := make([]models.StockAdjustmentTransactionDetailPG, len(*doc.Transaction.Details))

	for i, detail := range *doc.Transaction.Details {
		stockDetail := models.StockAdjustmentTransactionDetailPG{
			TransactionDetailPG: models.TransactionDetailPG{
				GuidFixed:           doc.GuidFixed,
				DocNo:               doc.DocNo,
				ShopID:              doc.ShopID,
				LineNumber:          int8(detail.LineNumber),
				DocRef:              detail.DocRef,
				DocRefDateTime:      detail.DocRefDatetime,
				Barcode:             detail.Barcode,
				Qty:                 detail.Qty,
				Price:               detail.Price,
				PriceExcludeVat:     detail.PriceExcludeVat,
				Discount:            detail.Discount,
				DiscountAmount:      detail.DiscountAmount,
				SumAmount:           detail.SumAmount,
				SumAmountExcludeVat: detail.SumAmountExcludeVat,
				WhCode:              detail.WhCode,
				LocationCode:        detail.LocationCode,
				VatType:             detail.VatType,
				TaxType:             detail.TaxType,
				StandValue:          detail.StandValue,
				DivideValue:         detail.DivideValue,
				ItemType:            detail.ItemType,
				ItemGuid:            detail.ItemGuid,
				TotalValueVat:       detail.TotalValueVat,
				Remark:              detail.Remark,
				UnitCode:            detail.UnitCode,
				UnitNames:           *pkgModels.DefaultArrayNameX(detail.UnitNames),
				ItemNames:           *pkgModels.DefaultArrayNameX(detail.ItemNames),
				WhNames:             *pkgModels.DefaultArrayNameX(detail.WhNames),
				LocationNames:       *pkgModels.DefaultArrayNameX(detail.LocationNames),
				GroupCode:           detail.GroupCode,
				GroupNames:          *pkgModels.DefaultArrayNameX(detail.GroupNames),
				DocDate:             detail.DocDatetime,
			},
		}
		details[i] = stockDetail
	}

	transaction := models.StockAdjustmentTransactionPG{
		TransactionPG: models.TransactionPG{
			ShopIdentity: pkgModels.ShopIdentity{
				ShopID: doc.ShopID,
			},
			GuidFixed:      doc.GuidFixed,
			GuidRef:        doc.GuidRef,
			InquiryType:    doc.InquiryType,
			TransFlag:      int16(doc.TransFlag),
			DocNo:          doc.DocNo,
			DocDate:        doc.DocDatetime,
			BranchCode:     doc.Branch.Code,
			BranchNames:    *pkgModels.DefaultArrayNameX(doc.Branch.Names),
			TaxDocDate:     doc.Transaction.TaxDocDate,
			TaxDocNo:       doc.Transaction.TaxDocNo,
			VatType:        doc.Transaction.VatType,
			VatRate:        doc.Transaction.VatRate,
			DocRefType:     doc.Transaction.DocRefType,
			DocRefNo:       doc.Transaction.DocRefNo,
			DocRefDate:     doc.Transaction.DocRefDate,
			Description:    doc.Transaction.Description,
			TotalValue:     doc.Transaction.TotalValue,
			DiscountWord:   doc.Transaction.DiscountWord,
			TotalDiscount:  doc.Transaction.TotalDiscount,
			TotalBeforeVat: doc.Transaction.TotalBeforeVat,
			TotalVatValue:  doc.Transaction.TotalVatValue,
			TotalAfterVat:  doc.Transaction.TotalAfterVat,
			TotalExceptVat: doc.Transaction.TotalExceptVat,
			TotalAmount:    doc.Transaction.TotalAmount,
			IsCancel:       doc.IsCancel,
		},
		Items: &details,
	}
	return &transaction, nil
}
