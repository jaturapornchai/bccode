package appurchasereceive

import (
	"errors"
	"smlcloudplatform/internal/transaction/models"

	pkgModels "smlcloudplatform/internal/models"
)

type APPurchaseReceiveCreditorTransactionPhaser struct{}

func (p APPurchaseReceiveCreditorTransactionPhaser) PhaseSingleDoc(doc models.APPurchaseReceivePG) (*models.CreditorTransactionPG, error) {

	txn, err := p.PhaseAPPurchaseReceiveCreditor(doc)
	if err != nil {
		return nil, errors.New("Error on Convert AccrualReceiveDoc to StockTransaction : " + err.Error())
	}
	return txn, err
}

func (p APPurchaseReceiveCreditorTransactionPhaser) PhaseAPPurchaseReceiveCreditor(doc models.APPurchaseReceivePG) (*models.CreditorTransactionPG, error) {
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
