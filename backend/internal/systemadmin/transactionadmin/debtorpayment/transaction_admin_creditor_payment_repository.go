package debtorpayment

import (
	"context"
	debtorPaymentModels "smlcloudplatform/internal/transaction/paid/models"
	"smlcloudplatform/pkg/microservice"
)

type IDebtorPaymentTransactionAdminRepository interface {
	FindDebtorPaymentDocByHoldingCode(ctx context.Context, holdingCode string) ([]debtorPaymentModels.PaidDoc, error)
}

type DebtorPaymentTransactionAdminRepository struct {
	pst microservice.IPersisterMongo
}

func NewCreditorPaymentTransactionAdminRepository(pst microservice.IPersisterMongo) IDebtorPaymentTransactionAdminRepository {
	return &DebtorPaymentTransactionAdminRepository{
		pst: pst,
	}
}

func (r DebtorPaymentTransactionAdminRepository) FindDebtorPaymentDocByHoldingCode(ctx context.Context, holdingCode string) ([]debtorPaymentModels.PaidDoc, error) {

	docs := []debtorPaymentModels.PaidDoc{}

	err := r.pst.Find(ctx, &debtorPaymentModels.PaidDoc{}, map[string]interface{}{"holdingcode": holdingCode}, &docs)
	if err != nil {
		return nil, err
	}

	return docs, nil
}
