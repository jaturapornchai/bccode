package purchase

import (
	"encoding/json"
	"errors"
	pkgModels "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/models"
	purchaseModels "smlcloudplatform/internal/transaction/purchase/models"
)

type PurchaseTransactionPhaser struct{}

func (p PurchaseTransactionPhaser) PhaseSingleDoc(msg string) (*models.PurchaseTransactionPG, error) {

	doc := purchaseModels.PurchaseDoc{}
	err := json.Unmarshal([]byte(msg), &doc)
	if err != nil {
		return nil, errors.New("Cannot Unmarshal PurchaseDoc Message : " + err.Error())
	}
	trx, err := p.PhaseStockTransactionPurchaseDoc(doc)
	if err != nil {
		return nil, errors.New("Error on Convert PurchaseDoc to StockTransaction : " + err.Error())
	}
	return trx, err
}

func (p *PurchaseTransactionPhaser) PhaseStockTransactionPurchaseDoc(doc purchaseModels.PurchaseDoc) (*models.PurchaseTransactionPG, error) {

	details := []models.PurchaseTransactionDetailPG{}

	if doc.PurchaseData.Purchase.Details != nil {
		details = make([]models.PurchaseTransactionDetailPG, len(*doc.PurchaseData.Purchase.Details))

		for i, detail := range *doc.PurchaseData.Purchase.Details {
			stockDetail := models.PurchaseTransactionDetailPG{
				TransactionDetailPG: models.TransactionDetailPG{
					GuidFixed:           doc.GuidFixed,
					DocRef:              detail.DocRef,
					DocRefDateTime:      detail.DocRefDatetime,
					DocNo:               doc.DocNo,
					ShopID:              doc.ShopID,
					LineNumber:          int8(detail.LineNumber),
					Barcode:             detail.Barcode,
					Qty:                 detail.Qty,
					Price:               detail.Price,
					PriceExcludeVat:     detail.PriceExcludeVat,
					Discount:            detail.Discount,
					DiscountAmount:      detail.DiscountAmount,
					SumAmount:           detail.SumAmount,
					SumAmountExcludeVat: detail.SumAmountExcludeVat,
					TotalValueVat:       detail.TotalValueVat,
					WhCode:              detail.WhCode,
					LocationCode:        detail.LocationCode,
					VatType:             detail.VatType,
					TaxType:             detail.TaxType,
					StandValue:          detail.StandValue,
					DivideValue:         detail.DivideValue,
					ItemType:            detail.ItemType,
					ItemGuid:            detail.ItemGuid,
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
				ManufacturerGUID:  detail.ManufacturerGUID,
				ManufacturerCode:  detail.ManufacturerCode,
				ManufacturerNames: *pkgModels.DefaultArrayNameX(detail.ManufacturerNames),
			}
			details[i] = stockDetail
		}
	}

	totalPayCreditAmount := float64(0)
	totalPayTransfer := float64(0)
	if doc.PaymentDetail.PaymentCreditCards != nil {

		for _, creditCard := range *doc.PaymentDetail.PaymentCreditCards {
			totalPayCreditAmount += creditCard.Amount
		}
	}
	if doc.PaymentDetail.PaymentTransfers != nil {

		for _, transfer := range *doc.PaymentDetail.PaymentTransfers {
			totalPayTransfer += transfer.Amount
		}
	}

	transaction := models.PurchaseTransactionPG{
		CreditorCode:  doc.CustCode,
		CreditorNames: *pkgModels.DefaultArrayNameX(doc.CustNames),
		TransactionPG: models.TransactionPG{
			GuidFixed: doc.GuidFixed,
			GuidRef:   doc.GuidRef,
			ShopIdentity: pkgModels.ShopIdentity{
				ShopID: doc.ShopID,
			},
			TransFlag:      12,
			DocNo:          doc.DocNo,
			DocDate:        doc.DocDatetime,
			BranchCode:     doc.Branch.Code,
			BranchNames:    *pkgModels.DefaultArrayNameX(doc.Branch.Names),
			TaxDocNo:       doc.TaxDocNo,
			TaxDocDate:     doc.TaxDocDate,
			Description:    doc.Description,
			InquiryType:    doc.InquiryType,
			VatType:        doc.VatType,
			VatRate:        doc.VatRate,
			DocRefType:     doc.DocRefType,
			DocRefNo:       doc.DocRefNo,
			DocRefDate:     doc.DocRefDate,
			TotalValue:     doc.TotalValue,
			DiscountWord:   doc.DiscountWord,
			TotalDiscount:  doc.TotalDiscount,
			TotalBeforeVat: doc.TotalBeforeVat,
			TotalVatValue:  doc.TotalVatValue,
			TotalExceptVat: doc.TotalExceptVat,
			TotalAfterVat:  doc.TotalAfterVat,
			TotalAmount:    doc.TotalAmount,
			IsCancel:       doc.IsCancel,
		},
		TotalPayCash:     doc.PaymentDetail.CashAmount,
		TotalPayCredit:   totalPayCreditAmount,
		TotalPayTransfer: totalPayTransfer,
		Items:            &details,
	}
	return &transaction, nil
}

func (p PurchaseTransactionPhaser) PhaseMultipleDoc(input string) (*[]models.PurchaseTransactionPG, error) {

	docs := []purchaseModels.PurchaseDoc{}
	err := json.Unmarshal([]byte(input), &docs)
	if err != nil {
		//t.ms.Logger.Errorf("Cannot Unmarshal PurchaseDoc Message : %v", err.Error())
		// fmt.Printf("Cannot Unmarshal PurchaseDoc Message : %v", err.Error())
		return nil, errors.New("Cannot Unmarshal PurchaseDoc Message : " + err.Error())
	}

	stockTransactions := make([]models.PurchaseTransactionPG, len(docs))

	for i, doc := range docs {
		trx, err := p.PhaseStockTransactionPurchaseDoc(doc)
		if err != nil {
			//t.ms.Logger.Errorf("Cannot Phase PurchaseDoc to StockTransaction : %v", err.Error())
			// fmt.Printf("Cannot Phase PurchaseDoc to StockTransaction : %v", err.Error())
			return nil, errors.New("Error on Convert PurchaseDoc to StockTransaction : " + err.Error())
		}
		stockTransactions[i] = *trx
	}

	return &stockTransactions, nil
}
