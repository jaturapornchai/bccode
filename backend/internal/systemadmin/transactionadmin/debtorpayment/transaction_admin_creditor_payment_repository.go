package debtorpayment

import (
	"context"
	debtorPaymentModels "smlcloudplatform/internal/transaction/paid/models"
	"smlcloudplatform/pkg/microservice"
)

type IDebtorPaymentTransactionAdminRepository interface {
	FindDebtorPaymentDocByShopID(ctx context.Context, shopID string) ([]debtorPaymentModels.PaidDoc, error)
}

type DebtorPaymentTransactionAdminRepository struct {
	pst microservice.IPersisterMongo
}

func NewCreditorPaymentTransactionAdminRepository(pst microservice.IPersisterMongo) IDebtorPaymentTransactionAdminRepository {
	return &DebtorPaymentTransactionAdminRepository{
		pst: pst,
	}
}

func (r DebtorPaymentTransactionAdminRepository) FindDebtorPaymentDocByShopID(ctx context.Context, shopID string) ([]debtorPaymentModels.PaidDoc, error) {

	docs := []debtorPaymentModels.PaidDoc{}

	err := r.pst.Find(ctx, &debtorPaymentModels.PaidDoc{}, map[string]interface{}{"shopid": shopID}, &docs)
	if err != nil {
		return nil, err
	}

	return docs, nil
}
