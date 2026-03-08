package stockbalance

import (
	"encoding/json"
	"errors"
	pkgModels "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/models"
	stockBalanceModels "smlcloudplatform/internal/transaction/stockbalance/models"
)

type StockBalanceTransactionPhaser struct{}

func (p StockBalanceTransactionPhaser) PhaseSingleDoc(input string) (*models.StockBalanceTransactionPG, error) {

	doc := stockBalanceModels.StockBalanceMessage{}
	err := json.Unmarshal([]byte(input), &doc)
	if err != nil {
		return nil, errors.New("Cannot Unmarshal StockBalance Doc Error: " + err.Error())
	}

	transaction, err := p.PhaseStockBalanceTransaction(doc)
	if err != nil {
		return nil, errors.New("Cannot Phase StockBalance Doc Error: " + err.Error())
	}
	return transaction, nil

}

func (p StockBalanceTransactionPhaser) PhaseMultipleDoc(input string) (*[]models.StockBalanceTransactionPG, error) {
	docs := []stockBalanceModels.StockBalanceMessage{}
	err := json.Unmarshal([]byte(input), &docs)
	if err != nil {
		return nil, errors.New("Cannot Unmarshal StockBalance Doc Error: " + err.Error())
	}

	transactions := make([]models.StockBalanceTransactionPG, len(docs))
	for i, doc := range docs {
		transaction, err := p.PhaseStockBalanceTransaction(doc)
		if err != nil {
			return nil, errors.New("Cannot Phase StockBalance Doc Error: " + err.Error())
		}
		transactions[i] = *transaction
	}
	return &transactions, nil
}

func (p StockBalanceTransactionPhaser) PhaseStockBalanceTransaction(doc stockBalanceModels.StockBalanceMessage) (*models.StockBalanceTransactionPG, error) {

	// return nil, errors.New("Not Implement Yet")

	var details []models.StockBalanceTransactionDetailPG

	if doc.Details != nil {

		details = make([]models.StockBalanceTransactionDetailPG, len(*doc.Details))
		// details := make([]models.StockBalanceTransactionDetailPG, len(*doc.Details))

		for i, detail := range *doc.Details {

			stockDetail := models.StockBalanceTransactionDetailPG{
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

			//
			// WhNames:             *(detail.WhNames),
			// LocationNames:       *(detail.LocationNames),
			// UnitNames:           *(detail.UnitNames),
			if detail.WhNames != nil {
				stockDetail.WhNames = *(detail.WhNames)
			}

			if detail.LocationNames != nil {
				stockDetail.LocationNames = *(detail.LocationNames)
			}

			if detail.UnitNames != nil {
				stockDetail.UnitNames = *(detail.UnitNames)
			}

			details[i] = stockDetail
		}
	}

	stockTransaction := models.StockBalanceTransactionPG{
		TransactionPG: models.TransactionPG{
			ShopIdentity: pkgModels.ShopIdentity{
				ShopID: doc.ShopID,
			},
			GuidFixed:      doc.GuidFixed,
			GuidRef:        doc.GuidRef,
			InquiryType:    doc.InquiryType,
			TransFlag:      54,
			DocNo:          doc.DocNo,
			DocDate:        doc.DocDatetime,
			BranchCode:     doc.Branch.Code,
			BranchNames:    *pkgModels.DefaultArrayNameX(doc.Branch.Names),
			TaxDocDate:     doc.TaxDocDate,
			TaxDocNo:       doc.TaxDocNo,
			VatType:        doc.VatType,
			VatRate:        doc.VatRate,
			DocRefType:     doc.DocRefType,
			DocRefNo:       doc.DocRefNo,
			DocRefDate:     doc.DocRefDate,
			Description:    doc.Description,
			TotalValue:     doc.TotalValue,
			DiscountWord:   doc.DiscountWord,
			TotalDiscount:  doc.TotalDiscount,
			TotalBeforeVat: doc.TotalBeforeVat,
			TotalVatValue:  doc.TotalVatValue,
			TotalAfterVat:  doc.TotalAfterVat,
			TotalExceptVat: doc.TotalExceptVat,
			TotalAmount:    doc.TotalAmount,
			IsCancel:       doc.IsCancel,
		},
		Items: &details,
	}
	return &stockTransaction, nil

}
