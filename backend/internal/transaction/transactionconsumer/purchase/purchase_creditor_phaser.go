package purchase

import (
	"errors"
	pkgModels "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/models"
)

type PurchaseCreditorTransactionPhaser struct {
}

func (p PurchaseCreditorTransactionPhaser) PhaseSingleDoc(doc models.PurchaseTransactionPG) (*models.CreditorTransactionPG, error) {
	trx, err := p.PhasePurchaseCreditor(doc)
	if err != nil {
		return nil, errors.New("Error on Convert PurchaseDoc to StockTransaction : " + err.Error())
	}
	return trx, err
}

func (p PurchaseCreditorTransactionPhaser) PhasePurchaseCreditor(doc models.PurchaseTransactionPG) (*models.CreditorTransactionPG, error) {
	transaction := models.CreditorTransactionPG{
		ShopIdentity: pkgModels.ShopIdentity{
			ShopID: doc.ShopID,
		},
		GuidFixed:      doc.GuidFixed,
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
		PaidAmount:     0,
	}

	return &transaction, nil
}
