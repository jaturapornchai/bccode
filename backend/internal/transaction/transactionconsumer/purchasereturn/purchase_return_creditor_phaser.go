package purchasereturn

import (
	pkgModels "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/models"
)

// type IPurchaseReturnTransactionCreditorPhaser interface {
// 	HasStockEffectDoc(doc *purchasereturnmodels.PurchaseReturnDoc) bool
// 	HasCreditorEffectDoc(doc *purchasereturnmodels.PurchaseReturnDoc) bool
// }

type PurchaseReturnTransactionCreditorPhaser struct{}

func (p PurchaseReturnTransactionCreditorPhaser) PhaseSingleDoc(doc models.PurchaseReturnTransactionPG) (*models.CreditorTransactionPG, error) {

	transaction, err := p.PhasePurchaseReturnCreditor(doc)
	if err != nil {
		return nil, err
	}
	return transaction, nil
}

func (p PurchaseReturnTransactionCreditorPhaser) PhasePurchaseReturnCreditor(doc models.PurchaseReturnTransactionPG) (*models.CreditorTransactionPG, error) {

	transaction := models.CreditorTransactionPG{
		GuidFixed: doc.GuidFixed,
		ShopIdentity: pkgModels.ShopIdentity{
			ShopID: doc.ShopID,
		},
		DocNo:          doc.DocNo,
		DocDate:        doc.DocDate,
		CreditorCode:   doc.CreditorCode,
		CreditorNames:  doc.CreditorNames,
		InquiryType:    doc.InquiryType,
		TransFlag:      int(doc.TransFlag),
		TotalValue:     doc.TotalValue,
		TotalBeforeVat: doc.TotalBeforeVat,
		TotalAfterVat:  doc.TotalAfterVat,
		TotalVatValue:  doc.TotalVatValue,
		TotalExceptVat: doc.TotalExceptVat,
		TotalAmount:    doc.TotalAmount,
		BalanceAmount:  doc.TotalAmount,
	}
	return &transaction, nil
}
