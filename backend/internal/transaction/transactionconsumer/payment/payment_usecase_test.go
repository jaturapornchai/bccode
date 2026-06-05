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

	transMq.HoldingCode = "shop1"
	transMq.DocNo = "doc1"
	transMq.TransFlag = 50
	transMq.PaymentDetailRaw = "[{\"docmode\":0,\"transflag\":1,\"bankcode\":\"ชำระด้วยบัตรเครดิต\",\"bankname\":\"ชำระด้วยบัตรเครดิต\",\"bookbankcode\":\"ชำระด้วยบัตรเครดิต\",\"cardnumber\":\"ชำระด้วยบัตรเครดิต\",\"approvedcode\":\"\",\"docdatetime\":\"2024-02-05T13:26:56.358108\",\"branchnumber\":\"\",\"bankreference\":\"\",\"duedate\":\"2024-02-05T13:26:56.358107\",\"chequenumber\":\"\",\"code\":\"code1x\",\"description\":\"CreditCard\",\"number\":\"\",\"referenceone\":\"\",\"referencetwo\":\"\",\"providercode\":\"\",\"providername\":\"\",\"amount\":1570.0}]"
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
