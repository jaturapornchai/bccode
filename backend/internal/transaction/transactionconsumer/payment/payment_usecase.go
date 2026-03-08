package payment

import (
	trans_models "smlcloudplatform/internal/transaction/models"
	"smlcloudplatform/pkg/microservice"

	transaction_payment_repositories "smlcloudplatform/internal/transaction/payment/repositories"
	transaction_payment_usecase "smlcloudplatform/internal/transaction/payment/usecase"
	transaction_payment_detail_repositories "smlcloudplatform/internal/transaction/paymentdetail/repositories"
	transaction_payment_detail_usecase "smlcloudplatform/internal/transaction/paymentdetail/usecase"
)

type IPaymentUsecase interface {
	Upsert(doc trans_models.TransactionMessageQueue) error
	Delete(shopID string, docNo string) error
}

type PaymentUsecase struct {
	transactionPaymentUsecase       transaction_payment_usecase.IPaymentUsecase
	transactionPaymentDetailUsecase transaction_payment_detail_usecase.IPaymentDetailUsecase
}

func InitPayment(pst microservice.IPersister) *PaymentUsecase {

	transactionPaymentRepo := transaction_payment_repositories.NewPaymentRepository(pst)
	transaction_payment_usecase := transaction_payment_usecase.NewPaymentUsecase(transactionPaymentRepo)

	transactionPaymentDetailRepo := transaction_payment_detail_repositories.NewPaymentDetailRepository(pst)
	transactionPaymentDetailUsecase := transaction_payment_detail_usecase.NewPaymentDetailUsecase(transactionPaymentDetailRepo)

	return NewPayment(transaction_payment_usecase, transactionPaymentDetailUsecase)

}

func NewPayment(
	transaction_payment_usecase transaction_payment_usecase.IPaymentUsecase,
	transaction_payment_detail_usecase transaction_payment_detail_usecase.IPaymentDetailUsecase,
) *PaymentUsecase {
	return &PaymentUsecase{
		transactionPaymentUsecase:       transaction_payment_usecase,
		transactionPaymentDetailUsecase: transaction_payment_detail_usecase,
	}
}

func (u PaymentUsecase) Upsert(transDoc trans_models.TransactionMessageQueue) error {

	transPayment, err := transaction_payment_usecase.ParseTransactionToPayment(transDoc)

	if err != nil {
		return err
	}

	err = u.transactionPaymentUsecase.Upsert(transDoc.ShopID, transDoc.DocNo, transPayment)
	if err != nil {
		return err
	}

	transPaymentDetails, err := transaction_payment_detail_usecase.ParseTransactionToPaymentDetail(transDoc)
	if err != nil {
		return err
	}

	for transPaymentDetail := range transPaymentDetails {
		err = u.transactionPaymentDetailUsecase.Upsert(transDoc.ShopID, transDoc.DocNo, transPaymentDetails[transPaymentDetail])
		if err != nil {
			return err
		}
	}

	return nil

}

func (u PaymentUsecase) Delete(shopID string, docNo string) error {

	err := u.transactionPaymentUsecase.Delete(shopID, docNo)
	if err != nil {
		return err
	}

	err = u.transactionPaymentDetailUsecase.Delete(shopID, docNo)
	if err != nil {
		return err
	}

	return nil
}
