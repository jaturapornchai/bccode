package creditortransaction

import (
	"smlcloudplatform/internal/transaction/models"
	"smlcloudplatform/internal/transaction/transactionconsumer/repositories"
	"smlcloudplatform/pkg/microservice"
)

type ICreditorTransactionPGRepository interface {
	Get(shopID string, docNo string) (*models.CreditorTransactionPG, error)
	Create(doc models.CreditorTransactionPG) error
	Update(shopID string, docNo string, doc models.CreditorTransactionPG) error
	Delete(shopID string, docNo string, doc models.CreditorTransactionPG) error
}

type CreditorTransactionPGRepository struct {
	pst microservice.IPersister
	repositories.ITransactionConsumerRepository[models.CreditorTransactionPG]
}

func NewCreditorTransactionPGRepository(pst microservice.IPersister) ICreditorTransactionPGRepository {

	repo := &CreditorTransactionPGRepository{
		pst: pst,
	}

	repo.ITransactionConsumerRepository = repositories.NewTransactionConsumerRepository[models.CreditorTransactionPG](pst)
	return repo
}
