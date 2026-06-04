package debtorpayment

import (
	models "smlcloudplatform/internal/transaction/models"
	"smlcloudplatform/internal/transaction/transactionconsumer/repositories"
	"smlcloudplatform/pkg/microservice"
)

type IDebtorPaymentTransactionPGRepository interface {
	Get(holdingCode string, docNo string) (*models.DebtorPaymentTransactionPG, error)
	Create(doc models.DebtorPaymentTransactionPG) error
	Update(holdingCode string, docNo string, doc models.DebtorPaymentTransactionPG) error
	Delete(holdingCode string, docNo string, doc models.DebtorPaymentTransactionPG) error
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
