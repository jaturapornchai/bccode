package creditorpayment

import (
	"encoding/json"
	models "smlcloudplatform/internal/transaction/models"

	pkgModels "smlcloudplatform/internal/models"
	creditorpaymentmodels "smlcloudplatform/internal/transaction/pay/models"
)

type CreditorPaymentTransactionPhaser struct{}

// implement all method in ITransactionPhaser
func (c CreditorPaymentTransactionPhaser) PhaseSingleDoc(input string) (*models.CreditorPaymentTransactionPG, error) {

	doc := creditorpaymentmodels.PayDoc{}
	err := json.Unmarshal([]byte(input), &doc)
	if err != nil {
		return nil, err
	}

	transaction, err := c.PhaseCreditPaymentTransactionDoc(doc)
	if err != nil {
		return nil, err
	}

	return transaction, nil
}

func (c CreditorPaymentTransactionPhaser) PhaseMultipleDoc(input string) (*[]models.CreditorPaymentTransactionPG, error) {

	docs := []creditorpaymentmodels.PayDoc{}
	err := json.Unmarshal([]byte(input), &docs)
	if err != nil {
		return nil, err
	}

	transactions := []models.CreditorPaymentTransactionPG{}
	for _, doc := range docs {
		transaction, err := c.PhaseCreditPaymentTransactionDoc(doc)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, *transaction)
	}

	return &transactions, nil
}

func (c *CreditorPaymentTransactionPhaser) PhaseCreditPaymentTransactionDoc(doc creditorpaymentmodels.PayDoc) (*models.CreditorPaymentTransactionPG, error) {

	details := make([]models.CreditorPaymentTransactionDetailPG, len(*doc.Details))

	for i, detail := range *doc.Details {
		d := models.CreditorPaymentTransactionDetailPG{
			DocNo:         doc.DocNo,
			ShopID:        doc.ShopID,
			LineNumber:    int8(i),
			BillingNo:     detail.DocNo,
			BillType:      detail.TransFlag,
			BillAmount:    detail.Value,
			BalanceAmount: detail.Balance,
			PayAmount:     detail.PaymentAmount,
		}

		details[i] = d
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

	if doc.Branch.Names == nil {
		doc.Branch.Names = &[]pkgModels.NameX{}
	}

	transaction := models.CreditorPaymentTransactionPG{
		ShopIdentity: pkgModels.ShopIdentity{
			ShopID: doc.ShopID,
		},
		GuidFixed:        doc.GuidFixed,
		DocNo:            doc.DocNo,
		DocDate:          doc.DocDatetime,
		BranchCode:       doc.Branch.Code,
		BranchNames:      pkgModels.JSONB(*doc.Branch.Names),
		CreditorCode:     doc.CustCode,
		CreditorNames:    *doc.CustNames,
		Details:          &details,
		TotalAmount:      doc.TotalAmount,
		TotalPayCash:     doc.PaymentDetail.CashAmount,
		TotalPayTransfer: totalPayTransfer,
		TotalPayCredit:   totalPayCreditAmount,
	}

	return &transaction, nil
}
