package debtorpayment

import (
	models "smlcloudplatform/internal/transaction/models"
	"smlcloudplatform/internal/transaction/transactionconsumer/repositories"
	"smlcloudplatform/pkg/microservice"
)

type IDebtorPaymentTransactionPGRepository interface {
	Get(shopID string, docNo string) (*models.DebtorPaymentTransactionPG, error)
	Create(doc models.DebtorPaymentTransactionPG) error
	Update(shopID string, docNo string, doc models.DebtorPaymentTransactionPG) error
	Delete(shopID string, docNo string, doc models.DebtorPaymentTransactionPG) error
}

type DebtorPaymentTransactionPGRepository struct {
	pst microservice.IPersister
	repositories.ITransactionConsumerRepository[models.DebtorPaymentTransactionPG]
}

func NewDebtorPaymentTransactionPGRepository(pst microservice.IPersister) IDebtorPaymentTransactionPGRepository {

	repo := &DebtorPaymentTransactionPGRepository{
		pst: pst,
	}

	repo.ITransactionConsumerRepository = repositories.NewTransactionConsumerRepository[models.DebtorPaymentTransactionPG](pst)

	return repo
}
