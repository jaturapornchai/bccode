package repositories

import (
	"smlcloudplatform/internal/transaction/paymentdetail/models"
	"smlcloudplatform/internal/transaction/transactionconsumer/repositories"
	"smlcloudplatform/pkg/microservice"
)

type IPaymentDetailRepository interface {
	Get(shopID string, docNo string) (*models.TransactionPaymentDetail, error)
	Create(doc models.TransactionPaymentDetail) error
	Update(shopID string, docNo string, doc models.TransactionPaymentDetail) error
	Delete(shopID string, docNo string, doc models.TransactionPaymentDetail) error
}

type PaymentDetailRepository struct {
	pst microservice.IPersister
	repositories.ITransactionConsumerRepository[models.TransactionPaymentDetail]
}

func NewPaymentDetailRepository(pst microservice.IPersister) *PaymentDetailRepository {

	repo := &PaymentDetailRepository{
		pst: pst,
	}

	repo.ITransactionConsumerRepository = repositories.NewTransactionConsumerRepository[models.TransactionPaymentDetail](pst)
	return repo
}

func (repo PaymentDetailRepository) Create(doc models.TransactionPaymentDetail) error {
	err := repo.pst.Create(&doc)
	if err != nil {
		return err
	}
	return nil
}
