package repositories

import (
	"smlcloudplatform/internal/transaction/payment/models"
	"smlcloudplatform/internal/transaction/transactionconsumer/repositories"
	"smlcloudplatform/pkg/microservice"
)

type IPaymentRepository interface {
	Get(shopID string, docNo string) (*models.TransactionPayment, error)
	Create(doc models.TransactionPayment) error
	Update(shopID string, docNo string, doc models.TransactionPayment) error
	Delete(shopID string, docNo string, doc models.TransactionPayment) error
}

type PaymentRepository struct {
	pst microservice.IPersister
	repositories.ITransactionConsumerRepository[models.TransactionPayment]
}

func NewPaymentRepository(pst microservice.IPersister) *PaymentRepository {

	repo := &PaymentRepository{
		pst: pst,
	}

	repo.ITransactionConsumerRepository = repositories.NewTransactionConsumerRepository[models.TransactionPayment](pst)
	return repo
}

func (repo PaymentRepository) Create(doc models.TransactionPayment) error {
	err := repo.pst.Create(&doc)
	if err != nil {
		return err
	}
	return nil
}

func (repo PaymentRepository) Update(shopID string, docNo string, doc models.TransactionPayment) error {

	err := repo.pst.Update(&doc, map[string]interface{}{
		"shopid": shopID,
		"docno":  docNo,
	})

	if err != nil {
		return err
	}
	return nil
}
