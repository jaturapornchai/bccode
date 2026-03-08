package purchase

import (
	"errors"
	"smlcloudplatform/internal/transaction/models"
	payment_models "smlcloudplatform/internal/transaction/payment/models"
)

type PaymentTransactionPhaser struct{}

func (p PaymentTransactionPhaser) PhaseSingleDoc(doc models.PurchaseTransactionPG) (*payment_models.TransactionPayment, error) {

	trx, err := p.PhaseStockTransactionPaymentDoc(doc)
	if err != nil {
		return nil, errors.New("Error on Convert PaymentDoc to StockTransaction : " + err.Error())
	}
	return trx, err
}

func (p *PaymentTransactionPhaser) PhaseStockTransactionPaymentDoc(doc models.PurchaseTransactionPG) (*payment_models.TransactionPayment, error) {

	transaction := payment_models.TransactionPayment{
		ShopID:    doc.ShopID,
		DocNo:     doc.DocNo,
		DocDate:   doc.DocDate,
		GuidRef:   doc.GuidRef,
		TransFlag: doc.TransFlag,
	}
	return &transaction, nil
}
