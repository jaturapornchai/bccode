package purchasereturn

import (
	"encoding/json"
	pkgModels "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/models"
	purchasereturnmodels "smlcloudplatform/internal/transaction/purchasereturn/models"
)

// type IPurchaseReturnTransactionPhaser interface {
// 	PhaseSingleDoc(input string) (*models.PurchaseReturnTransactionPG, error)
// 	PhaseMultipleDoc(input string) (*[]models.PurchaseReturnTransactionPG, error)
// 	HasStockEffectDoc(doc *models.PurchaseReturnTransactionPG) bool
// 	HasCreditorEffectDoc(doc *models.PurchaseReturnTransactionPG) bool
// }

type PurchaseReturnTransactionPhaser struct{}

// implement ITransactionPhaser

func (p PurchaseReturnTransactionPhaser) PhaseSingleDoc(input string) (*models.PurchaseReturnTransactionPG, error) {

	doc := purchasereturnmodels.PurchaseReturnDoc{}
	err := json.Unmarshal([]byte(input), &doc)
	if err != nil {
		return nil, err
	}

	transaction, err := p.PhasePurchaseReturnTransaction(&doc)
	if err != nil {
		return nil, err
	}

	return transaction, nil
}

func (p PurchaseReturnTransactionPhaser) PhaseMultipleDoc(input string) (*[]models.PurchaseReturnTransactionPG, error) {

	docs := []purchasereturnmodels.PurchaseReturnDoc{}
	err := json.Unmarshal([]byte(input), &docs)
	if err != nil {
		return nil, err
	}

	transactions := make([]models.PurchaseReturnTransactionPG, len(docs))

	for i, doc := range docs {
		transaction, err := p.PhasePurchaseReturnTransaction(&doc)
		if err != nil {
			return nil, err
		}

		transactions[i] = *transaction
	}

	return &transactions, nil
}

func (p PurchaseReturnTransactionPhaser) PhasePurchaseReturnTransaction(doc *purchasereturnmodels.PurchaseReturnDoc) (*models.PurchaseReturnTransactionPG, error) {

	details := make([]models.PurchaseReturnTransactionDetailPG, len(*doc.Details))

	for i, detail := range *doc.Details {

		transactionDetail := models.PurchaseReturnTransactionDetailPG{
			TransactionDetailPG: models.TransactionDetailPG{
				GuidFixed:           doc.GuidFixed,
				DocNo:               doc.DocNo,
				ShopID:              doc.ShopID,
				LineNumber:          int8(detail.LineNumber),
				ItemGuid:            detail.ItemGuid,
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
				Remark:              detail.Remark,
				DocRef:              detail.DocRef,
				DocRefDateTime:      detail.DocRefDatetime,
				VatCal:              int8(detail.VatCal),
				ItemType:            detail.ItemType,
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

		details[i] = transactionDetail
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

	transaction := models.PurchaseReturnTransactionPG{
		CreditorCode:  doc.CustCode,
		CreditorNames: *pkgModels.DefaultArrayNameX(doc.CustNames),
		TransactionPG: models.TransactionPG{
			GuidFixed: doc.GuidFixed,
			ShopIdentity: pkgModels.ShopIdentity{
				ShopID: doc.ShopID,
			},
			TransFlag:      16,
			DocNo:          doc.DocNo,
			DocDate:        doc.DocDatetime,
			GuidRef:        doc.GuidRef,
			DocRefType:     doc.DocRefType,
			DocRefNo:       doc.Transaction.DocRefNo,
			DocRefDate:     doc.Transaction.DocRefDate,
			BranchCode:     doc.Branch.Code,
			BranchNames:    *pkgModels.DefaultArrayNameX(doc.Branch.Names),
			TaxDocDate:     doc.Transaction.TaxDocDate,
			TaxDocNo:       doc.Transaction.TaxDocNo,
			Description:    doc.Description,
			InquiryType:    doc.InquiryType,
			VatType:        doc.Transaction.VatType,
			VatRate:        doc.Transaction.VatRate,
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
