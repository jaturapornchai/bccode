package creditortransaction

import (
	"smlcloudplatform/internal/transaction/models"
	"smlcloudplatform/internal/transaction/transactionconsumer/repositories"
	"smlcloudplatform/pkg/microservice"
)

type ICreditorTransactionPGRepository interface {
	Get(holdingCode string, docNo string) (*models.CreditorTransactionPG, error)
	Create(doc models.CreditorTransactionPG) error
	Update(holdingCode string, docNo string, doc models.CreditorTransactionPG) error
	Delete(holdingCode string, docNo string, doc models.CreditorTransactionPG) error
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
