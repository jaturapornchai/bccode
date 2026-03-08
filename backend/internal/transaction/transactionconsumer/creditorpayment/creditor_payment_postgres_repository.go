package creditorpayment

import (
	models "smlcloudplatform/internal/transaction/models"
	"smlcloudplatform/internal/transaction/transactionconsumer/repositories"
	"smlcloudplatform/pkg/microservice"
)

type ICreditorPaymentTransactionPGRepository interface {
	Get(shopID string, docNo string) (*models.CreditorPaymentTransactionPG, error)
	Create(doc models.CreditorPaymentTransactionPG) error
	Update(shopID string, docNo string, doc models.CreditorPaymentTransactionPG) error
	Delete(shopID string, docNo string, doc models.CreditorPaymentTransactionPG) error
	MigrationDatabase() error
}

type CreditorPaymentTransactionPGRepository struct {
	pst microservice.IPersister
	repositories.ITransactionConsumerRepository[models.CreditorPaymentTransactionPG]
}

func NewCreditorPaymentTransactionPGRepository(pst microservice.IPersister) ICreditorPaymentTransactionPGRepository {

	repo := &CreditorPaymentTransactionPGRepository{
		pst: pst,
	}

	repo.ITransactionConsumerRepository = repositories.NewTransactionConsumerRepository[models.CreditorPaymentTransactionPG](pst)

	return repo
}

func (r *CreditorPaymentTransactionPGRepository) MigrationDatabase() error {

	err := r.pst.AutoMigrate(
		models.CreditorPaymentTransactionPG{},
		models.CreditorPaymentTransactionDetailPG{},
	)

	return err
}
