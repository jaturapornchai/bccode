package payment_test

import (
	"smlcloudplatform/internal/config"
	"smlcloudplatform/internal/models"
	payment_repositories "smlcloudplatform/internal/transaction/payment/repositories"
	payment_usecase "smlcloudplatform/internal/transaction/payment/usecase"
	"smlcloudplatform/internal/transaction/transactionconsumer/payment"

	trans_models "smlcloudplatform/internal/transaction/models"
	paymentdetail_repositories "smlcloudplatform/internal/transaction/paymentdetail/repositories"
	paymentdetail_usecase "smlcloudplatform/internal/transaction/paymentdetail/usecase"
	"smlcloudplatform/pkg/microservice"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPaymentUpsert(t *testing.T) {

	pst := microservice.NewPersister(config.NewConfig().PersisterConfig())

	paymentRepo := payment_repositories.NewPaymentRepository(pst)
	paymentUsecase := payment_usecase.NewPaymentUsecase(paymentRepo)

	paymentDetailRepo := paymentdetail_repositories.NewPaymentDetailRepository(pst)
	paymentDetailUsecase := paymentdetail_usecase.NewPaymentDetailUsecase(paymentDetailRepo)

	paymentConsumeUsecase := payment.NewPayment(paymentUsecase, paymentDetailUsecase)

	transMq := trans_models.TransactionMessageQueue{}

	transMq.ShopID = "shop1"
	transMq.DocNo = "doc1"
	transMq.TransFlag = 50
	transMq.PaymentDetailRaw = "[{\"doc_mode\":0,\"trans_flag\":1,\"bank_code\":\"ชำระด้วยบัตรเครดิต\",\"bank_name\":\"ชำระด้วยบัตรเครดิต\",\"book_bank_code\":\"ชำระด้วยบัตรเครดิต\",\"card_number\":\"ชำระด้วยบัตรเครดิต\",\"approved_code\":\"\",\"doc_date_time\":\"2024-02-05T13:26:56.358108\",\"branch_number\":\"\",\"bank_reference\":\"\",\"due_date\":\"2024-02-05T13:26:56.358107\",\"cheque_number\":\"\",\"code\":\"code1x\",\"description\":\"CreditCard\",\"number\":\"\",\"reference_one\":\"\",\"reference_two\":\"\",\"provider_code\":\"\",\"provider_name\":\"\",\"amount\":1570.0}]"
	transMq.Branch.Code = "0001"

	branchLangCode := "en"
	branchLangName := "BranchTest"

	transMq.Branch.Names = &[]models.NameX{
		{
			Code: &branchLangCode,
			Name: &branchLangName,
		},
	}

	err := paymentConsumeUsecase.Upsert(transMq)

	assert.NoError(t, err)

}
