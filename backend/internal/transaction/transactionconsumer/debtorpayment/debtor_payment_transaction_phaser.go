package debtorpayment

import (
	"encoding/json"
	"smlcloudplatform/internal/transaction/models"
	debtorpaymentmodels "smlcloudplatform/internal/transaction/paid/models"

	pkgModels "smlcloudplatform/internal/models"
)

type DebtorPaymentTransactionPhaser struct{}

func (p DebtorPaymentTransactionPhaser) PhaseSingleDoc(input string) (*models.DebtorPaymentTransactionPG, error) {

	doc := debtorpaymentmodels.PaidDoc{}
	err := json.Unmarshal([]byte(input), &doc)
	if err != nil {
		return nil, err
	}

	transaction, err := p.PhaseDebtorPaymentTransactionDoc(doc)
	if err != nil {
		return nil, err
	}

	return transaction, nil
}

func (p DebtorPaymentTransactionPhaser) PhaseMultipleDoc(input string) (*[]models.DebtorPaymentTransactionPG, error) {

	docs := []debtorpaymentmodels.PaidDoc{}
	err := json.Unmarshal([]byte(input), &docs)
	if err != nil {
		return nil, err
	}

	transactions := []models.DebtorPaymentTransactionPG{}
	for _, doc := range docs {
		transaction, err := p.PhaseDebtorPaymentTransactionDoc(doc)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, *transaction)
	}

	return &transactions, nil

}

func (p *DebtorPaymentTransactionPhaser) PhaseDebtorPaymentTransactionDoc(doc debtorpaymentmodels.PaidDoc) (*models.DebtorPaymentTransactionPG, error) {

	details := make([]models.DebtorPaymentTransactionDetailPG, len(*doc.Details))

	for i, detail := range *doc.Details {

		d := models.DebtorPaymentTransactionDetailPG{
			ShopID:        doc.ShopID,
			DocNo:         doc.DocNo,
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

	transaction := models.DebtorPaymentTransactionPG{
		ShopIdentity: pkgModels.ShopIdentity{
			ShopID: doc.ShopID,
		},
		GuidFixed:        doc.GuidFixed,
		DocNo:            doc.DocNo,
		DocDate:          doc.DocDatetime,
		BranchCode:       doc.Branch.Code,
		BranchNames:      pkgModels.JSONB(*doc.Branch.Names),
		DebtorCode:       doc.CustCode,
		DebtorNames:      *doc.CustNames,
		TotalAmount:      doc.TotalAmount,
		TotalPayCash:     doc.PaymentDetail.CashAmount,
		TotalPayCredit:   totalPayCreditAmount,
		TotalPayTransfer: totalPayTransfer,
		Details:          &details,
	}

	return &transaction, nil
}
